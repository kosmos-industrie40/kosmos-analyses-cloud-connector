package models

import (
	"database/sql"
	"fmt"
	"strings"

	"k8s.io/klog"
)

type Container struct {
	Url         string   `json:"url"`
	Tag         string   `json:"tag"`
	Arguments   []string `json:"arguments"`
	Environment []string `json:"environment"`
}

func (c Container) arrayToString(data []string) string {
	if len(data) == 0 {
		return "{}"
	}
	return "{\"" + strings.Join(data, "\",\"") + "\"}"
}

func (c Container) stringToArray(data string) []string {
	if data == "{}" {
		return []string{}
	}

	data = strings.TrimLeft(data, "{\"")
	data = strings.TrimRight(data, "\"}")
	return strings.Split(data, "\",\"")
}

func (c Container) Exists(db *sql.DB) (bool, int64, error) {
	result, err := db.Query("SELECT id FROM containers WHERE url = $1 AND tag = $2 AND arguments = $3 AND environment = $4", c.Url, c.Tag, c.arrayToString(c.Arguments), c.arrayToString(c.Environment))
	if err != nil {
		return false, 0, err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("could not close row type: %s\n", err)
		}
	}()

	var id int64

	if !result.Next() {
		return false, 0, nil
	}

	result.Scan(&id)

	return true, id, nil
}

func (c Container) Insert(db *sql.DB) (int64, error) {
	result, err := db.Exec("INSERT INTO containers (url, tag, arguments, environment) VALUES ($1, $2, $3, $4) RETURNING id", c.Url, c.Tag, c.arrayToString(c.Arguments), c.arrayToString(c.Environment))
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c *Container) Query(db *sql.DB, id int64) error {
	result, err := db.Query("SELECT url, tag, arguments, environment FROM containers WHERE id = $1", id)
	if err != nil {
		return err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("could not close result: %v\n", err)
		}
	}()

	var url, tag, arguments, environment string

	if !result.Next() {
		return fmt.Errorf("no container found to id: %d\n", id)
	}

	if err := result.Scan(&url, &tag, &arguments, &environment); err != nil {
		return err
	}

	c.Url = url
	c.Tag = tag
	c.Arguments = c.stringToArray(arguments)
	c.Environment = c.stringToArray(environment)

	return nil
}
