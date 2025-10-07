# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование go mod файлов
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Финальный образ
FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata

# Создание пользователя для безопасности
RUN adduser -D -s /bin/sh appuser

# Установка рабочей директории
WORKDIR /root/

# Копирование бинарного файла из builder stage
COPY --from=builder /app/main .

# # Копирование конфигурационных файлов (если есть)
# COPY --from=builder /app/.env.example .env.example

# Изменение владельца файлов
RUN chown -R appuser:appuser /root/

# Переключение на непривилегированного пользователя
USER appuser

# Открытие портов
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запуск приложения
CMD ["./main"]
