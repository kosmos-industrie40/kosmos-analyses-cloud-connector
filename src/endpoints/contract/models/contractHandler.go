package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/klog"
)

type ContractHandler interface {
	// InsertContract write the contract to a persistent storage
	InsertContract(contract Contract) error

	// DeleteContract delete a contract identified by the the id from the persistent storage
	DeleteContract(contract string) error

	// GetContract get a specific contract from the persistent storage based on the id
	GetContract(contract string) (Contract, error)
}

type contractHandler struct {
	db     *sql.DB
	system string
}

type sensorId struct {
	Sensor                int64
	MachineSensor         int64
	ContractMachineSensor int64
}

func (c contractHandler) InsertContract(contract Contract) error {
	contractJson, err := json.Marshal(contract)
	if err != nil {
		klog.Errorf("marshal error: %v", err)
		return err
	}

	query, err := c.db.Query("SELECT * FROM contracts WHERE id = $1", contract.Body.Contract.ID)
	if err != nil {
		klog.Errorf("cannot query contracts")
		return err
	}

	if query.Next() {
		return fmt.Errorf("the contract with id: %s already exists", contract.Body.Contract.ID)
	}

	_, err = c.db.Exec("INSERT INTO contracts (id, start_time, end_time, creation, validate_signature, contract) VALUES  ($1, $2, $3, $4, $5, $6)",
		contract.Body.Contract.ID,
		contract.Body.Contract.Valid.Start,
		contract.Body.Contract.Valid.End,
		contract.Body.Contract.CreationTime,
		contract.Body.CheckSignature,
		contractJson,
	)
	if err != nil {
		return err
	}

	klog.Infof("insert partners")
	if err := c.insertPartners(contract.Body.Contract.ID, contract.Body.Contract.Partners); err != nil {
		return err
	}

	klog.Infof("insert permission read")
	if err := c.insertPermission(false, contract.Body.Contract.ID, contract.Body.Contract.Permissions.Read); err != nil {
		return err
	}
	klog.Infof("insert permission write")
	if err := c.insertPermission(true, contract.Body.Contract.ID, contract.Body.Contract.Permissions.Write); err != nil {
		return err
	}

	klog.Infof("insert machine")
	if err := c.insertMachine(contract.Body.Machine); err != nil {
		return err
	}

	klog.Infof("insert kosmos local")
	systemMap := make(map[string]int64)
	contract.Body.KosmosLocalSystems = append(contract.Body.KosmosLocalSystems, c.system) // adding kosmos local system in insertion
	for _, system := range contract.Body.KosmosLocalSystems {
		err := func() error {

			systemIdQuery, err := c.db.Query("SELECT id FROM systems WHERE name = $1", system)
			if err != nil {
				return err
			}

			defer func() {
				if err := systemIdQuery.Close(); err != nil {
					klog.Errorf("cannot close query object: %s", err)
				}
			}()

			if systemIdQuery.Next() {
				var id int64
				if err := systemIdQuery.Scan(&id); err != nil {
					return err
				}

				systemMap[system] = id
				return nil
			}

			systemId, err := c.db.Query("INSERT INTO systems (name) VALUES ($1) RETURNING id", system)
			if err != nil {
				return err
			}
			defer func() {
				if err := systemId.Close(); err != nil {
					klog.Errorf("cannot close query object: %s", err)
				}
			}()

			if !systemId.Next() {
				return fmt.Errorf("insert was not successfull in inserting systems")
			}

			var id int64
			if err := systemId.Scan(&id); err != nil {
				return err
			}

			systemMap[system] = id

			return nil
		}()
		if err != nil {
			return err
		}
	}

	klog.Infof("insert technical container")
	for _, tc := range contract.Body.TechnicalContainers {
		if err := c.insertTechnicalContainer(systemMap[tc.System], tc.Containers, contract.Body.Contract.ID); err != nil {
			return err
		}
	}

	klog.Infof("insert sensors")
	sensorMap := make(map[string]sensorId)
	for _, sensor := range contract.Body.Sensors {

		var si sensorId

		klog.Infof("insert only sensor")
		id, err := c.insertOnlySensor(sensor.Name, sensor.Meta)
		if err != nil {
			return err
		}
		si.Sensor = id

		klog.Infof("insert machine sensor")
		ms, err := c.insertMachineSensor(contract.Body.Machine, id)
		if err != nil {
			return err
		}
		si.MachineSensor = ms

		klog.Infof("insert contract machine sensor")
		cms, err := c.insertContractMachineSensor(contract.Body.Contract.ID, ms)
		if err != nil {
			return err
		}

		si.ContractMachineSensor = cms

		klog.Infof("insert storage duration")
		for _, duration := range sensor.StorageDuration {
			err = c.insertStorageDuration(cms, duration.Duration, systemMap[duration.SystemName])
			if err != nil {
				return err
			}
		}

		sensorMap[sensor.Name] = si
	}

	klog.Infof("insert analysis")
	if err := c.insertAnalysis(contract.Body.Analysis, systemMap, sensorMap); err != nil {
		return err
	}

	return nil
}

func (c contractHandler) insertAnalysis(analysis Analysis, systemMap map[string]int64, sensorId map[string]sensorId) error {
	if !analysis.Enable {
		return nil
	}

	for _, system := range analysis.Systems {
		if system.Name != c.system {
			continue
		}

		if !system.Enable {
			continue
		}

		for _, pipeline := range system.Pipelines {

			var pIds []int64
			for _, sensor := range pipeline.Sensors {
				klog.Infof("analysis of system: %s", system.Name)
				pId, err := c.insertPipeline(sensorId[sensor].ContractMachineSensor, systemMap[system.Name], pipeline.Trigger)
				if err != nil {
					return err
				}

				pIds = append(pIds, pId)
			}

			modelContainerMap := make(map[string]int64)
			for _, pipe := range pipeline.Pipeline {
				exec, err := c.insertModel(pipe.Container)
				if err != nil {
					return err
				}

				bytes, err := json.Marshal(pipe.Container)
				if err != nil {
					return err
				}
				modelContainerMap[string(bytes)] = exec

			}

			for _, pipe := range pipeline.Pipeline {
				bytes, err := json.Marshal(pipe.Container)
				if err != nil {
					return err
				}

				exec := modelContainerMap[string(bytes)]

				if pipe.From == nil && pipe.To == nil {
					for _, pId := range pIds {
						_, err := c.db.Exec("INSERT INTO analysis (pipeline, persist, execute) VALUES ($1, $2, $3)",
							pId,
							pipe.Persist,
							exec,
						)
						if err != nil {
							return err
						}
					}
				} else if pipe.From == nil {
					to, err := c.getModelId(*(pipe.To))
					if err != nil {
						return err
					}

					for _, pId := range pIds {
						_, err := c.db.Exec("INSERT INTO analysis (pipeline, next_model, persist, execute) VALUES ($1, $2, $3, $4)",
							pId,
							to,
							pipe.Persist,
							exec,
						)
						if err != nil {
							return err
						}
					}
				} else if pipe.To == nil {
					from, err := c.getModelId(*(pipe.From))
					if err != nil {
						return err
					}

					for _, pId := range pIds {
						_, err := c.db.Exec("INSERT INTO analysis (pipeline, prev_model, persist, execute) VALUES ($1, $2, $3, $4)",
							pId,
							from,
							pipe.Persist,
							exec,
						)
						if err != nil {
							return err
						}
					}
				} else {
					from, err := c.getModelId(*(pipe.From))
					if err != nil {
						return err
					}

					to, err := c.getModelId(*(pipe.To))
					if err != nil {
						return err
					}

					for _, pId := range pIds {
						_, err := c.db.Exec("INSERT INTO analysis (pipeline, prev_model, next_model, persist, execute) VALUES ($1, $2, $3, $4, $5)",
							pId,
							from,
							to,
							pipe.Persist,
							exec,
						)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (c contractHandler) getModelId(model Model) (int64, error) {
	query, err := c.db.Query("SELECT m.id FROM models AS m JOIN containers c on c.id = m.container WHERE c.url = $1 AND c.tag = $2",
		model.Url,
		model.Tag,
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err = query.Scan(&id)
		return id, err
	}

	return 0, fmt.Errorf("could not found model with url %s and tag %s", model.Url, model.Tag)
}

func (c contractHandler) insertPipeline(cms, system int64, trigger Trigger) (int64, error) {
	var tr string

	if trigger.Definition == nil {
		tr = "NULL"
	} else {
		tr = trigger.Definition.After
	}

	queryID, err := c.db.Query("SELECT id FROM pipelines WHERE contract_machine_sensor = $1 AND system = $2 AND time_trigger = $3",
		cms,
		system,
		tr,
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := queryID.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if queryID.Next() {
		var id int64
		err = queryID.Scan(&id)
		return id, err
	}

	var query *sql.Rows
	if tr == "NULL" {
		klog.Infof("tr == null\nsystemid is %d", system)
		query, err = c.db.Query("INSERT INTO pipelines (contract_machine_sensor, system) VALUES ($1, $2) RETURNING id",
			cms,
			system,
		)
	} else {
		klog.Infof("tr != null")
		query, err = c.db.Query("INSERT INTO pipelines (contract_machine_sensor, system, time_trigger) VALUES ($1, $2, $3) RETURNING id",
			cms,
			system,
			tr,
		)
	}
	if err != nil {
		return 0, err
	}

	klog.Infof("%v, %v, %v", cms, system, tr)

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !query.Next() {
		return 0, fmt.Errorf("insertion was not successfull in insert pipeline")
	}

	var id int64
	err = query.Scan(&id)
	return id, err
}

func (c contractHandler) insertModel(container Container) (int64, error) {
	cId, err := c.insertContainer(container)
	if err != nil {
		return 0, err
	}

	query, err := c.db.Query("SELECT id FROM models WHERE container = $1", cId)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err = query.Scan(&id)
		return id, err
	}

	insertQuery, err := c.db.Query("INSERT INTO models (container) VALUES ($1) RETURNING id", cId)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := insertQuery.Close(); err != nil {
			klog.Errorf("cannot close query object")
		}
	}()

	if !insertQuery.Next() {
		return 0, fmt.Errorf("insertion failed")
	}

	var id int64
	err = insertQuery.Scan(&id)
	return id, err
}

func (c contractHandler) insertMachine(machine string) error {
	query, err := c.db.Query("SELECT id FROM machines WHERE id = $1", machine)
	if err != nil {
		return err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		return nil
	}

	_, err = c.db.Exec("INSERT INTO machines (id) VALUES ($1)", machine)
	return err
}

func (c contractHandler) insertTechnicalContainer(system int64, containers []Container, contract string) error {
	for _, container := range containers {
		id, err := c.insertContainer(container)
		if err != nil {
			return err
		}

		if _, err := c.db.Exec("INSERT INTO technical_containers (contract, container, system) VALUES ($1, $2, $3)", contract, id, system); err != nil {
			return err
		}
	}

	return nil
}

func (c contractHandler) stringArrayToString(arg []string) string {
	if len(arg) == 0 {
		return "{}"
	}
	return fmt.Sprintf("{'%s'}", strings.Join(arg, "','"))
}

func (c contractHandler) insertContainer(container Container) (int64, error) {
	query, err := c.db.Query("SELECT id FROM containers WHERE url = $1 AND tag = $2 AND arguments = $3 AND environment = $4",
		container.Url,
		container.Tag,
		c.stringArrayToString(container.Arguments),
		c.stringArrayToString(container.Environment),
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err := query.Scan(&id)
		return id, err
	}

	queryInsert, err := c.db.Query("INSERT INTO containers (url, tag, arguments, environment) VALUES ($1, $2, $3, $4) RETURNING  id",
		container.Url,
		container.Tag,
		c.stringArrayToString(container.Arguments),
		c.stringArrayToString(container.Environment),
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := queryInsert.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !queryInsert.Next() {
		return 0, fmt.Errorf("insertion was not successfull in insertContainer")
	}

	var id int64
	err = queryInsert.Scan(&id)
	return id, err

}

func (c contractHandler) insertPermission(write bool, contract string, orgs []string) error {
	for _, org := range orgs {
		id, err := c.insertOrganisations(org)
		if err != nil {
			return err
		}

		var table string
		if write {
			table = "write_permissions"
		} else {
			table = "read_permissions"
		}

		queryString := fmt.Sprintf("INSERT INTO %s (contract, organisation) VALUES ($1, $2)", table)
		if _, err := c.db.Exec(queryString, contract, id); err != nil {
			return err
		}
	}

	return nil
}

func (c contractHandler) insertPartners(contract string, partners []string) error {
	for _, partner := range partners {
		org, err := c.insertOrganisations(partner)
		if err != nil {
			return err
		}

		_, err = c.db.Exec("INSERT INTO partners (contract, organisation) VALUES ($1, $2)", contract, org)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c contractHandler) insertOrganisations(organisation string) (int64, error) {
	query, err := c.db.Query("SELECT id FROM organisations WHERE name = $1", organisation)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err := query.Scan(&id)
		return id, err
	}

	queryInsert, err := c.db.Query("INSERT INTO organisations (name) VALUES ($1) RETURNING id", organisation)
	if err != nil {
		return 0, err
	}
	klog.Infof("insert organisation %s", organisation)

	defer func() {
		if err := queryInsert.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !queryInsert.Next() {
		return 0, fmt.Errorf("insertion was not successfull in insertOrganisation")
	}

	var id int64
	err = queryInsert.Scan(&id)
	return id, err
}

func (c contractHandler) insertMachineSensor(machine string, sensor int64) (int64, error) {
	query, err := c.db.Query("SELECT id FROM machine_sensors WHERE machine = $1 AND sensor = $2", machine, sensor)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err := query.Scan(&id)
		return id, err
	}

	queryInsert, err := c.db.Query("INSERT INTO machine_sensors (machine, sensor) VALUES ($1, $2) RETURNING id", machine, sensor)
	if err != nil {
		return 0, err
	}

	if !queryInsert.Next() {
		return 0, fmt.Errorf("inertion was not successfull")
	}

	var id int64
	err = queryInsert.Scan(&id)
	return id, err
}

func (c contractHandler) insertContractMachineSensor(contract string, machineSensor int64) (int64, error) {
	query, err := c.db.Query("SELECT id FROM contract_machine_sensors WHERE contract = $1 AND machine_sensor = $2", contract, machineSensor)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err := query.Scan(&id)
		return id, err
	}

	queryInsert, err := c.db.Query("INSERT INTO contract_machine_sensors (contract, machine_sensor) VALUES ($1, $2) RETURNING id", contract, machineSensor)
	if err != nil {
		return 0, err
	}

	if !queryInsert.Next() {
		return 0, fmt.Errorf("inertion was not successfull")
	}

	var id int64
	err = queryInsert.Scan(&id)
	return id, err
}

func (c contractHandler) insertStorageDuration(contractMachineSensor int64, duration string, system int64) error {
	_, err := c.db.Exec("INSERT INTO storage_duration (system, contract_machine_sensor, duration) VALUES ($1, $2, $3)",
		system,
		contractMachineSensor,
		duration,
	)
	return err
}

func (c contractHandler) insertOnlySensor(sensor string, meta interface{}) (int64, error) {
	data, err := json.Marshal(meta)
	if err != nil {
		klog.Errorf("crasy 0")
		return 0, err
	}

	var query, insertQuery *sql.Rows
	if string(data) == "{}" {
		query, err = c.db.Query("SELECT id FROM sensors WHERE transmitted_id = $1", sensor)
	} else {
		query, err = c.db.Query("SELECT id FROM sensors WHERE transmitted_id = $1 AND meta = $2", sensor, string(data))
	}
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if query.Next() {
		var id int64
		err := query.Scan(&id)
		return id, err
	}

	if string(data) == "{}" {
		insertQuery, err = c.db.Query("INSERT INTO sensors (transmitted_id) VALUES ($1) RETURNING id", sensor)
	} else {
		insertQuery, err = c.db.Query("INSERT INTO sensors (transmitted_id, meta) VALUES ($1, $2) RETURNING id", sensor, string(data))
	}
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := insertQuery.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !insertQuery.Next() {
		return 0, fmt.Errorf("the id of the sensor in db will not be returned")
	}

	var id int64
	err = insertQuery.Scan(&id)
	return id, err
}

func (c contractHandler) DeleteContract(contract string) error {
	_, err := c.db.Exec("UPDATE contracts SET active = false WHERE id = $1", contract)
	return err
}

func (c contractHandler) GetContract(contract string) (Contract, error) {
	query, err := c.db.Query("SELECT contract FROM contracts WHERE id = $1", contract)
	if err != nil {
		return Contract{}, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var contractJson string
	var con Contract

	if !query.Next() {
		return con, nil
	}

	if err = query.Scan(&contractJson); err != nil {
		return Contract{}, err
	}

	err = json.Unmarshal([]byte(contractJson), &con)

	return con, err
}

func NewContractHandler(db *sql.DB, system string) ContractHandler {
	return contractHandler{db: db, system: system}
}
