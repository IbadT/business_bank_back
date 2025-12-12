// @title           Matematika API
// @version         1.0
// @description     API для генерации финансовых выписок и транзакций. Сервис позволяет генерировать реалистичные финансовые выписки на основе заданных параметров (оборот, желаемая прибыль, модель бизнеса).
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@matematika.com
// @contact.url    https://github.com/IbadT/business_bank_back

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @schemes  http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите только токен (без Bearer). Получите токен через /api/login или /api/register. Swagger автоматически добавит префикс "Bearer "

// @tag.name auth
// @tag.description API для аутентификации и регистрации пользователей

// @tag.name holidays
// @tag.description API для управления праздниками

// @tag.name transactions
// @tag.description API для управления транзакциями

package main

import (
	"log"

	_ "github.com/IbadT/business_bank_back/services/matematika/docs" // swagger docs
	"github.com/IbadT/business_bank_back/services/matematika/internal/app"
	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
)

func main() {
	// Загружаем конфигурацию из переменных окружения
	cfg := &app.Config{
		Port:     database.GetEnv("PORT", "8080"),
		GRPCPort: database.GetEnv("GRPC_PORT", "9090"),
	}

	// Создаем и запускаем приложение
	application := app.NewApp(cfg)
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
