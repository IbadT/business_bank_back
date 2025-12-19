package helpers

import "errors"

var (
	// Balance errors
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrNegativeBalance     = errors.New("transaction would result in negative balance")
	ErrNegativeBalanceStillExists = errors.New("negative balance still exists after all adjustments")

	// Date and time errors
	ErrFutureMonth     = errors.New("cannot generate statement for future month")
	ErrFailedToParseDate = errors.New("failed to parse date")
	ErrInvalidDate     = errors.New("invalid date")
	ErrInvalidTime     = errors.New("invalid time")
	ErrInvalidDateTime = errors.New("invalid date and time")
	ErrNoAvailableDate = errors.New("no available date found within month")

	// Validation errors
	ErrPositiveNumber = errors.New("must be a positive number")
	ErrRequired       = errors.New("is required")
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidFormat  = errors.New("invalid format")
	ErrInvalidLength  = errors.New("invalid length")
	ErrInvalidValue   = errors.New("invalid value")
	ErrInvalidEnum    = errors.New("invalid enum value")
	ErrInvalidRequestBody = errors.New("invalid request body")

	// Request ID errors
	ErrInvalidRequestID = errors.New("invalid requestId format")
	ErrRequestIDEmpty   = errors.New("requestID cannot be empty")
	ErrRequestIDNotFound = errors.New("requestID not found in database")

	// Transaction errors
	ErrInvalidTransactionType = errors.New("invalid transaction type. Must be 'income' or 'expense'")
	ErrInvalidAmount         = errors.New("amount must be greater than 0")
	ErrInvalidTransactionAmount = errors.New("invalid transaction amount")
	ErrEmptyCategory         = errors.New("category is required")
	ErrInvalidCategory       = errors.New("invalid transaction category")
	ErrEmptyMethod           = errors.New("method is required")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrTransactionsArrayEmpty = errors.New("transactions array cannot be empty")
	ErrAllTransactionsSameRequestID = errors.New("all transactions must have the same requestId, but found different")

	// Date format errors
	ErrInvalidDateFormat = errors.New("invalid date format. Expected ISO8601 (YYYY-MM-DDTHH:MM:SSZ) or YYYY-MM-DD")
	ErrInvalidTransactionDateFormat = errors.New("invalid transaction date format")
	ErrInvalidPostingDateFormat     = errors.New("invalid posting date format")

	// User errors
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserIDRequired     = errors.New("userID is required")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserNotFoundOrNoChanges = errors.New("user not found or no changes made")
	ErrInvalidRole        = errors.New("invalid role")

	// Email and password validation errors
	ErrEmailRequired           = errors.New("email is required")
	ErrPasswordRequired        = errors.New("password is required")
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters long")
	ErrInvalidEmail            = errors.New("email field must be a valid email address")

	// Associated card errors
	ErrAssociatedCardRequired      = errors.New("associatedCard is required")
	ErrAssociatedCardInvalidLength = errors.New("associatedCard must be 16 digits")
	ErrAssociatedCardInvalidFormat = errors.New("associatedCard must contain only digits")

	// Gateway errors
	ErrGatewayNotFound = errors.New("gateway not found")
	ErrGatewayIDRequired = errors.New("id is required")
	ErrGatewayNameRequired = errors.New("name is required")

	// Holiday errors
	ErrHolidayAlreadyExists = errors.New("holiday already exists")
	ErrInvalidCountry       = errors.New("invalid country")

	// Generator errors
	ErrInvalidModel   = errors.New("invalid business model")
	ErrInvalidRequestParams = errors.New("invalid request parameters")
	ErrUnauthorized   = errors.New("user authentication required")
	ErrFailedToCreateGenerationRequest = errors.New("failed to create generation request")
	ErrFailedToAdjustTransactions = errors.New("failed to adjust transactions")
	ErrFailedToRecalculateBalances = errors.New("failed to recalculate balances")
	ErrCannotPostponeTransaction = errors.New("cannot postpone transaction")
	ErrCannotReduceTransactionAmount = errors.New("cannot reduce transaction amount: available balance too low")

	// Validator errors
	ErrTurnoverMustBeGreaterThanZero = errors.New("turnover must be greater than 0")
	ErrDesiredProfitPercentInvalid   = errors.New("desired profit percent must be between 0 and 100")
	ErrModelMustBeB2COrB2B           = errors.New("model must be either B2C or B2B")
	ErrInitialBalanceCannotBeNegative = errors.New("initial balance cannot be negative")
	ErrB2CIncomeTransactionsCountInvalid = errors.New("B2C income transactions count must be 4-5")
	ErrB2BIncomeTransactionsCountInvalid = errors.New("B2B income transactions count must be 10-20")
	ErrExpenseTransactionsCountInvalid   = errors.New("expense transactions count must be 35-55 (requirement: ~45 ± 10)")
	ErrTotalTransactionsCountInvalid     = errors.New("total transactions count must be 39-75")

	// Base amount errors
	ErrInvalidUserIDFormat           = errors.New("invalid userID format")
	ErrFailedToSaveMobileBaseAmount    = errors.New("failed to save mobile base amount")
	ErrFailedToGetMobileBaseAmount     = errors.New("failed to get mobile base amount")
	ErrMobileBaseAmountNotFound        = errors.New("mobile base amount not found. Generate first month first")
	ErrFailedToSaveUtilitiesBaseAmount = errors.New("failed to save utilities base amount")
	ErrFailedToGetUtilitiesBaseAmount  = errors.New("failed to get utilities base amount")
	ErrUtilitiesBaseAmountNotFound     = errors.New("utilities base amount not found. Generate first month first")
	ErrFailedToSaveLeasingBaseAmount  = errors.New("failed to save leasing base amount")
	ErrFailedToGetLeasingBaseAmount    = errors.New("failed to get leasing base amount")
	ErrLeasingBaseAmountNotFound       = errors.New("leasing base amount not found. Generate first month first")
	ErrTurnoverMustBeGreaterThanZeroForLeasing = errors.New("turnover must be greater than 0 for first month leasing calculation")

	// Balance adjustment errors
	ErrGenerationRequestNotFound = errors.New("generation request not found")
	ErrNoTransactionsFound       = errors.New("no transactions found for request_id")
	ErrFailedToGetAdjustedTransactions = errors.New("failed to get adjusted transactions")
	ErrFailedToGetIncomeTransactions   = errors.New("failed to get income transactions")
	ErrFailedToGetExpenseTransactions  = errors.New("failed to get expense transactions")

	// JWT errors
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidTokenClaims       = errors.New("invalid token claims")
	ErrUserIDNotFoundInClaims   = errors.New("user ID not found in claims")
	ErrUserIDEmptyInClaims      = errors.New("user ID is empty in claims")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrFailedToExtractClaims    = errors.New("failed to extract claims")
	ErrFailedToGetUserData      = errors.New("failed to get user data")
	ErrFailedToGenerateTokens   = errors.New("failed to generate tokens")

	// Redis errors
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Cache errors
	ErrCacheHolidaysTTLZero = errors.New("cache holidays ttl is 0")
	ErrCacheGatewaysTTLZero = errors.New("cache gateways ttl is 0")

	// Not found errors
	ErrNotFound          = errors.New("not found")
	ErrStatementNotFound = errors.New("statement not found")

	// Operation errors
	ErrFailedToKafkaStatement    = errors.New("failed to generate statement to Kafka")
	ErrFailedToCheckHealth       = errors.New("failed to check health")
	ErrFailedToGenerateStatement = errors.New("failed to generate statement")
	ErrFailedToLoadConfiguration = errors.New("failed to get admin config")
	ErrFailedToGetTransactions   = errors.New("failed to get transactions")
	ErrFailedToGetBusinessRules  = errors.New("failed to get business rules")
	ErrFailedToGetDailyBalances  = errors.New("failed to get daily balances")
	ErrFailedToGetStatements     = errors.New("failed to get statements")

	// Kafka errors
	ErrNoHandlerFoundForTopic = errors.New("no handler found for topic")
	ErrFailedToProcessMessage = errors.New("failed to process message")
	ErrFailedToCreateConsumerGroup = errors.New("failed to create consumer group")
	ErrFailedToUnmarshalMessage = errors.New("failed to unmarshal message")
	ErrFailedToCreateKafkaProducer = errors.New("failed to create kafka producer")
	ErrFailedToMarshalMessage = errors.New("failed to marshal message")
	ErrFailedToPublishMessage = errors.New("failed to publish message")

	// Database errors
	ErrFailedToGetSQLDB = errors.New("failed to get sql.DB")
	ErrFailedToPingDatabase = errors.New("failed to ping database")

	// GRPC client errors
	ErrFailedToConnect = errors.New("failed to connect")

	// Seeds errors
	ErrFailedToHashPassword = errors.New("failed to hash password")
	ErrFailedToGetGenerationRequests = errors.New("failed to get generation requests")

	// Internal errors
	ErrInternalServerError = errors.New("internal server error")
)
