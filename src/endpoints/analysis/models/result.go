package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog"
)

type ResultListHandler interface {
	Get(string, map[string][]string) ([]byte, error)
}

type resultList struct {
	Id      int64  `json:"resultID"`
	Machine string `json:"machine"`
	Date    string `json:"date"`
}

type resultListHandler struct {
	db *sql.DB
}

func NewResultList(db *sql.DB) ResultListHandler {
	return resultListHandler{db: db}
}

func (r resultListHandler) Get(contractID string, queryParams map[string][]string) ([]byte, error) {
	var queryWhere []string
	var argWhere []interface{}

	queryWhere = append(queryWhere, "contract = $1")
	argWhere = append(argWhere, contractID)

	var counter = 2
	for i, v := range queryParams {
		if len(v) != 1 {
			return nil, fmt.Errorf("unexpected length of the query parameters")
		}

		switch i {
		case "machine":
			ids, err := func() ([]string, error) {
				query, err := r.db.Query("SELECT cms.id FROM contract_machine_sensors as cms JOIN machine_sensors ms on cms.machine_sensor = ms.id WHERE ms.machine = $1",
					v[0],
				)
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
					var id int64
					if err := query.Scan(&id); err != nil {
						return nil, err
					}

					ids = append(ids, fmt.Sprintf("%d", id))
				}

				return ids, nil
			}()

			if err != nil {
				return nil, err
			}

			queryWhere = append(queryWhere, fmt.Sprintf("contract_machine_sensor in ($%d)", counter))
			argWhere = append(argWhere, strings.Join(ids, ","))
		case "sensor":
			ids, err := func() ([]string, error) {
				query, err := r.db.Query("SELECT cms.id FROM contract_machine_sensors as cms JOIN machine_sensors ms on cms.machine_sensor = ms.id JOIN sensors s on ms.sensor = s.id WHERE s.transmitted_id = $1",
					v[0],
				)
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
					var id int64
					if err := query.Scan(&id); err != nil {
						return nil, err
					}
					ids = append(ids, fmt.Sprintf("%d", id))
				}

				return ids, nil
			}()

			if err != nil {
				return nil, err
			}

			queryWhere = append(queryWhere, fmt.Sprintf("contract_machine_sensor in ($%d)", counter))
			argWhere = append(argWhere, strings.Join(ids, ","))
		case "start":
			_, err := time.Parse(time.RFC3339, v[0])
			if err != nil {
				return nil, err
			}
			queryWhere = append(queryWhere, fmt.Sprintf("time >= $%d", counter))
			argWhere = append(argWhere, v[0])
		case "end":
			_, err := time.Parse(time.RFC3339, v[0])
			if err != nil {
				return nil, err
			}
			queryWhere = append(queryWhere, fmt.Sprintf("time <= $%d", counter))
			argWhere = append(argWhere, v[0])
		}
		counter++
	}

	where := fmt.Sprintf("WHERE %s", strings.Join(queryWhere, " AND "))

	query, err := r.db.Query(fmt.Sprintf("SELECT ar.id, ar.time, ms.machine FROM analyse_result AS ar JOIN contract_machine_sensors cms on cms.id = ar.contract_machine_sensor JOIN machine_sensors ms on cms.machine_sensor = ms.id %s", where), argWhere...)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var res []resultList
	for !query.Next() {
		var id int64
		var machineID, timestamp string

		if err := query.Scan(&id, &timestamp, &machineID); err != nil {
			return nil, err
		}

		res = append(res, resultList{
			Id:      id,
			Machine: machineID,
			Date:    timestamp,
		})
	}

	return json.Marshal(res)
}
