package logic

import (
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

type Authentication interface {
	Authentication(database.Postgres)
	Login(string, string) (string, error)
	User(string) (string, error)
	Logout(string) error
}

type Analyses interface {
	Analyses(database.Postgres)
	InsertResult(string, string, string, []models.UploadResult) error
	GetSpecificResult(string, string) ([]byte, error)
	GetResultSet(string, map[string][]string) ([]models.ResultList, error)
}

type Contract interface {
	GetContract(string) (models.Contract, error)
	GetAllContracts() ([]string, error)
	InsertContract(models.Contract) error
	Contract(database.Postgres)
	DeleteContract(string) error
}

type Model interface {
	Model(database.Postgres)
	GetModel(string) ([]models.Model, error)
	UpdateModel(string, models.UpdateModelState) error
}
