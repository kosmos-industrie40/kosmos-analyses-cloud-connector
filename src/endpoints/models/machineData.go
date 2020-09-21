package models

type MachineData struct {
	Machine   string `json:"machine"`
	Sensor    string `json:"sensor"`
	Timestamp string `json:"timestamp"`
	Columns   []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	} `json:"columns"`
	Data     [][]string `json:"data"`
	Metadata []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Value       string `json:"value"`
	} `json:"meta"`
	Signature string `json:"signature"`
}
