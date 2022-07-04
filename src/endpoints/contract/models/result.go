package models

import (
	"database/sql"

	"k8s.io/klog"
)

type ResultList interface {
	GetAllContracts(string) ([]string, error)
}

type resultList struct {
	db *sql.DB
}

func (r resultList) GetAllContracts(token string) ([]string, error) {
	query, err := r.db.Query("SELECT id FROM contracts")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var ids []string
	for query.Next() {
		var id string
		err := query.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func NewResultList(db *sql.DB) ResultList {
	return resultList{db: db}
}
