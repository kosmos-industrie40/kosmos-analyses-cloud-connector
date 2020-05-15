package logic

import (
	"fmt"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"

	"k8s.io/klog"
)

func GetAllContracts(db database.Postgres) ([]string, error) {
	var ret []string
	var query interface{}
	query = ret

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
	ret := models.Contract{ContractId: contract}

	var cloudId, edgeId []int
	var cModelId interface{}
	cModelId = cloudId
	var val []*interface{}
	val = append(val, &cModelId)

	// get model ids from cloud and edge model
	if err := db.Query("model_cloud", []string{"model"}, []string{"contracts"}, val, []interface{}{contract}); err != nil {
		return models.Contract{}, err
	}
	cloudId = cModelId.([]int)

	cModelId = edgeId
	if err := db.Query("model_cloud", []string{"model"}, []string{"contracts"}, val, []interface{}{contract}); err != nil {
		return models.Contract{}, err
	}
	edgeId = cModelId.([]int)

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

	return ret, nil
}

func queryModel(modelId int, db database.Postgres) (models.Model, error) {
	var ret models.Model

	var url string
	var tag string
	var cUrl interface{}
	var cTag interface{}
	cUrl = url
	cTag = tag

	value := []*interface{}{&cUrl, &cTag}
	err := db.Query("Models", []string{"url", "tag"}, []string{"id"}, value, []interface{}{modelId})
	if err != nil {
		return models.Model{}, err
	}

	ret.Tag = cTag.(string)
	ret.Url = cUrl.(string)

	return ret, nil
}

func queryMachine(modelId int, db database.Postgres) ([]models.ContractMachines, error) {
	var ret []models.ContractMachines

	var id []string
	var cId interface{}
	cId = id
	val := []*interface{}{&cId}

	if err := db.Query("machine_contract", []string{"machine"}, []string{"contract"}, val, []interface{}{modelId}); err != nil {
		return nil, err
	}

	// query machine here if there are more informations given than the id

	id = cId.([]string)
	for _, v := range id {
		ret = append(ret, models.ContractMachines{MachineId: v})
	}

	return ret, nil
}

func querySensor(modelId int, db database.Postgres) ([]models.ContractSensors, error) {
	var ret []models.ContractSensors

	var id []int
	var sensorId []int
	var cId interface{}
	var cSensorId interface{}
	cId = id
	cSensorId = sensorId
	val := []*interface{}{&cId, &cSensorId}

	if err := db.Query("machine_sensor", []string{"sensor", "id"}, []string{"machine"}, val, []interface{}{modelId}); err != nil {
		return nil, err
	}

	// query machine here if there are more informations given than the id

	id = cId.([]int)
	sensorId = cSensorId.([]int)

	if len(id) != len(sensorId) {
		return nil, fmt.Errorf("the return value from id to sendor id are not equal")
	}

	for i, j := range id {

		var sensor models.ContractSensors

		var transId string
		var cTransId interface{}
		cTransId = transId
		val := []*interface{}{&cTransId}

		if err := db.Query("sensor", []string{"transmitted_id"}, []string{"id"}, val, []interface{}{sensorId[i]}); err != nil {
			return nil, fmt.Errorf("could not query sensor id")
		}

		sensor.SensorId = cTransId.(string)

		var modelsId []int
		var cModelsId interface{}
		cModelsId = modelsId
		valu := []*interface{}{&cModelsId}

		if err := db.Query("machine_sensor", []string{"model"}, []string{"machine_sensor"}, valu, []interface{}{j}); err != nil {
			return nil, fmt.Errorf("could not query sensor id")
		}

		modelsId = cModelsId.([]int)

		for _, k := range modelsId {
			mod, err := queryModel(k, db)
			if err != nil {
				return nil, fmt.Errorf("could not query model")
			}
			sensor.Model = append(sensor.Model, mod)
		}

		ret = append(ret, sensor)

	}
	return ret, nil
}

func InsertContract(contract models.Contract, db database.Postgres) error {
	if err := db.Insert("contracts", []string{"id"}, []interface{}{contract.ContractId}); err != nil {
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
	err := db.Update("contract", []string{"contract"}, []interface{}{contract}, []string{}, []interface{}{false})
	return err
}

func insertModel(model models.Model, db database.Postgres) (int, error) {
	id := -1
	var cId interface{}
	cId = id
	value := []*interface{}{&cId}

	if err := db.Query("models", []string{"id"}, []string{"url", "tag"}, value, []interface{}{model.Url, model.Tag}); err != nil {
		return 0, err
	}

	id = cId.(int)
	if id != -1 {
		return id, nil
	}

	if err := db.Insert("models", []string{"url", "tag"}, []interface{}{model.Url, model.Tag}); err != nil {
		return 0, err
	}

	if err := db.Query("models", []string{"id"}, []string{"url", "tag"}, value, []interface{}{model.Url, model.Tag}); err != nil {
		return 0, err
	}

	id = cId.(int)

	return id, nil
}

func insertModelsCloud(contractId string, models []models.Model, db database.Postgres) error {
	for _, v := range models {
		id, err := insertModel(v, db)
		if err != nil {
			return err
		}
		if err := db.Insert("models_cloud", []string{"contract", "model"}, []interface{}{contractId, id}); err != nil {
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
		if err := db.Insert("models_edge", []string{"contract", "model"}, []interface{}{contractId, id}); err != nil {
			return err
		}
	}
	return nil
}

func insertMachine(contractId string, machines []models.ContractMachines, db database.Postgres) error {
	for _, v := range machines {
		if err := db.Insert("machine", []string{"id"}, []interface{}{v}); err != nil {
			return err
		}

		if err := db.Insert("machine_contract", []string{"contract", "machine"}, []interface{}{contractId, v}); err != nil {
			return err
		}

		if err := insertSensor(v.MachineId, v.Sensors, db); err != nil {
			return err
		}
	}
	return nil
}

func insertSensor(machineId string, sensors []models.ContractSensors, db database.Postgres) error {
	for _, v := range sensors {

		id := -1
		var cId interface{}
		cId = id
		value := []*interface{}{&cId}

		if err := db.Query("sensors", []string{"id"}, []string{"transmitted_id"}, value, []interface{}{v.SensorId}); err != nil {
			return err
		}

		id = cId.(int)
		if id != -1 {
			if err := db.Insert("sensors", []string{"transmitted_id"}, []interface{}{v.SensorId}); err != nil {
				return err
			}
			if err := db.Query("sensors", []string{"id"}, []string{"transmitted_id"}, value, []interface{}{v.SensorId}); err != nil {
				return err
			}
			id = cId.(int)
		}

		msId := -1
		var cMsId interface{}
		cMsId = msId
		val := []*interface{}{&cMsId}

		if err := db.Query("machine_sensor", []string{"id"}, []string{"machine", "sensor"}, val, []interface{}{machineId, id}); err != nil {
			return err
		}

		msId = cMsId.(int)
		if msId == -1 {
			if err := db.Insert("machine_sensor", []string{"machine", "sensor"}, []interface{}{machineId, v.SensorId}); err != nil {
				return err
			}
			if err := db.Query("machine_sensor", []string{"id"}, []string{"machine", "sensor"}, val, []interface{}{machineId, id}); err != nil {
				return err
			}

			msId = cMsId.(int)
		}

		for _, j := range v.Model {
			mId, err := insertModel(j, db)
			if err != nil {
				return err
			}

			if err := db.Insert("sensor_model", []string{"machine_sensor", "model"}, []interface{}{msId, mId}); err != nil {
				return err
			}
		}
	}
	return nil
}
