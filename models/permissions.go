package models

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type Permission struct {
	Contract      string
	Organisations []string
}

func (p *Permission) Query(db *sql.DB, contract, permission string) error {
	var table string
	switch permission {
	case "read":
		table = "read_permission"
	case "write":
		table = "write_permission"
	default:
		return fmt.Errorf("unexpected permission")
	}
	result, err := db.Query(fmt.Sprintf("SELECT organisations FROM %s WHERE contract = $1", table), contract)
	if err != nil {
		return err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("can not close query object: %s\n", err)
		}
	}()

	for result.Next() {
		var org int64
		if err := result.Scan(&org); err != nil {
			return err
		}
		var organisation *Organisation
		if err := organisation.Query(db, org); err != nil {
			return err
		}

		p.Organisations = append(p.Organisations, organisation.Name)
	}

	if len(p.Organisations) == 0 {
		return fmt.Errorf("no permission found")
	}

	p.Contract = contract
	return nil
}

func (p *Permission) Insert(db *sql.DB, permission string) error {
	var table string
	switch permission {
	case "read":
		table = "read_permission"
	case "write":
		table = "write_permission"
	default:
		return fmt.Errorf("unexpected permission")
	}
	var orgs []int64

	for _, v := range p.Organisations {
		org := Organisation{Name: v}
		exists, id, err := org.Exists(db)
		if err != nil {
			return err
		}

		if !exists {
			id, err = org.Insert(db)
		}

		orgs = append(orgs, id)
	}

	for _, v := range orgs {
		_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (contract, organisaion) VALUES ($1, $2)", table), p.Contract, v)
		if err != nil {
			return err
		}
	}
	return nil
}
