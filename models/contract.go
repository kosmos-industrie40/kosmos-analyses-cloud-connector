package models

type Contract struct {
	ModelsCloud []Model            `json:"modelsCloud"`
	ModelsEdge  []Model            `json:"modelsEdge"`
	ContractId  string             `json:"contractId"`
	Machines    []ContractMachines `json:"machines"`
}

type ContractMachines struct {
	MachineId string            `json:"machineId"`
	Sensors   []ContractSensors `json:"sensors"`
}

type ContractSensors struct {
	SensorId string  `json:"sensorId"`
	Model    []Model `json:"model"`
}
