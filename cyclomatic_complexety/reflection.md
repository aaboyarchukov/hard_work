# Цикломатическая Сложность

В данном уроке мы узнаем:
- что такое Цикломатическая сложность (ЦС)
- узнаем способы ее снижения в проекте и покажем на примерах

### Цикломатическая Сложность

**Цикломатическая Сложность (ЦС)** - метрика, которая показывает насколько запутанной является функция. Ее показатель фактически говорит о том, сколькими способами может быть обработана данная функция.

Но иногда и слишком маленькая ЦС невозможна, так как требуется довольно сложная алгоритмическая логика.

*Правило:* **стремиться к ЦС, равной 1, надо тем сильнее и полнее, чем дальше мы уходим в проектирование и бизнес-логику от чистого алгоритмического кодинга**.

Чем выше алгоритмический кодинг (и сложнее) - **тем больше ЦС**, и чем плотнее проектирование и безнес-логика - **тем меньше ЦС**


### Приемы снижения ЦС:

Правила:
1. Никаких условных операторов и switch/case
2. Никаких циклов (for, while)
3. Никаких null/None (условная проверка на nullable значения) -> лучше использовать nullable типы и операторы, которые поддерживают работу с таким типом

Принцип:
1. Open-Closed Principle:
	**Модуль считается открытым, если его можно продолжать расширять.** 
	**Модуль считается закрытым, когда он выложен в продакшен, и им можно только пользоваться.**
То есть чем выше ЦС - тем более закрытым и сложным для расширения становится модуль.

Для избавления от условных операторов можно использовать решение, реализованные в стандартных библиотеках ЯП - компактные формы условий.

Часто распространяющееся ошибка:

```python
if условие
   return true
else
   return false
   
==============>

return условие
```

**Причем от else можно избавиться всегда!!!!**

Всегда надо стараться продумывать логику работы системы так, чтобы компоненты системы были максимально открыты)

### Шаги к решению

1. Полиморфизм
	1.  Параметрический
	2. ad-hoc (специальный)
	3. Полиморфизм подтипов

- Параметрический (generics):
Данный метод позволяет работать с разными типами
```go
func AppendItems[T any](a, b T)
```


- ad-hoc полиморфизм
В данном методе мы перегружаем один и тот же метод - разными типами

```go
func AppendItems(a, b string ) {
	return fmt.Sprintf("%s-%s", a, b) 
}

func AppendItems(a, b int) {
	return a + b
}

==========================>

// В Go нет перегрузки, поэтму можно использовать TypeClass Pattern

type Appender[T any] interface {
	Append(a, b T) T
}

type StringAppender struct {}

func (sa *StringAppender) Append(a, b string) string {
	return fmt.Sprintf("%s-%s", a, b)
}

type IntAppender struct {}

func (ia *IntAppender) Append(a, b int) int {
	return a + b
}

// общая функция
func AppendItems[T any](a, b T, impl Appender[T]) T {
    return impl.Append(a, b)
}

func main() {
	intResult := AppendItems(1, 2, IntAppender{})
	// 3
	
	stringResult := AppendItems("1", "2", StringAppender{})
	// "1-2"
}
```

Таким образом мы в функцию передаем абсолютно разные типы и имплементацию так называемого контейнерного типа для них, который реализует интерфейс `Appender`. Тем самым мы получаем снижение ЦС.


- Полиморфизм подтипов (утиная типизация)
```go
type Animal interface {
	Sound()
}

type Cow struct {}

func (c *Cow) Sound()

type Cat struct {}

func (c *Cat) Sound()
```

Вывод:
> 	**если в некотором методе имеется условие, по которому выполняется качественно разный код, то тут полезно применить ad-hoc полиморфизм: вынести каждую ветку условия в отдельный метод (у всех у них будет одно и то же имя), который будет вызываться для каждого варианта со своей реализацией и со своим уникальным набором параметров, который у оригинально метода чрезмерен: охватывает все варианты работы в зависимости от условий, но при этом реально будут использованы далеко не все аргументы. Очевидно, в такой схеме автоматически снизится и ЦС.**

2. Автомат состояний - работа методов зависит от текущего состояния!

Паттерн: [State pattern](https://refactoring.guru/ru/design-patterns/state)
Суть данного паттерна в том, что программа может находиться в одном из нескольких состояний, которые всё время сменяют друг друга.

В данном паттерне предлагается каждое новое состояние вынести в отдельный класс, где определить поведение именно этого состояния! Тогда при переходе очередного состояния можно не использовать условные операторы и снизить ЦС)

Покажем на примере Работника:

Состояния:
- В работе
- На обеде
- На больничном
- Закончил работу

Действия:
- Пойти на обед
- Прийти на работу
- Выйти на больничный
- Уйти с работы

```go
type Employer struct {
	OnWork State
	EndWork State
	OnSickLeave State
	OffSickLeave State
	
	currentState State
	
}

func (e *Employer) setState(state State) {
	e.currentState = state
}

type State interface {
	goWork() error
	endWork() error
	goOnSickLeave() error
	goOffSickLeave() error
}

// для каждого из состояний реализуем интерфейс
// и соответствующие переходы между состояниями
// например: если раюотник на больничном он не может выйти на работу
// пока не перейдет в состояние выйти с больничного
type OnWorkState struct {
	employer *Employer
}

type EndWorkState struct {
	employer *Employer
}

type OnSickLeaveState struct {
	employer *Employer
}

// реализуем данный интерфейс
type OffSickLeaveState struct {
	employer *Employer
}

func (osl *OffSickLeaveState) goWork() error {
	return fmt.Errorf("you are sick!")
}

func (osl *OffSickLeaveState) endWork() error {
	return fmt.Errorf("you are sick!")
}

func (osl *OffSickLeaveState) goOnSickLeave() error {
	osl.employer.setState(osl.employer.OnSickLeave)
	// othre logic
	return nil
}

func (osl *OffSickLeaveState) goOffSickLeave() error {
	osl.employer.setState(osl.employer.OnWork)
	// othre logic
	return nil
}

```

Также стоит упомянуть **NullObject Pattern** - суть его в том, чтобы создать отдельный класс по работе с NullObject нашего текущего объекта, чтобы лишиться лишних проверок внутри кода.

Паттерн: [NullObject Pattern](https://refactoring.guru/introduce-null-object)
Пример использования:

```go
type Logger interface {
	Info(msg string)
	Error(msg string) 
}

type STDLogger struct {
	logger *logger
}

func (std *STDLogger) Info(msg string) {
	std.logger.Info(msg)
}

func (std *STDLogger) Error(msg string) {
	std.logger.Error(msg)
}

type NullLogger struct {}

func (n *NullLogger) Info(msg string) {}

func (n *NullLogger) Error(msg string) {}


// some service
func NewService(logger Logger) *Service {
    if logger == nil {
        logger = NullLogger{} // Null Object
    }

    return &UserService{
        logger: logger,
    }
}

func main() {
	serviceWithLogger := NewService(&STDLogger{
		logger: zerolog.New()
	},)
	
	serviceWithLogger.Call() // some logic -> logging
	
	serviceWithoutLogger := NewService(&NullLogger{})
	
	serviceWithoutLogger.Call() // some logic -> no logging
}
```

В итоге у нас получилась всего лишь одна проверка, в противном случае понадобилось бы вызывать проверки внутри имплементированных методов для сервиса, все время проверяя на `nil` наш логгер

В некоторых языках программирования (в основном функциональных) существует тип Optional ->который возвращает некоторое значение, с которым можно валидно работать.

3. Table-Driven Logic (табличная логика)

С помощью данного подхода мы используем структуру с полями для работы с  входными данными, например, нам необходимо валидировать входные данные:

Без паттерна:

```go
if field == "email" {
    validate email
} else if field == "age" {
    validate age
} else if field == "phone" { ... }

```

С паттерном:

```go
type FieldRule struct {
    Field     string  `json:"field"`
    Regex     string  `json:"regex"`
    Min       *int    `json:"min"`
    Max       *int    `json:"max"`
    Required  bool    `json:"required"`
}

var rules = []FieldRule{
    {
        Field:    "email",
        Regex:    `^.+@.+$`,
        Min:      nil,
        Max:      nil,
        Required: true,
    },
    {
        Field:    "age",
        Regex:    `^[0-9]+$`,
        Min:      intPtr(0),
        Max:      intPtr(120),
        Required: false,
    },
    {
        Field:    "username",
        Regex:    `^[a-z0-9_]+$`,
        Min:      intPtr(3),
        Max:      intPtr(20),
        Required: true,
    },
}

func ValidateField(value string, rule FieldRule) error {
    // 1. Проверка required
    if rule.Required && value == "" {
        return fmt.Errorf("field %s is required", rule.Field)
    }

    // 2. Проверка regex
    if rule.Regex != "" {
        if ok, _ := regexp.MatchString(rule.Regex, value); !ok {
            return fmt.Errorf("field %s does not match regex", rule.Field)
        }
    }

    // 3. Проверка min/max (если value — число)
    if rule.Min != nil || rule.Max != nil {
        v, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("field %s must be number", rule.Field)
        }

        if rule.Min != nil && v < *rule.Min {
            return fmt.Errorf("field %s < min %d", rule.Field, *rule.Min)
        }
        if rule.Max != nil && v > *rule.Max {
            return fmt.Errorf("field %s > max %d", rule.Field, *rule.Max)
        }
    }

    return nil
}

func ValidateForm(input map[string]string) error {
    for _, rule := range rules {
        value := input[rule.Field]
        if err := ValidateField(value, rule); err != nil {
            return err
        }
    }
    return nil
}


form := map[string]string{
    "email":    "test@example.com",
    "age":      "32",
    "username": "john_doe",
}

if err := ValidateForm(form); err != nil {
    fmt.Println("Validation error:", err)
} else {
    fmt.Println("OK")
}
```

С этим паттерном у нас отпадает необходимость в изменении логики валидации, ведь правила задаются в структуру, а для того, чтобы добавить новое правило - достаточно добавить новый элемент в структуру!

4. Стратегия (Strategy Pattern)
5. Цепочка обязанностей (Chain Pattern - Функциональная композиция)
6. Dependency Injection

### Мягкие правила

Более мягкая версия снижения ЦС может быть такой:

1. Запрещены else и любые цепочки else if.  
2. Запрещены if, вложенные в if, и циклы, вложенные в if.  
3. if внутри цикла допускается только один, и только для прерывания его работы (break/continue), выхода из функции (return) или генерации исключения.
4. Если внутри условия сложная логика -> вынос в отдельную функцию
5. В одном ветвлении обработка только одного аргумента функции, если обработка вообще есть
6. Соблюдать принцип SRP

### Дополнительные правила

1. Не определять методы, которые ничего не вычисляют
> Рекомендация такая, что в принципе можно допустить возвращение методами, меняющими состояние объекта, значения некоторого типа "статус" (код ошибки, условно); при этом геттеры, возвращающие статус операций, такжа продолжают работать как и раньше. Фактически мы просто добавляем в самый конец метода вызов соответствующего геттера.

2. Перейти к работе только с иммутабельными значениями
> Да, идея уже в том, что мы движемся в функциональное программирование, и методов, меняющих внутреннее состояние родного объекта, уже не должно быть. Однако они могут выполнять самые разные трансформации над текущим состоянием (которое само по себе неизменяемо, т.к. нету методов, модифицирующих атрибуты объекта), возвращая новый объект.

3. ФП => использование контейнерных типов и уход от использования циклов
4. Еще можно разработать свой микро-язык (DSL) под конкретную задачу, который на уровне компилятора будет обрабатывать определенные значения 

### Примеры:

1. Распаковка архива
Было:
```go
func IsArchive(file []byte) string {
    fileType, err_get_type := filetype.Match(file)
    if err_get_type != nil {

        return ""
    }

    return strings.ToLower(fileType.Extension)

}

func FindFileInArchive(file []byte) ([]byte, error) {
	return UnpackArchive(file)
}

func UnpackArchive(file []byte) ([]byte, error) {
	var result []byte
	switch archiveType := IsArchive(file); archiveType {
    case "zip":
        data, err_decode := DecodeZipArchive(file)
        if err_decode != nil {

            return nil, err_decode
        }
        result = data
    case "7z":
        data, err_decode := Decode7ZArchive(file)
        if err_decode != nil {

            return nil, err_decode
        }
        result = data
    case "rar":
        data, err_decode := DecodeRARArchive(file)
        if err_decode != nil {

            return nil, err_decode
        }
        result = data
    case "tar":
        data, err_decode := DecodeTARArchive(file)
        if err_decode != nil {

            return nil, err_decode
        }
        result = data
    }
    
    return result, nil
}
```

Стало:

```go
type Archivator interface {
	DecodeArhive(file []byte) ([]byte, error)
}

const arhiveTypes map[string]Archivator = map[string]Archivator {
	"zip": ZIPArhive{},
	"tar": TarArhive{},
	"rar": RARArhive{},
}

func FindFileInArchive(file []byte, ext string) ([]byte, error) {
	var resultArchivator Archivator = UnkonwnArchive{} 
	
	if archivator, ok := arhiveTypes[ext]; ok {
		resultArchivator = archivator 
	}
	
	return UnpackArchive(resultArchivator, file)
}

type ZIPArhive struct {}
func (zip *ZIPArhive) DecodeArhive(file []byte) ([]byte, error)

type TarArhive struct {}
func (tar *TarArhive) DecodeArhive(file []byte) ([]byte, error)

type RARArhive struct {}
func (rar *RARArhive) DecodeArhive(file []byte) ([]byte, error)

type UnkonwnArchive struct {}
func (nullable *UnkonwnArchive) DecodeArhive(file []byte) ([]byte, error) {
	// for nullable value
	return nil, errors.New("unkown format")
}

func UnpackArchive(arhive Archivator, file []byte) ([]byte, error) {
	return arhive.DecodeArhive(file)
}
```

**Заметка:**
В данном примере мы использовали полиморфизм на основе TypeClass Pattern, NullObject Pattern, а также использовали mapping для реализации паттерна Strategy, тем самым мы снизили ЦС с 4 до 1.

2. Обработка платежа
Было:

```go
func ProcessPayment(method string, amount float64, currency string, country string, userType string, device string) error {
    if amount <= 0 {
        return fmt.Errorf("invalid amount")
    }

    if currency == "USD" {
        if country == "US" {
            if method == "card" {
                // 1
                if userType == "guest" {
                    // 2
                    if amount > 1000 {
                        return fmt.Errorf("guest limit exceeded")
                    }
                } else if userType == "registered" {
                    // 3
                    if amount > 5000 {
                        return fmt.Errorf("registered limit exceeded")
                    }
                }
            } else if method == "paypal" {
                // 4
                if device == "mobile" {
                    if amount > 300 {
                        return fmt.Errorf("mobile paypal limit")
                    }
                } else {
                    if amount > 1000 {
                        return fmt.Errorf("paypal desktop limit")
                    }
                }
            }
        } else {
            if method == "crypto" {
                // 5
                if amount < 50 {
                    return fmt.Errorf("crypto minimum")
                }
                if amount > 20000 {
                    return fmt.Errorf("crypto max")
                }
            }
        }
    } else if currency == "EUR" {
        if country == "DE" {
            if method == "sepa" {
                if amount > 10000 {
                    return fmt.Errorf("SEPA limit")
                }
            }
        } else {
            if method == "card" {
                if amount > 2000 {
                    return fmt.Errorf("card limit EUR")
                }
            }
        }
    } else {
        if method == "crypto" {
            if amount > 50000 {
                return fmt.Errorf("crypto max global")
            }
        } else {
            return fmt.Errorf("unsupported method")
        }
    }

    // 6
    if device == "smart_tv" {
        return fmt.Errorf("TV not supported")
    }

    // 7
    if userType == "banned" {
        return fmt.Errorf("user banned")
    }

    return nil
}

```

Стало:
```go
// implement validator for string, int and etc.
type Validator interface {
	ValidateValue(value any) error
}

type RegexValidator struct {
	Pattern *regexp.Regexp
}

func (v RegexValidator) Validate(value any) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	if !v.Pattern.MatchString(s) {
		return fmt.Errorf("value '%s' does not match pattern", s)
	}
	return nil
}

type RangeValidator struct {
	Min, Max float64
}

func (v RangeValidator) Validate(value any) error {
	n, ok := value.(float64)
	if !ok {
		return fmt.Errorf("expected number, got %T", value)
	}

	if n < v.Min || n > v.Max {
		return fmt.Errorf("number %.2f is outside the range %.2f–%.2f", n, v.Min, v.Max)
	}
	return nil
}


// each method implement own payment
// with own strategies
type MethodPayment interface {
	Apply(payment PaymentContext) error
	Validate(tableStrategies map[string]Validator) error
	Validators() map[string]Validator
}

type PayPal struct{}

func (PayPal) Validators() map[string]Validator {
	return map[string]Validator{
		"userType": RegexValidator{
			Pattern: regexp.MustCompile(`^(regular|premium|vip)$`),
		},
	}
}

func (PayPal) Validate(payment PaymentContext) error {
	for field, validator := range PayPal{}.Validators() {

		var value any

		switch field {
		case "userType":
			value = payment.UserType
		default:
			return fmt.Errorf("unknown field %s", field)
		}

		if err := validator.Validate(value); err != nil {
			return fmt.Errorf("PayPal validation error (%s): %w", field, err)
		}
	}

	return nil
}

func (PayPal) Apply(payment PaymentContext) error {
	fmt.Println("Applying PayPal...")
	return nil
}


type DevicePayment interface {
	Apply(payment PaymentContext) error
	Validate(tableStrategies map[string]Validator) error
	Validators() map[string]Validator
}

type MobileDevice struct{}

func (MobileDevice) Validators() map[string]Validator {
	return map[string]Validator{
		"amount": RangeValidator{Min: 1, Max: 2000},
	}
}

func (MobileDevice) Validate(payment PaymentContext) error {
	for field, validator := range MobileDevice{}.Validators() {

		var value any

		switch field {
		case "amount":
			value = payment.Amount
		default:
			return fmt.Errorf("unknown field %s", field)
		}

		if err := validator.Validate(value); err != nil {
			return fmt.Errorf("MobileDevice validation (%s): %w", field, err)
		}
	}

	return nil
}

func (MobileDevice) Apply(payment PaymentContext) error {
	fmt.Println("Using MobileDevice...")
	return nil
}




type PaymentContext struct {
	Method MethodPayment
	Device DevicePayment
	Amount   float64
    Currency string
    Country  string
    UserType string
}

func ProcessPayment(payment PaymentContext) error {
	
	// 1. валидация девайса
	if err := payment.Device.Validate(payment.Device.Validators(), payment); err != nil {
		return fmt.Errorf("device validation failed: %w", err)
	}
	
	// 2. валидация метода
	if err := payment.Method.Validate(payment.Method.Validators(), payment); err != nil {
		return fmt.Errorf("method validation failed: %w", err)
	}


	// 3. логика метода
	if err := payment.Method.Apply(payment); err != nil {
		return fmt.Errorf("method apply failed: %w", err)
	}

	// 4. логика девайса
	if err := payment.Device.Apply(payment); err != nil {
		return fmt.Errorf("device apply failed: %w", err)
	}

	fmt.Println("Payment processed successfully!")
	return nil
}


```

**Заметка:**
В данном примере мы использовали полиморфизм на основе TypeClass Pattern, также использовали mapping для реализации паттерна Strategy, паттерн Chain of Responsibility ну и Dependency Injection, тем самым мы снизили ЦС с >20 до ~4 в итоговой функции.

3. Пример, который смотрели выше

Было:

```go
type FieldRule struct {
    Field     string  `json:"field"`
    Regex     string  `json:"regex"`
    Min       *int    `json:"min"`
    Max       *int    `json:"max"`
    Required  bool    `json:"required"`
}

var rules = []FieldRule{
    {
        Field:    "email",
        Regex:    `^.+@.+$`,
        Min:      nil,
        Max:      nil,
        Required: true,
    },
    {
        Field:    "age",
        Regex:    `^[0-9]+$`,
        Min:      intPtr(0),
        Max:      intPtr(120),
        Required: false,
    },
    {
        Field:    "username",
        Regex:    `^[a-z0-9_]+$`,
        Min:      intPtr(3),
        Max:      intPtr(20),
        Required: true,
    },
}

func ValidateField(value string, rule FieldRule) error {
    // 1. Проверка required
    if rule.Required && value == "" {
        return fmt.Errorf("field %s is required", rule.Field)
    }

    // 2. Проверка regex
    if rule.Regex != "" {
        if ok, _ := regexp.MatchString(rule.Regex, value); !ok {
            return fmt.Errorf("field %s does not match regex", rule.Field)
        }
    }

    // 3. Проверка min/max (если value — число)
    if rule.Min != nil || rule.Max != nil {
        v, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("field %s must be number", rule.Field)
        }

        if rule.Min != nil && v < *rule.Min {
            return fmt.Errorf("field %s < min %d", rule.Field, *rule.Min)
        }
        if rule.Max != nil && v > *rule.Max {
            return fmt.Errorf("field %s > max %d", rule.Field, *rule.Max)
        }
    }

    return nil
}

func ValidateForm(input map[string]string) error {
    for _, rule := range rules {
        value := input[rule.Field]
        if err := ValidateField(value, rule); err != nil {
            return err
        }
    }
    return nil
}


form := map[string]string{
    "email":    "test@example.com",
    "age":      "32",
    "username": "john_doe",
}

if err := ValidateForm(form); err != nil {
    fmt.Println("Validation error:", err)
} else {
    fmt.Println("OK")
}
```

Стало:

```go
type FieldRule struct {
    Field     string  `json:"field"`
    Regex     string  `json:"regex"`
    Min       *int    `json:"min"`
    Max       *int    `json:"max"`
    Required  bool    `json:"required"`
}

var rules = []FieldRule{
    {
        Field:    "email",
        Regex:    `^.+@.+$`,
        Min:      nil,
        Max:      nil,
        Required: true,
    },
    {
        Field:    "age",
        Regex:    `^[0-9]+$`,
        Min:      intPtr(0),
        Max:      intPtr(120),
        Required: false,
    },
    {
        Field:    "username",
        Regex:    `^[a-z0-9_]+$`,
        Min:      intPtr(3),
        Max:      intPtr(20),
        Required: true,
    },
}

type FieldValidator func(value string, rule FieldRule) error

const validators = []FieldValidator{
    validateRequired,
    validateRegex,
    validateMinMax,
}

func validateRequired(value string, rule FieldRule) error {
    if !rule.Required {
        return nil
    }
    if value == "" {
        return fmt.Errorf("field %s is required", rule.Field)
    }
    return nil
}

func validateRequired(value string, rule FieldRule) error {
    if !rule.Required {
        return nil
    }
    if value == "" {
        return fmt.Errorf("field %s is required", rule.Field)
    }
    return nil
}

func validateMinMax(value string, rule FieldRule) error {
    if rule.Min == nil && rule.Max == nil {
        return nil
    }
    v, err := strconv.Atoi(value)
    if err != nil {
        return fmt.Errorf("field %s must be a number", rule.Field)
    }

    if rule.Min != nil && v < *rule.Min {
        return fmt.Errorf("field %s < min %d", rule.Field, *rule.Min)
    }
    if rule.Max != nil && v > *rule.Max {
        return fmt.Errorf("field %s > max %d", rule.Field, *rule.Max)
    }
    return nil
}

func ValidateField(value string, rule FieldRule) error {
    for _, v := range validators {
        if err := v(value, rule); err != nil {
            return err
        }
    }
    return nil
}

func ValidateForm(input map[string]string) error {
    for _, rule := range rules {
        value := input[rule.Field]
        if err := ValidateField(value, rule); err != nil {
            return err
        }
    }
    return nil
}
```

**Заметка:**
В данном примере мы использовали Table Driven Pattern для реализации mapping различных правил, также мы вынесли валидацию полей в отдельную функцию в которой применяется паттерн Strategy Pattern, тем самым мы снизили ЦС, убрали выложенные if и перешли на линейные циклы.
