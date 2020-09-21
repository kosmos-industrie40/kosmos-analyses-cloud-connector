package models

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type Model struct {
	Container Container
}

func (m *Model) Insert(db *sql.DB) (int64, error) {
	exists, cId, err := m.Container.Exists(db)
	if err != nil {
		return 0, err
	}

	if !exists {
		cId, err = m.Container.Insert(db)
		if err != nil {
			return 0, err
		}
	}

	ret, err := db.Query("INSERT INTO models (container) VALUES ($1) RETURNING id", cId)
	if err != nil {
		return 0, err
	}

	defer func(){
		if err := ret.Close(); err != nil  {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !ret.Next() {
		return 0, fmt.Errorf("no id is returned")
	}
	var id int64

	err = ret.Scan(&id)
	return id, err
}

func (m *Model) Query(db *sql.DB, id int64) error {
	query, err := db.Query("SELECT container FROM models WHERE id = $1", id)
	if err != nil {
		return err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("could not close query object: %s", err)
		}
	}()

	if !query.Next() {
		return fmt.Errorf("no matching model found")
	}

	var containerId int64
	if err := query.Scan(&containerId); err != nil {
		return err
	}

	return (&m.Container).Query(db, containerId)
}

func (m *Model) Exists(db *sql.DB) (bool, int64, error) {
	exists, cID, err := m.Container.Exists(db)
	if err != nil {
		return false, 0, err
	}

	if !exists {
		return false, 0, nil
	}

	query, err := db.Query("SELECT id FROM models WHERE container = $1", cID)
	if err != nil {
		return false, 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("could not close query object: %s\n", err)
		}
	}()

	if !query.Next() {
		return false, 0, nil
	}

	var id int64
	if err := query.Scan(&id); err != nil {
		return false, 0, err
	}

	return true, id, nil
}