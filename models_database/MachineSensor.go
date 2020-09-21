package models_database

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type MachineSensor struct {
	Machine Machine
	Sensor  Sensor
}

func (m MachineSensor) Exists (db *sql.DB, sensorID int64, machineID string) (bool, int64, error) {
	query, err := db.Query("SELECT id FROM machine_sensors WHERE machine = $1 AND sensor = $2", machineID, sensorID)
	if err != nil {
		return false, 0, err
	}

	defer func(){
		if err := query.Close(); err != nil {
			klog.Errorf("could not close query object: %s\n", err)
		}
	}()

	if !query.Next() {
		return false, 0, nil
	}

	var id int64
	err = query.Scan(&id)
	return true, id, err
}

func (m MachineSensor) Insert (db *sql.DB,) (int64, error) {
	machine, err := m.Machine.Insert(db)
	if err != nil {
		return 0, err
	}

	sensor, err := m.Sensor.Insert(db)
	if err != nil {
		return 0, err
	}

	query, err := db.Query("INSERT INTO machine_sensors (machine, sensor) VALUES ($1, $2) RETURNING id", machine, sensor)
	if err != nil {
		return 0, err
	}
	defer func (){
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var id int64
	if !query.Next() {
		return 0, fmt.Errorf("no returning id can be found")
	}

	err = query.Scan(&id)
	return id, err
}

func (m MachineSensor) Query (db *sql.DB, id int64) error {
	query, err := db.Query("SELECT machine, sensor FROM machine_sensors WHERE id = $1", id)
	if err != nil {
		return err
	}
	defer func(){
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var machineId string
	var sensorId int64

	if !query.Next() {
		return fmt.Errorf("no matching object can be found")
	}

	err = query.Scan(&machineId, &sensorId)

	if err := m.Machine.Query(db, machineId); err != nil {
		return err
	}

	if err := m.Sensor.Query(db, sensorId); err != nil {
		return err
	}

	return nil
}