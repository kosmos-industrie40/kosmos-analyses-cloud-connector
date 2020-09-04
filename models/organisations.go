package models

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type Organisation struct {
	Name string
}

func (o *Organisation) Query(db *sql.DB, id int64) error {
	result, err := db.Query("SELECT name FROM organisations WHERE id = $1", id)
	if err != nil {
		return err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("could not close result: %v\n", err)
		}
	}()

	var name string
	if !result.Next() {
		return fmt.Errorf("no organisation found to id: %d\n", id)
	}

	if err := result.Scan(&name); err != nil {
		return err
	}

	o.Name = name

	return nil
}

func (o *Organisation) Insert(db *sql.DB) (int64, error) {
	result, err := db.Exec("INSERT INTO organisations (name) VALUES ($1)", o.Name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (o *Organisation) Exists(db *sql.DB) (bool, int64, error) {
	result, err := db.Query("SELECT id FROM organisations WHERE name = $1", o.Name)
	if err != nil {
		return false, 0, err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("could not close result: %v\n", err)
		}
	}()

	var id int64
	if !result.Next() {
		return false, 0, nil
	}

	if err := result.Scan(&id); err != nil {
		return false, 0, err
	}

	return true, id, nil
}
