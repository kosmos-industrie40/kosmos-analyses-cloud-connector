package logic

import (
	"fmt"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"strings"

	"github.com/google/uuid"
	"k8s.io/klog"
)

// Login this function is been used to Login a user
func Login(user, password string, db database.Postgres) (string, error) {
	//TODO check against user list (example LDAP)

	token := strings.Split(uuid.New().URN(), ":")
	var columns []string
	columns = append(columns, "token")
	columns = append(columns, "name")

	var data []interface{}
	data = append(data, token[2])
	data = append(data, user)

	err := db.Insert("token", columns, data)

	return token[2], err

}

func User(token string, db database.Postgres) (string, error) {
	var parameterValue []interface{}
	parameterValue = append(parameterValue, token)

	var para []string
	para = append(para, "token")

	var value string
	var inVal interface{}
	inVal = value
	var val []*interface{}
	val = append(val, &inVal)

	var columns []string
	columns = append(columns, "name")

	err := db.Query("token", columns, para, val, parameterValue)
	klog.Infof("user is: %v", inVal)

	switch inVal.(type) {
	default:
		return "", fmt.Errorf("unexpected data type")
	case string:
		value = inVal.(string)
	}

	return value, err
}

func Logout(token string, db database.Postgres) error {
	var parameter []string
	var paramValue []interface{}

	parameter = append(parameter, "token")
	paramValue = append(paramValue, token)

	return db.Delete("token", parameter, paramValue)
}
