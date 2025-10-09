#!/bin/bash

# Тестовый скрипт для проверки Nginx API Gateway

echo "=========================================="
echo "🔧 ТЕСТИРОВАНИЕ NGINX API GATEWAY"
echo "=========================================="
echo ""

# URL через Nginx
NGINX_URL="http://localhost"

echo "📍 Gateway URL: $NGINX_URL"
echo ""
echo "=========================================="
echo ""

# 1. Проверка Nginx health
echo "1️⃣ Проверка Nginx Health Check:"
echo "GET $NGINX_URL/nginx-health"
echo ""
curl -s "$NGINX_URL/nginx-health"
echo ""
echo ""

# 2. Проверка root endpoint
echo "2️⃣ Проверка Root Endpoint (API Info):"
echo "GET $NGINX_URL/"
echo ""
curl -s "$NGINX_URL/" | jq '.'
echo ""
echo ""

# 3. Проверка Matematika health через Nginx
echo "3️⃣ Проверка Matematika Health через Nginx:"
echo "GET $NGINX_URL/api/matematika/health"
echo ""
curl -s "$NGINX_URL/api/matematika/health" | jq '.'
echo ""
echo ""

# 4. Генерация выписки через Nginx
echo "4️⃣ Генерация выписки через Nginx API Gateway:"
echo "POST $NGINX_URL/api/matematika/generate-statement"
echo ""

JSON_DATA='{
  "accountId": "ACC_NGINX_TEST",
  "month": "2025-04",
  "businessType": "B2B",
  "initialBalance": 35000
}'

echo "📋 Данные запроса:"
echo "$JSON_DATA" | jq '.'
echo ""

RESPONSE=$(curl -s -X POST "$NGINX_URL/api/matematika/generate-statement" \
  -H "Content-Type: application/json" \
  -d "$JSON_DATA")

echo "📥 Ответ от сервера:"
echo "$RESPONSE" | jq '.'
echo ""
echo ""

# 5. Проверка несуществующего endpoint
echo "5️⃣ Проверка несуществующего endpoint (404):"
echo "GET $NGINX_URL/api/unknown"
echo ""
curl -s "$NGINX_URL/api/unknown" | jq '.'
echo ""
echo ""

# 6. Информация о доступных endpoints
echo "6️⃣ Проверка API Docs:"
echo "GET $NGINX_URL/api/docs"
echo ""
curl -s "$NGINX_URL/api/docs" | jq '.'
echo ""
echo ""

echo "=========================================="
echo "✅ ТЕСТИРОВАНИЕ ЗАВЕРШЕНО"
echo "=========================================="
echo ""
echo "📊 Доступные endpoints через Nginx:"
echo "   • GET  http://localhost/                           - API Info"
echo "   • GET  http://localhost/nginx-health               - Nginx Health"
echo "   • GET  http://localhost/api/matematika/health      - Matematika Health"
echo "   • POST http://localhost/api/matematika/generate-statement"
echo "   • GET  http://localhost/api/matematika/statement/:id/status"
echo "   • GET  http://localhost/api/matematika/statement/:id/result"
echo "   • GET  http://localhost/kafdrop/                   - Kafka UI"
echo "   • GET  http://localhost/pgadmin/                   - Database UI"
echo ""
echo "🔧 Nginx Logs:"
echo "   docker logs nginx-gateway"
echo "=========================================="

