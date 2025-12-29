package holiday

import (
	"errors"
	"net/http"
	"time"

	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	holidayService holidayservice.HolidayService
}

func NewHandler(s holidayservice.HolidayService) *Handler {
	return &Handler{s}
}

// ========================= HOLIDAYS =========================

// AddHoliday - добавление праздника
// @Summary      Добавление праздника
// @Description  Добавляет новый праздник в базу данных. Требуется авторизация. Поддерживаемые страны: RU, US.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.HolidayRequest  true  "Данные для добавления праздника"
// @Success      200      {object}  dto.MessageResponse  "Успешное добавление праздника"
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/holidays [post]
func (h *Handler) AddHoliday(c echo.Context) error {
	op := "http.handler.holiday.addHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.HolidayRequest

	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidRequestBody,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"name":    req.Name,
		"date":    req.HolidayDate,
		"country": req.Country,
	})
	log.Info("Adding holiday")

	// Парсим дату из формата YYYY-MM-DD
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		log.Error(err, "Invalid date format: %s", req.HolidayDate)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidDateFormatFull,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	if err := h.holidayService.AddHoliday(holidayDate, req.Name, req.Country); err != nil {
		log.Error(err, "Failed to add holiday")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrHolidayAlreadyExists) {
			statusCode = http.StatusConflict
		} else if errors.Is(err, helpers.ErrInvalidCountry) || errors.Is(err, helpers.ErrInvalidDate) {
			statusCode = http.StatusBadRequest
		}
		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToAddHoliday,
			Details: err.Error(),
			Code:    statusCode,
		})
	}
	
	log.Success("Holiday added successfully")
	
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday added successfully",
		Code:    http.StatusOK,
	})
}

// GetHolidays - получить список всех праздников для указанного года
// @Summary      Получить список всех праздников для указанного года
// @Description  Получает список всех праздников для указанного года. Формат года: YYYY (например, 2024). Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        year  query      string  true  "Год для получения праздников в формате YYYY" example:"2024"
// @Success      200      {object}  dto.GetHolidaysResponse  "Успешное получение списка праздников"
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/holidays [get]
func (h *Handler) GetHolidays(c echo.Context) error {
	op := "http.handler.holiday.getHolidays"
	log := logger.GetLogger().WithOperation(op)
	
	year := c.QueryParam("year")
	if year == "" {
		log.Warn("year parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgYearParameterRequired,
			Code:  http.StatusBadRequest,
		})
	}
	
	log = log.WithFields(logger.Fields{"year": year})
	log.Info("Getting holidays for year")
	
	yearTime, err := time.Parse("2006", year)
	if err != nil {
		log.Error(err, "Invalid year format: %s", year)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidYearFormat,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}
	holidays, err := h.holidayService.GetHolidays(yearTime)
	if err != nil {
		log.Error(err, "Failed to get holidays")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetHolidays,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	// Конвертируем domain.Holiday в dto.HolidayResponse
	holidayResponses := make([]dto.HolidayResponse, len(holidays))
	for i, holiday := range holidays {
		holidayResponses[i] = dto.HolidayResponse{
			HolidayDate: holiday.HolidayDate,
			Name:        holiday.Name,
			Country:     holiday.Country,
		}
	}

	log.WithFields(logger.Fields{"count": len(holidayResponses)}).Success("Holidays retrieved")

	return c.JSON(http.StatusOK, dto.GetHolidaysResponse{
		Holidays: holidayResponses,
		Year:     year,
	})
}

// IsHoliday - проверка является ли дата праздником
// @Summary      Проверка является ли дата праздником
// @Description  Проверяет, является ли указанная дата праздником. Формат даты: YYYY-MM-DD (например, 2024-12-15). Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        date  query      string  true  "Дата для проверки в формате YYYY-MM-DD" example:"2024-12-15"
// @Success      200      {object}  dto.IsHolidayResponse  "Результат проверки даты"
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/holidays/is-holiday [get]
func (h *Handler) IsHoliday(c echo.Context) error {
	op := "http.handler.holiday.isHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	reqDate := c.QueryParam("date")
	if reqDate == "" {
		log.Warn("date parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgDateParameterRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"date": reqDate})
	log.Info("Checking if date is holiday")

	date, err := time.Parse("2006-01-02", reqDate)
	if err != nil {
		log.Error(err, "Invalid date format: %s", reqDate)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidDateFormat,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}
	isHoliday := h.holidayService.IsHoliday(date)
	
	log.WithFields(logger.Fields{"is_holiday": isHoliday}).Info("Holiday check completed")
	
	return c.JSON(http.StatusOK, dto.IsHolidayResponse{
		IsHoliday: isHoliday,
		Date:      reqDate,
	})
}

// UpdateHoliday - обновление праздника
// @Summary      Обновление праздника
// @Description  Обновляет существующий праздник в базе данных по его ID. Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        id  path      string  true  "UUID праздника" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param        request  body      dto.HolidayRequest  true  "Данные для обновления праздника"
// @Success      200      {object}  dto.MessageResponse  "Успешное обновление праздника"
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse     "Праздник не найден"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/holidays/{id} [put]
func (h *Handler) UpdateHoliday(c echo.Context) error {
	op := "http.handler.holiday.updateHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	holidayID := c.Param("id")
	if holidayID == "" {
		log.Warn("id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgIDParameterRequired,
			Code:  http.StatusBadRequest,
		})
	}
	id, err := helpers.ParseHolidayID(holidayID)
	if err != nil {
		log.Error(err, "Invalid id format: %s", holidayID)
		errorMsg := helpers.ErrMsgInvalidIDFormat
		if errors.Is(err, helpers.ErrRequired) {
			errorMsg = helpers.ErrMsgIDParameterRequired
		}
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   errorMsg,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}
	var req dto.HolidayRequest
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidRequestBody,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"holiday_id": holidayID})
	log.Info("Updating holiday")

	// Парсим дату из формата YYYY-MM-DD
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		log.Error(err, "Invalid date format: %s", req.HolidayDate)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidDateFormatFull,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	if err := h.holidayService.UpdateHoliday(id, holidayDate, req.Name, req.Country); err != nil {
		log.Error(err, "Failed to update holiday")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToUpdateHoliday,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}
	
	log.Success("Holiday updated successfully")
	
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday updated successfully",
		Code:    http.StatusOK,
	})
}

// DeleteHoliday - удаление праздника
// @Summary      Удаление праздника
// @Description  Удаляет праздник из базы данных по его ID. Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        id  path      string  true  "UUID праздника" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.MessageResponse  "Успешное удаление праздника"
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse     "Праздник не найден"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/holidays/{id} [delete]
func (h *Handler) DeleteHoliday(c echo.Context) error {
	op := "http.handler.holiday.deleteHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	holidayID := c.Param("id")
	if holidayID == "" {
		log.Warn("id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgIDParameterRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"holiday_id": holidayID})
	log.Info("Deleting holiday")

	id, err := helpers.ParseHolidayID(holidayID)
	if err != nil {
		log.Error(err, "Invalid id format: %s", holidayID)
		errorMsg := helpers.ErrMsgInvalidIDFormat
		if errors.Is(err, helpers.ErrRequired) {
			errorMsg = helpers.ErrMsgIDParameterRequired
		}
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   errorMsg,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	if err := h.holidayService.DeleteHoliday(id); err != nil {
		log.Error(err, "Failed to delete holiday")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToDeleteHoliday,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}
	
	log.Success("Holiday deleted successfully")
	
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday deleted successfully",
		Code:    http.StatusOK,
	})
}
