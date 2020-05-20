package models

type Data struct {
	Machine   string        `json:"machine:noempty"`
	Sensor    string        `json:"string:noempty"`
	MessageId int           `json:"message_id:noempty"`
	From      string        `json:"from"`
	Columns   []DataColumnn `json:"columns"`
	Data      [][]string    `json:"data"`
	Meta      []DataMeta    `json:"meta"`
}

type DataColumnn struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	AllowedValues string `json:"allowed_values:noempty"`
}

type DataMeta struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description:noempty"`
	Value       string `json:"value:noempty"`
}

type SendData struct {
	Machine string        `json:"machine:noempty"`
	Sensor  string        `json:"string:noempty"`
	From    string        `json:"from"`
	Columns []DataColumnn `json:"columns"`
	Data    [][]string    `json:"data"`
	Meta    []DataMeta    `json:"meta"`
}
