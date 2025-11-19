package helpers

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient balance")

	ErrFutureMonth = errors.New("cannot generate statement for future month")

	ErrPositiveNumber = errors.New("must be a positive number")

	ErrRequired = errors.New("is required")

	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidFormat      = errors.New("invalid format")
	ErrInvalidLength      = errors.New("invalid length")
	ErrInvalidValue       = errors.New("invalid value")
	ErrInvalidEnum        = errors.New("invalid enum value")
	ErrInvalidDate        = errors.New("invalid date")
	ErrInvalidTime        = errors.New("invalid time")
	ErrInvalidRequestBody = errors.New("invalid request body")
	ErrInvalidDateTime    = errors.New("invalid date and time")

	ErrNotFound          = errors.New("not found")
	ErrStatementNotFound = errors.New("statement not found")

	ErrFailedToKafkaStatement    = errors.New("failed to generate statement to Kafka")
	ErrFailedToCheckHealth       = errors.New("failed to check health")
	ErrFailedToGenerateStatement = errors.New("failed to generate statement")
	ErrFailedToLoadConfiguration = errors.New("failed to get admin config")
	ErrFailedToGetTransactions   = errors.New("failed to get transactions")
	ErrFailedToGetBusinessRules  = errors.New("failed to get business rules")
	ErrFailedToGetDailyBalances  = errors.New("failed to get daily balances")
	ErrFailedToGetStatements     = errors.New("failed to get statements")

	ErrFailedToParseDate = errors.New("failed to parse date")

	ErrInternalServerError = errors.New("internal server error")
)
