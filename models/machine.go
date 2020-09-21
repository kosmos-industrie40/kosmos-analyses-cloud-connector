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
	Sensors []Sensor
}

func (m *Machine) Exists(db *sql.DB) (bool, []int64, error) {
	metaB, err := json.Marshal(m.Meta)
	if err != nil {
		return false, nil, err
	}

	meta := string(metaB)

	result, err := db.Query("SELECT exists(SELECT true FROM machines WHERE id = $1 AND meta = $2)", m.ID, meta)
	if err != nil {
		return false, nil, err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	var exists bool
	if !result.Next() {
		return false, nil, fmt.Errorf("can not check existens in query")
	}

	if err := result.Scan(&exists); err != nil {
		return false, nil, nil
	}

	if !exists {
		// no machines found
		return false, nil, nil
	}

	ids, err := db.Query("SELECT sensor FROM machine_sensors WHERE machine = $1", m.ID)
	if err != nil {
		return true, nil, err
	}

	defer func() {
		if err := ids.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	var machineSensorIds []int64

	for !ids.Next() {
		var id int64
		if err := ids.Scan(&id); err != nil {
			return true, nil, err
		}

		machineSensorIds = append(machineSensorIds, id)
	}

	return true, machineSensorIds, nil
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

	querySensors, err := db.Query("SELECT sensor FROM machine_sensors WHERE id = $1 GROUP BY sensor", id)
	if err != nil {
		return err
	}

	defer func() {
		if err := querySensors.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	var sensorIds []int64

	for !querySensors.Next() {
		var id int64
		if err := querySensors.Scan(&id); err != nil {
			return err
		}

		sensorIds = append(sensorIds, id)
	}

	for _, v := range sensorIds {
		sensor := Sensor{}
		if err := sensor.Query(db, v); err != nil {
			return err
		}
		m.Sensors = append(m.Sensors, sensor)
	}

	return nil
}

func (m *Machine) Insert(db *sql.DB) ([]int64, error) {
	metaB, err := json.Marshal(m.Meta)
	if err != nil {
		return nil, err
	}

	meta := string(metaB)

	if _, err := db.Query("INSERT INTO machines (id, meta) VALUES ($1, $2)", m.ID, meta); err != nil {
		return nil, err
	}

	var msIds []int64
	for _, sensor := range m.Sensors {
		msId, err := func() (int64, error) {
			sensorExists, sensorId, err := sensor.Exists(db)
			if err != nil {
				return 0, err
			}

			if !sensorExists {
				sensorId, err = sensor.Insert(db)
				if err != nil {
					return 0, err
				}
			}

			res, err := db.Query("SELECT id FROM machine_sensor WHERE machine = %1 AND sensor = $2", m.ID, sensorId)
			if err != nil {
				return 0, err
			}

			defer func() {
				if err := res.Close(); err != nil {
					klog.Errorf("cannot close query object: %s", err)
				}
			}()

			var msID int64

			if res.Next() {
				err := res.Scan(&msID)
				return msID, err
			} else {
				result, err := db.Query("INSERT INTO machine_sensor (machine, sensor) VALUES ($1, $2) RETURNING id", m.ID, sensorId)
				if err != nil {
					return 0, err
				}

				defer func() {
					if err := result.Close(); err != nil {
						klog.Errorf("cannot close query object: %s", err)
					}
				}()

				if !result.Next() {
					return 0, fmt.Errorf("no returned id found")
				}

				err = result.Scan(&msID)
				return msID, err
			}
		}()
		if err != nil {
			return nil, err
		}

		msIds = append(msIds, msId)
	}

	return msIds, nil
}
