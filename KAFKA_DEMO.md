# 🚀 KAFKA INTEGRATION DEMO

## 📋 Что это?

Полный пример работы с Kafka в микросервисе Matematika:
- **Producer** - отправка сообщений в Kafka
- **Consumer** - чтение сообщений из Kafka
- **Handler → Service → Kafka** - полная цепочка

---

## 🔧 Как запустить?

### 1. Запустить все сервисы

```bash
docker compose up --build
```

Будут запущены:
- ✅ Matematika (port 8080, 9090)
- ✅ PostgreSQL (port 5432)
- ✅ Kafka (port 9092, 9093)
- ✅ Zookeeper (port 2181)
- ✅ Kafdrop UI (port 9000)
- ✅ pgAdmin (port 8085)

### 2. Подождать пока все запустится (~30 секунд)

Проверить готовность:
```bash
# Проверка Matematika
curl http://localhost:8080/health

# Должен вернуть: {"service":"matematika","status":"healthy"}
```

### 3. Отправить тестовый запрос

#### Вариант 1: Через скрипт
```bash
./test-kafka.sh
```

#### Вариант 2: Через curl
```bash
curl -X POST http://localhost:8080/generate-statement \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "ACC_12345",
    "month": "2025-01",
    "businessType": "B2C",
    "initialBalance": 10000.50
  }'
```

#### Вариант 3: Через Postman/Insomnia
```
POST http://localhost:8080/generate-statement
Content-Type: application/json

{
  "accountId": "ACC_12345",
  "month": "2025-01",
  "businessType": "B2C",
  "initialBalance": 10000.50
}
```

---

## 📊 Что произойдет?

### 1. **HTTP Request** → Handler
```
POST /generate-statement
↓
CalculationHandler.GenerateStatement()
```

### 2. **Handler** → Service
```
Handler вызывает Service
↓
CalculationService.GenerateStatement()
```

### 3. **Service** выполняет:
```
1. ✅ Валидация данных
2. ✅ Создание Statement ID
3. ✅ Симуляция расчетов
4. ✅ Отправка в Kafka (Producer)
5. ✅ Возврат ответа клиенту
```

### 4. **Kafka Consumer** получает сообщение:
```
1. ✅ Читает из топика statement.calculation.completed
2. ✅ Десериализует JSON
3. ✅ Выводит в консоль (логи Docker)
```

---

## 🔍 Как проверить?

### 1. Логи Docker (САМЫЙ ВАЖНЫЙ!)
```bash
docker compose logs -f matematika
```

Увидишь:
```
========================================
📥 ПОЛУЧЕН ЗАПРОС на генерацию выписки
   AccountID: ACC_12345
   Month: 2025-01
   BusinessType: B2C
   InitialBalance: 10000.50
   StatementID: stmt_2025-01_ACC_12345
========================================
⚙️  Выполняем расчеты...
✓ Расчеты завершены
📤 Отправляем результаты в Kafka...
✓ Сообщение успешно отправлено в Kafka!
   Topic: statement.calculation.completed
   StatementID: stmt_2025-01_ACC_12345
========================================
📨 ПОЛУЧЕНО СООБЩЕНИЕ ИЗ KAFKA
   Topic: statement.calculation.completed
   Partition: 0
   Offset: 0
   ...
📊 РАСПАРСЕННЫЕ ДАННЫЕ:
   StatementID: stmt_2025-01_ACC_12345
   AccountID: ACC_12345
   Month: 2025-01
   Status: completed
========================================
```

### 2. Kafdrop UI
```
http://localhost:9000
```

Увидишь:
- Топик: `statement.calculation.completed`
- Количество сообщений
- Содержимое каждого сообщения

### 3. API Response
```json
{
  "statementId": "stmt_2025-01_ACC_12345",
  "status": "processing",
  "message": "Statement generation started and sent to Kafka"
}
```

---

## 📁 Структура кода

### Handler (HTTP Layer)
```
/services/matematika/internal/calculation/handler.go
- Принимает HTTP запросы
- Валидирует JSON
- Вызывает Service
```

### Service (Business Logic)
```
/services/matematika/internal/calculation/service.go
- GenerateStatement() - отправляет в Kafka (Producer)
- StartConsumer() - читает из Kafka (Consumer)
```

### Kafka Layer
```
/services/matematika/internal/kafka/
- producer.go - отправка сообщений
- consumer.go - чтение сообщений
- messages.go - структуры и топики
- config.go - конфигурация
```

### Main
```
/services/matematika/cmd/server/main.go
- Инициализация Kafka Producer
- Инициализация Services
- Запуск Kafka Consumer
- Запуск HTTP сервера
```

---

## 🎯 Топики Kafka

### 1. statement.calculation.completed
- **Producer:** Matematika Service
- **Consumer:** Matematika Service (для демо)
- **Формат:** CalculationCompletedMessage
- **Назначение:** Результаты расчетов

---

## 🔄 Полный Workflow

```
1. Внешний клиент
   ↓ (HTTP POST)
2. Handler.GenerateStatement()
   ↓ (вызов)
3. Service.GenerateStatement()
   ↓ (расчеты)
4. Kafka Producer
   ↓ (публикация)
5. Kafka Topic: statement.calculation.completed
   ↓ (подписка)
6. Kafka Consumer
   ↓ (обработка)
7. Console Output (логи)
```

---

## 🛠️ Troubleshooting

### Kafka не подключается
```bash
# Проверь что Kafka запущена
docker compose ps

# Перезапусти
docker compose restart kafka
```

### Consumer не получает сообщения
```bash
# Проверь логи
docker compose logs -f matematika

# Проверь топики в Kafdrop
open http://localhost:9000
```

### Порт занят
```bash
# Освободи порты
docker compose down
lsof -ti:8080 | xargs kill -9
```

---

## 📚 Что дальше?

1. **Изучи код** в `service.go` - там все методы с комментариями
2. **Поэкспериментируй** - измени JSON, посмотри что в Kafka
3. **Добавь логику** - в `GenerateStatement()` напиши реальные расчеты
4. **Масштабируй** - запусти несколько consumer'ов для параллелизма

---

## 🎓 Ключевые паттерны

1. **Dependency Injection** - Kafka передается через конструктор
2. **Interface-based** - `Producer` и `Consumer` это интерфейсы
3. **Context propagation** - `ctx` передается во все методы
4. **Graceful shutdown** - корректное закрытие всех соединений
5. **Structured logging** - четкие логи для отладки
6. **Error handling** - все ошибки обрабатываются

**Удачи! 🚀**

