package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"k8s.io/klog"
)

// Postgres is the type which provides the connection to a postgresql database
type Postgres struct {
	db *sql.DB
}

func (p *Postgres) Connect(server, user, password, database string, port int) error {
	conStr := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable dbname=%s", server, user, password, port, database)
	var err error
	(*p).db, err = sql.Open("postgres", conStr)
	if err != nil {
		return err
	}
	return err
}

// Insert you can insert data into a defined table with this function
func (p *Postgres) Insert(table string, columns []string, val []interface{}) error {
	var valuesPlaceholder string

	if len(columns) != len(val) {
		return fmt.Errorf("length of columns and values are different; count columns: %d and count values %d\n", len(columns), len(val))
	}

	for i := range columns {
		if valuesPlaceholder == "" {
			valuesPlaceholder = fmt.Sprintf("$%d", i+1)
		} else {
			valuesPlaceholder += fmt.Sprintf(", $%d", i+1)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), valuesPlaceholder)
	klog.Infof("database query insert: %s\n", query)
	_, err := p.db.Exec(query, val...)

	return err
}

func (p *Postgres) QueryTime(table string, columns []string, parameters []string, timeColumn string, start, end time.Time, values []*interface{}, parameterValue []interface{}) error {
	var parameter string

	if len(parameters) != len(parameterValue) {
		return fmt.Errorf("length of parameter and parameterValue are not equal; count parameters: %d and values: %d", len(parameter), len(parameterValue))
	}

	if len(columns) != len(values) {
		return fmt.Errorf("lenght of columns and return values are not equal")
	}

	var numberParameter int = 1
	for i, v := range parameters {
		if parameter == "" {
			parameter = fmt.Sprintf("%v = $%d", v, i+1)
		} else {
			parameter += fmt.Sprintf(" AND %v = $%d", v, i+1)
		}
		numberParameter++
	}

	if start.Equal(time.Time{}) {
		parameter = fmt.Sprintf("time > $%d", numberParameter)
		parameterValue = append(parameterValue, start)
		numberParameter++
	}

	if end.Equal(time.Time{}) {
		parameter = fmt.Sprintf("time < $%d", numberParameter)
		parameterValue = append(parameterValue, start)
		numberParameter++
	}


	var query string
	if len(parameters) == 0 {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(columns, ", "), table, parameter)
	}

	klog.Infof("Database query: %s", query)
	quResult, err := p.db.Query(query, parameterValue...)
	if err != nil {
		return err
	}
	defer func() {
		if err := quResult.Close(); err != nil {
			klog.Errorf("could not close database query object")
		}
	}()

	var qValue []interface{}
	mulValue := false
	var scanString []string
	var scanInt []int64
	var scanTime []time.Time
	klog.Infof("len of values: %d\n", len(values))
	for i := 0; i < len(values); i++ {
		val := *values[i]
		switch val.(type) {
		default:
			if mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			qValue = append(qValue, values[i])
		case []string:
			mulValue = true
			if len(qValue) > 0 && !mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			scanString = append(scanString, "")
			qValue = append(qValue, &scanString[len(scanString)-1])
		case []int64:
			mulValue = true
			if len(qValue) > 0 && !mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			scanInt = append(scanInt, int64(-1))
			klog.Infof("current length of scanInt: %d\n", len(scanInt))
			klog.Infof("address of scanInt: %v\n", &scanInt[len(scanInt)-1])
			qValue = append(qValue, &scanInt[len(scanInt)-1])
		case []time.Time:
			mulValue = true
			if len(qValue) > 0 && !mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			scanTime = append(scanTime, time.Time{})
			klog.Infof("current length of scanTime: %d\n", len(scanTime))
			klog.Infof("address of scanTime: %v\n", &scanTime[len(scanTime)-1])
			qValue = append(qValue, &scanTime[len(scanTime)-1])
		}

	}

	if mulValue {
		numString := 0
		numInt := 0
		for quResult.Next() {
			if err := quResult.Scan(qValue...); err != nil {
				return err
			}
			for i := 0; i < len(qValue); i++ {
				val := *values[i]
				switch val.(type) {
				case []string:
					dat := append(val.([]string), scanString[numString])
					*values[i] = dat
					numString++
				case []int64:
					klog.Infof("address from scanInt is: %s\n", &scanInt[numInt])
					klog.Infof("value from scanInt is: %d\n", scanInt[numInt])
					klog.Infof("data %v\n", val.([]int64))
					dat := append(val.([]int64), scanInt[numInt])
					*values[i] = dat
					numInt++
				}
			}
			numString = 0
			numInt = 0
		}
	} else {
		if !quResult.Next() {
			return nil
		}
		if err := quResult.Scan(qValue...); err != nil {
			return err
		}

	}

	return nil
}

func (p *Postgres) Query(table string, columns []string, parameters []string, values []*interface{}, parameterValue []interface{}) error {
	var parameter string

	if len(parameters) != len(parameterValue) {
		return fmt.Errorf("length of parameter and parameterValue are not equal; count parameters: %d and values: %d", len(parameter), len(parameterValue))
	}

	if len(columns) != len(values) {
		return fmt.Errorf("lenght of columns and return values are not equal")
	}

	for i, v := range parameters {
		if parameter == "" {
			parameter = fmt.Sprintf("%v = $%d", v, i+1)
		} else {
			parameter += fmt.Sprintf(" AND %v = $%d", v, i+1)
		}
	}

	var query string
	if len(parameters) == 0 {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(columns, ", "), table, parameter)
	}

	klog.Infof("Database query: %s", query)
	quResult, err := p.db.Query(query, parameterValue...)
	if err != nil {
		return err
	}
	defer func() {
		if err := quResult.Close(); err != nil {
			klog.Errorf("could not close database query object")
		}
	}()

	var qValue []interface{}
	mulValue := false
	var scanString []string
	var scanInt []int64
	klog.Infof("len of values: %d\n", len(values))
	for i := 0; i < len(values); i++ {
		val := *values[i]
		switch val.(type) {
		default:
			if mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			qValue = append(qValue, values[i])
		case []string:
			mulValue = true
			if len(qValue) > 0 && !mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			scanString = append(scanString, "")
			qValue = append(qValue, &scanString[len(scanString)-1])
		case []int64:
			mulValue = true
			if len(qValue) > 0 && !mulValue {
				return fmt.Errorf("using different types with multivalue and single value")
			}
			scanInt = append(scanInt, int64(-1))
			klog.Infof("current length of scanInt: %d\n", len(scanInt))
			klog.Infof("address of scanInt: %v\n", &scanInt[len(scanInt)-1])
			qValue = append(qValue, &scanInt[len(scanInt)-1])
		}

	}

	if mulValue {
		numString := 0
		numInt := 0
		for quResult.Next() {
			if err := quResult.Scan(qValue...); err != nil {
				return err
			}
			for i := 0; i < len(qValue); i++ {
				val := *values[i]
				switch val.(type) {
				case []string:
					dat := append(val.([]string), scanString[numString])
					*values[i] = dat
					numString++
				case []int64:
					klog.Infof("address from scanInt is: %s\n", &scanInt[numInt])
					klog.Infof("value from scanInt is: %d\n", scanInt[numInt])
					klog.Infof("data %v\n", val.([]int64))
					dat := append(val.([]int64), scanInt[numInt])
					*values[i] = dat
					numInt++
				}
			}
			numString = 0
			numInt = 0
		}
	} else {
		if !quResult.Next() {
			return nil
		}
		if err := quResult.Scan(qValue...); err != nil {
			return err
		}

	}

	return nil
}

func (p *Postgres) Update(table string, parameter []string, paramValues []interface{}, updateParameter []string, updateValues []interface{}) error {
	var spec, update string
	var params []interface{}

	if len(parameter) != len(paramValues) {
		return fmt.Errorf("parameter and paramValue haven't the same length")
	}

	if len(updateParameter) != len(updateValues) {
		return fmt.Errorf("updateParameter and updateValues haven't the same length")
	}

	for i, v := range parameter {
		if spec == "" {
			spec = fmt.Sprintf("%s = $%d", v, i+1+len(updateValues))
		} else {
			spec += fmt.Sprintf(", %s = $%d", v, i+1+len(updateValues))
		}
	}

	for i, v := range updateParameter {
		if update == "" {
			update = fmt.Sprintf("%s = $%d", v, i+1)
		} else {
			update += fmt.Sprintf(", %s = $%d", v, i+1)
		}
	}

	for _, v := range updateValues {
		switch v.(type) {
		default:
			return fmt.Errorf("unextepced type in updateValues")
		case bool:
			params = append(params, v.(bool))
		}
	}

	for _, v := range paramValues {
		switch v.(type) {
		default:
			return fmt.Errorf("unextepced type in updateValues")
		case string:
			params = append(params, v.(string))
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, update, spec)
	klog.Infof("database query: %s\n", query)

	_, err := p.db.Exec(query, params...)

	return err
}

func (p *Postgres) Delete(table string, paramters []string, paramValues []interface{}) error {
	var clause, query string
	if len(paramters) != len(paramValues) {
		return fmt.Errorf("len of paramters is not equal to len of paramValues")
	}

	for i, v := range paramters {
		if clause == "" {
			clause = fmt.Sprintf("%s = $%d", v, i+1)
		} else {
			clause += fmt.Sprintf(" AND %s = $%d", v, i+1)
		}
	}

	if clause == "" {
		query = fmt.Sprintf("DELETE FROM %s", table)
	} else {
		query = fmt.Sprintf("DELETE FROM %s WHERE %s", table, clause)
	}

	_, err := p.db.Exec(query, paramValues...)
	return err
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
