package models_database

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type Partners struct {
	Contract      string
	Organisations []string
}

func (p *Partners) Query(db *sql.DB, contract, permission string) error {
	result, err := db.Query("SELECT organisations FROM partners WHERE contract = $1", contract)
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
		return fmt.Errorf("no partners found")
	}

	p.Contract = contract
	return nil
}

func (p *Partners) Insert(db *sql.DB, permission string) error {
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
		_, err := db.Exec("INSERT INTO partners (contract, organisaion) VALUES ($1, $2)", p.Contract, v)
		if err != nil {
			return err
		}
	}
	return nil
}
