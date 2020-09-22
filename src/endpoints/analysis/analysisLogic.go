package analysis

import (
	"encoding/json"
	"fmt"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/analysis/models"
)

type AnalyseLogic interface {
	GetResultSet(string, map[string][]string) ([]byte, error)
	InsertResult(string, string, string, []models.Analysis) error
	GetSpecificResult(string, int64) ([]byte, error)
}

type analyseLogic struct {
	analysisHandler models.AnalysisHandler
	resultHandler models.ResultListHandler
}

func NewAnalyseLogic(resultHandler models.ResultListHandler, analysisHandler models.AnalysisHandler) AnalyseLogic {
	return analyseLogic{resultHandler: resultHandler, analysisHandler: analysisHandler}
}

func (a analyseLogic) GetResultSet(contractID string, queryOptions map[string][]string) ([]byte, error) {
	return a.resultHandler.Get(contractID, queryOptions)
}

func (a analyseLogic) GetSpecificResult(contractID string, resultID int64) ([]byte, error) {
	data, err := a.analysisHandler.Query(contractID, resultID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}


func (a analyseLogic) InsertResult(contractID, machineID, sensorId string, models []models.Analysis) error {
	for _, model := range models {
		if !model.Validate() {
			return fmt.Errorf("on of the transmitted models is not valid")
		}

		if err :=  a.analysisHandler.Insert(contractID, machineID, sensorId, model); err != nil {
			return err
		}
	}

	return nil
}

