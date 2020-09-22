package models

type TextResult struct {
		Total   string `json:"total"`
		Predict int    `json:"predict"`
		Parts   []struct {
			Machine string `json:"machine"`
			Result  string `json:"result"`
			Predict int    `json:"predict"`
			Sensors []struct {
				Sensor  string `json:"sensor"`
				Result  string `json:"result"`
				Predict int    `json:"predict"`
			} `json:"sensors"`
		} `json:"parts"`
}
