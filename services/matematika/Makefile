migrate-new:
	migrate create -ext sql -dir ./migrations ${NAME}

migrate-up:
	migrate -path ./migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" up 

migrate-down:
	migrate -path ./migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" down 
