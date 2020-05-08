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
			valuesPlaceholder = fmt.Sprintf("$%d", i)
		} else {
			valuesPlaceholder = fmt.Sprintf(", $%d", i)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), valuesPlaceholder)
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
			parameter = fmt.Sprintf("%v = $%d", v, i)
		} else {
			parameter += fmt.Sprintf("AND %v = $%d", v, i)
		}
	}

	var query string
	if len(parameters) == 0 {
		query = fmt.Sprintf("SELECT %s FROM %s")
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(columns, ", "), table, parameters)
	}

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
		}
		valueCounter++
	}

	return err
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
