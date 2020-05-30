package logic

import (
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

type Mod struct {
	db database.Postgres
}

func (m Mod) Model(db database.Postgres) {
	m.db = db
}

// GetModel returns all upgradable models for a specific contract
// the state which are upgradable models have is 'UPDATE'
func (m Mod) GetModel(contractId string) ([]models.Model, error) {
	var id []int64
	var cId interface{} = id
	value := []*interface{}{&cId}

	if err := m.db.Query("model_update", []string{"model"}, []string{"contract", "status"}, value, []interface{}{contractId, "UPDATE"}); err != nil {
		return nil, err
	}

	id = cId.([]int64)

	var ret []models.Model
	for _, mId := range id {
		var tag, url string
		var cTag interface{} = tag
		var cUrl interface{} = url
		value := []*interface{}{&cTag, &cUrl}

		if err := m.db.Query("model", []string{"tag", "url"}, []string{"id"}, value, []interface{}{mId}); err != nil {
			return nil, err
		}

		ret = append(ret, models.Model{Tag: cTag.(string), Url: cUrl.(string)})
	}

	return ret, nil
}

// UpdateModel is the logic to update a model state in the database
func (m Mod) UpdateModel(contractId string, mods models.UpdateModelState) error {
	for _, model := range mods.Models {
		var id int64
		var cId interface{} = id
		value := []*interface{}{&cId}

		if err := m.db.Query("model", []string{"id"}, []string{"tag", "url"}, value, []interface{}{model.Tag, model.Url}); err != nil {
			return err
		}

		id = cId.(int64)
		if err := m.db.Update("model_update", []string{"model", "contract"}, []interface{}{id, contractId}, []string{"status"}, []interface{}{mods.State}); err != nil {
			return err
		}
	}
	return nil
}
