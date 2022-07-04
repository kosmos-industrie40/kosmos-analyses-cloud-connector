package auth

import (
	"github.com/google/uuid"
)

type TokenGenerate interface {
	Generate() string
}

func NewTokenGeneratorUuid() TokenGenerate {
	return uuidTokenGenerate{}
}

type uuidTokenGenerate struct{}

func (uuidTokenGenerate) Generate() string {
	return uuid.New().String()
}
