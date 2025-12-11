// internal/domain/entities/gateway.go
package entities

// Gateway - доменная сущность платежного шлюза [35]
type Gateway struct {
	ID   string
	Name string
}

// NewGateway создает новый платежный шлюз
func NewGateway(id string, name string) *Gateway {
	return &Gateway{
		ID:   id,
		Name: name,
	}
}
