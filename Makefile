lint:
	cd services/matematika && golangci-lint run --color=always
	npm run lint --prefix ./services/maska
	cd services/shared && golangci-lint run --color=always

run:
	docker compose up --build -d

stop:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose down
	docker compose up --build -d

migrate-new-matematika:
	migrate create -ext sql -dir ./services/matematika/migrations ${NAME}

# migrate-new-maska:
# 	migrate create -ext sql -dir ./services/maska/migrations ${NAME}

migrate-new-shared:
	migrate create -ext sql -dir ./services/shared/migrations ${NAME}

migrate-up-matematika:
	migrate -path ./services/matematika/migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" up 

# migrate-up-maska:
# 	migrate -path ./services/maska/migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" up 

migrate-up-shared:
	migrate -path ./services/shared/migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" up 

migrate-down-matematika:
	migrate -path ./services/matematika/migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" down 

# migrate-down-maska:
# 	migrate -path ./services/maska/migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" down 

migrate-down-shared:
	migrate -path ./services/shared/migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" down 

test-work-kafka:
	./test-kafka.sh

test-work-nginx:
	./test-nginx.sh

# Seeding
seed-matematika:
	@echo "🌱 Seeding database through Docker network..."
	docker run --rm --network business_bank_back_business_bank_network \
		-e POSTGRES_HOST=postgres \
		-e POSTGRES_PORT=5432 \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=matematika \
		-v $(PWD)/services/matematika:/app \
		-w /app \
		golang:1.24-alpine \
		sh -c "go run cmd/seed/main.go"

# Swagger documentation
swagger-matematika:
	@echo "📚 Generating Swagger documentation for matematika..."
	cd services/matematika && swag init -g cmd/server/main.go -o docs
	@echo "✓ Swagger documentation generated in services/matematika/docs/"

# tests

# help
help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  run - Run the project"
	@echo "  stop - Stop the project"
	@echo "  logs - Show logs"
	@echo "  restart - Restart the project"
	@echo "  migrate-new-matematika - Create a new migration for matematika"
	@echo "  migrate-new-maska - Create a new migration for maska"
	@echo "  migrate-new-shared - Create a new migration for shared"
	@echo "  migrate-up-matematika - Apply all migrations for matematika"
	@echo "  seed-matematika - Seed database with mock data"
	@echo "  swagger-matematika - Generate Swagger documentation for matematika"
	@echo "  test-work-kafka - Test work kafka"
	@echo "  test-work-nginx - Test work nginx"