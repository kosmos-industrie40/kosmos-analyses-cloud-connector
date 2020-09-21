package models

type Analyse struct {
	Machine MachineSensor
	prev Model
	next Model
}