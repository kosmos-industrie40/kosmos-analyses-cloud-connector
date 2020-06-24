package logic

import (
	"fmt"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"strings"

	"github.com/google/uuid"
	"k8s.io/klog"
)

type Auth struct {
	Db database.Postgres
}

// Authentication comparing to a constructor
func (a Auth) Authentication(db database.Postgres) {
	a.Db = db
}

// Login this function is been used to Login a user
func (a Auth) Login(user, password string) (string, error) {
	//TODO check against user list (example LDAP)

	token := strings.Split(uuid.New().URN(), ":")
	columns := []string{"token", "name"}
	data := []interface{}{token[2], user}

	err := a.Db.Insert("token", columns, data)

	return token[2], err

}

func (a Auth) User(token string) (string, error) {
	parameterValue := []interface{}{token}

	para := []string{"token"}

	var value string
	var inVal interface{} = value
	val := []*interface{}{&inVal}

	columns := []string{"name"}

	err := a.Db.Query("token", columns, para, val, parameterValue)
	klog.Infof("user is: %v", inVal)

	switch v := inVal.(type) {
	default:
		return "", fmt.Errorf("unexpected data type")
	case string:
		value = v
	}

	return value, err
}

func (a Auth) Logout(token string) error {
	parameter := []string{"token"}
	paramValue := []interface{}{token}

	return a.Db.Delete("token", parameter, paramValue)
}
