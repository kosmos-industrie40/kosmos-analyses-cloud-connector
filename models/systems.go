package models

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type System struct {
	Name string
}

func (s *System) Exists(db *sql.DB) (bool, int64, error) {
	result, err := db.Query("SELECT id FROM systems WHERE name = $1", s.Name)
	if err != nil {
		return false, 0, err
	}
	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	if !result.Next() {
		return false, 0, nil
	}

	var id int64
	if err := result.Scan(&id); err != nil {
		return false, 0, err
	}

	return true, id, nil
}

func (s *System) Insert(db *sql.DB) (int64, error) {
	result, err := db.Exec("INSERT INTO systems (name) VALUES ($1)", s.Name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *System) Query(db *sql.DB, id int64) error {
	result, err := db.Query("SELECT name FROM systems WHERE id = $1", id)
	if err != nil {
		return err
	}

	if !result.Next() {
		return fmt.Errorf("no system found")
	}

	var name string
	if err := result.Scan(&name); err != nil {
		return err
	}

	s.Name = name
	return nil
}
