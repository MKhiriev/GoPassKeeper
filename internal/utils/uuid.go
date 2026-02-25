package utils

import "github.com/google/uuid"

type UUIDGenerator struct {
}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (g *UUIDGenerator) Generate() string {
	v7, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}

	return v7.String()
}
