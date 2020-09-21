package models_database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"k8s.io/klog"
)

type Sensor struct {
	ID   string
	Meta interface{}
}

func (s *Sensor) Insert(db *sql.DB) (int64, error) {
	metaB, err := json.Marshal(s.Meta)
	if err != nil {
		return 0, err
	}
	meta := string(metaB)

	result, err := db.Query("INSERT INTO sensors (transmitted_id, meta) VALUES ($1, $2) RETURNING id", s.ID, meta)
	if err != nil {
		return 0, err
	}

	defer func(){
		if err := result.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !result.Next() {
		return 0, fmt.Errorf("no id is returned")
	}

	var id int64
	err = result.Scan(&id)
	return id, err
}

func (s *Sensor) Query(db *sql.DB, id int64) error {
	result, err := db.Query("SELECT transmitted_id, meta FROM sensors WHERE id = $1", id)
	if err != nil {
		return err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("can not close query object: %s\n", err)
		}
	}()

	var transmitted, meta string
	if err := result.Scan(&transmitted, &meta); err != nil {
		return err
	}

	s.ID = transmitted
	s.Meta = meta

	return nil
}

func (s *Sensor) Exists(db *sql.DB) (bool, int64, error) {
	metaB, err := json.Marshal(s.Meta)
	if err != nil {
		return false, 0, err
	}
	meta := string(metaB)

	result, err := db.Query("SELECT id FROM sensors WHERE transmitted_id = $1 AND meta = $2", s.ID, meta)
	if err != nil {
		return false, 0, err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("can not close query object: %s\n", err)
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
