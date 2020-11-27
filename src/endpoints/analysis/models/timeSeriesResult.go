package models

type TimeSeriesResult struct {
	Columns []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns"`
	Data      [][]string `json:"data"`
	Signature string     `json:"signature"`
}
