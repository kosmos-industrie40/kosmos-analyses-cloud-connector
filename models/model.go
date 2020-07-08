package models

type Model struct {
	Tag string `json:"tag"`
	Url string `json:"url"`
}

type UpdateModelState struct {
	State  string  `json:"state"`
	Models []Model `json:"models"`
}
