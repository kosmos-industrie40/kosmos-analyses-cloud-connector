package machineData

type Model struct {
	Body struct {
		MachineID string `json:"machineID"`
		Sensor    string `json:"sensor"`
		Timestamp string `json:"timestamp"`
		Columns   []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Meta struct {
				Future      interface{} `json:"future,omitempty"`
				Description string      `json:"description"`
				Unit        string      `json:"unit"`
			} `json:"meta"`
		} `json:"columns"`
		Data     [][]string `json:"data"`
		Metadata []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Value       string `json:"value"`
		} `json:"meta"`
	} `json:"body"`
	Signature string `json:"signature"`
}
