package domain

import (
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
)

type Gateway struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewGateway(id string, name string) (*Gateway, error) {
	if id == "" {
		return nil, helpers.ErrGatewayIDRequired
	}
	if name == "" {
		return nil, helpers.ErrGatewayNameRequired
	}
	return &Gateway{
		ID:   id,
		Name: name,
	}, nil
}

func (g *Gateway) Validate() error {
	if g.ID == "" {
		return helpers.ErrGatewayIDRequired
	}
	if g.Name == "" {
		return helpers.ErrGatewayNameRequired
	}
	return nil
}
