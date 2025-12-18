# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование pkg/proto для replace директивы
COPY pkg/proto/go.mod ./pkg/proto/

# Копирование go mod файлов сервиса
COPY services/matematika/go.mod services/matematika/go.sum ./services/matematika/

# Загрузка зависимостей
WORKDIR /app/services/matematika
RUN go mod download

# Копирование исходного кода
WORKDIR /app
COPY pkg/ ./pkg/
COPY services/matematika/ ./services/matematika/

# Генерация Swagger документации
WORKDIR /app/services/matematika
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/server/main.go -o docs

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/server

# Копирование конфигурационных файлов
COPY services/matematika/config/ /app/config/

# Финальный образ
FROM alpine:latest

# Установка необходимых пакетов (включая netcat-openbsd для healthcheck)
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# Создание пользователя для безопасности
RUN adduser -D -s /bin/sh appuser

# Установка рабочей директории
WORKDIR /root/

# Копирование бинарного файла из builder stage
COPY --from=builder /app/main .

# Копирование конфигурационных файлов
COPY --from=builder /app/config ./config

# Копирование Swagger документации (включая docs.go для echo-swagger)
COPY --from=builder /app/services/matematika/docs ./docs

# Изменение владельца файлов
RUN chown -R appuser:appuser /root/

# Переключение на непривилегированного пользователя
USER appuser

# Открытие портов
EXPOSE 8080 9090

# Health check - используем простую проверку через /dev/tcp
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD sh -c "echo -e 'GET /health HTTP/1.1\r\nHost: localhost\r\n\r\n' | nc 127.0.0.1 8080 | grep -q '200 OK' || exit 1"

# Запуск приложения
CMD ["./main"]
