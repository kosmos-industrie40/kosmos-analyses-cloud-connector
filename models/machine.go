package models

type Data struct {
	Machine   string        `json:"machine,omitempty"`
	Timestamp string        `json:"timestamp"`
	Sensor    string        `json:"sensor"`
	MessageId int           `json:"message_id,omitempty"`
	From      string        `json:"from"`
	Columns   []DataColumnn `json:"columns"`
	Data      [][]string    `json:"data"`
	Meta      []DataMeta    `json:"meta"`
	Signature string        `json:"signature,omitempty"`
}

type DataColumnn struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	AllowedValues string `json:"allowed_values,omitempty"`
}

type DataMeta struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
}

type SendData struct {
	//	Machine   string        `json:"machine,omitempty"`
	Timestamp string `json:"timestamp"`
	//	Sensor    string        `json:"sensor,omitempty"`
	From    string        `json:"from"`
	Columns []DataColumnn `json:"columns"`
	Data    [][]string    `json:"data"`
	Meta    []DataMeta    `json:"meta,omitempty"`
}
