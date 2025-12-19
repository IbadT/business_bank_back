// internal/repository/config_repository.go
package repository

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
)

type ConfigRepository interface {
	GetHolidays() ([]*domain.Holiday, error)
	GetTransactionTemplates() ([]*entities.TransactionTemplate, error)
	GetGateways() ([]*entities.Gateway, error)
	GetCustomers() ([]*entities.Customer, error)
	GetDefaultCustomers() ([]dto.DefaultCustomer, error) // Для обратной совместимости
}

type fileConfigRepository struct {
	configPath string
}

func NewConfigRepository(configPath string) ConfigRepository {
	return &fileConfigRepository{
		configPath: configPath,
	}
}

func (r *fileConfigRepository) GetHolidays() ([]*domain.Holiday, error) {
	op := "repository.config.getHolidays"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting holidays from config")

	filePath := filepath.Join(r.configPath, "holidays.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err, "Failed to open holidays file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	var holidayModels []dto.Holiday
	if err := json.NewDecoder(file).Decode(&holidayModels); err != nil {
		log.Error(err, "Failed to decode holidays JSON")
		return nil, err
	}

	// Конвертируем dto.Holiday в entities.Holiday
	holidays := make([]*domain.Holiday, len(holidayModels))
	for i, h := range holidayModels {
		date, err := time.Parse("2006-01-02", h.Date)
		if err != nil {
			log.Error(err, "Invalid date format in holidays: %s", h.Date)
			return nil, err
		}
		holiday, err := domain.NewHoliday(date, h.Name, h.Country)
		if err != nil {
			log.Error(err, "Failed to create holiday domain object")
			return nil, err
		}
		holidays[i] = holiday
	}
	
	log.WithFields(logger.Fields{"count": len(holidays)}).Success("Holidays retrieved from config")
	return holidays, nil
}

func (r *fileConfigRepository) GetTransactionTemplates() ([]*entities.TransactionTemplate, error) {
	op := "repository.config.getTransactionTemplates"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting transaction templates from config")

	filePath := filepath.Join(r.configPath, "templates.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err, "Failed to open templates file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	var templateModels []dto.TransactionTemplate
	if err := json.NewDecoder(file).Decode(&templateModels); err != nil {
		log.Error(err, "Failed to decode templates JSON")
		return nil, err
	}

	// TODO: убрать это из репозиторного слоя
	// Конвертируем dto.TransactionTemplate в entities.TransactionTemplate
	templates := make([]*entities.TransactionTemplate, len(templateModels))
	for i, tm := range templateModels {
		transactionType, _ := value_objects.NewTransactionType(tm.Type)
		paymentMethod, _ := value_objects.NewPaymentMethod(tm.Method)

		schedule := entities.Schedule{
			Frequency:      tm.Frequency,
			PreferredDay:   tm.PreferredDay,
			WeekOfMonth:    tm.WeekOfMonth,
			MinOccurrences: tm.MinTransactions,
			MaxOccurrences: tm.MaxTransactions,
		}

		businessHours := entities.BusinessHours{
			Start: tm.BusinessHours.Start,
			End:   tm.BusinessHours.End,
		}

		template := &entities.TransactionTemplate{
			ID:            fmt.Sprintf("tm_%d", i+1),
			Category:      tm.Category,
			Type:          transactionType,
			IsPercentage:  tm.IsPercentage,
			FixedAmount:   tm.FixedAmount,
			Schedule:      schedule,
			BusinessHours: businessHours,
			IsOptional:    tm.IsOptional,
			PaymentMethod: paymentMethod,
			TransactionRange: entities.TransactionRange{
				Min: tm.MinTransactions,
				Max: tm.MaxTransactions,
			},
		}

		if tm.IsPercentage {
			template.PercentageRange = entities.PercentageRange{
				Min: tm.PercentageMin,
				Max: tm.PercentageMax,
			}
		}

		templates[i] = template
	}
	
	log.WithFields(logger.Fields{"count": len(templates)}).Success("Transaction templates retrieved from config")
	return templates, nil
}

func (r *fileConfigRepository) GetGateways() ([]*entities.Gateway, error) {
	op := "repository.config.getGateways"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting gateways from config")

	filePath := filepath.Join(r.configPath, "gateways.csv")
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err, "Failed to open gateways file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Error(err, "Failed to read gateways CSV")
		return nil, err
	}

	var gateways []*entities.Gateway
	for i, record := range records[1:] { // Пропускаем заголовок
		if len(record) > 0 {
			gateways = append(gateways, entities.NewGateway(
				fmt.Sprintf("gw_%d", i+1),
				record[0],
			))
		}
	}
	
	log.WithFields(logger.Fields{"count": len(gateways)}).Success("Gateways retrieved from config")
	return gateways, nil
}

func (r *fileConfigRepository) GetCustomers() ([]*entities.Customer, error) {
	op := "repository.config.getCustomers"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting customers from config")

	filePath := filepath.Join(r.configPath, "customers.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err, "Failed to open customers file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	var customerModels []dto.DefaultCustomer
	if err := json.NewDecoder(file).Decode(&customerModels); err != nil {
		log.Error(err, "Failed to decode customers JSON")
		return nil, err
	}

	// Конвертируем dto.DefaultCustomer в entities.Customer
	customers := make([]*entities.Customer, len(customerModels))
	for i, cm := range customerModels {
		customers[i] = entities.NewCustomer(
			fmt.Sprintf("cust_%d", i+1),
			cm.Name,
			cm.Category,
			cm.MinPercent,
			cm.MaxPercent,
			cm.MinTransactions,
			cm.MaxTransactions,
		)
	}
	
	log.WithFields(logger.Fields{"count": len(customers)}).Success("Customers retrieved from config")
	return customers, nil
}

func (r *fileConfigRepository) GetDefaultCustomers() ([]dto.DefaultCustomer, error) {
	op := "repository.config.getDefaultCustomers"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting default customers from config")

	filePath := filepath.Join(r.configPath, "customers.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err, "Failed to open customers file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	var customers []dto.DefaultCustomer
	if err := json.NewDecoder(file).Decode(&customers); err != nil {
		log.Error(err, "Failed to decode customers JSON")
		return nil, err
	}
	
	log.WithFields(logger.Fields{"count": len(customers)}).Success("Default customers retrieved from config")
	return customers, nil
}
