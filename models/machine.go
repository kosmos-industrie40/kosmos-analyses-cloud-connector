package models

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"k8s.io/klog"
)

type Machine struct {
	ID      string
	Meta    interface{}
}

func (m *Machine) Exists(db *sql.DB, id string, meta interface{}) (bool, string, error) {
	metaB, err := json.Marshal(meta)
	if err != nil {
		return false, "", err
	}
	uMeta := string(metaB)

	res, err := db.Query("SELECT * FROM machines WHERE id = $1 AND meta = $2", id, uMeta)
	if err != nil {
		return false, "", err
	}

	if !res.Next() {
		return false, id, nil
	}

	return true, id, nil
}

func (m *Machine) Query(db *sql.DB, id string) error {
	result, err := db.Query("SELECT meta FROM machines WHERE id = $1", id)
	if err != nil {
		return err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	var meta string
	if !result.Next() {
		return fmt.Errorf("no matching machine found")
	}

	if err := result.Scan(&meta); err != nil {
		return err
	}

	var metaDe interface{}
	if err := json.Unmarshal([]byte(meta), &metaDe); err != nil {
		return err
	}

	m.Meta = metaDe
	m.ID = id
	return nil
}

func (m *Machine) Insert(db *sql.DB) (string, error) {

	exists, _, err := m.Exists(db, m.ID, m.Meta)
	if err != nil {
		return "", err
	}

	if exists {
		return m.ID, nil
	}

	metaB, err := json.Marshal(m.Meta)
	if err != nil {
		return "", err
	}

	meta := string(metaB)

	if _, err := db.Exec("INSERT INTO machines (id, meta) VALUES ($1, $2)", m.ID, meta); err != nil {
		return "", err
	}

	return m.ID, nil
}
