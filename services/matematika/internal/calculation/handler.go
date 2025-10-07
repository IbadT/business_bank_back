package calculation

import "github.com/labstack/echo"

type CalculationHandler struct {
	calcService CalculationService
}

func NewCalculationHandler(calcService CalculationService) *CalculationHandler {
	return &CalculationHandler{calcService: calcService}
}

// - POST /generate-statement
// - GET /statement/{id}/status
// - GET /statement/{id}/result

func (h *CalculationHandler) GenerateStatement(c echo.Context) error {

	return nil
}

func (h *CalculationHandler) GetStatementStatusByID(c echo.Context) error {
	return nil
}

func (h *CalculationHandler) GetStatementResultByID(c echo.Context) error {
	return nil
}
