package models

type Analyse struct {
	From string `json:"from"`
	Timestamp string `json:"timestamp"`
	Model struct{
		Tag string `json:"tag"`
		Url string `json:"url"`
	} `json:"model"`
	Type string `json:"type"`
	Calculated struct{
		Message struct {
			Machine string `json:"machine"`
			Sensor string `json:"sensor"`
		} `json:"message"`
		Received string `json:"received"`
	} `json:"calculated"`
	Results interface{} `json:"results"`
	Signature string `json:"signature"`
}

type AnalysisText struct {
	Total string `json:"total"`
	Predict int `json:"predict"`
	Parts []*struct{
		Machine string `json:"machine"`
		Result string `json:"result"`
		Predict int `json:"predict"`
		Sensors []*struct{
			Sensor string `json:"sensor"`
			Result string `json:"result"`
			Predict int `json:"predict"`
		} `json:"sensors"`
	}
}

type AnalysisTimeSeries struct {
	Columns []struct{
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns"`
	Data [][]string `json:"data"`
	Signature string `json:"signature"`
}