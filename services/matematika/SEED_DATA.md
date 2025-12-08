# Mock Data для Matematika Service

## Описание

Файл `internal/database/seeds.go` содержит моковые данные для тестирования и разработки Matematika сервиса.

## Структура данных

### 5 Компаний с реалистичными выписками:

#### 1. **Srb Autos LLC** (B2C - Automotive)
- Account: `201290125551`
- Card: `2091222000102910`
- Периоды: January, February 2025
- Особенности:
  - B2C модель (пятничные поступления от шлюза)
  - Custom customers: Super LLC, Lulu Inc.
  - Custom contractors: Jakson Sam CPA (бухгалтер), LumNuft Inc (топливо)
  - Прибыль: ~63k январь, ~41k февраль

#### 2. **TechCorp Industries** (B2B - Technology)
- Account: `301892345678`
- Card: `4532123456789012`
- Периоды: January 2025
- Особенности:
  - B2B модель (множественные клиенты)
  - Custom customers: GlobalTech Solutions, DataStream Corp, CloudNine Systems
  - Custom contractors: DevSquad LLC (IT-dev), FinPro Accounting (бухгалтер)
  - Прибыль: ~68k

#### 3. **FastFood LLC** (B2C - Restaurant)
- Account: `402156789012`
- Card: `5412987654321098`
- Периоды: January, February, March 2025
- Особенности:
  - B2C ресторанная модель
  - Низкая прибыль (~8k) - типично для ресторанов
  - Высокие расходы на продукты и персонал

#### 4. **Construction LLC** (B2B - Construction)
- Account: `503789456123`
- Периоды: January 2025
- Особенности:
  - B2B строительная компания
  - Крупные транзакции (project payments)
  - Custom customers: BuildRight Corp, HomeConstruct Inc

#### 5. **RetailStore Inc** (B2C - Retail)
- Account: `604567890123`
- Периоды: January 2025
- Особенности:
  - B2C розничная торговля
  - Расходы на inventory (закупка товаров)

## Использование

### Запуск seeding

```bash
# Из корня проекта
make seed-matematika

# Или напрямую
cd services/matematika
go run cmd/seed/main.go
```

### Что происходит:

1. **Очистка старых mock данных** - удаляет записи с `account_id LIKE 'MOCK_%'`
2. **Базовый seed** - создает 3 основные компании (8 выписок)
3. **Расширенный seed** - добавляет еще 2 компании

### Итого после seeding:

```
5 компаний × 8 statements = ~200+ транзакций
```

## Примеры запросов после seeding

### Получить выписку Srb Autos (январь):
```bash
curl http://localhost/api/matematika/statement/stmt_2025-01_201290125551/result
```

### Получить статус:
```bash
curl http://localhost/api/matematika/statement/stmt_2025-01_201290125551/status
```

### Получить выписку TechCorp:
```bash
curl http://localhost/api/matematika/statement/stmt_2025-01_301892345678/result
```

### Получить все выписки FastFood (3 месяца):
```bash
curl http://localhost/api/matematika/statement/stmt_2025-01_402156789012/result
curl http://localhost/api/matematika/statement/stmt_2025-02_402156789012/result
curl http://localhost/api/matematika/statement/stmt_2025-03_402156789012/result
```

## Структура данных в БД

### Таблица: `statements`

| ID | AccountID | Month | Status | InitialBalance | FinalBalance | TotalRevenue | TotalExpenses | NetProfit |
|---|---|---|---|---|---|---|---|---|
| stmt_2025-01_201290125551 | 201290125551 | 2025-01 | completed | 100000.00 | 163149.16 | 100000.00 | -36850.84 | 63149.16 |
| stmt_2025-02_201290125551 | 201290125551 | 2025-02 | completed | 163149.16 | 119569.16 | 100000.00 | -58762.25 | 41237.75 |
| ... | ... | ... | ... | ... | ... | ... | ... | ... |

Полный JSON каждой выписки хранится в поле `data` (JSONB).

## Добавление собственных mock данных

Редактируй `internal/database/seeds.go`:

```go
func seedMyCompany(ctx context.Context, repo calculation.CalculationRepository) error {
    accountNumber := "YOUR_ACCOUNT_NUMBER"
    
    data := calculation.MatematikaResponse{
        "JANUARY 2025": calculation.MonthlyStatement{
            FinancialSummary: calculation.FinancialSummary{
                CompanyName:    "My Company Inc",
                AccountNumber:  accountNumber,
                Period:         "2025-01-01 - 2025-01-31",
                InitialBalance: 100000.00,
                FinalBalance:   120000.00,
                TotalRevenue:   50000.00,
                TotalExpenses:  -30000.00,
                NetProfit:      20000.00,
            },
            Transactions: []calculation.TransactionResponse{
                // Твои транзакции
            },
            ForwardingInfo: calculation.ForwardingInfo{
                AssociatedCard: "1234567890123456",
                OwnerName:      "Your Name",
            },
            DailyClosingBalances: generateDailyBalances("2025-01", 100000.00, 120000.00, 31),
        },
    }
    
    return saveStatement(ctx, repo, "stmt_2025-01_"+accountNumber, data)
}
```

Затем добавь вызов в `SeedDatabase()`:

```go
if err := seedMyCompany(ctx, repo); err != nil {
    return err
}
```

## Очистка данных

Seed автоматически очищает старые mock данные перед добавлением новых. Если нужно вручную очистить:

```sql
DELETE FROM statements WHERE account_id LIKE 'MOCK_%';
DELETE FROM statements WHERE account_id IN ('201290125551', '301892345678', '402156789012', '503789456123', '604567890123');
```

## Production замечания

⚠️ **Важно:** Эти данные только для development/testing!

- Не использовать в production
- Account numbers вымышленные
- Card numbers сгенерированы случайно
- Все компании и имена фиктивные

