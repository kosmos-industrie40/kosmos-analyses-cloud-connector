package logic

import (
	"fmt"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"

	"k8s.io/klog"
)

func GetAllContracts(db database.Postgres) ([]string, error) {
	var ret []string
	var query interface{} = ret

	var val []*interface{}
	val = append(val, &query)

	columns := []string{"id"}
	var parameters []string
	var parameterValue []interface{}

	err := db.Query("contract", columns, parameters, val, parameterValue)

	ret = query.([]string)

	return ret, err
}

func GetContract(contract string, db database.Postgres) (models.Contract, error) {
	klog.Infof("contract is %s\n", contract)
	ret := models.Contract{ContractId: contract}

	var cloudId, edgeId []int64
	var cModelId interface{}
	cModelId = cloudId
	var val []*interface{}
	val = append(val, &cModelId)

	// get model ids from cloud and edge model
	if err := db.Query("model_cloud", []string{"model"}, []string{"contract"}, val, []interface{}{contract}); err != nil {
		return models.Contract{}, err
	}
	cloudId = cModelId.([]int64)

	cModelId = edgeId
	if err := db.Query("model_edge", []string{"model"}, []string{"contract"}, val, []interface{}{contract}); err != nil {
		return models.Contract{}, err
	}
	edgeId = cModelId.([]int64)

	// get complete model
	for _, v := range edgeId {
		mod, err := queryModel(v, db)
		if err != nil {
			klog.Errorf("could not query model with id %d\n", v)
			return models.Contract{}, err
		}
		ret.ModelsEdge = append(ret.ModelsEdge, mod)
	}
	for _, v := range cloudId {
		mod, err := queryModel(v, db)
		if err != nil {
			klog.Errorf("could not query model with id %d\n", v)
			return models.Contract{}, err
		}
		ret.ModelsCloud = append(ret.ModelsCloud, mod)
	}

	mach, err := queryMachine(contract, db)
	if err != nil {
		return models.Contract{}, err
	}

	for i, v := range mach {
		sensor, err := querySensor(v.MachineId, db)
		if err != nil {
			return models.Contract{}, err
		}
		mach[i].Sensors = sensor
	}

	ret.Machines = mach

	if ret.ModelsCloud == nil && ret.ModelsEdge == nil && ret.Machines == nil {
		return models.Contract{}, nil
	}

	return ret, nil
}

func queryModel(modelId int64, db database.Postgres) (models.Model, error) {
	var ret models.Model

	var url string
	var tag string
	var cUrl interface{}
	var cTag interface{}
	cUrl = url
	cTag = tag

	value := []*interface{}{&cUrl, &cTag}
	err := db.Query("model", []string{"url", "tag"}, []string{"id"}, value, []interface{}{modelId})
	if err != nil {
		return models.Model{}, err
	}

	ret.Tag = cTag.(string)
	ret.Url = cUrl.(string)

	return ret, nil
}

func queryMachine(contractId string, db database.Postgres) ([]models.ContractMachines, error) {
	var ret []models.ContractMachines

	var id []string
	var cId interface{} = id
	val := []*interface{}{&cId}

	if err := db.Query("machine_contract", []string{"machine"}, []string{"contract"}, val, []interface{}{contractId}); err != nil {
		return nil, err
	}

	// query machine here if there are more informations given than the id

	id = cId.([]string)
	for _, v := range id {
		ret = append(ret, models.ContractMachines{MachineId: v})
	}

	return ret, nil
}

func querySensor(machine string, db database.Postgres) ([]models.ContractSensors, error) {
	var ret []models.ContractSensors

	var id, sensorId []int64
	var cId interface{} = id
	val := []*interface{}{&cId}

	if err := db.Query("machine_sensor", []string{"id"}, []string{"machine"}, val, []interface{}{machine}); err != nil {
		return nil, err
	}
	id = cId.([]int64)

	// query machine here if there are more informations given than the id

	for _, v := range id {
		var sensId []int64
		var cSensorId interface{} = sensId
		val := []*interface{}{&cSensorId}
		if err := db.Query("machine_sensor", []string{"sensor"}, []string{"id"}, val, []interface{}{v}); err != nil {
			return nil, err
		}
		sensorId = append(sensorId, cSensorId.([]int64)...)
	}

	if len(id) != len(sensorId) {
		return nil, fmt.Errorf("the return value from id to sendor id are not equal")
	}

	klog.Infof("the ids from the machine_sensor: %v", id)
	klog.Infof("the sensorIds from the machine_sensor: %v", sensorId)

	for i, j := range id {

		var sensor models.ContractSensors

		var transId string
		var cTransId interface{} = transId
		val := []*interface{}{&cTransId}

		if err := db.Query("sensor", []string{"transmitted_id"}, []string{"id"}, val, []interface{}{sensorId[i]}); err != nil {
			return nil, fmt.Errorf("could not query sensor id %s\n", err)
		}

		sensor.SensorId = cTransId.(string)

		var modelsId []int64
		var cModelsId interface{} = modelsId
		valu := []*interface{}{&cModelsId}

		klog.Infof("machine_sensor id: %d\n", j)
		if err := db.Query("sensor_model", []string{"model"}, []string{"sensor"}, valu, []interface{}{j}); err != nil {
			return nil, fmt.Errorf("could not model: %s\n", err)
		}

		modelsId = cModelsId.([]int64)

		for _, k := range modelsId {
			klog.Infof("model id: %d", k)
			mod, err := queryModel(k, db)
			if err != nil {
				return nil, fmt.Errorf("could not query model")
			}
			sensor.Model = append(sensor.Model, mod)
		}

		klog.Infof("sensor: %v\n", sensor)
		ret = append(ret, sensor)

	}
	return ret, nil
}

func InsertContract(contract models.Contract, db database.Postgres) error {
	if err := db.Insert("contract", []string{"id"}, []interface{}{contract.ContractId}); err != nil {
		return err
	}
	if err := insertModelsCloud(contract.ContractId, contract.ModelsCloud, db); err != nil {
		return err
	}
	if err := insertModelsEdge(contract.ContractId, contract.ModelsEdge, db); err != nil {
		return err
	}
	if err := insertMachine(contract.ContractId, contract.Machines, db); err != nil {
		return err
	}
	return nil
}

func DeleteContract(contract string, db database.Postgres) error {
	var upV bool = true
	err := db.Update("contract", []string{"id"}, []interface{}{contract}, []string{"delete"}, []interface{}{upV})
	return err
}

func insertModel(model models.Model, db database.Postgres) (int64, error) {
	var id int64 = -1
	var cId interface{} = id
	value := []*interface{}{&cId}

	if err := db.Query("model", []string{"id"}, []string{"url", "tag"}, value, []interface{}{model.Url, model.Tag}); err != nil {
		return 0, err
	}

	id = cId.(int64)
	if id != -1 {
		return id, nil
	}

	if err := db.Insert("model", []string{"url", "tag"}, []interface{}{model.Url, model.Tag}); err != nil {
		return 0, err
	}

	if err := db.Query("model", []string{"id"}, []string{"url", "tag"}, value, []interface{}{model.Url, model.Tag}); err != nil {
		return 0, err
	}

	id = cId.(int64)

	return id, nil
}

func insertModelsCloud(contractId string, models []models.Model, db database.Postgres) error {
	for _, v := range models {
		id, err := insertModel(v, db)
		if err != nil {
			return err
		}
		if err := db.Insert("model_cloud", []string{"contract", "model"}, []interface{}{contractId, id}); err != nil {
			return err
		}
	}
	return nil
}

func insertModelsEdge(contractId string, models []models.Model, db database.Postgres) error {
	for _, v := range models {
		id, err := insertModel(v, db)
		if err != nil {
			return err
		}
		if err := db.Insert("model_edge", []string{"contract", "model"}, []interface{}{contractId, id}); err != nil {
			return err
		}
	}
	return nil
}

func insertMachine(contractId string, machines []models.ContractMachines, db database.Postgres) error {
	for _, v := range machines {
		if err := db.Insert("machine", []string{"id"}, []interface{}{v.MachineId}); err != nil {
			return err
		}

		if err := db.Insert("machine_contract", []string{"contract", "machine"}, []interface{}{contractId, v.MachineId}); err != nil {
			return err
		}

		if err := insertSensor(v.MachineId, v.Sensors, db); err != nil {
			return err
		}
	}
	return nil
}

func insertSensor(machineId string, sensors []models.ContractSensors, db database.Postgres) error {
	for _, sensor := range sensors {

		var id int64 = -1
		var cId interface{} = id
		value := []*interface{}{&cId}

		if err := db.Query("sensor", []string{"id"}, []string{"transmitted_id"}, value, []interface{}{sensor.SensorId}); err != nil {
			return err
		}

		id = cId.(int64)
		if id == -1 {
			if err := db.Insert("sensor", []string{"transmitted_id"}, []interface{}{sensor.SensorId}); err != nil {
				return err
			}
			if err := db.Query("sensor", []string{"id"}, []string{"transmitted_id"}, value, []interface{}{sensor.SensorId}); err != nil {
				return err
			}
			id = cId.(int64)
		}

		var msId int64 = -1
		var cMsId interface{} = msId
		val := []*interface{}{&cMsId}

		if err := db.Query("machine_sensor", []string{"id"}, []string{"machine", "sensor"}, val, []interface{}{machineId, id}); err != nil {
			return err
		}

		msId = cMsId.(int64)
		if msId == -1 {
			if err := db.Insert("machine_sensor", []string{"machine", "sensor"}, []interface{}{machineId, id}); err != nil {
				return err
			}
			if err := db.Query("machine_sensor", []string{"id"}, []string{"machine", "sensor"}, val, []interface{}{machineId, id}); err != nil {
				return err
			}

			msId = cMsId.(int64)
		}

		klog.Infof("there are %d models", len(sensor.Model))
		for _, j := range sensor.Model {
			klog.Infof("model tag: %s and url: %s\n", j.Tag, j.Url)
			mId, err := insertModel(j, db)
			if err != nil {
				return err
			}

			if err := db.Insert("sensor_model", []string{"sensor", "model"}, []interface{}{msId, mId}); err != nil {
				return err
			}
		}
	}
	return nil
}
