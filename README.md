# Система генерации банковских выписок

## Архитектурные требования

### Архитектурные принципы

#### 1. Строгий SRP и высокая связность внутри сервисов
Каждый микросервис выполняет одну четко определенную функцию:
- **Matematika** - только расчеты финансовых показателей
- **Maska** - только формирование описаний/масок
- **Shared** - хранение конфигов и справочников

#### 2. Слабое зацепление между сервисами
- Микросервисы взаимодействуют через четко определенные интерфейсы (REST API, gRPC)
- Асинхронные сообщения через Kafka
- Никто не обращается напрямую к внутренней базе данных другого сервиса

#### 3. Коммуникация через API и асинхронные очереди
- **REST/gRPC** для быстрых запросов (получить готовую выписку)
- **Kafka** для тяжелых операций (генерация целого месяца)
- Строго описанные и версионируемые форматы

#### 4. Модульность и изоляция
- Каждый сервис контейнизирован (Docker)
- Независимое развертывание
- Общие утилиты в Shared-сервисе или общих библиотеках

## Архитектура микросервисов

### 1. Matematika Service (Go)
**Назначение:** Числовое ядро - генерация финансовых данных выписки

**Технологии:** Go + Gin + PostgreSQL + Redis + Kafka
- **Почему Go:** Высокая производительность для расчетов, отличная работа с числами, конкуренция

**Функции:**
- Расчеты распределения входящих/исходящих платежей по дням
- Вычисление итоговых сумм (оборот, доходы, расходы, чистая прибыль)
- Подсчет балансов на каждый день и конечного баланса периода
- Применение бизнес-правил (распределение по неделям, учет праздников, процент прибыли)
- Генерация структурированного JSON с финансовым итогом

**Входные данные:**
- Параметры выписки (месяц, тип бизнеса B2B/B2C, начальный баланс)
- Пользовательские данные о транзакциях
- Конфигурации из Shared-сервиса

**Выходные данные:**
- JSON с финансовым итогом
- Список транзакций в технических категориях
- Транзакции помечены категориями и типами (ACH, Wire, расходы на маркетинг)

**API Endpoints:**
```
POST /generate-statement
GET /statement/{id}/status
GET /statement/{id}/result
```

**База данных:**
- PostgreSQL - шаблонные транзакции, результаты расчетов
- Redis - кэш результатов, блокировки

---

### 2. Maska Service (NestJS)
**Назначение:** Формирование описаний и форматирование

**Технологии:** NestJS + TypeScript + PostgreSQL + Redis + Kafka
- **Почему NestJS:** Быстрая разработка, встроенная валидация, работа с шаблонами

**Функции:**
- Преобразование "сырых" данных Matematika в финальный вид банковской выписки
- Наложение текстовых шаблонов, масок и описаний
- Формирование человекочитаемых описаний транзакций
- Применение пользовательских настроек (контрагенты)
- Генерация PDF/HTML выписок
- Форматирование дат, номеров счетов, названий компаний

**Входные данные:**
- JSON от Matematika (через REST или Kafka)
- Шаблоны и справочники из Shared-сервиса
- Пользовательские настройки

**Выходные данные:**
- Финальный JSON готовый для отображения
- PDF/HTML выписки
- Форматированные описания транзакций

**API Endpoints:**
```
GET /statement/{id}/formatted
POST /statement/{id}/generate-pdf
GET /statement/{id}/download
PUT /user/{id}/settings
```

**База данных:**
- PostgreSQL - метаданные выписок, пользовательские настройки
- Redis - кэш готовых выписок

---

### 3. Shared Service (Go)
**Назначение:** Общие конфигурации и данные

**Технологии:** Go + Gin + PostgreSQL + File Storage
- **Почему Go:** Простота, высокая производительность для обслуживания конфигов

**Функции:**
- Хранение конфигураций паттернов (шаблоны для разных отраслей)
- Справочники и вспомогательная информация
- Списки контрагентов и контекстных данных из CSV
- Таблица праздников
- Шаблоны строк (маски описаний транзакций)
- Общие утилиты (генерация ID, валидация JSON)

**Хранимые данные:**
- Конфигурации паттернов (gateways.csv, retails_ca.csv и др.)
- Таблица праздников
- Шаблоны масок описаний транзакций
- Списки контрагентов по категориям
- Бизнес-правила и константы

**API Endpoints:**
```
GET /config/patterns
GET /config/holidays
GET /config/masks
GET /config/contractors/{category}
GET /utils/generate-id
POST /utils/validate-json
```

**База данных:**
- PostgreSQL - конфигурации, метаданные
- File Storage - CSV файлы, шаблоны

## Технологический стек

### Backend Services
- **Go:** Matematika Service, Shared Service
- **NestJS:** Maska Service

### Базы данных
- **PostgreSQL:** Основная реляционная БД для всех сервисов
- **Redis:** Кэширование, сессии, очереди
- **File Storage:** CSV файлы, шаблоны документов

### Инфраструктура
- **Docker:** Контейнеризация сервисов
- **Docker Compose:** Локальная разработка
- **Kafka:** Асинхронная коммуникация
- **Nginx:** API Gateway и load balancer

### Мониторинг
- **Prometheus:** Метрики
- **Grafana:** Дашборды
- **Jaeger:** Distributed tracing
- **Loki:** Централизованное логирование

## Коммуникация между сервисами

### 🔄 **Полный поток взаимодействия от клиента:**

#### **1. Запрос генерации выписки (Синхронный)**
```
Client → Nginx → Matematika Service
POST /api/matematika/generate-statement
{
  "accountId": "12345",
  "month": "2025-01",
  "businessType": "B2C",
  "initialBalance": 10000.00
}
```

#### **2. Получение конфигураций (Синхронный)**
```
Matematika → Shared Service
GET /api/shared/config/patterns/{businessType}
GET /api/shared/config/holidays/{month}
```

#### **3. Расчеты и генерация (Внутренний процесс)**
```
Matematika Service:
- Применяет бизнес-правила
- Генерирует транзакции
- Рассчитывает балансы
- Сохраняет результаты в БД
```

#### **4. Уведомление о завершении расчетов (Асинхронный)**
```
Matematika → Kafka Topic: statement.calculation.completed
{
  "statementId": "stmt_12345",
  "accountId": "12345",
  "month": "2025-01",
  "status": "completed",
  "data": { /* JSON с результатами */ }
}
```

#### **5. Получение шаблонов (Синхронный)**
```
Maska → Shared Service
GET /api/shared/config/masks/{statementType}
GET /api/shared/config/contractors/{category}
```

#### **6. Форматирование выписки (Внутренний процесс)**
```
Maska Service:
- Получает данные от Matematika
- Применяет шаблоны и маски
- Генерирует PDF/HTML/Excel
- Сохраняет готовую выписку
```

#### **7. Уведомление о готовности (Асинхронный)**
```
Maska → Kafka Topic: statement.formatting.completed
{
  "statementId": "stmt_12345",
  "status": "ready",
  "downloadUrl": "/api/maska/statements/stmt_12345/download"
}
```

#### **8. Получение готовой выписки (Синхронный)**
```
Client → Nginx → Maska Service
GET /api/maska/statements/{statementId}/download
Response: PDF/HTML/Excel файл
```

### 📡 **Детальная схема коммуникации:**

#### **Синхронная коммуникация (HTTP/gRPC):**
- **Client ↔ Nginx** - Единая точка входа
- **Nginx ↔ Matematika** - Запросы на генерацию
- **Nginx ↔ Maska** - Получение готовых выписок
- **Matematika ↔ Shared** - Получение конфигов и паттернов
- **Maska ↔ Shared** - Получение шаблонов и справочников

#### **Асинхронная коммуникация (Kafka):**
- **Топики:**
  - `statement.generation.request` - Запросы на генерацию выписки
  - `statement.calculation.completed` - Результаты расчетов Matematika
  - `statement.formatting.completed` - Готовые выписки Maska
  - `statement.error` - Ошибки обработки

#### **Поток данных:**
```
1. Client → Nginx → Matematika (HTTP POST)
2. Matematika → Shared (HTTP GET) - конфиги
3. Matematika → Kafka (PUBLISH) - результаты расчетов
4. Maska ← Kafka (CONSUME) - получение результатов
5. Maska → Shared (HTTP GET) - шаблоны
6. Maska → Kafka (PUBLISH) - готовые выписки
7. Client ← Nginx ← Maska (HTTP GET) - скачивание
```

### 📊 **ASCII диаграмма потока данных:**

```
┌─────────┐    HTTP POST     ┌─────────┐    HTTP GET     ┌─────────┐
│ Client  │ ──────────────→ │  Nginx  │ ──────────────→ │Matematika│
│         │                 │         │                 │ Service │
└─────────┘                 └─────────┘                 └─────────┘
                                                              │
                                                              │ HTTP GET
                                                              ▼
                                                        ┌─────────┐
                                                        │ Shared  │
                                                        │ Service │
                                                        └─────────┘
                                                              ▲
                                                              │ HTTP GET
                                                              │
┌─────────┐    Kafka PUBLISH    ┌─────────┐    Kafka CONSUME ┌─────────┐
│Matematika│ ─────────────────→ │  Kafka  │ ←──────────────── │ Maska  │
│ Service  │                    │         │                   │ Service │
└─────────┘                     └─────────┘                   └─────────┘
                                                                 │
                                                                 │ HTTP GET
                                                                 ▼
                                                            ┌─────────┐
                                                            │ Shared  │
                                                            │ Service │
                                                            └─────────┘
                                                                 ▲
                                                                 │
                                                                 │
┌─────────┐    HTTP GET      ┌─────────┐    Kafka PUBLISH ┌─────────┐
│ Client  │ ←─────────────── │  Nginx  │ ←──────────────── │ Maska  │
│         │                  │         │                   │ Service │
└─────────┘                  └─────────┘                   └─────────┘

```

### 🔄 **Детальная схема взаимодействия:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ПОТОК ГЕНЕРАЦИИ ВЫПИСКИ                            │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────┐                                                                     
│ Client  │ ── 1. POST /api/matematika/generate-statement ──┐                  
└─────────┘                                                 │                  
                                                           │                  
┌─────────┐                                                 │                  
│  Nginx  │ ←───────────────────────────────────────────────┘                  
│Gateway  │                                                 │                  
└─────────┘                                                 │                  
     │                                                      │                  
     │ 2. POST /generate-statement                          │                  
     ▼                                                      │                  
┌─────────┐                                                 │                  
│Matematika│ ── 3. GET /api/shared/config/patterns ──────┐  │                  
│ Service  │   4. GET /api/shared/config/holidays ──────┐ │  │                  
└─────────┘                                           │ │  │                  
     │                                                │ │  │                  
     │ 5. Расчеты и генерация транзакций              │ │  │                  
     │                                                │ │  │                  
     │ 6. PUBLISH statement.calculation.completed ────┼─┼──┼─────────────────┐
     ▼                                                │ │  │                 │
┌─────────┐                                           │ │  │                 │
│  Kafka  │ ←─────────────────────────────────────────┘ │  │                 │
│ Message │                                             │  │                 │
│  Queue  │                                             │  │                 │
└─────────┘                                             │  │                 │
     │                                                  │  │                 │
     │ 7. CONSUME statement.calculation.completed ──────┼──┼─────────────────┘
     ▼                                                  │  │                  
┌─────────┐                                             │  │                  
│ Maska   │ ── 8. GET /api/shared/config/masks ───────┐ │  │                  
│ Service │   9. GET /api/shared/config/contractors ──┼─┘  │                  
└─────────┘                                           │    │                  
     │                                                │    │                  
     │ 10. Форматирование и генерация PDF/HTML/Excel │    │                  
     │                                                │    │                  
     │ 11. PUBLISH statement.formatting.completed ────┼────┼─────────────────┐
     ▼                                                │    │                 │
┌─────────┐                                           │    │                 │
│  Kafka  │ ←─────────────────────────────────────────┘    │                 │
│ Message │                                                │                 │
│  Queue  │                                                │                 │
└─────────┘                                                │                 │
     │                                                     │                 │
     │ 12. CONSUME statement.formatting.completed ────────┼─────────────────┘
     ▼                                                     │                  
┌─────────┐                                                │                  
│  Nginx  │ ←── 13. GET /api/maska/statements/{id}/download ┘                  
│Gateway  │                                                │                  
└─────────┘                                                │                  
     │                                                     │                  
     │ 14. Response: PDF/HTML/Excel файл                  │                  
     ▼                                                     │                  
┌─────────┐                                                │                  
│ Client  │ ←──────────────────────────────────────────────┘                  
└─────────┘                                                                    

┌─────────────────────────────────────────────────────────────────────────────┐
│                            ВСПОМОГАТЕЛЬНЫЕ СЕРВИСЫ                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────┐                HTTP GET                ┌─────────┐
│Matematika│ ──────────────────────────────────────→ │ Shared  │
│ Service  │                                        │ Service │
└─────────┘                                        └─────────┘
     │                                                   │
     │ • Конфигурации паттернов                         │
     │ • Таблица праздников                             │
     │ • Бизнес-правила                                 │
     │                                                   │
     │                                                   │
┌─────────┐                HTTP GET                ┌─────────┐
│ Maska   │ ──────────────────────────────────────→ │ Shared  │
│ Service │                                        │ Service │
└─────────┘                                        └─────────┘
     │                                                   │
     │ • Шаблоны масок                                   │
     │ • Справочники контрагентов                        │
     │ • Форматы описаний                                │
     │                                                   │
```

### 🔧 **Технические детали:**

#### **HTTP Endpoints:**
```
Matematika Service:
- POST /generate-statement
- GET /statement/{id}/status
- GET /statement/{id}/result

Maska Service:
- GET /statement/{id}/formatted
- POST /statement/{id}/generate-pdf
- GET /statement/{id}/download
- PUT /user/{id}/settings

Shared Service:
- GET /config/patterns
- GET /config/holidays
- GET /config/masks
- GET /config/contractors/{category}
```

#### **Kafka Topics:**
```
statement.generation.request:
{
  "statementId": "string",
  "accountId": "string", 
  "month": "string",
  "businessType": "B2B|B2C",
  "initialBalance": "number"
}

statement.calculation.completed:
{
  "statementId": "string",
  "accountId": "string",
  "status": "completed|failed",
  "data": { /* расчетные данные */ }
}

statement.formatting.completed:
{
  "statementId": "string",
  "status": "ready|failed",
  "downloadUrl": "string",
  "formats": ["pdf", "html", "excel"]
}
```

## Структура проекта

```
business_bank_back/
├── .github/workflows/             # CI/CD
│   ├── ci.yml                     # Continuous Integration
│   ├── deploy.yml                 # Deployment pipeline
│   └── security.yml               # Security scanning
├── services/
│   ├── matematika/                # Go - Числовое ядро (расчеты)
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go        # Точка входа приложения
│   │   ├── internal/
│   │   │   ├── calculation/       # Логика расчетов
│   │   │   │   ├── handler.go     # HTTP обработчики
│   │   │   │   ├── service.go     # Бизнес-логика расчетов
│   │   │   │   ├── repository.go  # Доступ к данным
│   │   │   │   ├── models.go      # Структуры данных
│   │   │   │   └── calculator.go  # Математические расчеты
│   │   │   ├── transactions/      # Генерация транзакций
│   │   │   │   ├── generator.go   # Генератор транзакций
│   │   │   │   ├── validator.go   # Валидация транзакций
│   │   │   │   ├── types.go       # Типы транзакций
│   │   │   │   └── patterns.go    # Паттерны генерации
│   │   │   ├── business-rules/    # Бизнес-правила
│   │   │   │   ├── rules.go       # Основные правила
│   │   │   │   ├── constraints.go # Ограничения
│   │   │   │   ├── holidays.go    # Обработка праздников
│   │   │   │   └── validation.go  # Валидация правил
│   │   │   ├── kafka/             # Kafka интеграция
│   │   │   │   ├── producer.go    # Отправка сообщений
│   │   │   │   ├── consumer.go    # Получение сообщений
│   │   │   │   ├── messages.go   # Структуры сообщений
│   │   │   │   └── config.go     # Конфигурация Kafka
│   │   │   ├── middleware/        # Middleware
│   │   │   │   ├── auth.go        # Аутентификация
│   │   │   │   ├── logging.go     # Логирование
│   │   │   │   ├── cors.go        # CORS
│   │   │   │   └── recovery.go    # Обработка паник
│   │   │   ├── config/            # Конфигурация
│   │   │   │   ├── config.go      # Основная конфигурация
│   │   │   │   ├── database.go    # Настройки БД
│   │   │   │   └── kafka.go       # Настройки Kafka
│   │   │   └── database/          # База данных
│   │   │       ├── postgres.go    # PostgreSQL клиент
│   │   │       ├── redis.go       # Redis клиент
│   │   │       └── migrations/    # Миграции БД
│   │   ├── pkg/                   # Публичные пакеты
│   │   │   ├── logger/            # Логирование
│   │   │   │   ├── logger.go      # Основной логгер
│   │   │   │   └── structured.go  # Структурированные логи
│   │   │   ├── validator/         # Валидация
│   │   │   │   ├── validator.go   # Валидатор
│   │   │   │   └── rules.go       # Правила валидации
│   │   │   └── calculator/        # Калькулятор
│   │   │       ├── math.go        # Математические функции
│   │   │       └── financial.go  # Финансовые расчеты
│   │   ├── test/                  # Тесты
│   │   │   ├── unit/              # Юнит тесты
│   │   │   │   ├── calculation_test.go
│   │   │   │   ├── transactions_test.go
│   │   │   │   └── business_rules_test.go
│   │   │   ├── integration/       # Интеграционные тесты
│   │   │   │   ├── api_test.go
│   │   │   │   └── kafka_test.go
│   │   │   └── fixtures/          # Тестовые данные
│   │   │       ├── statements.json
│   │   │       └── transactions.json
│   │   ├── go.mod                 # Go модули
│   │   ├── go.sum                 # Зависимости
│   │   ├── Dockerfile             # Docker образ
│   │   └── README.md              # Документация сервиса
│   ├── maska/                     # NestJS - Форматирование
│   │   ├── src/
│   │   │   ├── formatting/        # Основная логика форматирования
│   │   │   │   ├── formatting.controller.ts    # HTTP контроллер
│   │   │   │   ├── formatting.service.ts       # Бизнес-логика
│   │   │   │   ├── formatting.module.ts        # NestJS модуль
│   │   │   │   └── dto/                        # Data Transfer Objects
│   │   │   │       ├── format-statement.dto.ts
│   │   │   │       ├── generate-pdf.dto.ts
│   │   │   │       └── user-settings.dto.ts
│   │   │   ├── templates/         # Работа с шаблонами
│   │   │   │   ├── templates.service.ts        # Сервис шаблонов
│   │   │   │   ├── templates.module.ts         # Модуль шаблонов
│   │   │   │   └── processors/                  # Процессоры
│   │   │   │       ├── pdf.processor.ts        # PDF генерация
│   │   │   │       ├── html.processor.ts       # HTML генерация
│   │   │   │       ├── json.processor.ts       # JSON обработка
│   │   │   │       └── excel.processor.ts     # Excel генерация
│   │   │   ├── masks/             # Маски описаний
│   │   │   │   ├── masks.service.ts            # Сервис масок
│   │   │   │   ├── masks.module.ts             # Модуль масок
│   │   │   │   └── processors/                 # Процессоры масок
│   │   │   │       ├── description.processor.ts
│   │   │   │       ├── company.processor.ts
│   │   │   │       └── transaction.processor.ts
│   │   │   ├── kafka/             # Kafka интеграция
│   │   │   │   ├── consumer.service.ts         # Kafka consumer
│   │   │   │   ├── consumer.module.ts          # Модуль consumer
│   │   │   │   └── processors/                 # Процессоры сообщений
│   │   │   │       ├── statement.processor.ts
│   │   │   │       └── calculation.processor.ts
│   │   │   ├── settings/          # Пользовательские настройки
│   │   │   │   ├── settings.controller.ts      # Контроллер настроек
│   │   │   │   ├── settings.service.ts         # Сервис настроек
│   │   │   │   ├── settings.module.ts          # Модуль настроек
│   │   │   │   └── entities/
│   │   │   │       └── user-settings.entity.ts
│   │   │   ├── common/            # Общие компоненты
│   │   │   │   ├── decorators/                 # Декораторы
│   │   │   │   │   ├── correlation-id.decorator.ts
│   │   │   │   │   └── logging.decorator.ts
│   │   │   │   ├── filters/                    # Фильтры
│   │   │   │   │   ├── http-exception.filter.ts
│   │   │   │   │   └── validation.filter.ts
│   │   │   │   ├── interceptors/              # Интерцепторы
│   │   │   │   │   ├── logging.interceptor.ts
│   │   │   │   │   └── correlation.interceptor.ts
│   │   │   │   └── pipes/                     # Пайпы
│   │   │   │       ├── validation.pipe.ts
│   │   │   │       └── transform.pipe.ts
│   │   │   ├── config/            # Конфигурация
│   │   │   │   ├── database.config.ts          # Настройки БД
│   │   │   │   ├── redis.config.ts            # Настройки Redis
│   │   │   │   ├── kafka.config.ts            # Настройки Kafka
│   │   │   │   └── app.config.ts               # Основные настройки
│   │   │   ├── entities/          # TypeORM сущности
│   │   │   │   ├── statement.entity.ts         # Сущность выписки
│   │   │   │   ├── template.entity.ts         # Сущность шаблона
│   │   │   │   └── user-settings.entity.ts    # Настройки пользователя
│   │   │   └── main.ts            # Точка входа приложения
│   │   ├── test/                  # Тесты
│   │   │   ├── unit/              # Юнит тесты
│   │   │   │   ├── formatting/
│   │   │   │   ├── templates/
│   │   │   │   ├── masks/
│   │   │   │   └── settings/
│   │   │   ├── integration/       # Интеграционные тесты
│   │   │   │   ├── api/
│   │   │   │   └── kafka/
│   │   │   └── e2e/               # End-to-End тесты
│   │   │       ├── statement-generation.e2e-spec.ts
│   │   │       └── pdf-generation.e2e-spec.ts
│   │   ├── package.json           # NPM зависимости
│   │   ├── package-lock.json      # Зафиксированные версии
│   │   ├── tsconfig.json          # TypeScript конфигурация
│   │   ├── nest-cli.json          # NestJS CLI конфигурация
│   │   ├── Dockerfile             # Docker образ
│   │   └── README.md              # Документация сервиса
│   └── shared/                    # Go - Общие конфигурации
│       ├── cmd/
│       │   └── server/
│       │       └── main.go        # Точка входа приложения
│       ├── internal/
│       │   ├── config/            # Управление конфигами
│       │   │   ├── handler.go     # HTTP обработчики
│       │   │   ├── service.go     # Бизнес-логика
│       │   │   ├── repository.go  # Доступ к данным
│       │   │   └── models.go      # Структуры данных
│       │   ├── patterns/          # Паттерны бизнеса
│       │   │   ├── patterns.service.go        # Сервис паттернов
│       │   │   ├── patterns.repository.go     # Репозиторий паттернов
│       │   │   ├── patterns.go                 # Структуры паттернов
│       │   │   └── loader.go                   # Загрузчик паттернов
│       │   ├── contractors/       # Справочники контрагентов
│       │   │   ├── contractors.service.go      # Сервис контрагентов
│       │   │   ├── contractors.repository.go   # Репозиторий контрагентов
│       │   │   ├── csv/                       # CSV обработка
│       │   │   │   ├── parser.go              # Парсер CSV
│       │   │   │   └── loader.go              # Загрузчик CSV
│       │   │   └── models.go                   # Модели контрагентов
│       │   ├── holidays/          # Таблица праздников
│       │   │   ├── holidays.service.go        # Сервис праздников
│       │   │   ├── holidays.repository.go     # Репозиторий праздников
│       │   │   └── models.go                  # Модели праздников
│       │   ├── masks/             # Шаблоны масок
│       │   │   ├── masks.service.go           # Сервис масок
│       │   │   ├── masks.repository.go        # Репозиторий масок
│       │   │   └── models.go                  # Модели масок
│       │   ├── utils/             # Общие утилиты
│       │   │   ├── id-generator.go            # Генератор ID
│       │   │   ├── json-validator.go         # Валидатор JSON
│       │   │   ├── csv-parser.go              # Парсер CSV
│       │   │   └── file-loader.go              # Загрузчик файлов
│       │   ├── middleware/        # Middleware
│       │   │   ├── auth.go                    # Аутентификация
│       │   │   ├── logging.go                # Логирование
│       │   │   ├── cors.go                    # CORS
│       │   │   └── recovery.go                # Обработка паник
│       │   ├── config/            # Конфигурация
│       │   │   ├── config.go                  # Основная конфигурация
│       │   │   ├── database.go                # Настройки БД
│       │   │   └── file-storage.go            # Настройки файлового хранилища
│       │   └── database/          # База данных
│       │       ├── postgres.go                # PostgreSQL клиент
│       │       └── migrations/                # Миграции БД
│       ├── data/                  # CSV файлы и шаблоны
│       │   ├── contractors/                   # Контрагенты
│       │   │   ├── gateways.csv              # Шлюзы
│       │   │   ├── retails_ca.csv            # Розница Калифорния
│       │   │   ├── retails_ny.csv            # Розница Нью-Йорк
│       │   │   └── suppliers.csv             # Поставщики
│       │   ├── holidays/                      # Праздники
│       │   │   ├── holidays.json             # JSON праздников
│       │   │   └── business-days.json        # Рабочие дни
│       │   └── masks/                         # Шаблоны масок
│       │       ├── templates.json            # JSON шаблонов
│       │       ├── descriptions.json         # Описания
│       │       └── formats.json              # Форматы
│       ├── pkg/                   # Публичные пакеты
│       │   ├── logger/            # Логирование
│       │   │   ├── logger.go      # Основной логгер
│       │   │   └── structured.go  # Структурированные логи
│       │   ├── validator/         # Валидация
│       │   │   ├── validator.go   # Валидатор
│       │   │   └── rules.go       # Правила валидации
│       │   └── utils/             # Утилиты
│       │       ├── csv.go         # CSV утилиты
│       │       ├── json.go        # JSON утилиты
│       │       └── file.go        # Файловые утилиты
│       ├── test/                  # Тесты
│       │   ├── unit/              # Юнит тесты
│       │   │   ├── config_test.go
│       │   │   ├── patterns_test.go
│       │   │   ├── contractors_test.go
│       │   │   ├── holidays_test.go
│       │   │   └── masks_test.go
│       │   ├── integration/       # Интеграционные тесты
│       │   │   ├── api_test.go
│       │   │   └── data_test.go
│       │   └── fixtures/          # Тестовые данные
│       │       ├── test-patterns.json
│       │       └── test-contractors.csv
│       ├── go.mod                 # Go модули
│       ├── go.sum                 # Зависимости
│       ├── Dockerfile             # Docker образ
│       └── README.md              # Документация сервиса
├── contracts/                     # API контракты и схемы
│   ├── openapi/                   # OpenAPI спецификации
│   │   ├── matematika-service.yaml    # API Matematika
│   │   ├── maska-service.yaml         # API Maska
│   │   └── shared-service.yaml        # API Shared
│   ├── kafka/                     # Kafka схемы
│   │   ├── statement-generation.json  # Схема генерации выписки
│   │   ├── calculation-completed.json # Схема завершения расчетов
│   │   └── formatting-completed.json  # Схема завершения форматирования
│   └── types/                     # Общие типы
│       ├── typescript/            # TypeScript типы
│       │   ├── statement.types.ts      # Типы выписок
│       │   ├── transaction.types.ts    # Типы транзакций
│       │   ├── user.types.ts          # Типы пользователей
│       │   └── common.types.ts         # Общие типы
│       └── go/                    # Go типы
│           ├── statement/         # Типы выписок
│           │   ├── statement.go
│           │   └── transaction.go
│           ├── user/              # Типы пользователей
│           │   └── user.go
│           └── common/             # Общие типы
│               ├── response.go
│               └── error.go
├── infrastructure/                # Инфраструктура
│   ├── nginx/                     # Nginx конфигурация
│   │   ├── nginx.conf             # Основная конфигурация
│   │   ├── default.conf            # Конфигурация по умолчанию
│   │   └── ssl/                    # SSL сертификаты
│   ├── monitoring/                # Мониторинг
│   │   ├── prometheus/             # Prometheus
│   │   │   ├── prometheus.yml     # Конфигурация Prometheus
│   │   │   └── rules/              # Правила алертов
│   │   ├── grafana/               # Grafana
│   │   │   ├── dashboards/         # Дашборды
│   │   │   │   ├── matematika.json
│   │   │   │   ├── maska.json
│   │   │   │   └── shared.json
│   │   │   └── datasources/        # Источники данных
│   │   ├── jaeger/                # Jaeger
│   │   │   └── jaeger.yml         # Конфигурация Jaeger
│   │   └── loki/                  # Loki
│   │       └── loki.yml           # Конфигурация Loki
│   └── scripts/                   # Скрипты развертывания
│       ├── deploy.sh              # Скрипт развертывания
│       ├── backup.sh              # Скрипт резервного копирования
│       ├── migrate.sh             # Скрипт миграций
│       └── health-check.sh        # Скрипт проверки здоровья
├── docs/                          # Документация
│   ├── api/                       # API документация
│   │   ├── matematika-api.md      # API Matematika
│   │   ├── maska-api.md           # API Maska
│   │   └── shared-api.md          # API Shared
│   ├── architecture/              # Архитектурная документация
│   │   ├── system-overview.md     # Обзор системы
│   │   ├── microservices.md       # Микросервисы
│   │   └── data-flow.md           # Поток данных
│   ├── deployment/                # Инструкции по развертыванию
│   │   ├── local-setup.md         # Локальная настройка
│   │   ├── production.md          # Продакшн развертывание
│   │   └── troubleshooting.md     # Решение проблем
│   └── development/               # Руководство разработчика
│       ├── getting-started.md     # Начало работы
│       ├── coding-standards.md    # Стандарты кодирования
│       └── testing.md             # Тестирование
├── docker-compose.yml             # Локальная разработка (ОСНОВНОЙ)
├── docker-compose.prod.yml        # Продакшн конфигурация (ОСНОВНОЙ)
├── Makefile                       # Управление проектом
├── .env.example                   # Пример переменных окружения
├── .env.local                     # Локальные переменные (не в git)
├── .env.prod                      # Продакшн переменные (не в git)
├── .gitignore                     # Игнорируемые файлы
└── README.md                      # Этот файл
```

## Docker конфигурация

### 🐳 **Правильная организация Docker файлов:**

#### **Корневые файлы (ОСНОВНЫЕ):**
- `docker-compose.yml` - **Локальная разработка**
- `docker-compose.prod.yml` - **Продакшн развертывание**
- `.env.example` - Пример переменных окружения
- `.env.local` - Локальные переменные (не в git)
- `.env.prod` - Продакшн переменные (не в git)

#### **Сервисы (индивидуальные Dockerfile):**
- `services/matematika/Dockerfile` - Go сервис
- `services/maska/Dockerfile` - NestJS сервис  
- `services/shared/Dockerfile` - Go сервис

#### **Инфраструктура (конфигурации):**
- `infrastructure/nginx/` - Nginx конфигурация
- `infrastructure/monitoring/` - Prometheus, Grafana, Jaeger, Loki

### 📁 **Структура Docker файлов:**

```
business_bank_back/
├── docker-compose.yml              # Локальная разработка
├── docker-compose.prod.yml         # Продакшн
├── .env.example                    # Пример переменных
├── .env.local                      # Локальные переменные (gitignore)
├── .env.prod                       # Продакшн переменные (gitignore)
├── services/
│   ├── matematika/
│   │   └── Dockerfile              # Go сервис
│   ├── maska/
│   │   └── Dockerfile              # NestJS сервис
│   └── shared/
│       └── Dockerfile              # Go сервис
└── infrastructure/
    ├── nginx/
    │   ├── nginx.conf
    │   └── default.conf
    └── monitoring/
        ├── prometheus/
        ├── grafana/
        ├── jaeger/
        └── loki/
```

### 🔧 **Использование:**

#### **Локальная разработка:**
```bash
# Использует docker-compose.yml + .env.local
docker-compose up -d
```

#### **Продакшн развертывание:**
```bash
# Использует docker-compose.prod.yml + .env.prod
docker-compose -f docker-compose.prod.yml up -d
```

#### **Переменные окружения:**
```bash
# Копируем пример и настраиваем
cp .env.example .env.local
cp .env.example .env.prod

# Редактируем под свои нужды
nano .env.local
nano .env.prod
```

## Команды для управления проектом

### Makefile команды:
```makefile
.PHONY: dev dev-build test build deploy clean logs restart migrate setup

# Настройка окружения
setup:
	@echo "Setting up environment..."
	@if [ ! -f .env.local ]; then cp .env.example .env.local; fi
	@if [ ! -f .env.prod ]; then cp .env.example .env.prod; fi
	@echo "Environment files created. Please edit .env.local and .env.prod"

# Локальная разработка
dev:
	@echo "Starting development environment..."
	docker-compose up -d
	@echo "All services running:"
	@echo "Matematika Service: http://localhost:8080"
	@echo "Maska Service: http://localhost:3000"
	@echo "Shared Service: http://localhost:8081"
	@echo "Kafka: http://localhost:9092"
	@echo "Grafana: http://localhost:3001"
	@echo "Nginx Gateway: http://localhost:80"

# Разработка с пересборкой
dev-build:
	@echo "Building and starting development environment..."
	docker-compose up -d --build
	@echo "All services running with latest changes"

# Тестирование
test:
	@echo "Running tests..."
	@cd services/matematika && go test ./...
	@cd services/maska && npm test
	@cd services/shared && go test ./...
	@echo "All tests completed"

# Сборка всех сервисов
build:
	@echo "Building all services..."
	docker-compose build
	@echo "Build completed"

# Продакшн развертывание
deploy:
	@echo "Deploying to production..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "Production deployment completed"

# Очистка
clean:
	@echo "Cleaning up..."
	docker-compose down -v
	docker-compose -f docker-compose.prod.yml down -v
	docker system prune -f
	@echo "Cleanup completed"

# Логи всех сервисов
logs:
	docker-compose logs -f

# Логи конкретного сервиса
logs-matematika:
	docker-compose logs -f matematika

logs-maska:
	docker-compose logs -f maska

logs-shared:
	docker-compose logs -f shared

# Перезапуск всех сервисов
restart:
	docker-compose restart

# Перезапуск конкретного сервиса
restart-matematika:
	docker-compose restart matematika

restart-maska:
	docker-compose restart maska

restart-shared:
	docker-compose restart shared

# Миграции БД
migrate:
	@echo "Running database migrations..."
	@cd services/matematika && go run cmd/migrate/main.go
	@cd services/maska && npm run migration:run
	@cd services/shared && go run cmd/migrate/main.go
	@echo "Migrations completed"

# Проверка здоровья сервисов
health:
	@echo "Checking service health..."
	@curl -f http://localhost:8080/health || echo "Matematika: DOWN"
	@curl -f http://localhost:3000/health || echo "Maska: DOWN"
	@curl -f http://localhost:8081/health || echo "Shared: DOWN"
	@echo "Health check completed"

# Полная остановка и очистка
stop:
	docker-compose down
	docker-compose -f docker-compose.prod.yml down

# Показать статус сервисов
status:
	docker-compose ps
```

## Тестирование

### 1. Юнит-тесты
- **Matematika:** Тесты расчетов балансов, процентов, бизнес-правил
- **Maska:** Тесты форматирования, парсинга масок, генерации PDF
- **Shared:** Тесты выдачи конфигов, валидации данных

### 2. Контрактные тесты
- JSON-схемы для взаимодействия между сервисами
- Тесты форматов Kafka-сообщений
- Валидация API контрактов

### 3. End-to-End тесты
- Сценарий "Генерация выписки за месяц с 5-й пятницей"
- Сценарий "Добавление пользовательских транзакций"
- Проверка корректности взаимодействия всех сервисов

## Логирование и трассировка

### Централизованное структурированное логирование
- Winston в NestJS для структурированного логирования (JSON)
- Централизованный сбор через Loki (Grafana Loki)
- Логирование ключевых событий и ошибок

### Correlation ID для сквозной трассировки
- Уникальный идентификатор запроса (UUID) на границе системы
- Передача correlationId через все сервисы и Kafka
- Возможность собрать всю цепочку событий по одному ID

### Distributed Tracing
- OpenTelemetry для трассировки запросов
- Jaeger для сбора и анализа трейсов
- Связь логов с трейсами через correlationId

## Обозреваемость и поддерживаемость

### Health-check эндпоинты
- GET /health для каждого сервиса
- Проверка доступности БД, Kafka, зависимостей
- Интеграция с оркестратором (Docker/K8s)

### Метрики и мониторинг
- Prometheus метрики для каждого сервиса
- Кастомные метрики: generated_statements_total, generation_duration
- Grafana дашборды для мониторинга

### Управление конфигурацией
- Централизованные настройки через переменные окружения
- NestJS ConfigModule для чтения конфигов
- Документированные .env файлы

## Масштабирование и отказоустойчивость

### Горизонтальное масштабирование
- Stateless микросервисы
- Параллельные экземпляры через Docker/K8s
- Kafka consumer groups для распределения нагрузки

### Отказоустойчивость
- Слабая связанность между сервисами
- Идемпотентность операций
- Обработка сбоев и повторов
- Автоматическое восстановление

### Гарантии доставки
- Kafka at-least-once доставка
- Идемпотентная обработка сообщений
- Порядок обработки через партиции

## Безопасность

### Валидация и санитария
- ValidationPipe с DTO и class-validator
- Проверка бизнес-условий на входе
- Санитизация пользовательских данных

### Ограничение доступа
- Network policy для внутренней сети
- Bearer token аутентификация
- Ограничение доступа по IP

### Защита от уязвимостей
- ORM для предотвращения SQL-инъекций
- Защита от XSS при отображении
- Контроль ошибок без утечки информации

## Документация

### Swagger / OpenAPI
- Описание всех REST endpoints
- Живая документация вместе с кодом
- Доступ к /docs endpoint для интеграции

### JSON-схемы
- Описание форматов данных между сервисами
- Схемы Kafka-сообщений
- Валидация контрактов

### Описание паттернов и конфигов
- Документация бизнес-паттернов
- Описание форматов конфигураций
- Инструкции по добавлению новых шаблонов

## Заключение

Данная архитектура обеспечивает:
- **Строгое разделение ответственности** между тремя микросервисами
- **Высокую производительность** через правильный выбор технологий
- **Масштабируемость** через stateless архитектуру и Kafka
- **Надежность** через fault tolerance и мониторинг
- **Поддерживаемость** через четкую структуру и документацию
- **Безопасность** через валидацию и контроль доступа













📈 Рекомендации по планированию:
1. Итеративная разработка:
2. MVP подход:
Сначала базовая функциональность
Потом расширенные возможности
В конце оптимизация
3. Приоритеты:
Высокий: Matematika (бизнес-логика)
Средний: Maska (форматирование)
Низкий: Shared (конфиги)