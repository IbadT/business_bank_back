Микросервис Matematika принимает на вход исходные финансовые параметры и формирует список транзакций с расчётом сумм и балансов. Его функциональность включает следующие аспекты:
•	Распределение доходов и расходов по категориям. На основе заданных финансовых параметров (оборот, желаемая прибыль и пр.) сервис генерирует перечень входящих и исходящих платежей. По умолчанию закладываются рамки: чистая прибыль компании около 6–9% от оборота[1], количество входящих транзакций ~4 (редко 5) для B2C-модели или 10–20 для B2B[2], исходящих ~45 ± 10 операций[3]. Общее число транзакций в месяце находится в диапазоне 39–75[4]. Эти диапазоны могут масштабироваться при необходимости (см. ниже). Сервис гарантирует, что суммарные доходы = 100% оборота, а распределение расходов соответствует целевой прибыли (оборот минус расходы = прибыль)[5][6].
•	Процентные и фиксированные транзакции. Matematika различает две категории финансовых операций:
•	Процентные операции – суммы вычисляются как процент от оборота. Например, в расходах Payroll суммарно ~27–27.5% от оборота за месяц (разбито на две выплаты)[7][8], топливо – ~15–17.5% оборота (разбито на 7–9 транзакций)[9][10], маркетинг – ~0.5–0.7% оборота[11][12] и т.д. Процентные доли заданы для большинства категорий и обеспечивают правдоподобное распределение расходов.
•	Фиксированные операции – суммы определены заранее или вычисляются по правилу, независимо от общего оборота. Например, подписки ПО имеют фиксированную цену из конфигурации (одна транзакция в месяц)[13][14], мобильная связь и коммунальные услуги – фиксируются в первом месяце в диапазоне \$200–500 и далее меняются ±15% от этой базы[15][16], оплата платных дорог – фиксированные значения \$20/\$35/\$50 за транзакцию[17][18], лизинг (Leasing) – сумма первого месяца ~11.5–12% оборота фиксируется и повторяется 1:1 в последующих месяцах[19]. Также особые категории рассчитываются по формулам: например, доплата за перегруз – небольшая сумма, полученная как вес (200–1000 lb) * ставку (\$0.011–\$0.039)[20][21]. Matematika вычисляет такие суммы и сохраняет детали расчёта (например, вес, ставка) в поле calculationDetails в JSON.
•	Расписание транзакций и учёт праздников. Сервис генерирует даты и времена операций согласно заданным шаблонам:
•	Для каждого типа расхода задана частота и предпочтительный день. Например, перевод владельцу – 1 раз в месяц, в будний день (не праздничный)[22], IRS-налоги – 1 раз в обычные месяцы и 2 раза в квартальные месяцы (апр, июнь, сен, янв) 15-го числа[23][24], Payroll – 2 раза: во 2-ю и 4-ю пятницу[7][8], топливо – 7–9 раз в месяц по будним дням[9][10], подписка ПО – в случайно выбранный день недели, который затем сохраняется для следующих месяцев[25][14], мобильная связь – 2-я пятница месяца[26][27], коммунальные – 3-я пятница[28][29], лизинг – последняя (4-я или 5-я) пятница месяца[30][31], и т.д. Matematika должна соблюдать эти шаблоны дат.
•	Праздники: операции по счёту (ACH, wire, internal transfers) и B2B-пополнения не проводятся в праздничные дни[32]. Если выплата должна была произойти в праздник, сервис переносит её на ближайший следующий рабочий день[32]. Список праздничных дней берётся из конфигурации (например, федеральные праздники США) и учитывается при планировании каждой транзакции.
•	Рабочее время: время транзакций генерируется в допустимых интервалах. Все транзакции по счёту происходят в рабочие часы 08:00–18:00, операции по банковским картам – с 09:00 до 20:00, а оплата подписок по карте специально ставится в 00:01 (полночь)[33]. Это правило обеспечивает реалистичность временных меток (например, покупки по карте в ночное время не происходят, кроме автоматических списаний подписок).
•	Учёт пятой недели (пятницы) в месяц. В некоторых месяцах бывает 5 недельных дней вместо стандартных 4 (например, 5 пятниц в августе 2025). Matematika должна это учитывать: если категория транзакций привязана к конкретному дню недели, и в данном месяце таких дней 5, то заранее рассчитанный «банк» суммы на эти операции делится не на 4 части, а на 5, с соответствующей корректировкой сумм[34]. Например, для B2C-модели пополнений 4 пятничных поступления делят 100% выручки примерно на четыре части по ~25% каждая; но если пятниц 5, будет пять поступлений по ~20% каждая, скорректированные случайным образом[34]. Сервис автоматизировано увеличивает количество транзакций и уменьшает каждую сумму, чтобы суммарно остаться в пределах 100% оборота.
•	Генерация поступлений (доходов). При моделировании входящих платежей Matematika опирается на два сценария:
•	B2C (Business-to-Consumer): Предполагается 4 (иногда 5) поступления в месяц от платёжного шлюза (например, Stripe) – каждую пятницу[35]. При генерации первого месяца случайно выбирается один шлюз из списка (gateways.csv), и во всех выписках далее используется он же[35]. Каждое поступление ~25% ± 4.5% от месячного оборота, суммарно 100% оборота[35]. Комиссия шлюза уже вычтена из этих сумм[5]. То есть Matematika формирует четыре транзакции типа income с категорией "Пополнение шлюз" (или аналогичным названием шлюза) равными долями оборота.
•	B2B (Business-to-Business): Поступления распределяются по нескольким категориям-клиентам. Например, retails, wholesale, agriculture, factoring_avance – каждая категория имеет 2–8 платежей в месяц, каждый на 5.5–8.5% от оборота[36][37]. Всего в сумме 10–20 входящих платежей[6]. Приблизительно 70% ± 10% из них должны быть ACH Credit, остальные ~30% – Electronic Payment (имитация разнообразия платёжных методов)[6]. Matematika выбирает случайное количество платежей в указанном диапазоне для каждой категории и вычисляет суммы в заданных процентах. Сумма всех доходов = 100% оборота (т.е. Matematika гарантирует что оборот полностью покрыт распределением входящих)[6]. Важно: Если пользователь через входные данные указал своих клиент-контрагентов, они должны заменить дефолтные названия из конфигурации[6] (подробнее см. раздел Maska).
•	Масштабирование количества операций. Сервис учитывает влияния пользовательских настроек на количество транзакций:
•	При добавлении ручных операций пользователем (см. CustomData во входном JSON) общее количество и суммы автоматических транзакций подстраиваются, чтобы сохранить рамки распределения[38]. Например, если пользователь вручную добавил 2 дополнительных расхода, Matematika может слегка уменьшить суммы других расходов или убрать некоторые мелкие транзакции (особенно из категорий, помеченных как необязательные – см. ниже), чтобы не выйти за пределы ~39–75 транзакций в сумме и не занизить прибыль чрезмерно.
•	Сервис предоставляет возможность масштабирования количества операций по коэффициенту. Например, если требуется сгенерировать вдвое более детализированную выписку, можно задать коэффициент 2 – тогда целевой диапазон операций станет 78–150 вместо 39–75, и Matematika удвоит количество транзакций в каждой категории (кроме тех, у которых фиксировано 1 в месяц, а также не удваивает выплаты Payroll ADP)[38]. При удвоении числа операций их суммы пропорционально уменьшаются, чтобы общие проценты по категориям сохранились[38].
•	Некоторые категории расходов помечены как опциональные (низкоприоритетные) – например, маркетинг, клининг (уборка), охрана, бухгалтер, юрист, IT-sec, IT-dev[12][39]. В случае необходимости эти категории могут быть полностью или частично урезаны (не сгенерированы) для соблюдения желаемой чистой прибыли или ограничения числа транзакций. Если пользователь указал повышенную цель по прибыли (меньший процент расходов), Matematika сначала сокращает расходы «красных» категорий (опциональных) до нуля[12][40], прежде чем уменьшать основные расходы. При генерации разных месяцев набор вырезаемых категорий может варьироваться для реалистичности (случайным образом из списка опциональных)[41][40].
•	Нормализация и балансировка итогов. После расчёта всех сумм Matematika выполняет округление и выравнивание балансов:
•	Округление до центов: Все суммы транзакций округляются до двух знаков после запятой (центов)[42]. Например, промежуточный результат \$11.809102 превращается в \$11.81[42].
•	Нормализация: Поскольку округления могут вводить погрешность, сервис в конце перераспределяет несколько центов между транзакциями (уравновешивает итог)[42]. Это необходимо, чтобы точно соблюсти заданные процентные доли: сумма всех сгенерированных доходов минус сумма всех расходов ровно равнялась рассчитанной прибыли. Если после округления обнаруживается избыточный или недостающий цент, Matematika корректирует одну из транзакций (например, слегка увеличивает или уменьшает последний платеж) так, чтобы баланс сошёлся.
•	Недопустимость отрицательного баланса: Matematika следит, чтобы на всём протяжении месяца баланс счёта не уходил в минус[43]. Начальный баланс задаётся пользователем или вычисляется автоматически; если в процессе распределения выясняется, что расходов слишком много (или начальный баланс слишком мал) и после списания баланс стал бы отрицательным, это рассматривается как ошибка (см. Валидация). В корректных сценариях сервис ограничивает суммы расходов в рамках доступного баланса на дату транзакции. Например, если на счёте недостаточно средств к какому-то дню, Matematika может перенести запланированный расход на день позже, после поступления, или уменьшить сумму транзакции.
•	Выходные данные Matematika. Результатом работы сервиса является JSON, содержащий рассчитанные транзакции и сопутствующую информацию:
•	Список транзакций месяца (массив объектов). Каждая транзакция включает поля: transactionId (уникальный ID), даты (transactionDate – дата и время совершения, postingDate – дата проводки в выписке), type (income или expense), category (категория операции), method (метод платежа: напр. ACH_CREDIT, Electronic Payment, card, account и т.п.), amount (сумма, положительная для доходов и отрицательная для расходов), balanceAfter (баланс после выполнения операции), а также флаги технической информации – например, isManual (пометка, что транзакция добавлена вручную пользователем) и FixAsFirst (используется для сортировки или пометки первой операции определённого типа). Пример фрагмента JSON с транзакцией от Matematika:
 	json { "transactionId": "t_inc_003", "transactionDate": "2025-01-06T11:00:00", "postingDate": "2025-01-06", "type": "income", "category": "retails_ca.csv", "method": "ACH_CREDIT", "amount": 8300.00, "isManual": false, "balanceAfter": 129149.12 }[44]
•	Сводные финансовые показатели месяца: начальный баланс, конечный баланс, итого доходов, итого расходов и чистая прибыль за месяц. Эти данные включаются в JSON (например, в объекте financialSummary или аналогичном). В примере за январь 2025: initialBalance: 100000.00, finalBalance: 163149.16, totalRevenue: 100000.00, totalExpenses: -36850.84, netProfit: 63149.16[45].
•	Разбивка доходов и расходов по методам/видам (для справки): например, сколько всего пришло ACH-платежами, Wire, Zelle, через шлюз; и сколько ушло по карте vs со счёта[46][47]. Эта информация необязательна, но может включаться для контроля (см. поля revenueBreakdown, expensesBreakdown в JSON).
•	Количество транзакций: общее число, в том числе подмножества (сколько депозитов и сколькими методами, сколько списаний и по каким каналам)[48][49].
•	Ежедневные итоговые балансы: массив dailyClosingBalances с балансом на конец каждого дня (например, 23:59 каждого дня месяца)[50]. Это позволяет восстановить дневное сальдо.
•	Данные для маскировки (forwardingInfo): Matematika передаёт в Maska некоторые данные, необходимые для генерации текстовых описаний. В JSON присутствует объект, содержащий:
o	associatedCard – номер банковской карты, привязанной к счёту (16 цифр, генерируется, если не задано). Последние 4 цифры этой карты используются Maska при генерации строк для транзакций по карте (подставляются вместо XXXX XXXX XXXX 9017 и т.п. в шаблонах)[51][52].
o	ownerName – имя владельца счёта (из companyInfo входных данных). Может использоваться, например, в шаблонах payroll (ADP) или других.
o	companyName – название фирмы (дублируется из входных данных, нужно для шаблонов – в INDN полях ACH, в описаниях налогов и т.д.).
o	customCustomers – список пользовательских названий клиентов (контрагентов по доходам), если они заданы. Эти имена должны заменить соответствующих дефолтных клиентов в шаблонах B2B-пополнений[6].
o	customContractors – список пользовательских подрядчиков по расходам. Каждый элемент связывает transactionType (категорию/тип операции, например "Бухгалтер", "Топливо") с заданным именем name (например, "Jakson Sam CPA" для бухгалтера)[53]. Maska использует эту карту для подстановки: транзакции категории "Бухгалтер" будут отражены не как дефолтный контрагент из базы, а под именем "Jakson Sam CPA" и с соответствующим шаблоном.
 	Пример фрагмента forwardingInfo из JSON Matematika:
 	json "forwardingInfo": { "associatedCard": "2091222000102910", "ownerName": "John Doe", "customCustomers": [ "Super LLC", "Lulu Inc." ], "customContractors": [ { "transactionType": "Бухгалтер", "name": "Jakson Sam CPA" }, { "transactionType": "Топливо", "name": "LumNuft Inc" } ] }[54][53]
Таким образом, Matematika генерирует полный «скелет» выписки за месяц – все транзакции с точными суммами и метаданными, но без «человеко-читаемых» описаний. Далее эти данные поступают на вход микросервиса Maska для маскировки.

---

## ТЕХНИЧЕСКАЯ РЕАЛИЗАЦИЯ

### Архитектура кода

```
internal/
  calculation/
    handler.go          - HTTP endpoints (уже есть)
    service.go          - Основная бизнес-логика (требует полной реализации)
    repository.go       - Работа с БД (требует реализации)
    models.go           - Request/Response DTO (✅ готово)
    types.go            - Enums (✅ готово)
    orm.go              - Database entities (требует расширения)
    validator.go        - Валидация входных данных (создать)
    
  generator/            - СОЗДАТЬ - Генераторы транзакций
    income.go           - Генерация доходов (B2C/B2B)
    expenses.go         - Генерация расходов
    categories.go       - Конфигурация категорий расходов
    scaling.go          - Масштабирование и оптимизация
    
  scheduler/            - СОЗДАТЬ - Работа с датами и временем
    date_generator.go   - Генерация дат по правилам
    holiday.go          - Работа с праздниками
    time_generator.go   - Генерация времени транзакций
    
  normalizer/           - СОЗДАТЬ - Нормализация и балансировка
    balance.go          - Расчет балансов
    rounding.go         - Округление и распределение центов
    validator.go        - Проверка отсутствия отрицательного баланса
    
  clients/              - СОЗДАТЬ - HTTP клиенты для внешних сервисов
    shared_client.go    - Клиент для Shared Service
    
  business-rules/       - УЖЕ ЕСТЬ (но пустые файлы)
    constraints.go      - Бизнес-ограничения
    holidays.go         - Логика праздников
    rules.go            - Бизнес-правила
    validation.go       - Бизнес-валидация
```

---

### 1. РАСШИРЕНИЕ REQUEST МОДЕЛИ

**Файл:** `internal/calculation/models.go`

**Добавить в GenerateStatementRequest:**

```go
type GenerateStatementRequest struct {
    // ✅ Уже есть
    AccountID      string       `json:"account_id" validate:"required"`
    Month          string       `json:"month" validate:"required,datetime=2006-01"`
    BusinessType   BusinessType `json:"business_type" validate:"required"`
    InitialBalance float64      `json:"initial_balance" validate:"required,gte=0"`
    
    // ❌ ДОБАВИТЬ - Информация о компании
    CompanyInfo CompanyInfo `json:"company_info" validate:"required"`
    
    // ❌ ДОБАВИТЬ - Кастомные контрагенты
    CustomCustomers   []string           `json:"custom_customers,omitempty"`
    CustomContractors []CustomContractor `json:"custom_contractors,omitempty"`
    
    // ❌ ДОБАВИТЬ - Ручные транзакции пользователя
    ManualTransactions []ManualTransaction `json:"manual_transactions,omitempty"`
    
    // ❌ ДОБАВИТЬ - Настройки генерации
    TargetNetProfitPercent *float64 `json:"target_net_profit_percent,omitempty"` // 6-9% по умолчанию
    ScalingCoefficient     *float64 `json:"scaling_coefficient,omitempty"`       // 1.0 по умолчанию
    
    // ❌ ДОБАВИТЬ - Номер карты (опционально, генерируется если не указан)
    AssociatedCard *string `json:"associated_card,omitempty" validate:"omitempty,len=16,numeric"`
}

type CompanyInfo struct {
    CompanyName   string `json:"company_name" validate:"required"`
    AccountNumber string `json:"account_number" validate:"required"`
    OwnerName     string `json:"owner_name" validate:"required"`
}

type ManualTransaction struct {
    Date        string              `json:"date" validate:"required"`
    Type        TransactionType     `json:"type" validate:"required"`
    Category    string              `json:"category" validate:"required"`
    Method      TransactionMethod   `json:"method" validate:"required"`
    Amount      float64             `json:"amount" validate:"required"`
    Description string              `json:"description,omitempty"`
}
```

---

### 2. ГЕНЕРАТОР ДОХОДОВ (B2C/B2B)

**Файл:** `internal/generator/income.go` (СОЗДАТЬ)

**Функции для реализации:**

```go
package generator

import (
    "time"
)

// IncomeGenerator - генератор доходов
type IncomeGenerator struct {
    sharedClient clients.SharedServiceClient
    dateGen      *scheduler.DateGenerator
}

// GenerateB2CIncome - генерация доходов для B2C модели
// - 4-5 поступлений каждую пятницу
// - Каждое ~25% ± 4.5% от revenue
// - Метод: случайный шлюз из gateways.csv (один на все месяцы)
func (g *IncomeGenerator) GenerateB2CIncome(revenue float64, month time.Time) ([]Transaction, error) {
    // TODO:
    // 1. Найти все пятницы месяца (учесть 5-ю неделю!)
    // 2. GET запрос к Shared: /api/shared/config/patterns?category=gateways
    // 3. Выбрать случайный шлюз (сохранить в state для следующих месяцев)
    // 4. Разделить revenue на количество пятниц
    // 5. Добавить случайную вариацию ±4.5%
    // 6. Нормализовать суммы чтобы итого = revenue
    // 7. Для каждой транзакции:
    //    - type: income
    //    - category: название шлюза
    //    - method: "account" или "gateway"
    //    - date: пятница + случайное время 10:00-16:00
}

// GenerateB2BIncome - генерация доходов для B2B модели
// - 10-20 поступлений по категориям (retails, wholesale, agriculture, factoring)
// - Каждая категория: 2-8 платежей по 5.5-8.5% от revenue
// - 70% ACH Credit, 30% Electronic Payment
func (g *IncomeGenerator) GenerateB2BIncome(revenue float64, month time.Time, customCustomers []string) ([]Transaction, error) {
    // TODO:
    // 1. GET запрос к Shared: /api/shared/config/patterns?category=b2b_clients
    // 2. Для каждой категории (retails, wholesale, agriculture, factoring_avance):
    //    a. Случайное количество транзакций: 2-8
    //    b. Каждая транзакция: случайный процент 5.5-8.5% от revenue
    // 3. Распределить методы: 70% ACH_CREDIT, 30% Electronic Payment
    // 4. Заменить дефолтные названия на customCustomers (если указаны)
    // 5. Генерировать даты в рабочие дни (не праздники!)
    // 6. Нормализовать суммы чтобы итого = revenue
}
```

---

### 3. ГЕНЕРАТОР РАСХОДОВ

**Файл:** `internal/generator/expenses.go` (СОЗДАТЬ)

**Конфигурация категорий:**

```go
// ExpenseCategory - конфигурация категории расходов
type ExpenseCategory struct {
    Name              string          // "Payroll ADP", "Топливо", "Маркетинг"
    Type              ExpenseType     // Percentage или Fixed
    PercentRange      [2]float64      // [27.0, 27.5] для Payroll
    FixedRange        [2]float64      // [200, 500] для мобильной связи
    FrequencyPerMonth int             // 2 для Payroll, 7-9 для топлива
    DayRule           string          // "2nd_friday", "last_friday", "15th_day", "random_weekday"
    TimeRange         [2]int          // [8, 18] для часов
    PaymentMethod     TransactionMethod
    Optional          bool            // true для маркетинга, клининга, юриста
    FirstMonthOnly    bool            // true для фиксации базовой суммы (мобильная, коммунальные)
}

// ExpenseGenerator - генератор расходов
type ExpenseGenerator struct {
    categories   []ExpenseCategory
    dateGen      *scheduler.DateGenerator
    holidayCheck *scheduler.HolidayChecker
}

// GenerateExpenses - генерация всех расходов
func (g *ExpenseGenerator) GenerateExpenses(
    revenue float64,
    targetProfit float64,
    month time.Time,
    optional OptionalCategories,
) ([]Transaction, error) {
    // TODO:
    // 1. Загрузить конфигурацию категорий из файла/БД
    // 2. Для каждой категории:
    //    a. Если Percentage - вычислить сумму от revenue
    //    b. Если Fixed - взять/сгенерировать фиксированную сумму
    // 3. Сгенерировать даты согласно DayRule
    // 4. Проверить праздники, перенести на рабочие дни
    // 5. Сгенерировать время согласно TimeRange
    // 6. Если сумма расходов превышает (revenue - targetProfit):
    //    - Урезать Optional категории
    //    - Скорректировать суммы других категорий
}
```

**Примеры категорий (файл: `internal/generator/categories.go`):**

```go
var DefaultExpenseCategories = []ExpenseCategory{
    {
        Name:              "Payroll ADP",
        Type:              Percentage,
        PercentRange:      [2]float64{27.0, 27.5},
        FrequencyPerMonth: 2, // 2-я и 4-я пятница
        DayRule:           "2nd_4th_friday",
        TimeRange:         [2]int{17, 17}, // Всегда 17:00
        PaymentMethod:     TransactionMethodBankTransfer,
        Optional:          false,
    },
    {
        Name:              "Топливо / Fleet",
        Type:              Percentage,
        PercentRange:      [2]float64{15.0, 17.5},
        FrequencyPerMonth: 8, // случайное 7-9
        DayRule:           "random_weekday",
        TimeRange:         [2]int{9, 20},
        PaymentMethod:     "card",
        Optional:          false,
    },
    {
        Name:              "Маркетинг",
        Type:              Percentage,
        PercentRange:      [2]float64{0.5, 0.7},
        FrequencyPerMonth: 1,
        DayRule:           "random_weekday",
        PaymentMethod:     TransactionMethodBankTransfer,
        Optional:          true, // ❗ Можно урезать
    },
    {
        Name:              "Подписка ПО",
        Type:              Fixed,
        FixedRange:        [2]float64{150, 150}, // фиксированная
        FrequencyPerMonth: 1,
        DayRule:           "same_weekday", // сохраняется между месяцами
        TimeRange:         [2]int{0, 0},   // 00:01 (полночь)
        PaymentMethod:     "card",
        Optional:          false,
    },
    {
        Name:              "Мобильная связь",
        Type:              Fixed,
        FixedRange:        [2]float64{200, 500}, // первый месяц, потом ±15%
        FrequencyPerMonth: 1,
        DayRule:           "2nd_friday",
        TimeRange:         [2]int{10, 11},
        PaymentMethod:     "card",
        Optional:          false,
        FirstMonthOnly:    true, // фиксируется в первом месяце
    },
    {
        Name:              "Leasing",
        Type:              Percentage, // первый месяц
        PercentRange:      [2]float64{11.5, 12.0},
        FrequencyPerMonth: 1,
        DayRule:           "last_friday",
        TimeRange:         [2]int{16, 17},
        PaymentMethod:     TransactionMethodBankTransfer,
        Optional:          false,
        FirstMonthOnly:    true, // потом 1:1 повторяется
    },
    {
        Name:              "IRS-налоги",
        Type:              Percentage,
        PercentRange:      [2]float64{1.5, 2.0},
        FrequencyPerMonth: 1, // 2 раза в квартальные месяцы
        DayRule:           "15th_day",
        PaymentMethod:     TransactionMethodBankTransfer,
        Optional:          false,
    },
    // ... ещё ~10 категорий
}
```

---

### 4. РАБОТА С ДАТАМИ И ВРЕМЕНЕМ

**Файл:** `internal/scheduler/date_generator.go` (СОЗДАТЬ)

```go
package scheduler

import "time"

type DateGenerator struct {
    holidayChecker *HolidayChecker
}

// FindNthWeekday - найти N-ю пятницу/понедельник месяца
// Пример: FindNthWeekday(month, time.Friday, 2) → 2-я пятница
func (g *DateGenerator) FindNthWeekday(month time.Time, weekday time.Weekday, n int) time.Time {
    // TODO:
    // 1. Найти первый weekday месяца
    // 2. Добавить (n-1) недель
    // 3. Проверить что дата в пределах месяца
}

// FindAllWeekdays - найти все пятницы/понедельники месяца
// Возвращает slice из 4-5 дат
func (g *DateGenerator) FindAllWeekdays(month time.Time, weekday time.Weekday) []time.Time {
    // TODO:
    // 1. Итерироваться по всем дням месяца
    // 2. Собрать все даты с нужным weekday
}

// FindLastWeekday - найти последнюю пятницу месяца
// Может быть 4-я или 5-я неделя
func (g *DateGenerator) FindLastWeekday(month time.Time, weekday time.Weekday) time.Time {
    // TODO:
    // 1. Найти все weekday месяца
    // 2. Взять последний
}

// RandomWeekday - случайный рабочий день (не праздник!)
func (g *DateGenerator) RandomWeekday(month time.Time) time.Time {
    // TODO:
    // 1. Собрать все будние дни месяца (Пн-Пт)
    // 2. Исключить праздники
    // 3. Выбрать случайный
}

// NextWorkingDay - следующий рабочий день после date
func (g *DateGenerator) NextWorkingDay(date time.Time) time.Time {
    // TODO:
    // 1. Проверить date+1 день
    // 2. Если выходной или праздник → date+1
    // 3. Рекурсивно пока не найдем рабочий день
}
```

**Файл:** `internal/scheduler/time_generator.go` (СОЗДАТЬ)

```go
// GenerateTime - случайное время в диапазоне часов
func GenerateTime(date time.Time, hourRange [2]int) time.Time {
    // TODO:
    // 1. Случайный час между hourRange[0] и hourRange[1]
    // 2. Случайные минуты 0-59
    // 3. Случайные секунды 0-59
    // Возвращаем time.Time с датой date и сгенерированным временем
}
```

---

### 5. РАБОТА С ПРАЗДНИКАМИ

**Файл:** `internal/scheduler/holiday.go` (СОЗДАТЬ)

```go
package scheduler

import (
    "time"
    "context"
)

type HolidayChecker struct {
    client    clients.SharedServiceClient
    cache     map[string][]time.Time // кэш: "2025-01" -> [праздники]
}

// LoadHolidays - загрузить праздники из Shared Service
func (h *HolidayChecker) LoadHolidays(ctx context.Context, year int, month int) error {
    // TODO:
    // 1. GET запрос: shared:8081/api/shared/config/holidays/{year}/{month}
    // 2. Парсить ответ в []time.Time
    // 3. Сохранить в cache["2025-01"]
}

// IsHoliday - проверка является ли дата праздником
func (h *HolidayChecker) IsHoliday(date time.Time) bool {
    // TODO:
    // 1. Проверить cache для месяца date
    // 2. Если нет в cache → LoadHolidays()
    // 3. Проверить наличие date в списке праздников
}

// IsWorkingDay - проверка рабочего дня (не выходной и не праздник)
func (h *HolidayChecker) IsWorkingDay(date time.Time) bool {
    weekday := date.Weekday()
    return weekday != time.Saturday && 
           weekday != time.Sunday && 
           !h.IsHoliday(date)
}
```

---

### 6. НОРМАЛИЗАЦИЯ И БАЛАНСИРОВКА

**Файл:** `internal/normalizer/balance.go` (СОЗДАТЬ)

```go
package normalizer

type BalanceCalculator struct{}

// CalculateBalances - расчет balanceAfter для всех транзакций
func (b *BalanceCalculator) CalculateBalances(
    initialBalance float64,
    transactions []Transaction,
) ([]Transaction, error) {
    // TODO:
    // 1. Отсортировать transactions по дате
    // 2. Для каждой транзакции:
    //    balance = balance + amount (income) или balance - amount (expense)
    // 3. Проверить что balance >= 0 на каждом шаге
    // 4. Установить balanceAfter для каждой транзакции
}

// CalculateDailyClosingBalances - баланс на конец каждого дня
func (b *BalanceCalculator) CalculateDailyClosingBalances(
    month time.Time,
    transactions []Transaction,
) []DailyClosingBalance {
    // TODO:
    // 1. Сгруппировать транзакции по дням
    // 2. Для каждого дня месяца (1-31):
    //    - Найти последнюю транзакцию дня
    //    - Взять её balanceAfter
    //    - Если в день нет транзакций → balance предыдущего дня
}
```

**Файл:** `internal/normalizer/rounding.go` (СОЗДАТЬ)

```go
// RoundToCents - округление до 2 знаков после запятой
func RoundToCents(amount float64) float64 {
    return math.Round(amount*100) / 100
}

// NormalizeAmounts - нормализация сумм для точного соблюдения процентов
func NormalizeAmounts(transactions []Transaction, targetSum float64) []Transaction {
    // TODO:
    // 1. Округлить все суммы: RoundToCents()
    // 2. Вычислить actualSum = сумма всех округленных
    // 3. Погрешность = targetSum - actualSum
    // 4. Распределить погрешность по транзакциям:
    //    - Если погрешность +0.05 → добавить центы к 5 транзакциям
    //    - Если погрешность -0.03 → вычесть центы из 3 транзакций
    // 5. Пересчитать actualSum, проверить = targetSum
}
```

**Файл:** `internal/normalizer/validator.go` (СОЗДАТЬ)

```go
// ValidateNoNegativeBalance - проверка отсутствия отрицательного баланса
func ValidateNoNegativeBalance(transactions []Transaction) error {
    // TODO:
    // 1. Пройтись по всем транзакциям
    // 2. Найти минимальный balanceAfter
    // 3. Если минимум < 0 → return error с деталями (дата, сумма)
}

// ValidateRevenueDistribution - проверка что доходы = 100% revenue
func ValidateRevenueDistribution(transactions []Transaction, targetRevenue float64) error {
    // TODO:
    // 1. Сумма всех income транзакций
    // 2. Допустимая погрешность: ±0.01 (1 цент)
    // 3. Если разница > 0.01 → return error
}
```

---

### 7. КЛИЕНТ ДЛЯ SHARED SERVICE

**Файл:** `internal/clients/shared_client.go` (СОЗДАТЬ)

```go
package clients

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type SharedServiceClient struct {
    baseURL    string // "http://shared:8081"
    httpClient *http.Client
}

// GetHolidays - получить праздники за месяц
func (c *SharedServiceClient) GetHolidays(ctx context.Context, year int, month int) ([]Holiday, error) {
    url := fmt.Sprintf("%s/api/shared/config/holidays/%d/%d", c.baseURL, year, month)
    // TODO: HTTP GET запрос, парсинг JSON
}

// GetPatterns - получить шаблоны для бизнес-типа
func (c *SharedServiceClient) GetPatterns(ctx context.Context, businessType string) (*Patterns, error) {
    url := fmt.Sprintf("%s/api/shared/config/patterns?type=%s", c.baseURL, businessType)
    // TODO: HTTP GET запрос, парсинг JSON
}

// GetGateways - получить список платежных шлюзов
func (c *SharedServiceClient) GetGateways(ctx context.Context) ([]Gateway, error) {
    url := fmt.Sprintf("%s/api/shared/config/gateways", c.baseURL)
    // TODO: HTTP GET запрос, парсинг gateways.csv
}

// GetContractors - получить контрагентов по категории
func (c *SharedServiceClient) GetContractors(ctx context.Context, category string) ([]string, error) {
    url := fmt.Sprintf("%s/api/shared/config/contractors/%s", c.baseURL, category)
    // TODO: HTTP GET запрос
}
```

---

### 8. ОСНОВНОЙ SERVICE СЛОЙ

**Файл:** `internal/calculation/service.go` (ПЕРЕПИСАТЬ)

**Текущая заглушка заменить на:**

```go
func (s *calculationService) GenerateStatement(
    ctx context.Context,
    req *GenerateStatementRequest,
) (*GenerateStatementResponse, error) {
    
    // ШАГ 1: Валидация входных данных
    if err := s.validator.ValidateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // ШАГ 2: Генерация уникального ID
    statementID := generateStatementID(req.AccountID, req.Month)
    
    // ШАГ 3: Загрузка праздников из Shared Service
    year, month := parseMonth(req.Month)
    if err := s.holidayChecker.LoadHolidays(ctx, year, month); err != nil {
        return nil, fmt.Errorf("failed to load holidays: %w", err)
    }
    
    // ШАГ 4: Расчет целевых показателей
    targetRevenue := calculateRevenue(req) // если не задан явно
    targetProfit := calculateTargetProfit(targetRevenue, req.TargetNetProfitPercent)
    targetExpenses := targetRevenue - targetProfit
    
    // ШАГ 5: Генерация ДОХОДОВ
    var incomeTransactions []Transaction
    if req.BusinessType == BusinessTypeB2C {
        incomeTransactions, err = s.incomeGen.GenerateB2CIncome(targetRevenue, monthTime)
    } else {
        incomeTransactions, err = s.incomeGen.GenerateB2BIncome(targetRevenue, monthTime, req.CustomCustomers)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to generate income: %w", err)
    }
    
    // ШАГ 6: Генерация РАСХОДОВ
    expenseTransactions, err := s.expenseGen.GenerateExpenses(
        targetRevenue,
        targetProfit,
        monthTime,
        req.ManualTransactions,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to generate expenses: %w", err)
    }
    
    // ШАГ 7: Добавление ручных транзакций
    allTransactions := append(incomeTransactions, expenseTransactions...)
    allTransactions = append(allTransactions, req.ManualTransactions...)
    
    // ШАГ 8: Нормализация сумм
    allTransactions = s.normalizer.NormalizeAmounts(allTransactions, targetRevenue, targetExpenses)
    
    // ШАГ 9: Расчет балансов
    allTransactions, err = s.balanceCalc.CalculateBalances(req.InitialBalance, allTransactions)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate balances: %w", err)
    }
    
    // ШАГ 10: Валидация результата
    if err := s.normalizer.ValidateNoNegativeBalance(allTransactions); err != nil {
        return nil, fmt.Errorf("negative balance detected: %w", err)
    }
    
    // ШАГ 11: Формирование FinancialSummary
    summary := buildFinancialSummary(req, allTransactions)
    
    // ШАГ 12: Расчет dailyClosingBalances
    dailyBalances := s.balanceCalc.CalculateDailyClosingBalances(monthTime, allTransactions)
    
    // ШАГ 13: Формирование forwardingInfo
    forwardingInfo := buildForwardingInfo(req, allTransactions)
    
    // ШАГ 14: Формирование итогового JSON
    monthKey := formatMonthKey(req.Month) // "JANUARY 2025"
    result := MatematikaResponse{
        monthKey: MonthlyStatement{
            FinancialSummary:     summary,
            Transactions:         convertToTransactionResponse(allTransactions),
            ForwardingInfo:       forwardingInfo,
            DailyClosingBalances: dailyBalances,
            // Опционально
            Totals:            buildTotals(allTransactions),
            RevenueBreakdown:  buildRevenueBreakdown(allTransactions),
            ExpensesBreakdown: buildExpensesBreakdown(allTransactions),
            TransactionCounts: buildTransactionCounts(allTransactions),
        },
    }
    
    // ШАГ 15: Сохранение в БД
    if err := s.repo.SaveStatement(ctx, statementID, result); err != nil {
        return nil, fmt.Errorf("failed to save to database: %w", err)
    }
    
    // ШАГ 16: Публикация в Kafka (асинхронно)
    go s.publishToKafka(ctx, statementID, result)
    
    // ШАГ 17: Возврат ответа клиенту (202 Accepted)
    return &GenerateStatementResponse{
        StatementID: statementID,
        Status:      StatusPending,
        Message:     "Statement generation started",
    }, nil
}
```

---

### 9. REPOSITORY СЛОЙ

**Файл:** `internal/calculation/repository.go` (РЕАЛИЗОВАТЬ)

**Добавить методы:**

```go
type CalculationRepository interface {
    // Сохранение statement
    SaveStatement(ctx context.Context, id string, statement MatematikaResponse) error
    
    // Получение statement
    GetStatementByID(ctx context.Context, id string) (*MatematikaResponse, error)
    
    // Обновление статуса
    UpdateStatus(ctx context.Context, id string, status StatementStatus) error
    
    // Получение статуса
    GetStatus(ctx context.Context, id string) (StatementStatus, error)
    
    // Проверка существования
    Exists(ctx context.Context, id string) (bool, error)
}
```

---

### 10. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ

**Файл:** `internal/calculation/validator.go` (СОЗДАТЬ)

```go
package calculation

import (
    "fmt"
    "time"
    "regexp"
)

type RequestValidator struct{}

func (v *RequestValidator) ValidateRequest(req *GenerateStatementRequest) error {
    // 1. Валидация месяца
    if err := v.validateMonth(req.Month); err != nil {
        return err
    }
    
    // 2. Валидация номера карты (если указан)
    if req.AssociatedCard != nil {
        if err := v.validateCardNumber(*req.AssociatedCard); err != nil {
            return err
        }
    }
    
    // 3. Валидация scalingCoefficient
    if req.ScalingCoefficient != nil && *req.ScalingCoefficient <= 0 {
        return fmt.Errorf("scaling coefficient must be positive")
    }
    
    // 4. Валидация ручных транзакций
    for _, tx := range req.ManualTransactions {
        if err := v.validateManualTransaction(tx, req.Month); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *RequestValidator) validateMonth(month string) error {
    // Формат: YYYY-MM
    regex := regexp.MustCompile(`^\d{4}-\d{2}$`)
    if !regex.MatchString(month) {
        return fmt.Errorf("invalid month format, expected YYYY-MM")
    }
    
    // Проверка что месяц не в будущем
    t, err := time.Parse("2006-01", month)
    if err != nil {
        return err
    }
    
    if t.After(time.Now()) {
        return fmt.Errorf("cannot generate statement for future month")
    }
    
    return nil
}

func (v *RequestValidator) validateCardNumber(card string) error {
    if len(card) != 16 {
        return fmt.Errorf("card number must be 16 digits")
    }
    
    regex := regexp.MustCompile(`^\d{16}$`)
    if !regex.MatchString(card) {
        return fmt.Errorf("card number must contain only digits")
    }
    
    return nil
}
```

---

### 11. ORM MODELS (DATABASE)

**Файл:** `internal/calculation/orm.go` (РАСШИРИТЬ)

**Добавить таблицы:**

```go
// StatementEntity - главная таблица выписок
type StatementEntity struct {
    ID             string    `gorm:"primaryKey"`
    AccountID      string    `gorm:"index"`
    Month          string    `gorm:"index"`
    Status         string    // pending/processing/completed/failed
    BusinessType   string
    InitialBalance float64
    FinalBalance   float64
    TotalRevenue   float64
    TotalExpenses  float64
    NetProfit      float64

    // ??????? (не сделал)
    Data           []byte    `gorm:"type:jsonb"` // Полный MatematikaResponse JSON
    // ???????

    CreatedAt      time.Time `gorm:"autoCreateTime"`
    UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// TransactionEntity - таблица транзакций
type TransactionEntity struct {
    ID              string    `gorm:"primaryKey"`
    StatementID     string    `gorm:"index"`
    TransactionDate time.Time `gorm:"index"`
    PostingDate     time.Time
    Type            string
    Category        string
    Method          string
    Amount          float64
    BalanceAfter    float64
    IsManual        bool
    CreatedAt       time.Time `gorm:"autoCreateTime"`
}

// DailyBalanceEntity - таблица дневных балансов
type DailyBalanceEntity struct {
    ID          string    `gorm:"primaryKey"`
    StatementID string    `gorm:"index"`
    Date        time.Time `gorm:"index"`
    Balance     float64
}
```

---

### 12. МАСШТАБИРОВАНИЕ И ОПТИМИЗАЦИЯ

**Файл:** `internal/generator/scaling.go` (СОЗДАТЬ)

```go
package generator

// ScaleTransactions - масштабирование количества транзакций
func ScaleTransactions(transactions []Transaction, coefficient float64) []Transaction {
    // TODO:
    // 1. Для каждой категории:
    //    - Если категория не "Payroll ADP" и не фиксированная 1 раз/месяц
    //    - Умножить количество транзакций на coefficient
    // 2. Пропорционально уменьшить суммы каждой транзакции
    // 3. Проверить что общие проценты по категориям сохранились
}

// HandleFifthWeek - обработка пятой недели
func HandleFifthWeek(transactions []Transaction, month time.Time) []Transaction {
    // TODO:
    // 1. Проверить количество пятниц (или других дней) в месяце
    // 2. Если пятниц 5 вместо 4:
    //    - Разделить сумму категории на 5 вместо 4
    //    - Сгенерировать 5-ю транзакцию
    // 3. Нормализовать суммы
}

// OptimizeForTargetProfit - оптимизация расходов для целевой прибыли
func OptimizeForTargetProfit(
    expenses []Transaction,
    targetProfit float64,
    revenue float64,
) []Transaction {
    // TODO:
    // 1. Вычислить currentExpenses = сумма всех расходов
    // 2. Вычислить targetExpenses = revenue - targetProfit
    // 3. Если currentExpenses > targetExpenses:
    //    a. Найти все Optional категории
    //    b. Удалить Optional транзакции до достижения targetExpenses
    //    c. Если не хватает - пропорционально уменьшить другие расходы
}
```

---

## ПОРЯДОК РЕАЛИЗАЦИИ (Roadmap)

### Этап 1: Базовые утилиты (1-2 дня)
1. ✅ Расширить `models.go` - добавить все поля в Request
2. ✅ Создать `validator.go` - валидация входных данных
3. ✅ Создать `clients/shared_client.go` - HTTP клиент для Shared

### Этап 2: Работа с датами (2-3 дня)
4. ✅ Создать `scheduler/date_generator.go` - генерация дат
5. ✅ Создать `scheduler/holiday.go` - работа с праздниками
6. ✅ Создать `scheduler/time_generator.go` - генерация времени

### Этап 3: Генераторы транзакций (5-7 дней)
7. ✅ Создать `generator/categories.go` - конфигурация категорий
8. ✅ Создать `generator/income.go` - генерация доходов (B2C + B2B)
9. ✅ Создать `generator/expenses.go` - генерация расходов
10. ✅ Создать `generator/scaling.go` - масштабирование

### Этап 4: Нормализация (2-3 дня)
11. ✅ Создать `normalizer/rounding.go` - округление
12. ✅ Создать `normalizer/balance.go` - расчет балансов
13. ✅ Создать `normalizer/validator.go` - валидация результата

### Этап 5: Интеграция (3-4 дня)
14. ✅ Переписать `service.go` - главный workflow
15. ✅ Реализовать `repository.go` - сохранение в БД
16. ✅ Расширить `orm.go` - database entities
17. ✅ Обновить `handler.go` - обработка ошибок

### Этап 6: Тестирование (5-7 дней)
18. ✅ Unit тесты для генераторов
19. ✅ Integration тесты для service
20. ✅ E2E тесты для API endpoints

**ИТОГО: ~20-30 дней разработки**

---

## ДОПОЛНИТЕЛЬНЫЕ РЕКОМЕНДАЦИИ

### Приоритеты реализации:
1. **Критично:** Генераторы доходов/расходов (без них ничего не работает)
2. **Важно:** Работа с датами и праздниками (без этого даты будут неправильные)
3. **Важно:** Нормализация балансов (без этого суммы не сойдутся)
4. **Средне:** Масштабирование и оптимизация (можно отложить на v2)

### Тестирование:
- Каждый генератор должен иметь unit тесты
- Проверять что суммы доходов = 100% revenue
- Проверять отсутствие транзакций в праздники
- Проверять отсутствие отрицательного баланса

### Performance:
- Кэшировать праздники из Shared Service
- Использовать goroutines для параллельной генерации категорий
- Оптимизировать БД запросы (bulk insert для транзакций)

---

## ПОЛНАЯ СТРУКТУРА ВХОДНОГО JSON

### Обзор структуры

Входной JSON состоит из трех основных разделов:
1. **companyInfo** - информация о компании и счете
2. **financials** - финансовые параметры генерации
3. **customData** - пользовательские данные и кастомизация

### Пример полного входного JSON:

```json
{
  "companyInfo": {
    "companyName": "Srb Autos LLC.",
    "ownerName": "John Doe",
    "accountNumber": "201290125551",
    "associatedCard": "2091222000102910",
    "model": "B2B",
    "state": "CA",
    "industry": "automotive"
  },
  "financials": {
    "startBalance": 100000.00,
    "turnover": 100000.00,
    "profitPercent": 7.5,
    "months": 3,
    "startMonth": "2025-01",
    "operationMultiplier": 1.0
  },
  "customData": {
    "manualIncomes": [
      {
        "date": "2025-01-01",
        "amount": 5000.00,
        "category": "Пополнение шлюз",
        "description": "Initial gateway deposit"
      }
    ],
    "manualExpenses": [
      {
        "date": "2025-01-03",
        "amount": 1500.00,
        "category": "Маркетинг",
        "description": "Marketing campaign"
      }
    ],
    "customCustomers": [
      "Super LLC",
      "Lulu Inc."
    ],
    "customContractors": [
      {
        "transactionType": "Бухгалтер",
        "name": "Jakson Sam CPA"
      },
      {
        "transactionType": "Топливо",
        "name": "LumNuft Inc"
      }
    ],
    "disableCategories": ["Клининг", "Охрана"]
  }
}
```

---

### ДЕТАЛЬНОЕ ОПИСАНИЕ ПОЛЕЙ

#### 1. COMPANY INFO

**Файл:** `internal/calculation/models.go`

**Структура:**

```go
type CompanyInfo struct {
    // Обязательные поля
    CompanyName   string `json:"company_name" validate:"required,min=3,max=100"`
    OwnerName     string `json:"owner_name" validate:"required,min=3,max=100"`
    
    // Опциональные поля
    AccountNumber  *string `json:"account_number,omitempty" validate:"omitempty,numeric,len=12"`
    AssociatedCard *string `json:"associated_card,omitempty" validate:"omitempty,numeric,len=16"`
    Model          *string `json:"model,omitempty" validate:"omitempty,oneof=B2B B2C"`
    State          *string `json:"state,omitempty" validate:"omitempty,len=2,uppercase"` // CA, NY, TX
    Industry       *string `json:"industry,omitempty"` // automotive, retail, agriculture
}
```

**Логика обработки:**

```go
// Файл: internal/calculation/service.go

func (s *calculationService) processCompanyInfo(req *GenerateStatementRequest) (*ProcessedCompanyInfo, error) {
    info := req.CompanyInfo
    
    // 1. Генерация accountNumber если не указан
    accountNumber := info.AccountNumber
    if accountNumber == nil {
        generated := generateAccountNumber() // 12 цифр
        accountNumber = &generated
    }
    
    // 2. Генерация associatedCard если не указан
    associatedCard := info.AssociatedCard
    if associatedCard == nil {
        generated := generateCardNumber() // 16 цифр, валидный по Luhn
        associatedCard = &generated
    }
    
    // 3. Определение модели бизнеса
    model := BusinessTypeB2C // по умолчанию
    if info.Model != nil {
        model = BusinessType(*info.Model)
    } else if len(req.CustomData.CustomCustomers) > 0 {
        // Если указаны кастомные клиенты → скорее всего B2B
        model = BusinessTypeB2B
    }
    
    return &ProcessedCompanyInfo{
        CompanyName:    info.CompanyName,
        OwnerName:      info.OwnerName,
        AccountNumber:  *accountNumber,
        AssociatedCard: *associatedCard,
        Model:          model,
        State:          info.State,
        Industry:       info.Industry,
    }, nil
}

// Генерация валидного номера счета
func generateAccountNumber() string {
    // TODO: 12 случайных цифр
    // Формат: XXXXXXXXXXXX
}

// Генерация валидного номера карты (Luhn algorithm)
func generateCardNumber() string {
    // TODO: 16 цифр, валидный по алгоритму Luhn
    // Формат: XXXX XXXX XXXX XXXX
}
```

---

#### 2. FINANCIALS

**Файл:** `internal/calculation/models.go`

**Структура:**

```go
type Financials struct {
    // Начальный баланс (обязательно для первого месяца)
    StartBalance *float64 `json:"start_balance,omitempty" validate:"omitempty,gte=0"`
    
    // Оборот/выручка (можно указать вместо расчета)
    Turnover *float64 `json:"turnover,omitempty" validate:"omitempty,gt=0"`
    
    // Целевая прибыль (в процентах от оборота)
    ProfitPercent *float64 `json:"profit_percent,omitempty" validate:"omitempty,gte=0,lte=50"`
    
    // Целевая прибыль (абсолютная сумма)
    TargetProfit *float64 `json:"target_profit,omitempty" validate:"omitempty,gte=0"`
    
    // Количество месяцев для генерации
    Months *int `json:"months,omitempty" validate:"omitempty,gte=1,lte=36"`
    
    // Начальный месяц
    StartMonth *string `json:"start_month,omitempty" validate:"omitempty,datetime=2006-01"`
    
    // Явные периоды (альтернатива months + startMonth)
    Periods []string `json:"periods,omitempty" validate:"omitempty,dive,datetime=2006-01"`
    
    // Коэффициент масштабирования операций
    OperationMultiplier *float64 `json:"operation_multiplier,omitempty" validate:"omitempty,gt=0,lte=5"`
}
```

**Логика обработки:**

```go
// Файл: internal/calculation/service.go

func (s *calculationService) processFinancials(req *GenerateStatementRequest) (*ProcessedFinancials, error) {
    fin := req.Financials
    
    // 1. Определение начального баланса
    startBalance := float64(0)
    if fin.StartBalance != nil {
        startBalance = *fin.StartBalance
    } else {
        // Получить из БД последний баланс для accountID
        lastBalance, err := s.repo.GetLastBalance(ctx, req.AccountID)
        if err != nil {
            return nil, fmt.Errorf("start balance required for first month")
        }
        startBalance = lastBalance
    }
    
    // 2. Определение оборота
    turnover := float64(100000) // дефолт
    if fin.Turnover != nil {
        turnover = *fin.Turnover
    }
    
    // 3. Определение целевой прибыли
    profitPercent := 7.5 // дефолт 7.5% (середина диапазона 6-9%)
    if fin.ProfitPercent != nil {
        profitPercent = *fin.ProfitPercent
    }
    
    targetProfit := turnover * (profitPercent / 100)
    if fin.TargetProfit != nil {
        targetProfit = *fin.TargetProfit
    }
    
    // 4. Валидация: прибыль не может быть больше оборота
    if targetProfit >= turnover {
        return nil, fmt.Errorf("target profit (%.2f) cannot exceed turnover (%.2f)", targetProfit, turnover)
    }
    
    // 5. Определение периодов для генерации
    var periods []string
    if len(fin.Periods) > 0 {
        periods = fin.Periods
    } else if fin.Months != nil && fin.StartMonth != nil {
        // Сгенерировать N месяцев начиная с startMonth
        periods = generateMonthSequence(*fin.StartMonth, *fin.Months)
    } else {
        return nil, fmt.Errorf("either 'periods' or 'months + startMonth' must be specified")
    }
    
    // 6. Коэффициент масштабирования
    multiplier := 1.0
    if fin.OperationMultiplier != nil {
        multiplier = *fin.OperationMultiplier
    }
    
    return &ProcessedFinancials{
        StartBalance:        startBalance,
        Turnover:            turnover,
        TargetProfit:        targetProfit,
        TargetExpenses:      turnover - targetProfit,
        Periods:             periods,
        OperationMultiplier: multiplier,
    }, nil
}

// Генерация последовательности месяцев
func generateMonthSequence(startMonth string, count int) []string {
    // TODO:
    // 1. Парсить startMonth: "2025-01"
    // 2. Генерировать count месяцев: ["2025-01", "2025-02", "2025-03"]
}
```

---

#### 3. CUSTOM DATA

**Файл:** `internal/calculation/models.go`

**Структура:**

```go
type CustomData struct {
    // Ручные доходы
    ManualIncomes []ManualIncome `json:"manual_incomes,omitempty" validate:"omitempty,dive"`
    
    // Ручные расходы
    ManualExpenses []ManualExpense `json:"manual_expenses,omitempty" validate:"omitempty,dive"`
    
    // Кастомные клиенты (для замены в B2B)
    CustomCustomers []string `json:"custom_customers,omitempty" validate:"omitempty,dive,min=2,max=100"`
    
    // Кастомные подрядчики
    CustomContractors []CustomContractor `json:"custom_contractors,omitempty" validate:"omitempty,dive"`
    
    // Отключенные категории
    DisableCategories []string `json:"disable_categories,omitempty"`
    
    // Заметки
    Notes *string `json:"notes,omitempty"`
}

type ManualIncome struct {
    Date        string              `json:"date" validate:"required,datetime=2006-01-02"`
    Amount      float64             `json:"amount" validate:"required,gt=0"`
    Category    *string             `json:"category,omitempty"`
    Type        *string             `json:"type,omitempty"` // Wire, Zelle, Gateway
    Description *string             `json:"description,omitempty"`
}

type ManualExpense struct {
    Date        string              `json:"date" validate:"required,datetime=2006-01-02"`
    Amount      float64             `json:"amount" validate:"required,gt=0"`
    Category    string              `json:"category" validate:"required"`
    Description *string             `json:"description,omitempty"`
}
```

**Логика обработки:**

```go
// Файл: internal/calculation/service.go

func (s *calculationService) processCustomData(customData *CustomData, month string) (*ProcessedCustomData, error) {
    
    // 1. Обработка manual incomes
    manualIncomes := []Transaction{}
    for _, income := range customData.ManualIncomes {
        // Валидация что дата в пределах месяца
        if !isDateInMonth(income.Date, month) {
            return nil, fmt.Errorf("manual income date %s is outside month %s", income.Date, month)
        }
        
        // Определение категории
        category := "Прочее"
        if income.Category != nil {
            category = *income.Category
        } else if income.Type != nil {
            category = mapIncomeTypeToCategory(*income.Type)
        }
        
        tx := Transaction{
            TransactionID: generateTransactionID("manual_inc"),
            Date:          parseDate(income.Date),
            Type:          TransactionTypeIncome,
            Category:      category,
            Amount:        income.Amount,
            IsManual:      true,
            Description:   income.Description,
        }
        
        manualIncomes = append(manualIncomes, tx)
    }
    
    // 2. Обработка manual expenses
    manualExpenses := []Transaction{}
    for _, expense := range customData.ManualExpenses {
        // Валидация даты
        if !isDateInMonth(expense.Date, month) {
            return nil, fmt.Errorf("manual expense date %s is outside month %s", expense.Date, month)
        }
        
        tx := Transaction{
            TransactionID: generateTransactionID("manual_exp"),
            Date:          parseDate(expense.Date),
            Type:          TransactionTypeExpense,
            Category:      expense.Category,
            Amount:        -expense.Amount, // Отрицательное значение для расходов
            IsManual:      true,
            Description:   expense.Description,
        }
        
        manualExpenses = append(manualExpenses, tx)
    }
    
    // 3. Валидация customCustomers
    if len(customData.CustomCustomers) > 20 {
        return nil, fmt.Errorf("too many custom customers, max 20")
    }
    
    // 4. Валидация customContractors
    for _, contractor := range customData.CustomContractors {
        if !isValidTransactionType(contractor.TransactionType) {
            return nil, fmt.Errorf("unknown transaction type: %s", contractor.TransactionType)
        }
    }
    
    // 5. Валидация disableCategories
    for _, category := range customData.DisableCategories {
        if !isOptionalCategory(category) {
            log.Printf("Warning: trying to disable mandatory category: %s", category)
        }
    }
    
    return &ProcessedCustomData{
        ManualIncomes:     manualIncomes,
        ManualExpenses:    manualExpenses,
        CustomCustomers:   customData.CustomCustomers,
        CustomContractors: customData.CustomContractors,
        DisableCategories: customData.DisableCategories,
    }, nil
}

// Проверка является ли дата в указанном месяце
func isDateInMonth(date string, month string) bool {
    // TODO:
    // 1. Парсить date: "2025-01-03"
    // 2. Парсить month: "2025-01"
    // 3. Проверить что date.Year() == month.Year() && date.Month() == month.Month()
}

// Маппинг типа дохода на категорию
func mapIncomeTypeToCategory(incomeType string) string {
    switch incomeType {
    case "Wire":
        return "Wire Transfer"
    case "Zelle":
        return "Zelle"
    case "Gateway":
        return "Пополнение шлюз"
    default:
        return "Прочее"
    }
}

// Проверка является ли категория опциональной
func isOptionalCategory(category string) bool {
    optionalCategories := []string{
        "Маркетинг", "Клининг", "Охрана", "Бухгалтер", 
        "Юрист", "IT-sec", "IT-dev",
    }
    
    for _, opt := range optionalCategories {
        if opt == category {
            return true
        }
    }
    return false
}
```

---

### ОБНОВЛЕННАЯ ПОЛНАЯ СТРУКТУРА REQUEST

**Файл:** `internal/calculation/handler.go`

**Заменить GenerateStatementRequest на:**

```go
type GenerateStatementRequest struct {
    CompanyInfo CompanyInfo `json:"company_info" validate:"required"`
    Financials  Financials  `json:"financials" validate:"required"`
    CustomData  *CustomData `json:"custom_data,omitempty"`
}
```

**Примечание:** Старые поля (AccountID, Month, BusinessType, InitialBalance) теперь внутри вложенных структур:
- `AccountID` → берется из `CompanyInfo.AccountNumber`
- `Month` → берется из `Financials.StartMonth` или `Financials.Periods[0]`
- `BusinessType` → берется из `CompanyInfo.Model`
- `InitialBalance` → берется из `Financials.StartBalance`

---

### ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ

**Файл:** `internal/calculation/validator.go` (РАСШИРИТЬ)

**Добавить методы:**

```go
type RequestValidator struct {
    sharedClient clients.SharedServiceClient
}

// ValidateRequest - полная валидация входного JSON
func (v *RequestValidator) ValidateRequest(req *GenerateStatementRequest) error {
    // 1. Валидация companyInfo
    if err := v.validateCompanyInfo(req.CompanyInfo); err != nil {
        return fmt.Errorf("company_info validation failed: %w", err)
    }
    
    // 2. Валидация financials
    if err := v.validateFinancials(req.Financials); err != nil {
        return fmt.Errorf("financials validation failed: %w", err)
    }
    
    // 3. Валидация customData (если указан)
    if req.CustomData != nil {
        if err := v.validateCustomData(req.CustomData, req.Financials); err != nil {
            return fmt.Errorf("custom_data validation failed: %w", err)
        }
    }
    
    // 4. Бизнес-валидация
    if err := v.validateBusinessRules(req); err != nil {
        return fmt.Errorf("business rules validation failed: %w", err)
    }
    
    return nil
}

func (v *RequestValidator) validateCompanyInfo(info CompanyInfo) error {
    // Валидация companyName
    if len(info.CompanyName) < 3 {
        return fmt.Errorf("company_name too short")
    }
    
    // Валидация accountNumber (если указан)
    if info.AccountNumber != nil {
        if len(*info.AccountNumber) != 12 {
            return fmt.Errorf("account_number must be 12 digits")
        }
        if !isNumeric(*info.AccountNumber) {
            return fmt.Errorf("account_number must contain only digits")
        }
    }
    
    // Валидация associatedCard (если указан)
    if info.AssociatedCard != nil {
        if err := v.validateCardNumber(*info.AssociatedCard); err != nil {
            return err
        }
    }
    
    // Валидация state (если указан)
    if info.State != nil {
        if !isValidUSState(*info.State) {
            return fmt.Errorf("invalid US state code: %s", *info.State)
        }
    }
    
    return nil
}

func (v *RequestValidator) validateFinancials(fin Financials) error {
    // Проверка обязательных полей
    if fin.StartMonth == nil && len(fin.Periods) == 0 {
        return fmt.Errorf("either start_month or periods must be specified")
    }
    
    if fin.Months != nil && fin.StartMonth == nil {
        return fmt.Errorf("start_month required when months specified")
    }
    
    // Валидация месяцев
    if fin.StartMonth != nil {
        if err := v.validateMonth(*fin.StartMonth); err != nil {
            return err
        }
    }
    
    for _, period := range fin.Periods {
        if err := v.validateMonth(period); err != nil {
            return fmt.Errorf("invalid period %s: %w", period, err)
        }
    }
    
    // Валидация profitPercent
    if fin.ProfitPercent != nil {
        if *fin.ProfitPercent < 0 || *fin.ProfitPercent > 50 {
            return fmt.Errorf("profit_percent must be between 0 and 50, got %.2f", *fin.ProfitPercent)
        }
    }
    
    // Валидация operationMultiplier
    if fin.OperationMultiplier != nil {
        if *fin.OperationMultiplier <= 0 || *fin.OperationMultiplier > 5 {
            return fmt.Errorf("operation_multiplier must be between 0 and 5")
        }
    }
    
    return nil
}

func (v *RequestValidator) validateCustomData(data *CustomData, fin Financials) error {
    // Получаем период для валидации дат
    month := getFirstMonth(fin)
    
    // 1. Валидация manual incomes
    totalManualIncome := 0.0
    for i, income := range data.ManualIncomes {
        if err := v.validateManualIncome(income, month); err != nil {
            return fmt.Errorf("manual_income[%d]: %w", i, err)
        }
        totalManualIncome += income.Amount
    }
    
    // 2. Валидация manual expenses
    totalManualExpense := 0.0
    for i, expense := range data.ManualExpenses {
        if err := v.validateManualExpense(expense, month); err != nil {
            return fmt.Errorf("manual_expense[%d]: %w", i, err)
        }
        totalManualExpense += expense.Amount
    }
    
    // 3. Проверка что manual операции не превышают разумные пределы
    if fin.Turnover != nil {
        if totalManualIncome > *fin.Turnover {
            return fmt.Errorf("manual incomes (%.2f) exceed turnover (%.2f)", totalManualIncome, *fin.Turnover)
        }
        
        maxExpenses := *fin.Turnover * 0.95 // максимум 95% оборота
        if totalManualExpense > maxExpenses {
            return fmt.Errorf("manual expenses (%.2f) too high", totalManualExpense)
        }
    }
    
    // 4. Валидация customContractors
    for i, contractor := range data.CustomContractors {
        if contractor.TransactionType == "" || contractor.Name == "" {
            return fmt.Errorf("custom_contractor[%d]: transaction_type and name required", i)
        }
    }
    
    return nil
}

func (v *RequestValidator) validateBusinessRules(req *GenerateStatementRequest) error {
    // Бизнес-правила
    fin := req.Financials
    
    // 1. Проверка достаточности начального баланса
    if fin.StartBalance != nil && fin.Turnover != nil {
        targetProfit := calculateTargetProfit(*fin.Turnover, fin.ProfitPercent)
        targetExpenses := *fin.Turnover - targetProfit
        
        // Начальный баланс должен покрывать хотя бы половину расходов
        minBalance := targetExpenses * 0.5
        if *fin.StartBalance < minBalance {
            return fmt.Errorf(
                "start_balance (%.2f) too low for expenses (%.2f), minimum %.2f recommended",
                *fin.StartBalance, targetExpenses, minBalance,
            )
        }
    }
    
    // 2. Проверка реалистичности прибыли
    if fin.ProfitPercent != nil && *fin.ProfitPercent > 20 {
        log.Printf("Warning: unusually high profit percent: %.2f%%", *fin.ProfitPercent)
    }
    
    // 3. Проверка количества ручных операций
    if req.CustomData != nil {
        totalManual := len(req.CustomData.ManualIncomes) + len(req.CustomData.ManualExpenses)
        if totalManual > 30 {
            return fmt.Errorf("too many manual transactions (%d), max 30", totalManual)
        }
    }
    
    return nil
}

// Валидация номера карты по алгоритму Luhn
func (v *RequestValidator) validateCardNumber(card string) error {
    if len(card) != 16 {
        return fmt.Errorf("card number must be 16 digits")
    }
    
    if !isNumeric(card) {
        return fmt.Errorf("card number must contain only digits")
    }
    
    // Проверка по алгоритму Luhn
    if !luhnCheck(card) {
        return fmt.Errorf("invalid card number (Luhn check failed)")
    }
    
    return nil
}

// Алгоритм Luhn для валидации номера карты
func luhnCheck(card string) bool {
    // TODO: реализовать алгоритм Luhn
    // https://en.wikipedia.org/wiki/Luhn_algorithm
}

// Проверка US state code
func isValidUSState(state string) bool {
    validStates := []string{
        "AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
        "HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
        "MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
        "NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
        "SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY",
    }
    
    for _, valid := range validStates {
        if valid == state {
            return true
        }
    }
    return false
}

func isNumeric(s string) bool {
    regex := regexp.MustCompile(`^\d+$`)
    return regex.MatchString(s)
}
```

---

### ИНТЕГРАЦИЯ MANUAL ТРАНЗАКЦИЙ

**Файл:** `internal/generator/merger.go` (СОЗДАТЬ)

**Функции:**

```go
package generator

// MergeManualTransactions - интеграция ручных транзакций с автоматическими
func MergeManualTransactions(
    autoTransactions []Transaction,
    manualTransactions []Transaction,
    targetRevenue float64,
    targetExpenses float64,
) ([]Transaction, error) {
    // TODO:
    // 1. Подсчитать суммы manual транзакций
    manualIncome := sumByType(manualTransactions, TransactionTypeIncome)
    manualExpense := sumByType(manualTransactions, TransactionTypeExpense)
    
    // 2. Скорректировать автоматические транзакции
    //    Например, если manual expense добавил 1500$ маркетинга,
    //    а автоматический маркетинг должен был быть 700$,
    //    то автоматический можно убрать или уменьшить
    
    // 3. Если manual income/expense уменьшает нужное количество авто-операций:
    //    - Удалить некоторые Optional категории
    //    - Пропорционально уменьшить суммы других операций
    
    // 4. Нормализовать итоговые суммы:
    //    totalIncome = targetRevenue
    //    totalExpense = targetExpenses
    
    // 5. Объединить manual + auto транзакции
    allTransactions := append(autoTransactions, manualTransactions...)
    
    // 6. Отсортировать по дате
    sortByDate(allTransactions)
    
    return allTransactions, nil
}

// AdjustForManualOperations - корректировка автоматических операций
func AdjustForManualOperations(
    autoTransactions []Transaction,
    manualCount int,
    targetCount int, // целевое количество транзакций (39-75)
) []Transaction {
    // TODO:
    // 1. Если (len(auto) + manualCount) > targetCount:
    //    - Удалить некоторые Optional операции
    //    - Уменьшить частоту некоторых категорий
    
    // 2. Пересчитать суммы оставшихся операций
    //    чтобы проценты сохранились
}
```

---

### ОПРЕДЕЛЕНИЕ КАТЕГОРИЙ ДЛЯ ЗАМЕНЫ

**Файл:** `internal/generator/replacements.go` (СОЗДАТЬ)

```go
package generator

// ApplyCustomContractors - замена дефолтных контрагентов на кастомные
func ApplyCustomContractors(
    transactions []Transaction,
    customContractors []CustomContractor,
) []Transaction {
    // TODO:
    // Создать map: transactionType -> customName
    replacements := make(map[string]string)
    for _, contractor := range customContractors {
        replacements[contractor.TransactionType] = contractor.Name
    }
    
    // Для каждой транзакции:
    // - Если её Category совпадает с transactionType
    // - Заменить contractor name (сохранить в метаданных для Maska)
    // - Или пометить что нужно использовать кастомного
    
    for i, tx := range transactions {
        if customName, found := replacements[tx.Category]; found {
            // Добавить метаданные для Maska
            tx.CustomContractor = &customName
            transactions[i] = tx
        }
    }
    
    return transactions
}

// ApplyCustomCustomers - замена дефолтных клиентов (для B2B доходов)
func ApplyCustomCustomers(
    transactions []Transaction,
    customCustomers []string,
) []Transaction {
    // TODO:
    // Для B2B доходов (retails, wholesale, agriculture):
    // 1. Найти все income транзакции этих категорий
    // 2. Заменить до len(customCustomers) клиентов на кастомные
    // 3. Сохранить в метаданных для Maska
    
    customIndex := 0
    for i, tx := range transactions {
        if tx.Type == TransactionTypeIncome && isB2BCategory(tx.Category) {
            if customIndex < len(customCustomers) {
                tx.CustomCustomer = &customCustomers[customIndex]
                transactions[i] = tx
                customIndex++
            }
        }
    }
    
    return transactions
}

func isB2BCategory(category string) bool {
    b2bCategories := []string{
        "retails_ca.csv", "wholesale_ca.csv", 
        "agriculture_ca.csv", "factoring_avance_ca.csv",
    }
    
    for _, cat := range b2bCategories {
        if cat == category {
            return true
        }
    }
    return false
}
```

---

### ОБРАБОТКА DISABLED КАТЕГОРИЙ

**Файл:** `internal/generator/expenses.go` (ДОБАВИТЬ)

```go
// FilterDisabledCategories - исключение отключенных категорий
func FilterDisabledCategories(
    categories []ExpenseCategory,
    disabledList []string,
) []ExpenseCategory {
    // TODO:
    // 1. Создать set из disabledList
    // 2. Отфильтровать categories, исключив те что в disabledList
    // 3. ВАЖНО: проверить что категория Optional перед удалением
    //    Нельзя удалить обязательную категорию (Payroll, IRS налоги)
    
    filtered := []ExpenseCategory{}
    disabledSet := makeSet(disabledList)
    
    for _, cat := range categories {
        if disabledSet[cat.Name] {
            // Пытаемся удалить эту категорию
            if !cat.Optional {
                log.Printf("WARNING: Cannot disable mandatory category: %s", cat.Name)
                filtered = append(filtered, cat) // оставляем
            } else {
                log.Printf("Disabled optional category: %s", cat.Name)
                // не добавляем в filtered
            }
        } else {
            filtered = append(filtered, cat)
        }
    }
    
    return filtered
}
```

---

## ПОЭТАПНАЯ РАЗРАБОТКА С УЧЕТОМ ВХОДНОГО JSON

### ЭТАП 1: Обновление моделей данных (1 день)

#### 1.1 Расширить `internal/calculation/models.go`
- ✅ Добавить `CompanyInfo` структуру
- ✅ Добавить `Financials` структуру
- ✅ Добавить `CustomData` структуру
- ✅ Добавить `ManualIncome` структуру
- ✅ Добавить `ManualExpense` структуру
- ✅ Обновить `GenerateStatementRequest` для использования вложенных структур

#### 1.2 Добавить вспомогательные типы в `types.go`
- ✅ `ExpenseType` enum (Percentage/Fixed)
- ✅ Константы для опциональных категорий

**Файлы:**
- `internal/calculation/models.go`
- `internal/calculation/types.go`

---

### ЭТАП 2: Валидация входных данных (2 дня)

#### 2.1 Создать `internal/calculation/validator.go`
- ✅ `ValidateRequest()` - главная функция
- ✅ `validateCompanyInfo()` - валидация компании
- ✅ `validateFinancials()` - валидация финансов
- ✅ `validateCustomData()` - валидация custom данных
- ✅ `validateBusinessRules()` - бизнес-правила

#### 2.2 Реализовать вспомогательные функции
- ✅ `validateMonth()` - формат YYYY-MM
- ✅ `validateCardNumber()` - Luhn algorithm
- ✅ `validateManualIncome()` - валидация ручного дохода
- ✅ `validateManualExpense()` - валидация ручного расхода
- ✅ `isValidUSState()` - проверка state кода
- ✅ `luhnCheck()` - алгоритм Luhn
- ✅ `isDateInMonth()` - дата в пределах месяца

**Файлы:**
- `internal/calculation/validator.go` (создать)

---

### ЭТАП 3: Генераторы служебных данных (1 день)

#### 3.1 Создать `internal/calculation/generators.go`
- ✅ `generateAccountNumber()` - генерация номера счета (12 цифр)
- ✅ `generateCardNumber()` - генерация номера карты (16 цифр + Luhn)
- ✅ `generateTransactionID()` - генерация ID транзакции
- ✅ `formatMonthKey()` - форматирование "JANUARY 2025"
- ✅ `parseMonth()` - парсинг "2025-01" → year, month
- ✅ `generateMonthSequence()` - генерация последовательности месяцев

**Файлы:**
- `internal/calculation/generators.go` (создать)

---

### ЭТАП 4: HTTP клиент для Shared Service (2 дня)

#### 4.1 Создать `internal/clients/shared_client.go`
- ✅ `SharedServiceClient` структура
- ✅ `NewSharedServiceClient()` конструктор
- ✅ `GetHolidays()` - получить праздники
- ✅ `GetPatterns()` - получить шаблоны
- ✅ `GetGateways()` - получить шлюзы
- ✅ `GetContractors()` - получить контрагентов
- ✅ Error handling и retry логика
- ✅ Кэширование результатов

**Файлы:**
- `internal/clients/shared_client.go` (создать)
- `internal/clients/models.go` (создать) - response модели от Shared

---

### ЭТАП 5: Работа с датами и праздниками (3 дня)

#### 5.1 Создать `internal/scheduler/date_generator.go`
- ✅ `FindNthWeekday()` - N-я пятница/понедельник
- ✅ `FindAllWeekdays()` - все пятницы месяца
- ✅ `FindLastWeekday()` - последняя пятница
- ✅ `RandomWeekday()` - случайный рабочий день
- ✅ `NextWorkingDay()` - следующий рабочий день
- ✅ `GetDayOfMonth()` - N-е число месяца
- ✅ Unit тесты для генерации дат

#### 5.2 Создать `internal/scheduler/holiday.go`
- ✅ `HolidayChecker` структура с кэшем
- ✅ `LoadHolidays()` - загрузка из Shared
- ✅ `IsHoliday()` - проверка праздника
- ✅ `IsWorkingDay()` - проверка рабочего дня
- ✅ `IsWeekend()` - проверка выходного

#### 5.3 Создать `internal/scheduler/time_generator.go`
- ✅ `GenerateTime()` - время в диапазоне
- ✅ `GenerateMidnight()` - специально 00:01
- ✅ `GenerateBusinessHours()` - 08:00-18:00
- ✅ `GenerateCardHours()` - 09:00-20:00

**Файлы:**
- `internal/scheduler/date_generator.go` (создать)
- `internal/scheduler/holiday.go` (создать)
- `internal/scheduler/time_generator.go` (создать)
- `internal/scheduler/types.go` (создать) - вспомогательные типы

---

### ЭТАП 6: Конфигурация категорий расходов (2 дня)

#### 6.1 Создать `internal/generator/categories.go`
- ✅ `ExpenseCategory` структура
- ✅ `DefaultExpenseCategories` - список всех ~15 категорий:
  - Payroll ADP (обязательный, 27-27.5%, 2-я и 4-я пятница)
  - Топливо / Fleet (обязательный, 15-17.5%, 7-9 раз)
  - Маркетинг (опциональный, 0.5-0.7%, 1 раз)
  - Подписка ПО (обязательный, фиксированная, 1 раз, same_weekday)
  - Мобильная связь (обязательный, фиксированная первый месяц, 2-я пятница)
  - Коммунальные (обязательный, фиксированная первый месяц, 3-я пятница)
  - Leasing (обязательный, 11.5-12% первый месяц, последняя пятница)
  - IRS-налоги (обязательный, 1.5-2%, 15-е число)
  - Перевод владельцу (обязательный, случайный будний)
  - Платная дорога (опциональный, фиксированная $20/35/50)
  - Доплата за перегруз (опциональный, формула weight*rate)
  - Авторемонт (опциональный, процентный)
  - Запчасти (опциональный, процентный)
  - Клининг (опциональный, фиксированная)
  - Охрана (опциональный, фиксированная)
  - Бухгалтер (опциональный, фиксированная)
  - Юрист (опциональный, фиксированная)

#### 6.2 Создать вспомогательные функции
- ✅ `GetOptionalCategories()` - список опциональных
- ✅ `GetMandatoryCategories()` - список обязательных
- ✅ `FilterDisabledCategories()` - исключение отключенных

**Файлы:**
- `internal/generator/categories.go` (создать)
- `internal/generator/types.go` (создать)

---

### ЭТАП 7: Генератор доходов (3 дня)

#### 7.1 Создать `internal/generator/income.go`
- ✅ `IncomeGenerator` структура
- ✅ `GenerateB2CIncome()` - генерация для B2C:
  - Найти все пятницы месяца
  - GET запрос к Shared за gateways
  - Выбрать один шлюз (сохранить в state/БД)
  - Разделить revenue на количество пятниц
  - Вариация ±4.5%
  - Нормализация до 100% revenue
- ✅ `GenerateB2BIncome()` - генерация для B2B:
  - GET запрос к Shared за B2B клиентов
  - Для каждой категории (retails, wholesale, agriculture, factoring)
  - 2-8 транзакций по 5.5-8.5%
  - 70% ACH Credit, 30% Electronic Payment
  - Применить customCustomers замены
  - Нормализация до 100% revenue

#### 7.2 Вспомогательные функции
- ✅ `distributeAmountWithVariation()` - распределение с вариацией
- ✅ `normalizeToTarget()` - нормализация сумм
- ✅ `selectPaymentMethod()` - выбор метода по проценту

**Файлы:**
- `internal/generator/income.go` (создать)

---

### ЭТАП 8: Генератор расходов (4 дня)

#### 8.1 Создать `internal/generator/expenses.go`
- ✅ `ExpenseGenerator` структура
- ✅ `GenerateExpenses()` - главная функция:
  - Загрузить категории
  - Применить disableCategories фильтр
  - Для каждой категории:
    - Вычислить сумму (percentage или fixed)
    - Сгенерировать даты по DayRule
    - Проверить праздники
    - Сгенерировать время
    - Создать транзакции
  - Оптимизировать под targetProfit
  - Нормализовать суммы

#### 8.2 Генераторы для специфичных категорий
- ✅ `generatePayroll()` - 2-я и 4-я пятница
- ✅ `generateFuel()` - 7-9 раз, random weekdays
- ✅ `generateLeasing()` - последняя пятница, fixed after first
- ✅ `generateMobile()` - 2-я пятница, fixed after first ±15%
- ✅ `generateUtilities()` - 3-я пятница, fixed after first ±15%
- ✅ `generateSoftware()` - same weekday, midnight
- ✅ `generateIRS()` - 15-е число, 2 раза в квартальные месяцы
- ✅ `generateOwnerTransfer()` - random weekday
- ✅ `generateTollRoad()` - фиксированные $20/35/50
- ✅ `generateOverweight()` - формула weight * rate

**Файлы:**
- `internal/generator/expenses.go` (создать)
- `internal/generator/fixed_expenses.go` (создать) - для фиксированных
- `internal/generator/percentage_expenses.go` (создать) - для процентных

---

### ЭТАП 9: State management для recurring транзакций (2 дня)

**Проблема:** Некоторые транзакции должны сохранять параметры между месяцами:
- Подписка ПО - день недели фиксируется
- Мобильная связь - сумма фиксируется ±15%
- Коммунальные - сумма фиксируется ±15%
- Leasing - сумма фиксируется 1:1
- B2C шлюз - один шлюз для всех месяцев

#### 9.1 Создать `internal/calculation/state.go`
- ✅ `StatementState` структура
- ✅ `SaveState()` - сохранение state в БД
- ✅ `LoadState()` - загрузка state
- ✅ `UpdateState()` - обновление параметров

```go
type StatementState struct {
    AccountID string
    
    // Фиксированные параметры
    SelectedGateway      *string  // Выбранный шлюз для B2C
    SoftwareWeekday      *int     // День недели для подписки ПО
    MobileBaseAmount     *float64 // Базовая сумма мобильной связи
    UtilitiesBaseAmount  *float64 // Базовая сумма коммунальных
    LeasingBaseAmount    *float64 // Базовая сумма лизинга
    
    LastGeneratedMonth   string   // Последний сгенерированный месяц
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

**Файлы:**
- `internal/calculation/state.go` (создать)
- `internal/calculation/orm.go` (добавить StateEntity)

---

### ЭТАП 10: Интеграция в service.go (3 дня)

#### 10.1 Переписать `internal/calculation/service.go`
- ✅ Добавить все зависимости в `calculationService`:
  ```go
  type calculationService struct {
      repo           CalculationRepository
      kafkaProducer  kafka.Producer
      validator      *RequestValidator
      sharedClient   *clients.SharedServiceClient
      incomeGen      *generator.IncomeGenerator
      expenseGen     *generator.ExpenseGenerator
      balanceCalc    *normalizer.BalanceCalculator
      normalizer     *normalizer.AmountNormalizer
      holidayChecker *scheduler.HolidayChecker
      dateGen        *scheduler.DateGenerator
  }
  ```

#### 10.2 Реализовать полный workflow в `GenerateStatement()`
- ✅ Валидация входных данных
- ✅ Обработка companyInfo (генерация номеров)
- ✅ Обработка financials (расчет целевых показателей)
- ✅ Загрузка праздников
- ✅ Загрузка/создание state
- ✅ Генерация доходов (B2C или B2B)
- ✅ Генерация расходов
- ✅ Интеграция manual транзакций
- ✅ Применение customContractors/customCustomers
- ✅ Фильтрация disabledCategories
- ✅ Нормализация сумм
- ✅ Расчет балансов
- ✅ Валидация результата
- ✅ Формирование response
- ✅ Сохранение в БД
- ✅ Обновление state
- ✅ Публикация в Kafka

**Файлы:**
- `internal/calculation/service.go` (переписать)

---

### ЭТАП 11: Builder функции (2 дня)

#### 11.1 Создать `internal/calculation/builders.go`
- ✅ `buildFinancialSummary()` - формирование финансового summary
- ✅ `buildForwardingInfo()` - формирование forwardingInfo
- ✅ `buildTotals()` - расчет totals
- ✅ `buildRevenueBreakdown()` - разбивка доходов
- ✅ `buildExpensesBreakdown()` - разбивка расходов
- ✅ `buildTransactionCounts()` - подсчет транзакций
- ✅ `convertToTransactionResponse()` - конвертация в DTO

```go
// buildFinancialSummary - формирование финансового summary
func buildFinancialSummary(
    companyInfo CompanyInfo,
    month string,
    initialBalance float64,
    transactions []Transaction,
) FinancialSummary {
    totalRevenue := sumByType(transactions, TransactionTypeIncome)
    totalExpenses := sumByType(transactions, TransactionTypeExpense)
    finalBalance := calculateFinalBalance(initialBalance, transactions)
    
    return FinancialSummary{
        CompanyName:    companyInfo.CompanyName,
        AccountNumber:  getAccountNumber(companyInfo),
        Period:         formatPeriod(month), // "2025-01-01 - 2025-01-31"
        InitialBalance: initialBalance,
        FinalBalance:   finalBalance,
        TotalRevenue:   totalRevenue,
        TotalExpenses:  totalExpenses,
        NetProfit:      totalRevenue - math.Abs(totalExpenses),
    }
}

// buildForwardingInfo - данные для Maska
func buildForwardingInfo(
    companyInfo CompanyInfo,
    customData *CustomData,
    generatedCard string,
) ForwardingInfo {
    // Определить номер карты
    card := generatedCard
    if companyInfo.AssociatedCard != nil {
        card = *companyInfo.AssociatedCard
    }
    
    // Собрать customCustomers
    customers := []string{}
    if customData != nil && len(customData.CustomCustomers) > 0 {
        customers = customData.CustomCustomers
    }
    
    // Собрать customContractors
    contractors := []CustomContractor{}
    if customData != nil && len(customData.CustomContractors) > 0 {
        contractors = customData.CustomContractors
    }
    
    return ForwardingInfo{
        AssociatedCard:    card,
        OwnerName:         companyInfo.OwnerName,
        CustomCustomers:   customers,
        CustomContractors: contractors,
    }
}

// buildTotals - для January с детальной статистикой
func buildTotals(transactions []Transaction) *Totals {
    totalRevenue := sumByType(transactions, TransactionTypeIncome)
    totalExpenses := sumByType(transactions, TransactionTypeExpense)
    
    return &Totals{
        TotalRevenue:  totalRevenue,
        TotalExpenses: totalExpenses,
        NetProfit:     totalRevenue - math.Abs(totalExpenses),
    }
}

// buildRevenueBreakdown - разбивка доходов по методам
func buildRevenueBreakdown(transactions []Transaction) *RevenueBreakdown {
    incomes := filterByType(transactions, TransactionTypeIncome)
    
    return &RevenueBreakdown{
        TotalAch:     sumByMethod(incomes, TransactionMethodACHCredit),
        TotalWire:    sumByMethod(incomes, TransactionMethodWireCredit),
        TotalZelle:   sumByCategory(incomes, "Zelle"),
        TotalGateway: sumByCategory(incomes, "Пополнение шлюз"),
        TotalOther:   sumOther(incomes),
    }
}

// buildExpensesBreakdown - разбивка расходов по методам
func buildExpensesBreakdown(transactions []Transaction) *ExpensesBreakdown {
    expenses := filterByType(transactions, TransactionTypeExpense)
    
    return &ExpensesBreakdown{
        ByCard:    sumByMethod(expenses, "card"),
        ByAccount: sumByMethod(expenses, TransactionMethodBankTransfer),
    }
}

// buildTransactionCounts - подсчет количества транзакций
func buildTransactionCounts(transactions []Transaction) *TransactionCounts {
    incomes := filterByType(transactions, TransactionTypeIncome)
    expenses := filterByType(transactions, TransactionTypeExpense)
    
    return &TransactionCounts{
        Total: len(transactions),
        Deposits: DepositCounts{
            Total: len(incomes),
            Ach:   countByMethod(incomes, TransactionMethodACHCredit),
            Wire:  countByMethod(incomes, TransactionMethodWireCredit),
            Zelle: countByCategory(incomes, "Zelle"),
        },
        Withdrawals: WithdrawalCounts{
            Total:       len(expenses),
            FromAccount: countByMethod(expenses, TransactionMethodBankTransfer),
            ByCard:      countByMethod(expenses, "card"),
        },
    }
}
```

**Файлы:**
- `internal/calculation/builders.go` (создать)

---

### ЭТАП 12: Handler обновление (1 день)

#### 12.1 Обновить `internal/calculation/handler.go`
- ✅ Обновить валидацию request body
- ✅ Добавить structured error responses
- ✅ Добавить validation middleware
- ✅ Улучшить error handling

```go
// Добавить validation middleware
func (h *CalculationHandler) validateMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Валидация перед обработкой
        return next(c)
    }
}

// Улучшенный error handling
func (h *CalculationHandler) GenerateStatement(c echo.Context) error {
    var req GenerateStatementRequest
    
    // 1. Parse JSON
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, ErrorResponse{
            Error:   "invalid_json",
            Message: "Failed to parse request body",
            Details: err.Error(),
        })
    }
    
    // 2. Validate using validator library
    if err := h.validator.Validate(&req); err != nil {
        return c.JSON(400, ErrorResponse{
            Error:   "validation_failed",
            Message: "Request validation failed",
            Details: formatValidationErrors(err),
        })
    }
    
    // 3. Call service
    result, err := h.calcService.GenerateStatement(c.Request().Context(), &req)
    if err != nil {
        // Различные типы ошибок
        if errors.Is(err, ErrInsufficientBalance) {
            return c.JSON(422, ErrorResponse{
                Error:   "insufficient_balance",
                Message: err.Error(),
            })
        }
        
        if errors.Is(err, ErrFutureMonth) {
            return c.JSON(422, ErrorResponse{
                Error:   "invalid_month",
                Message: err.Error(),
            })
        }
        
        return c.JSON(500, ErrorResponse{
            Error:   "internal_error",
            Message: "Failed to generate statement",
        })
    }
    
    return c.JSON(202, result) // 202 Accepted для async операции
}

type ErrorResponse struct {
    Error   string      `json:"error"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}
```

**Файлы:**
- `internal/calculation/handler.go` (обновить)
- `internal/calculation/errors.go` (создать) - кастомные ошибки

---

### ЭТАП 13: Repository implementation (2 дня)

#### 13.1 Реализовать `internal/calculation/repository.go`
- ✅ `SaveStatement()` - сохранение полного statement
- ✅ `GetStatementByID()` - получение statement
- ✅ `UpdateStatus()` - обновление статуса
- ✅ `GetStatus()` - получение статуса
- ✅ `SaveState()` - сохранение state
- ✅ `LoadState()` - загрузка state
- ✅ `GetLastBalance()` - последний баланс для accountID

```go
func (r *calculationRepository) SaveStatement(
    ctx context.Context,
    id string,
    statement MatematikaResponse,
) error {
    // 1. Сериализовать MatematikaResponse в JSON
    data, err := json.Marshal(statement)
    if err != nil {
        return err
    }
    
    // 2. Извлечь summary для быстрого доступа
    monthKey := getFirstMonthKey(statement)
    monthly := statement[monthKey]
    
    // 3. Создать StatementEntity
    entity := StatementEntity{
        ID:             id,
        AccountID:      monthly.FinancialSummary.AccountNumber,
        Month:          extractMonth(monthKey), // "2025-01"
        Status:         string(StatusCompleted),
        InitialBalance: monthly.FinancialSummary.InitialBalance,
        FinalBalance:   monthly.FinancialSummary.FinalBalance,
        TotalRevenue:   monthly.FinancialSummary.TotalRevenue,
        TotalExpenses:  monthly.FinancialSummary.TotalExpenses,
        NetProfit:      monthly.FinancialSummary.NetProfit,
        Data:           data,
    }
    
    // 4. Сохранить в БД
    if err := r.db.Create(&entity).Error; err != nil {
        return err
    }
    
    // 5. Опционально: сохранить транзакции отдельно для быстрых запросов
    // (можно пропустить, если достаточно JSONB)
    
    return nil
}
```

**Файлы:**
- `internal/calculation/repository.go` (реализовать)

---

### ЭТАП 14: Генератор служебных функций (1 день)

#### 14.1 Создать `internal/calculation/generators.go`
```go
// generateAccountNumber - генерация 12-значного номера счета
func generateAccountNumber() string {
    // TODO: 12 случайных цифр
    // Можно использовать формат: [2 цифры routing][10 цифр account]
}

// generateCardNumber - генерация 16-значного номера карты с Luhn
func generateCardNumber() string {
    // TODO:
    // 1. Сгенерировать 15 случайных цифр
    // 2. Вычислить 16-ю цифру по алгоритму Luhn
    // 3. Объединить в строку
}

// generateStatementID - уникальный ID выписки
func generateStatementID(accountID, month string) string {
    return fmt.Sprintf("stmt_%s_%s", month, accountID)
}

// generateTransactionID - уникальный ID транзакции
func generateTransactionID(prefix string) string {
    return fmt.Sprintf("%s_%s", prefix, uuid.New().String()[:8])
}

// formatMonthKey - "2025-01" → "JANUARY 2025"
func formatMonthKey(month string) string {
    t, _ := time.Parse("2006-01", month)
    return strings.ToUpper(t.Format("JANUARY 2006"))
}

// formatPeriod - "2025-01" → "2025-01-01 - 2025-01-31"
func formatPeriod(month string) string {
    t, _ := time.Parse("2006-01", month)
    start := t.Format("2006-01-02")
    end := t.AddDate(0, 1, -1).Format("2006-01-02")
    return fmt.Sprintf("%s - %s", start, end)
}
```

**Файлы:**
- `internal/calculation/generators.go` (создать)
- `internal/calculation/formatters.go` (создать) - formatting функции

---

## ОБНОВЛЕННЫЙ ROADMAP

### Этап 1: Модели и валидация (3 дня) ⭐ НАЧАТЬ ОТСЮДА
1. ✅ Расширить `models.go` - CompanyInfo, Financials, CustomData
2. ✅ Создать `validator.go` - полная валидация
3. ✅ Создать `generators.go` - служебные генераторы
4. ✅ Обновить `handler.go` - новый request format

### Этап 2: Клиенты и утилиты (2 дня)
5. ✅ Создать `clients/shared_client.go` - HTTP клиент
6. ✅ Создать `calculation/state.go` - state management
7. ✅ Обновить `orm.go` - StateEntity

### Этап 3: Scheduler (3 дня)
8. ✅ Создать `scheduler/date_generator.go` - даты
9. ✅ Создать `scheduler/holiday.go` - праздники
10. ✅ Создать `scheduler/time_generator.go` - время
11. ✅ Unit тесты для scheduler

### Этап 4: Категории (2 дня)
12. ✅ Создать `generator/categories.go` - 15+ категорий
13. ✅ Создать `generator/types.go` - типы

### Этап 5: Генераторы транзакций (6 дней)
14. ✅ Создать `generator/income.go` - B2C и B2B доходы
15. ✅ Создать `generator/expenses.go` - главная логика
16. ✅ Создать `generator/fixed_expenses.go` - фиксированные
17. ✅ Создать `generator/percentage_expenses.go` - процентные
18. ✅ Создать `generator/special_expenses.go` - формульные (overweight)
19. ✅ Unit тесты для генераторов

### Этап 6: Нормализация (3 дня)
20. ✅ Создать `normalizer/balance.go` - балансы
21. ✅ Создать `normalizer/rounding.go` - округление
22. ✅ Создать `normalizer/validator.go` - валидация
23. ✅ Unit тесты для normalizer

### Этап 7: Масштабирование (2 дня)
24. ✅ Создать `generator/scaling.go` - масштабирование
25. ✅ Создать `generator/merger.go` - слияние manual + auto
26. ✅ Создать `generator/replacements.go` - замены контрагентов

### Этап 8: Интеграция (3 дня)
27. ✅ Создать `calculation/builders.go` - builder функции
28. ✅ Переписать `service.go` - полный workflow
29. ✅ Реализовать `repository.go` - БД операции
30. ✅ Обновить `handler.go` - улучшенный error handling

### Этап 9: Тестирование (5 дней)
31. ✅ Integration тесты для service
32. ✅ E2E тесты для API
33. ✅ Тесты валидации
34. ✅ Тесты state management

### Этап 10: Оптимизация (2 дня)
35. ✅ Добавить кэширование
36. ✅ Оптимизировать БД запросы
37. ✅ Добавить metrics

**ИТОГО: ~31 день разработки**

---

## КРИТИЧЕСКИЕ ЗАМЕЧАНИЯ

### 1. Зависимость от Shared Service

**Проблема:** Matematika требует данные из Shared:
- Праздники
- Gateways список
- B2B клиенты
- Contractor шаблоны

**Решение:**
- Shared Service ДОЛЖЕН быть реализован параллельно
- Создать mock Shared для тестирования Matematika
- Добавить fallback на hardcoded данные если Shared недоступен

### 2. State между месяцами

**Проблема:** Recurring параметры (шлюз, день недели ПО, суммы мобильной) должны сохраняться.

**Решение:**
- Создать таблицу `statement_state` в БД
- При генерации первого месяца: создать state
- При генерации следующих: загрузить state

### 3. Обработка нескольких месяцев

**Вопрос:** Если `months: 3`, генерировать в одном запросе или отдельно?

**Рекомендация:**
- Генерировать ПО ОДНОМУ месяцу за раз
- Если `months: 3`:
  - Создать 3 отдельных statements в БД
  - Вернуть response с тремя ключами:
    ```json
    {
      "JANUARY 2025": {...},
      "FEBRUARY 2025": {...},
      "MARCH 2025": {...}
    }
    ```
- Каждый месяц зависит от предыдущего (finalBalance → initialBalance)

### 4. Асинхронность

**Важно:** Генерация 3 месяцев может занять 30-90 секунд.

**Решение:**
- POST возвращает `202 Accepted` сразу
- Расчеты через Kafka consumer или фоновый worker
- Клиент polling через GET /statement/{id}/status

---

## ПРИОРИТЕТНАЯ ПОСЛЕДОВАТЕЛЬНОСТЬ

### MVP (Minimum Viable Product) - 10 дней
1. ✅ Модели (CompanyInfo, Financials, CustomData)
2. ✅ Базовая валидация
3. ✅ Генератор дат (без праздников пока)
4. ✅ Простой генератор доходов B2C (hardcoded шлюз)
5. ✅ Простой генератор расходов (5 основных категорий)
6. ✅ Нормализация балансов
7. ✅ Builders
8. ✅ Repository (сохранение в БД)
9. ✅ Integration в service.go

### V1 (Production Ready) - 21 день дополнительно
10. ✅ Клиент для Shared Service
11. ✅ Праздники
12. ✅ Генератор B2B доходов
13. ✅ Все 15 категорий расходов
14. ✅ State management
15. ✅ Manual транзакции
16. ✅ Custom контрагенты
17. ✅ Масштабирование
18. ✅ Disable категории
19. ✅ Тесты

**Начни с Этапа 1!**

