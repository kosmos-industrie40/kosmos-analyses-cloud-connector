package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog"
)

// AuthHelper is an helper interface, which can check if an user is authenticated or not
type Helper interface {
	// IsAuthenticated consumes the token and checked if this token is valid
	// returns an error on error or true if the authentication is valid
	// in other cases (no error and not authenticated) you can use the status code
	IsAuthenticated(*http.Request, string, bool) (bool, int, error)

	// CreateSession will create a session on a specific token. This token will be used
	// to identify if the user, has the required permission or not
	CreateSession(string, []string, []string, time.Time) error

	// DeleteSession will delete a user session, which is identified by a the session token
	DeleteSession(string) error

	// CleanUp will run every hour and will remove all invalid tokens
	CleanUp()

	// TokenValid checks if a token is valid and can be used or not
	TokenValid(r *http.Request) (bool, error)
}

type helperOidc struct {
	db            *sql.DB
	contractWrite string
}

func (a helperOidc) testContractWrite(tokens []string) bool {
	for _, token := range tokens {
		if token == a.contractWrite {
			return true
		}
	}
	return false
}

func (a helperOidc) TokenValid(r *http.Request) (bool, error) {

	token := r.Header.Get("token")
	if token == "" {
		return false, nil
	}

	query, err := a.db.Query("SELECT * FROM token WHERE token = $1 AND valid >= NOW()", token)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !query.Next() {
		return false, nil
	}

	return true, nil
}

func (a helperOidc) cleanUp() error {
	res, err := a.db.Exec("DELETE FROM token WHERE valid < NOW()")
	if err != nil {
		return fmt.Errorf("cannot delete invalid tokens: %s", err)
	}

	columns, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("cannot get the number of rows: %s", err)
	}
	klog.Infof("removing %d invalid tokens", columns)
	return nil
}

func (a helperOidc) CleanUp() {
	for {
		if err := a.cleanUp(); err != nil {
			klog.Error(err)
		}
		time.Sleep(time.Hour)
	}
}

func (a helperOidc) CreateSession(token string, organisations, contractCreation []string, valid time.Time) error {
	klog.V(2).Infof("organisations: %s", fmt.Sprintf("'%s'", strings.Join(organisations, "','")))

	canCreateContract := a.testContractWrite(contractCreation)
	klog.V(2).Infof("the user of the added token has contract write rights: %t", canCreateContract)

	query, err := a.db.Query(fmt.Sprintf("SELECT id FROM organisations WHERE name in (%s)", fmt.Sprintf("'%s'", strings.Join(organisations, "','"))))
	if err != nil {
		return err
	}
	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var orgs []int64
	for query.Next() {
		var organisation int64
		if err := query.Scan(&organisation); err != nil {
			return err
		}

		orgs = append(orgs, organisation)
	}

	klog.Infof("orgs: %v", orgs)

	if len(orgs) == 0 {
		return fmt.Errorf("no matching organisations found")
	}

	if _, err := a.db.Exec("INSERT INTO token (token, valid, write_contract) VALUES ($1, $2, $3)", token, valid, canCreateContract); err != nil {
		return err
	}

	for _, org := range orgs {
		klog.Infof("insert token_permission with (%s, %d)", token, org)
		if _, err := a.db.Exec("INSERT INTO token_permission (token, organisation) VALUES ($1, $2)", token, org); err != nil {
			return err
		}
	}

	return nil
}

func (a helperOidc) IsAuthenticated(request *http.Request, contract string, write bool) (bool, int, error) {
	token := request.Header.Get("token")
	if token == "" {
		klog.Infof("no token can be found")
		return false, http.StatusUnauthorized, nil
	}

	query, err := a.db.Query("SELECT valid FROM token WHERE token = $1", token)
	if err != nil {
		return false, http.StatusInternalServerError, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !query.Next() {
		return false, http.StatusUnauthorized, nil
	}

	var valid time.Time
	if err := query.Scan(&valid); err != nil {
		klog.Infof("cannot scan valid time")
		return false, http.StatusInternalServerError, err
	}

	if time.Now().After(valid) {
		klog.Infof("timestamp.after doesn't match")
		return false, http.StatusUnauthorized, nil
	}

	var table string
	if write {
		table = "write_permissions rp"
	} else {
		table = "read_permissions rp"
	}
	hasPermission, err := a.db.Query(fmt.Sprintf("SELECT tp.organisation FROM token_permission as tp JOIN %s on tp.organisation = rp.organisation WHERE token = $1 AND contract = $2", table), token, contract)
	if err != nil {
		return false, http.StatusInternalServerError, err
	}

	defer func() {
		if err := hasPermission.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !hasPermission.Next() {
		return false, http.StatusUnauthorized, nil
	}

	return true, 0, nil
}

func (a helperOidc) DeleteSession(token string) error {
	_, err := a.db.Exec("DELETE FROM token WHERE token = $1", token)
	return err
}

// NewAuthHelper creates a new authentication auth helper
func NewAuthHelper(db *sql.DB, contractWrite string) Helper {
	return helperOidc{db: db, contractWrite: contractWrite}
}
