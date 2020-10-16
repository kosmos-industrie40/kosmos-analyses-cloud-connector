package models

type ContractRemoval struct {
	ContractID string `json:"contract"`
	Blame string `json:"blame"`
	Signature string `json:"signature"`
}
