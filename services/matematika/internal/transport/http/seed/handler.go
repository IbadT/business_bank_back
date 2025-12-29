package seed

import (
	"net/http"

	seedservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/seed"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	seedService seedservice.SeedService
}

func NewHandler(seedService seedservice.SeedService) *Handler {
	return &Handler{seedService}
}


// Seed выполняет заполнение базы данных seed данными
// @Summary      Заполнить базу данных seed данными
// @Description  Заполняет базу данных тестовыми данными (пользователи, праздники, транзакции и т.д.)
// @Tags         seed
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.MessageResponse  "База данных успешно заполнена"
// @Failure      500  {object}  dto.ErrorResponse  "Ошибка при заполнении базы данных"
// @Router       /api/seed [post]
func (h *Handler) Seed(c echo.Context) error {
	op := "http.handler.seed"
	log := logger.GetLogger().WithOperation(op)
	
	log.Info("Starting database seed")
	
	if h.seedService == nil {
		log.Error(nil, "Seed service not available")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: helpers.ErrMsgSeedServiceNotAvailable,
			Code:  http.StatusInternalServerError,
		})
	}

	if err := h.seedService.SeedDatabase(); err != nil {
		log.Error(err, "Failed to seed database")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToSeedDatabase,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	log.Success("Database seeded successfully")

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Database seeded successfully",
		Code:    http.StatusOK,
	})
}
