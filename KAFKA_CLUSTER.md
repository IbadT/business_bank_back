# 🏗️ KAFKA CLUSTER CONFIGURATION

## 📋 Обзор

Kafka кластер состоит из **2 брокеров** с репликацией для обеспечения отказоустойчивости и горизонтального масштабирования.

---

## 🔧 Архитектура кластера

```
┌─────────────────────────────────────────────────────────┐
│                     ZOOKEEPER                            │
│                  (Координатор)                           │
│                   Port: 2181                             │
└─────────────────────────────────────────────────────────┘
                           │
                           │ (Координация)
                           │
         ┌─────────────────┴─────────────────┐
         │                                   │
         ▼                                   ▼
┌─────────────────────┐         ┌─────────────────────┐
│   KAFKA BROKER 1    │◄───────►│   KAFKA BROKER 2    │
│   (kafka1)          │         │   (kafka2)          │
│                     │         │                     │
│   Internal: 9092    │         │   Internal: 9093    │
│   External: 19092   │         │   External: 19093   │
│   Broker ID: 1      │         │   Broker ID: 2      │
└─────────────────────┘         └─────────────────────┘
         │                                   │
         └─────────────────┬─────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │  TOPIC PARTITIONS    │
                │  (Распределены)      │
                └──────────────────────┘
```

---

## 🎯 Конфигурация брокеров

### Kafka Broker 1 (kafka1)

| Параметр | Значение | Описание |
|----------|----------|----------|
| **Container** | kafka1 | Имя контейнера |
| **Broker ID** | 1 | Уникальный ID в кластере |
| **Internal Port** | 9092 | Порт для межсервисной коммуникации |
| **External Port** | 19092 | Порт для host machine |
| **Listener** | INTERNAL://kafka1:9092 | Внутренний listener |
| **External Listener** | EXTERNAL://localhost:19092 | Внешний listener |
| **Volume** | kafka1_data | Персистентное хранилище |

### Kafka Broker 2 (kafka2)

| Параметр | Значение | Описание |
|----------|----------|----------|
| **Container** | kafka2 | Имя контейнера |
| **Broker ID** | 2 | Уникальный ID в кластере |
| **Internal Port** | 9093 | Порт для межсервисной коммуникации |
| **External Port** | 19093 | Порт для host machine |
| **Listener** | INTERNAL://kafka2:9093 | Внутренний listener |
| **External Listener** | EXTERNAL://localhost:19093 | Внешний listener |
| **Volume** | kafka2_data | Персистентное хранилище |

---

## 📊 Топики и распределение партиций

### 1. statement.calculation.completed

**Назначение:** Результаты расчетов от Matematika Service

| Параметр | Значение |
|----------|----------|
| Партиций | 6 |
| Реплик | 2 |
| Retention | 7 дней (604800000 ms) |
| Compression | Snappy |
| Min ISR | 1 |

**Распределение:**
- Broker 1: партиции 0, 2, 4 (leader) + реплики 1, 3, 5
- Broker 2: партиции 1, 3, 5 (leader) + реплики 0, 2, 4

### 2. statement.generation.request

**Назначение:** Запросы на генерацию выписок

| Параметр | Значение |
|----------|----------|
| Партиций | 4 |
| Реплик | 2 |
| Retention | 1 день (86400000 ms) |
| Compression | Snappy |
| Min ISR | 1 |

**Распределение:**
- Broker 1: партиции 0, 2 (leader) + реплики 1, 3
- Broker 2: партиции 1, 3 (leader) + реплики 0, 2

### 3. statement.formatting.completed

**Назначение:** Готовые выписки от Maska Service

| Параметр | Значение |
|----------|----------|
| Партиций | 4 |
| Реплик | 2 |
| Retention | 7 дней (604800000 ms) |
| Compression | Snappy |
| Min ISR | 1 |

**Распределение:**
- Broker 1: партиции 0, 2 (leader) + реплики 1, 3
- Broker 2: партиции 1, 3 (leader) + реплики 0, 2

### 4. statement.error

**Назначение:** Ошибки обработки (меньше нагрузки)

| Параметр | Значение |
|----------|----------|
| Партиций | 2 |
| Реплик | 2 |
| Retention | 30 дней (2592000000 ms) |
| Compression | Snappy |
| Min ISR | 1 |

**Распределение:**
- Broker 1: партиция 0 (leader) + реплика 1
- Broker 2: партиция 1 (leader) + реплика 0

---

## 🚀 Запуск кластера

### 1. Запустить все сервисы:

```bash
docker compose up -d
```

### 2. Проверить статус брокеров:

```bash
docker compose ps
```

Должны быть запущены:
- ✅ zookeeper (healthy)
- ✅ kafka1 (healthy)
- ✅ kafka2 (healthy)
- ✅ kafka-init (completed) - инициализирует топики
- ✅ kafdrop (healthy)

### 3. Проверить логи инициализации:

```bash
docker logs kafka-init
```

Увидишь:
```
✅ Топик: statement.calculation.completed (6 партиций, 2 реплики)
✅ Топик: statement.generation.request (4 партиции, 2 реплики)
✅ Топик: statement.formatting.completed (4 партиции, 2 реплики)
✅ Топик: statement.error (2 партиции, 2 реплики)
```

---

## 🔍 Мониторинг кластера

### Kafdrop UI

**URL:** http://localhost:9000

**Что увидишь:**

1. **Brokers:**
   - kafka1 (ID: 1) - healthy
   - kafka2 (ID: 2) - healthy

2. **Topics:**
   - statement.calculation.completed (6 партиций)
   - statement.generation.request (4 партиции)
   - statement.formatting.completed (4 партиции)
   - statement.error (2 партиции)

3. **Partitions per topic:**
   - Partition ID
   - Leader broker
   - Replicas
   - In-Sync Replicas (ISR)

4. **Consumer Groups:**
   - matematika-service-group
   - maska-service-group
   - Lag per partition

---

## 🔧 Подключение к кластеру

### Из микросервисов (внутри Docker network):

```go
// В matematika service
kafkaBrokers := []string{"kafka1:9092", "kafka2:9093"}

producerConfig := kafka.DefaultProducerConfig(kafkaBrokers)
kafkaProducer, err := kafka.NewProducer(producerConfig, log.Default())
```

### Из host machine (для локальной разработки):

```go
// Локальное подключение
kafkaBrokers := []string{"localhost:19092", "localhost:19093"}
```

### Environment Variables:

```yaml
# docker-compose.yml
matematika:
  environment:
    KAFKA_BROKERS: "kafka1:9092,kafka2:9093"  # Оба брокера
```

---

## 📈 Преимущества кластера

### 1. **Отказоустойчивость (High Availability)**

- **2 реплики каждой партиции** - данные дублируются
- **Автоматический failover** - при падении broker 1, broker 2 становится leader
- **Min ISR = 1** - минимум 1 реплика должна быть синхронизирована

**Пример:**
```
Если kafka1 упал:
- Broker 2 становится leader для партиций 0, 2, 4
- Данные не теряются (есть реплики)
- Consumer продолжают работать
```

### 2. **Горизонтальное масштабирование**

- **6 партиций в statement.calculation.completed** = до 6 consumer'ов параллельно
- **Равномерное распределение** - каждый broker обрабатывает половину нагрузки
- **Load balancing** - Kafka автоматически распределяет партиции

**Пример:**
```
Consumer Group "matematika-service-group" с 3 инстансами:
- Consumer 1: партиции 0, 1
- Consumer 2: партиции 2, 3
- Consumer 3: партиции 4, 5

Каждый инстанс обрабатывает ~33% сообщений
```

### 3. **Производительность**

- **Snappy compression** - уменьшает трафик на 50-70%
- **Batch processing** - Kafka группирует сообщения для эффективности
- **Параллельная обработка** - несколько партиций = параллелизм

---

## 🛠️ Управление кластером

### Просмотр топиков:

```bash
docker exec kafka1 kafka-topics --list --bootstrap-server localhost:9092
```

### Детали топика:

```bash
docker exec kafka1 kafka-topics --describe \
  --topic statement.calculation.completed \
  --bootstrap-server localhost:9092
```

Вывод:
```
Topic: statement.calculation.completed
Partition: 0    Leader: 1    Replicas: 1,2    Isr: 1,2
Partition: 1    Leader: 2    Replicas: 2,1    Isr: 2,1
Partition: 2    Leader: 1    Replicas: 1,2    Isr: 1,2
Partition: 3    Leader: 2    Replicas: 2,1    Isr: 2,1
Partition: 4    Leader: 1    Replicas: 1,2    Isr: 1,2
Partition: 5    Leader: 2    Replicas: 2,1    Isr: 2,1
```

### Создание нового топика вручную:

```bash
docker exec kafka1 kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic new-topic \
  --partitions 4 \
  --replication-factor 2
```

### Просмотр Consumer Groups:

```bash
docker exec kafka1 kafka-consumer-groups --list \
  --bootstrap-server localhost:9092
```

### Детали Consumer Group:

```bash
docker exec kafka1 kafka-consumer-groups --describe \
  --group matematika-service-group \
  --bootstrap-server localhost:9092
```

---

## 🔒 Настройки безопасности и надежности

### 1. Репликация:

```yaml
KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 2
KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 2
```

- Все служебные топики Kafka реплицируются на 2 брокера
- Защита от потери данных при падении брокера

### 2. Min In-Sync Replicas (ISR):

```yaml
KAFKA_MIN_INSYNC_REPLICAS: 1
min.insync.replicas=1  # в топиках
```

- Минимум 1 реплика должна быть синхронизирована
- Producer может писать, если хотя бы 1 реплика живая

### 3. Acks уровень:

```go
// В Producer конфигурации
RequiredAcks: sarama.WaitForAll  // Ждем подтверждения от всех ISR реплик
```

- Максимальная надежность доставки
- Запись считается успешной только когда все ISR реплики подтвердили

### 4. Idempotent Producer:

```go
IdempotentWrites: true
MaxOpenRequests: 1
```

- Защита от дубликатов при retry
- Гарантия порядка сообщений

---

## 📊 Сценарии отказоустойчивости

### Сценарий 1: Падение Broker 1 (kafka1)

**Что происходит:**
1. Zookeeper обнаруживает отказ kafka1
2. Kafka2 становится leader для партиций 0, 2, 4
3. Producer и Consumer автоматически переключаются на kafka2
4. Данные не теряются (есть реплики на kafka2)
5. Система продолжает работать с половинной производительностью

**Восстановление:**
1. Запустить kafka1: `docker compose up -d kafka1`
2. Kafka автоматически синхронизирует данные
3. Rebalancing - партиции перераспределяются

### Сценарий 2: Сетевая задержка между брокерами

**Что происходит:**
1. Реплики выходят из ISR (In-Sync Replicas)
2. Producer продолжает писать в leader партиции
3. После восстановления связи - автоматическая синхронизация
4. Реплики возвращаются в ISR

### Сценарий 3: Добавление третьего брокера

**Шаги:**
1. Добавить kafka3 в docker-compose.yml (broker_id=3)
2. Запустить: `docker compose up -d kafka3`
3. Перераспределить партиции:
   ```bash
   kafka-reassign-partitions --zookeeper zookeeper:2181 \
     --reassignment-json-file reassignment.json \
     --execute
   ```

---

## 🎯 Best Practices

### 1. Количество партиций:

```
Партиций ≥ Максимальное количество consumer'ов в группе
```

- statement.calculation.completed: 6 партиций → до 6 параллельных consumer'ов
- Больше партиций = больше параллелизм, но больше overhead

### 2. Replication Factor:

```
Replication Factor = min(количество брокеров, 3)
```

- 2 брокера → replication=2
- 3+ брокеров → replication=3 (оптимально)

### 3. Min ISR:

```
Min ISR = Replication Factor - 1
```

- Replication=2 → Min ISR=1
- Позволяет писать при падении 1 брокера

### 4. Retention:

- **Частые операции:** 1-7 дней
- **Аудит:** 30+ дней
- **Критичные данные:** unlimited (архивация в S3/MinIO)

---

## 🚨 Troubleshooting

### Проблема: Producer не может подключиться

**Проверка:**
```bash
docker logs kafka1 | grep ERROR
docker logs kafka2 | grep ERROR
```

**Решение:**
- Проверить что оба брокера healthy: `docker compose ps`
- Проверить сеть: `docker network inspect business_bank_network`
- Проверить KAFKA_BROKERS в matematika: `kafka1:9092,kafka2:9093`

### Проблема: Партиции не распределяются

**Проверка:**
```bash
docker exec kafka1 kafka-topics --describe \
  --topic statement.calculation.completed \
  --bootstrap-server localhost:9092
```

**Решение:**
- Убедиться что оба брокера в ISR
- Перезапустить kafka-init: `docker compose restart kafka-init`

### Проблема: Lag растет

**Проверка:**
```bash
docker exec kafka1 kafka-consumer-groups --describe \
  --group matematika-service-group \
  --bootstrap-server localhost:9092
```

**Решение:**
- Добавить больше consumer'ов (scale up matematika service)
- Оптимизировать обработку сообщений
- Увеличить batch size

---

## 📚 Дополнительные ресурсы

- **Kafdrop UI:** http://localhost:9000
- **Init Script:** `infrastructure/kafka/init-topics.sh`
- **Docker Compose:** `docker-compose.yml` (строки 91-239)
- **README Kafka Integration:** `README.md` (строки 1386-2111)

---

**Kafka Cluster полностью настроен и готов к production! 🚀**

