package logic

import (
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

func GetAllContracts(db database.Postgres) ([]string, error) {
	return nil, nil
}

func GetContract(contract string, db database.Postgres) (models.Contract, error){
	return models.Contract{}, nil
}

func InsertContract(contract models.Contract, db database.Postgres) error {
	return nil
}

func DeleteContract(contract string, db database.Postgres) error {
	return nil
}

func insertModels(models []models.Model) (int, error) {
	return -1, nil
}

func insertModelsCloud(contractId string, models []models.Model) error {
	return nil
}

func insertModelsEdge(contractId string, models []models.Model) error {
	return nil
}

func insertMachine(contractId string, machines []models.ContractMachines) error {
	return nil
}

func insertSensor(machineId string, machine []models.ContractSensors) error {
	return nil
}
