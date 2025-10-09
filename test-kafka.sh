#!/bin/bash

# Тестовый скрипт для проверки Kafka интеграции
# Отправляет запрос на генерацию выписки

echo "=========================================="
echo "🚀 ТЕСТИРОВАНИЕ KAFKA ИНТЕГРАЦИИ"
echo "=========================================="
echo ""

# Ждем пока сервис запустится
echo "⏳ Ожидание запуска сервиса..."
sleep 3

# URL сервиса
URL="http://localhost:8080/generate-statement"

# JSON данные для запроса
JSON_DATA='{
  "accountId": "ACC_12345",
  "month": "2025-01",
  "businessType": "B2C",
  "initialBalance": 10000.50
}'

echo "📤 Отправка запроса на: $URL"
echo ""
echo "📋 Данные запроса:"
echo "$JSON_DATA" | jq '.'
echo ""
echo "=========================================="
echo ""

# Отправляем запрос
RESPONSE=$(curl -s -X POST "$URL" \
  -H "Content-Type: application/json" \
  -d "$JSON_DATA")

# Выводим ответ
echo "📥 Ответ от сервера:"
echo "$RESPONSE" | jq '.'
echo ""
echo "=========================================="
echo "✅ Запрос отправлен!"
echo ""
echo "Теперь проверь логи Docker:"
echo "  docker compose logs -f matematika"
echo ""
echo "Или открой Kafdrop:"
echo "  http://localhost:9000"
echo "=========================================="

