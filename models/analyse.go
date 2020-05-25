package models

type ResultList struct {
	Id 		int64 	`json:"resultId"`
	Machine string 	`json:"machine"`
	Date 	int64 	`json:"date"`
}

type UploadResult struct {
	Date int64 `json:"date"`
}