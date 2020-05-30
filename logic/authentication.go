package logic

import (
	"fmt"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"strings"

	"github.com/google/uuid"
	"k8s.io/klog"
)

type Auth struct {
	db database.Postgres
}

// Authentication comparing to a constructor
func (a Auth) Authentication(db database.Postgres) {
	a.db = db
}

// Login this function is been used to Login a user
func (a Auth) Login(user, password string) (string, error) {
	//TODO check against user list (example LDAP)

	token := strings.Split(uuid.New().URN(), ":")
	var columns []string
	columns = append(columns, "token")
	columns = append(columns, "name")

	var data []interface{}
	data = append(data, token[2])
	data = append(data, user)

	err := a.db.Insert("token", columns, data)

	return token[2], err

}

func (a Auth) User(token string) (string, error) {
	var parameterValue []interface{}
	parameterValue = append(parameterValue, token)

	var para []string
	para = append(para, "token")

	var value string
	var inVal interface{} = value
	var val []*interface{}
	val = append(val, &inVal)

	var columns []string
	columns = append(columns, "name")

	err := a.db.Query("token", columns, para, val, parameterValue)
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
	var parameter []string
	var paramValue []interface{}

	parameter = append(parameter, "token")
	paramValue = append(paramValue, token)

	return a.db.Delete("token", parameter, paramValue)
}
