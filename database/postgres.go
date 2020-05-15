package database

import (
	"database/sql"
	"fmt"
	"strings"

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
			parameter += fmt.Sprintf("AND %v = $%d", v, i+1)
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

	valueCounter := 0
	for quResult.Next() {
		value := *values[valueCounter]
		if valueCounter == len(values) {
			return fmt.Errorf("to many columns in db return")
		}
		switch value.(type) {
		case int:
			var cache int
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case int64:
			var cache int64
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case string:
			var cache string
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case float32:
			var cache float32
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case float64:
			var cache float64
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case byte:
			var cache byte
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case []byte:
			var cache []byte
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case bool:
			var cache bool
			quResult.Scan(&cache)
			*values[valueCounter] = cache
		case []int:
			var cache int
			quResult.Scan(&cache)
			*values[valueCounter] = append(value.([]int), cache)
		case []string:
			var cache string
			quResult.Scan(&cache)
			*values[valueCounter] = append(value.([]string), cache)
		default:
			klog.Errorf("unexpected type")
			return fmt.Errorf("unexpected type, used in interface")
		}
		valueCounter++
	}

	return err
}

func (p *Postgres) Update(table string, parameter []string, paramValues []interface{}, updateParameter []string, updateValues []interface{}) error {
	var clause, update string

	if len(parameter) != len(paramValues) {
		return fmt.Errorf("parameter and paramValues haven't the same length")
	}

	if len(updateParameter) != len(updateValues) {
		return fmt.Errorf("updateParameter and updateValues haven't the same length")
	}

	for i, v := range parameter {
		if clause == "" {
			clause = fmt.Sprintf("%s = $%d", v, len(updateParameter)+i+1)
		} else {
			clause += fmt.Sprintf("AND %s = $%d", v, len(updateParameter)+i+1)
		}
	}

	for i, v := range updateValues {
		if update == "" {
			update = fmt.Sprintf("%s = $%d", v, i+1)
		} else {
			update += fmt.Sprintf(", %s = $%d", v, i+1)
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, update, clause)

	var params []interface{}
	for i := 0; i < len(updateParameter); i++ {
		params = append(params, updateParameter[i])
	}
	for i := 0; i < len(paramValues); i++ {
		params = append(params, paramValues[i])
	}

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
			clause += fmt.Sprintf("AND %s = $%d", v, i+1)
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
