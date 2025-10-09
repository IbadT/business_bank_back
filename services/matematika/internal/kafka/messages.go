package kafka

// ============================================================================
// KAFKA ТОПИКИ
// ============================================================================
// Топики используются для асинхронной коммуникации между микросервисами
// Соглашение об именовании: <entity>.<action>.<status>

// Kafka топики для системы генерации банковских выписок
const (
	// TopicStatementGenerationRequest - топик для запросов на генерацию выписки
	// Producer: API Gateway или внешний клиент
	// Consumer: Matematika Service
	// Формат сообщения: StatementMessage
	TopicStatementGenerationRequest = "statement.generation.request"

	// TopicCalculationCompleted - топик для результатов расчетов Matematika
	// Producer: Matematika Service (после завершения расчетов)
	// Consumer: Maska Service (для форматирования)
	// Формат сообщения: CalculationCompletedMessage
	TopicCalculationCompleted = "statement.calculation.completed"

	// TopicFormattingCompleted - топик для готовых выписок Maska
	// Producer: Maska Service (после генерации PDF/HTML)
	// Consumer: Notification Service или API для уведомления клиента
	// Формат сообщения: FormattingCompletedMessage
	TopicFormattingCompleted = "statement.formatting.completed"

	// TopicStatementError - топик для ошибок обработки
	// Producer: Любой сервис при критической ошибке
	// Consumer: Monitoring Service для алертов
	// Формат сообщения: ErrorMessage
	TopicStatementError = "statement.error"
)

// ============================================================================
// CONSUMER GROUPS
// ============================================================================
// Consumer Group ID используется для распределенной обработки
// Все consumers с одинаковым GroupID формируют группу и делят партиции

// Consumer Group IDs для каждого микросервиса
const (
	// ConsumerGroupMatematikaService - группа для Matematika Service
	// Подписывается на: TopicStatementGenerationRequest
	ConsumerGroupMatematikaService = "matematika-service-group"

	// ConsumerGroupMaskaService - группа для Maska Service
	// Подписывается на: TopicCalculationCompleted
	ConsumerGroupMaskaService = "maska-service-group"
)

// ============================================================================
// MESSAGE STATUSES
// ============================================================================
// Стандартные статусы для сообщений и записей в БД

// Message статусы используются в CalculationCompletedMessage и БД
const (
	StatusPending    = "pending"    // Ожидает обработки
	StatusProcessing = "processing" // В процессе обработки
	StatusCompleted  = "completed"  // Успешно завершено
	StatusFailed     = "failed"     // Обработка не удалась
)
