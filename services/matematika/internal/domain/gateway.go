package domain

import (
	"errors"
)

type Gateway struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewGateway(id string, name string) (*Gateway, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	return &Gateway{
		ID:   id,
		Name: name,
	}, nil
}

func (g *Gateway) Validate() error {
	if g.ID == "" {
		return errors.New("id is required")
	}
	if g.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
