package helpers

// HTTP error messages constants
const (
	// Common validation errors
	ErrMsgInvalidRequestBody = "Invalid request body"
	ErrMsgUnauthorized       = "Unauthorized"
	ErrMsgUserIDRequired     = "User ID is required"
	ErrMsgInvalidUserID     = "Invalid user ID"
	ErrMsgInvalidIDFormat   = "Invalid ID format"

	// Parameter validation errors
	ErrMsgRequestIDRequired     = "request_id parameter is required"
	ErrMsgRequestIDRequiredAlt  = "requestId is required"
	ErrMsgYearParameterRequired = "year parameter is required"
	ErrMsgDateParameterRequired = "date parameter is required"
	ErrMsgIDParameterRequired   = "id parameter is required"
	ErrMsgTypeParameterRequired = "type parameter is required"
	ErrMsgMethodParameterRequired = "method parameter is required"

	// Date format errors
	ErrMsgInvalidDateFormat      = "Invalid date format"
	ErrMsgInvalidDateFormatFull  = "Invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)"
	ErrMsgInvalidYearFormat      = "Invalid year format"

	// Validation errors
	ErrMsgTurnoverMustBeGreaterThanZero     = "turnover must be greater than 0"
	ErrMsgDesiredProfitPercentInvalid       = "desiredProfitPercent must be between 0 and 100"
	ErrMsgModelMustBeB2COrB2B               = "model must be either B2C or B2B"
	ErrMsgInitialBalanceCannotBeNegative    = "initialBalance cannot be negative"
	ErrMsgTurnoverParameterRequiredForFirstMonth = "turnover parameter is required for first month"
	ErrMsgTurnoverMustBePositiveNumber      = "turnover must be a positive number"

	// Authentication errors
	ErrMsgUserAuthenticationRequired = "User authentication required"
	ErrMsgInvalidEmailOrPassword     = "Invalid email or password"
	ErrMsgRegistrationFailed         = "Registration failed"

	// Gateway errors
	ErrMsgFailedToGetB2CGateway    = "Failed to get B2C gateway"
	ErrMsgB2CGatewayNotFound       = "B2C gateway not found"
	ErrMsgInvalidGatewayID         = "Invalid gateway ID"
	ErrMsgFailedToUpdateB2CGateway = "Failed to update B2C gateway"
	ErrMsgFailedToDeleteB2CGateways = "Failed to delete B2C gateways"
	ErrMsgInsufficientPermissions  = "Insufficient permissions. Required role: admin"

	// Generation errors
	ErrMsgFailedToGenerateStatement = "Failed to generate statement"
	ErrMsgTransactionWouldResultInNegativeBalance = "transaction would result in negative balance"
	ErrMsgInsufficientBalance = "insufficient balance"
	ErrMsgInvalidBusinessModel = "invalid business model"

	// Transaction errors
	ErrMsgFailedToCreateTransaction     = "Failed to create transaction"
	ErrMsgFailedToGetTransactions        = "Failed to get transactions"
	ErrMsgFailedToGetTransactionsCount   = "Failed to get transactions count"
	ErrMsgFailedToCreateBatchTransactions = "Failed to create batch transactions"

	// Balance errors
	ErrMsgFailedToValidateBalance      = "Failed to validate balance"
	ErrMsgFailedToGetBalanceAdjustment = "Failed to get balance adjustment"
	ErrMsgNoBalanceAdjustmentsFound    = "No balance adjustments found for the given request_id"

	// Base amount errors
	ErrMsgFailedToGetBaseAmounts        = "Failed to get base amounts"
	ErrMsgFailedToCalculateMobileAmount = "Failed to calculate mobile amount"
	ErrMsgFailedToCalculateUtilitiesAmount = "Failed to calculate utilities amount"
	ErrMsgFailedToCalculateLeasingAmount   = "Failed to calculate leasing amount"
	ErrMsgFailedToResetMobileBaseAmount   = "Failed to reset mobile base amount"
	ErrMsgFailedToResetUtilitiesBaseAmount = "Failed to reset utilities base amount"
	ErrMsgFailedToResetLeasingBaseAmount  = "Failed to reset leasing base amount"

	// Holiday errors
	ErrMsgFailedToAddHoliday    = "Failed to add holiday"
	ErrMsgFailedToGetHolidays   = "Failed to get holidays"
	ErrMsgFailedToUpdateHoliday = "Failed to update holiday"
	ErrMsgFailedToDeleteHoliday = "Failed to delete holiday"

	// User errors
	ErrMsgFailedToSaveAssociatedCard = "Failed to save associated card"

	// Breakdown errors
	ErrMsgFailedToCalculateRevenueBreakdown  = "Failed to calculate revenue breakdown"
	ErrMsgFailedToCalculateExpensesBreakdown = "Failed to calculate expenses breakdown"
	ErrMsgInvalidRequestIDFormat            = "Invalid request_id format. Expected UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"

	// Gateway detail messages
	ErrMsgB2CGatewayNotFoundDetails         = "No gateway has been saved for this user. A gateway will be automatically selected during the first B2C generation."
	ErrMsgInvalidGatewayIDDetails           = "The specified gateway ID does not exist in the available gateways list"
	ErrMsgInsufficientPermissionsDetails    = "Only administrators can access this resource"

	// Seed errors
	ErrMsgSeedServiceNotAvailable = "Seed service not available"
	ErrMsgFailedToSeedDatabase    = "Failed to seed database"
)

