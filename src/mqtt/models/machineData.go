package models

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Meta struct {
		Future      interface{} `json:"future,omitempty"`
		Unit        string      `json:"unit"`
		Description string      `json:"description"`
	} `json:"meta"`
}

type MachineData struct {
	Body struct {
		MachineID string     `json:"machineID"`
		Timestamp string     `json:"timestamp"`
		Columns   []Column    `json:"columns"`
		Data      [][]string `json:"data"`
		Metadata  []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Value       string `json:"value"`
		} `json:"meta"`
	} `json:"body"`
	Signature string `json:"signature"`
}
