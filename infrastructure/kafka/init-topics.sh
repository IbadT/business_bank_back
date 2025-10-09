#!/bin/bash

# Скрипт инициализации топиков Kafka с правильным распределением партиций
# Запускается после старта Kafka кластера

echo "=========================================="
echo "🚀 ИНИЦИАЛИЗАЦИЯ KAFKA ТОПИКОВ"
echo "=========================================="

# Ждем пока оба брокера запустятся
echo "⏳ Ожидание готовности Kafka кластера..."
sleep 15

# Kafka brokers
KAFKA_BROKERS="kafka1:9092,kafka2:9093"

echo ""
echo "📋 Создаем топики с распределением партиций..."
echo ""

# 1. statement.calculation.completed - результаты расчетов
# 6 партиций, 2 реплики (по 3 партиции на каждом брокере)
kafka-topics --create \
  --bootstrap-server $KAFKA_BROKERS \
  --topic statement.calculation.completed \
  --partitions 6 \
  --replication-factor 2 \
  --config retention.ms=604800000 \
  --config compression.type=snappy \
  --config min.insync.replicas=1 \
  --if-not-exists

echo "✅ Топик: statement.calculation.completed (6 партиций, 2 реплики)"

# 2. statement.generation.request - запросы на генерацию
# 4 партиции, 2 реплики
kafka-topics --create \
  --bootstrap-server $KAFKA_BROKERS \
  --topic statement.generation.request \
  --partitions 4 \
  --replication-factor 2 \
  --config retention.ms=86400000 \
  --config compression.type=snappy \
  --config min.insync.replicas=1 \
  --if-not-exists

echo "✅ Топик: statement.generation.request (4 партиции, 2 реплики)"

# 3. statement.formatting.completed - готовые выписки
# 4 партиции, 2 реплики
kafka-topics --create \
  --bootstrap-server $KAFKA_BROKERS \
  --topic statement.formatting.completed \
  --partitions 4 \
  --replication-factor 2 \
  --config retention.ms=604800000 \
  --config compression.type=snappy \
  --config min.insync.replicas=1 \
  --if-not-exists

echo "✅ Топик: statement.formatting.completed (4 партиции, 2 реплики)"

# 4. statement.error - ошибки обработки
# 2 партиции, 2 реплики (меньше нагрузки)
kafka-topics --create \
  --bootstrap-server $KAFKA_BROKERS \
  --topic statement.error \
  --partitions 2 \
  --replication-factor 2 \
  --config retention.ms=2592000000 \
  --config compression.type=snappy \
  --config min.insync.replicas=1 \
  --if-not-exists

echo "✅ Топик: statement.error (2 партиции, 2 реплики)"

echo ""
echo "=========================================="
echo "📊 ИНФОРМАЦИЯ О ТОПИКАХ"
echo "=========================================="
echo ""

# Выводим информацию о созданных топиках
kafka-topics --list --bootstrap-server $KAFKA_BROKERS

echo ""
echo "=========================================="
echo "📋 ДЕТАЛИ ТОПИКОВ"
echo "=========================================="
echo ""

# Детальная информация о каждом топике
kafka-topics --describe --bootstrap-server $KAFKA_BROKERS

echo ""
echo "=========================================="
echo "✅ ИНИЦИАЛИЗАЦИЯ ЗАВЕРШЕНА"
echo "=========================================="
echo ""
echo "📍 Распределение партиций:"
echo "   • statement.calculation.completed: 6 партиций × 2 реплики = 12 партиций на кластер"
echo "   • statement.generation.request:    4 партиции × 2 реплики = 8 партиций на кластер"
echo "   • statement.formatting.completed:  4 партиции × 2 реплики = 8 партиций на кластер"
echo "   • statement.error:                 2 партиции × 2 реплики = 4 партиций на кластер"
echo ""
echo "🔧 Kafka Cluster:"
echo "   • Broker 1 (kafka1:9092) - ID: 1"
echo "   • Broker 2 (kafka2:9093) - ID: 2"
echo ""
echo "🎯 Партиции равномерно распределены между брокерами"
echo "   для горизонтального масштабирования и отказоустойчивости"
echo ""

