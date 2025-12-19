// internal/domain/generation_request.go
package domain

import (
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
)

// GenerationRequestFactory создает models.GenerationRequest с валидацией и установкой значений по умолчанию
type GenerationRequestFactory struct{}

// NewGenerationRequestFactory создает новую фабрику для GenerationRequest
func NewGenerationRequestFactory() *GenerationRequestFactory {
	return &GenerationRequestFactory{}
}

// Create создает новый GenerationRequest из параметров
func (f *GenerationRequestFactory) Create(
	userID *uuid.UUID,
	month string,
	year int,
	turnover float64,
	desiredProfitPercent float64,
	model string,
	initialBalance float64,
	scaleFactor int,
	customData models.JSONB,
) *models.GenerationRequest {
	// Устанавливаем значения по умолчанию
	if scaleFactor <= 0 {
		scaleFactor = 1
	}
	if customData == nil {
		customData = make(models.JSONB)
	}

	return &models.GenerationRequest{
		UserID:               userID,
		Month:                month,
		Year:                 year,
		Turnover:             turnover,
		DesiredProfitPercent: desiredProfitPercent,
		Model:                model,
		InitialBalance:       initialBalance,
		ScaleFactor:          scaleFactor,
		CustomData:           customData,
		Status:               "processing",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// CreateFromCustomData создает GenerationRequest с преобразованием CustomData из DTO
func (f *GenerationRequestFactory) CreateFromCustomData(
	userID *uuid.UUID,
	month string,
	year int,
	turnover float64,
	desiredProfitPercent float64,
	model string,
	initialBalance float64,
	scaleFactor int,
	customDataRaw interface{},
) *models.GenerationRequest {
	customData := f.convertCustomData(customDataRaw)
	return f.Create(userID, month, year, turnover, desiredProfitPercent, model, initialBalance, scaleFactor, customData)
}

// convertCustomData конвертирует CustomData из различных форматов в models.JSONB
func (f *GenerationRequestFactory) convertCustomData(customDataRaw interface{}) models.JSONB {
	if customDataRaw == nil {
		return make(models.JSONB)
	}

	// Если уже JSONB, возвращаем как есть
	if jsonb, ok := customDataRaw.(models.JSONB); ok {
		return jsonb
	}

	// Если map[string]interface{}, конвертируем
	if m, ok := customDataRaw.(map[string]interface{}); ok {
		jsonb := make(models.JSONB)
		for k, v := range m {
			jsonb[k] = v
		}
		return jsonb
	}

	// По умолчанию пустой JSONB
	return make(models.JSONB)
}
