# Три правила простого проектного дизайна

-- **-избавляться от точек генерации исключений**, запрещая соответствующее ошибочное поведение на уровне интерфейса класса;

1. `Email and Sender`

```go
// у нас в приложении необходимо отправлять сообщения по SMTP
// соответственно у нас есть метод Send с передаваемым email

// email.go
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var ErrInvalidEmail = errors.New("email is invalid")

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if emailRegex.MatchString(email) {
		return Email{value: email}, nil
	}
	
	return Email{}, ErrInvalidEmail
}

func (e *Email) Value() string {
	return e.value
}

// sender.go
// этот модуль имплементирует интерфейс отправки сообщения через SMTP
type SmtpSender struct {
	conn *http.Smtp
}

func (s *Sender) Send(source Email, target Email) error {
	return s.conn.Send(source.Value(), target.Value())
}

// в этом примере мы исключили вариант создания некорректного Email
// валидируем в конструкторе
```

2. `Connection and state machine`

```go
type ClosedConnection struct {}

func NewClosedConnection() ClosedConnection{}
func (cc ClosedConnection) Open() (OpenedConnection, error) {}

type OpenedConnection struct {}

func (oc OpenedConnection) Close() (ClosedConnection, error) {}
func (oc OpenedConnection) Write() error {}
func (oc OpenedConnection) Read() ([]byte, error) {}

// здесь мы реализуем автомат, который перекидывает из одного состояния в другое
// если бы мы, например, просто реализовали тип Connection со всеми методами,
// тогда мы могли ошибочно вызвать Read(), Write(), Close() - 
// у закрытого соединения,
// а мы как раз должны такого избегать, теперь пользователь вызвать 
// строго те методы, которые только ему доступны, пример:

func fetchData(addr string) ([]byte, error) {
	closed := NewClosedConnection()
	// closed -> ClosedConnection, только метод Open
	
	opened, err := closed.Open()
	if err != nil {
		return nil, err
	}
	
	defer func () {
		_, err := opened.Close()
		if err != nil {
			logger.Warn("some accident with closing conn")
		}
		
		logger.Info("closing conn is succesfull")
	}()
	// open -> OpenedConnection, можем использовать все методы

	data, err := opened.Read()
	if err != nil {
		return nil, err
	}

	return data, nil
}
```

-- **отказаться от дефолтных конструкторов без параметров**, и передавать конструктору обязательные аргументы;

1. `User`

```go
type UUID struct {}
type Email struct {}
type Age struct {}

type User struct {
	id UUID
	name string
	email Email
	age Age
}

var ErrInvalidId = errors.New("id is invalid")
var ErrInvalidEmail = errors.New("email is invalid")
var ErrInvalidAge = errors.New("age is invalid")

func NewUser(
	id UUID,
	email Email,
	age Age,
	name string
) (*User, error) {
	if !id.Valid() {
		return nil, ErrInvalidId
	}
	
	if !email.Valid() {
		return nil, ErrInvalidEmail
	}
	
	if !age.Valid() {
		return nil, ErrInvalidAge
	}
	
	return &User{
		id: id,
		email: email,
		age: age,
		name: name,
	}, nil
}

// здесь мы строго даем набор обязательных аругментов, которые в частисности
// валидируются, что не даст право на ошибку при создании пользователя
```

2. `Price`

```go
type Price struct {
	amount Cost
}

var ErrInvalidCost = errors.New("cost is invalid")

func NewPrice(amount Cost) (Price, error) {
	if !amount.Valid {
		return Price{}, ErrInvalidCost
	}
	
	return Price{amount: amount}, nil
}

// в данном примере мы создаем в конструкторе, цену для товара
// если мы передавали необязательные аргументы, тогда могли
// создать некорректную цену для опредленного товара
```

-- **избегать увлечения примитивными типами данных**, разрабатывать прикладную систему типов, на смысловом уровне моделирующую предметную область (используйте типы данных Клетка и Фигура, а не строки или числа).

1. `RateLimiter`

```go
type Cost struct {
	value int
}

func NewCost(value int) (Cost, error) {
	cost := Cost{value: value}
	if err := cost.Validate(); err != nil {
		return Cost{}, err
	}
	return cost, nil
}

func (c Cost) Validate() error {
	if c.value <= 0 {
		return ErrNonPositiveCost
	}
	return nil
}

func (c Cost) Value() int {
	return c.value
}
```

```go
// IdentityKind — вид идентификатора клиента.
type IdentityKind string

const (
	KindUser      IdentityKind = "user"
	KindDevice    IdentityKind = "device"
	KindSession   IdentityKind = "session"
	KindPhoneHash IdentityKind = "phone_hash"
	KindIP        IdentityKind = "ip"
)

const ipv6PrefixBits = 64

type Identity struct {
	kind  IdentityKind
	value string
}

func NewUserIdentity(userID string) (Identity, error) {
	return newIdentity(KindUser, userID)
}

func NewDeviceIdentity(deviceID string) (Identity, error) {
	return newIdentity(KindDevice, deviceID)
}

func NewSessionIdentity(sessionID string) (Identity, error) {
	return newIdentity(KindSession, sessionID)
}

func NewPhoneHashIdentity(hash string) (Identity, error) {
	return newIdentity(KindPhoneHash, hash)
}

func NewIPIdentity(ip net.IP) (Identity, error) {
	if ip == nil {
		return Identity{}, ErrEmptyIdentityValue
	}

	value := ip.String()
	if ip.To4() == nil {
		value = ip.Mask(net.CIDRMask(ipv6PrefixBits, net.IPv6len*8)).String()
	}
	return newIdentity(KindIP, value)
}

func newIdentity(kind IdentityKind, value string) (Identity, error) {
	identity := Identity{kind: kind, value: value}
	if err := identity.Validate(); err != nil {
		return Identity{}, err
	}
	return identity, nil
}

func (i Identity) Validate() error {
	switch i.kind {
	case KindUser, KindDevice, KindSession, KindPhoneHash, KindIP:
	default:
		return ErrUnknownIdentityKind
	}
	if i.value == "" {
		return ErrEmptyIdentityValue
	}
	return nil
}

func (i Identity) Kind() IdentityKind {
	return i.kind
}

func (i Identity) Value() string {
	return i.value
}

func (i Identity) IsZero() bool {
	return i.kind == "" && i.value == ""
}

func (i Identity) String() string {
	return string(i.kind) + ":" + i.value
}

```

```go
type Policy struct {
	bucketName      string
	capacity        int
	refillRate      float64
	defaultCost     Cost
	rolloutStrategy RolloutStrategy
	failureStrategy FailureStrategy
}

func newPolicy(
	name string,
	capacity int,
	refillRate float64,
	defaultCost int,
	rollout RolloutStrategy,
	failure FailureStrategy,
) Policy {
	cost, err := NewCost(defaultCost)
	if err != nil {
		panic("ratelimit: invalid default cost for policy " + name + ": " + err.Error())
	}

	policy := Policy{
		bucketName:      name,
		capacity:        capacity,
		refillRate:      refillRate,
		defaultCost:     cost,
		rolloutStrategy: rollout,
		failureStrategy: failure,
	}
	if err := policy.Validate(); err != nil {
		panic("ratelimit: invalid policy " + name + ": " + err.Error())
	}
	return policy
}

func (p Policy) Validate() error {
	if p.bucketName == "" {
		return ErrEmptyPolicyName
	}
	if p.capacity <= 0 {
		return ErrNonPositiveCapacity
	}
	if p.refillRate < 0 {
		return ErrNegativeRefillRate
	}
	if p.rolloutStrategy == nil {
		return ErrNilRolloutStrategy
	}
	if p.failureStrategy == nil {
		return ErrNilFailureStrategy
	}
	return nil
}

var (
	PolicyDefault       = newPolicy("default", 100, 10, 1, EnforceStrategy{}, FailOpenStrategy{})
	PolicyAuthSensitive = newPolicy("auth_login", 5, 0.1, 1, EnforceStrategy{}, FailClosedStrategy{})
	PolicyAnonymous     = newPolicy("anonymous", 30, 3, 1, EnforceStrategy{}, FailOpenStrategy{})
	PolicyUnlimited     = newPolicy("unlimited", 1000000, 1000000, 1, EnforceStrategy{}, FailOpenStrategy{})

	PolicyShadow = newPolicy("shadow_demo", 100, 10, 1, ShadowStrategy{}, FailOpenStrategy{})
)

func (p Policy) Name() string {
	return p.bucketName
}

func (p Policy) Cap() int {
	return p.capacity
}

func (p Policy) RefillRate() float64 {
	return p.refillRate
}

func (p Policy) DefaultCost() Cost {
	return p.defaultCost
}

func (p Policy) RolloutStrategy() RolloutStrategy {
	return p.rolloutStrategy
}

func (p Policy) FailureStrategy() FailureStrategy {
	return p.failureStrategy
}

func (p Policy) RolloutMode() RolloutMode {
	return p.rolloutStrategy.Mode()
}

func (p Policy) FailureMode() FailureMode {
	return p.failureStrategy.Mode()
}

func (p Policy) IsZero() bool {
	return p.bucketName == ""
}

```

```go
type RateLimitRequest struct {
	Identity Identity
	Cost     Cost
	Policy   Policy
}

// в данном примере чать реализации RateLimiter, где каждую из составляющих
// запроса - я оборачивал в собственный тип, что нам это дает:
// 1. ограниченность в создании и управлении
// 2. валидация на уровне типа
// 3. ограниченность в типах в принципе, не даем создать свои типы
// или использовать примитивы
```

2. `Email`

```go
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var ErrInvalidEmail = errors.New("email is invalid")

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if emailRegex.MatchString(email) {
		return Email{value: email}, nil
	}
	
	return Email{}, InvalidEmail
}

func (e *Email) Value() string {
	return e.value
}

// из первого примера возьмем Email, который также оборачиваем в собсвтенный тип
// и не используем примитивы
```