# Неочевидные проектные ошибки

В материале были рассмотрены проектные ошибки, на которые необходимо обращать внимание, для того, чтобы выстраивался проект, который можно легко поддерживать.

## 1. Отладка, логи, проверки

Приведем 5 примеров, когда в работе страховали код, за счет чрезмерного логирования:

В данном примере на низком уровне - уровне репозитория, мы логируем ошибку, при условии того, что мы будем логировать ее выше - мы получим стектрейс из логов, что не хочется видеть, к тому же - если мы в логах будем указывать в явном виде, что за ошибка - это небезопасно. Как видно, здесь идет чрезмерное логирование и утечка.

Было:

```go
func (r *Repository) FileTypeNamesById(ctx context.Context, billID *domain.BillID) ([]string, error) {
	const op = "repository.bills.FileTypeNamesById"

	log.Info("entering method", "op", op, "billID", *billID)
	log.Debug("preparing query", "op", op, "query", "GetBillFileTypeNames")

	fileTypeNames, err := r.queries.GetBillFileTypeNames(ctx, *billID)

	if err != nil {
		log.Error("error details", "op", op, "type", fmt.Sprintf("%T", err), "msg", err.Error())
		log.Error("stack", "op", op, "stack", string(debug.Stack()))
		return nil, postgres.MapError(err, op)
	}

	r.log.Info("query succeeded", "op", op, "count", len(fileTypeNames))

	return fileTypeNames, nil
}
```

Теперь мы просто возвращаем ошибку, в которой есть также имя операции, чтобы при разборе понять, где именно проблема.

Стало:

```go
func (r *Repository) FileTypeNamesById(ctx context.Context, billID *domain.BillID) ([]string, error) {
	const op = "repository.bills.FileTypeNamesById"
	fileTypeNames, err := r.queries.GetBillFileTypeNames(ctx, *billID)
	if err != nil {
		return nil, postgres.MapError(err, op)
	}

	return fileTypeNames, nil
}
```

Следующий пример показывает, что мы очень много проверяем через `fmt.Println()`, что также является антипаттерном - лучше по максимуму испольовать тесты, а также при необходимости - есть дебаггер.

Было:

```go
// Аутентификация
func (s *Service) Login(ctx context.Context, email, password string) (domain.SID, error) {
	fmt.Println("Login start, email:", email)

	creds, err := s.UserRepository.AuthInfo(ctx, email)
	if err != nil {
		fmt.Println("AuthInfo err:", err)
		return "", fmt.Errorf("get user credentials: %w", err)
	}
	fmt.Println("got user id:", creds.ID)

	if err = s.verifyPassword(creds.PasswordHash, password); err != nil {
		fmt.Println("password mismatch for", creds.ID)
		return "", domain_errors.ErrInvalidCredentials
	}

	sid, err := s.Sessions.CreateSession(ctx, creds.ID)
	if err != nil {
		fmt.Println("CreateSession err:", err)
		return "", fmt.Errorf("create session: %w", err)
	}

	fmt.Println("Login ok for", creds.ID)
	return sid, nil
}
```

Стало:

```go
func (s *Service) Login(ctx context.Context, email, password string) (domain.SID, error) {
	creds, err := s.UserRepository.AuthInfo(ctx, email)
	if err != nil {
		return "", fmt.Errorf("get user credentials: %w", err)
	}

	if err = s.verifyPassword(creds.PasswordHash, password); err != nil {
		return "", domain_errors.ErrInvalidCredentials
	}

	sid, err := s.Sessions.CreateSession(ctx, creds.ID)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return sid, nil
}
```

Следующий пример

Было:

```go
package notifications

type senderService interface {
	Send(ctx context.Context, recipient, message string) error
}

type sendNotification struct {
	email senderService
	sms   senderService
	push  senderService
}

func newSendNotification(email, sms, push senderService) *sendNotification {
	return &sendNotification{email: email, sms: sms, push: push}
}

func (h *sendNotification) handle(ctx context.Context, request api.SendNotificationRequestObject) (api.SendNotificationResponseObject, error) {
	var err error

	switch request.Body.Channel {
	case api.SendNotificationRequestChannelEmail:
		err = h.email.Send(ctx, request.Body.Recipient, request.Body.Message)
	case api.SendNotificationRequestChannelSms:
		err = h.sms.Send(ctx, request.Body.Recipient, request.Body.Message)
	case api.SendNotificationRequestChannelPush:
		err = h.push.Send(ctx, request.Body.Recipient, request.Body.Message)
	default:
		return api.SendNotification400JSONResponse{Message: "unknown channel"}, nil
	}

	switch {
	case err == nil:
		return api.SendNotification204Response{}, nil
	case errors.Is(err, domain_errors.ErrInvalidRecipient):
		return api.SendNotification400JSONResponse{Message: err.Error()}, nil
	default:
		return nil, err
	}
}
```

Стало:

```go
package notifications

type notificationsService interface {
	Send(ctx context.Context, ch domain.Channel, recipient, message string) error
}

type sender struct {
	service notificationsService
	channel domain.Channel
}

func (s *sender) send(ctx context.Context, recipient, message string) error {
	return s.service.Send(ctx, s.channel, recipient, message)
}
```

Теперь мы просто избавляемся от выбора на множестве на уровне HTTP слоя, у нас просто роутятся запросы сами в зависимости от значения, а мы просто вызываем саму логику.

В следующем примере мы видим также много вывода, но к тому же есть и лишние проверки, следовательно от них необходимо избавиться.

Было:

```go
func (s *Service) checkPassword(password string) *dto.PasswordValidationResult {
	fmt.Println("checkPassword, len:", utf8.RuneCountInString(password))

	if password == "" {
		fmt.Println("password is empty, but continuing anyway")
	}

	passwordEntropy := password_validator.GetEntropy(password)
	fmt.Println("entropy:", passwordEntropy)

	strengthPercent := min(int((passwordEntropy/float64(maxPasswordEntropyBits))*100), 100)
	minStrengthPercent := int((float64(minPasswordEntropyBits) / float64(maxPasswordEntropyBits)) * 100)
	fmt.Println("strength:", strengthPercent, "min:", minStrengthPercent)

	result := &dto.PasswordValidationResult{
		StrengthPercent:    strengthPercent,
		MinStrengthPercent: minStrengthPercent,
		IsSuccess:          true,
		Reason:             nil,
	}

	setError := func(reason error) *dto.PasswordValidationResult {
		fmt.Println("setError:", reason)
		result.IsSuccess = false
		result.Reason = reason
		return result
	}

	if utf8.RuneCountInString(password) < minPasswordLength {
		return setError(domain_errors.ErrPasswordTooShort)
	}

	for _, r := range password {
		if r > unicode.MaxASCII {
			return setError(domain_errors.ErrPasswordNonASCII)
		}
	}
	fmt.Println("ascii ok")

	if err := password_validator.Validate(password, minPasswordEntropyBits); err != nil {
		fmt.Println("validate err:", err)
		return setError(domain_errors.ErrPasswordWeakEntropy)
	}

	fmt.Println("password ok, strength:", strengthPercent)
	return result
}
```

Стало:

```go
func (s *Service) checkPassword(password string) *dto.PasswordValidationResult {
	passwordEntropy := password_validator.GetEntropy(password)

	strengthPercent := min(int((passwordEntropy/float64(maxPasswordEntropyBits))*100), 100)

	minStrengthPercent := int((float64(minPasswordEntropyBits) / float64(maxPasswordEntropyBits)) * 100)

	result := &dto.PasswordValidationResult{
		StrengthPercent:    strengthPercent,
		MinStrengthPercent: minStrengthPercent,
		IsSuccess:          true,
		Reason:             nil,
	}

	setError := func(reason error) *dto.PasswordValidationResult {
		result.IsSuccess = false
		result.Reason = reason
		return result
	}

	if utf8.RuneCountInString(password) < minPasswordLength {
		return setError(domain_errors.ErrPasswordTooShort)
	}

	for _, r := range password {
		if r > unicode.MaxASCII {
			return setError(domain_errors.ErrPasswordNonASCII)
		}
	}

	if err := password_validator.Validate(password, minPasswordEntropyBits); err != nil {
		return setError(domain_errors.ErrPasswordWeakEntropy)
	}

	return result
}
```

Следующий пример также показывает, что вывод данных будет здесь лишним, также уменьшили цикломатическую сложность.

Было:

```go
func (r *Repository) Get(ctx context.Context, email string) (string, error) {
	const op = "redis.otp.Get"

	key := otpKey(email)
	fmt.Println(op, "start, email:", email)
	fmt.Println(op, "key:", key)

	otp, err := r.client.Get(ctx, key).Result()

	if err != nil {
		if redis.IsNil(err) {
			fmt.Println(op, "OTP NOT FOUND for", email)
			return "", fmt.Errorf("%s: %w", op, domain_errors.ErrOTPNotFound)
		}
		fmt.Printf("%s: unexpected err type: %T\n", op, err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println(op, "found, returning")
	return otp, nil
}
```

Стало:

```go
func (r *Repository) Get(ctx context.Context, email string) (string, error) {
	const op = "redis.otp.Get"

	otp, err := r.client.Get(ctx, otpKey(email)).Result()

	if redis.IsNil(err) {
		return "", fmt.Errorf("%s: %w", op, domain_errors.ErrOTPNotFound)
	}

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return otp, nil
}
```

## 2. Рефакторинг – это не коробка для мусора

Многие представляют процесс рефакторинга, как перетаскивание кусочков кода, в различные модули, объединяя их, или декомпозируя. Но о рефакториге важно думать как о способе нахождения общих связей и их последующего упрощения, для того, чтобы по итогу получилось простое приложение, которое будет легко сопровождать в будущем. То есть необходимо очень сильно задумываться именно о простом структурировании кода, о его дизайне, таким образом, чтобы система получалась максимально структуризированной, эффективной и простой.

### Примеры

Отвязал конкретные реализации хендлеров и перешел к интерфейсам, также поменял структуру в проекте - разделив на домены хендлеры.

До:

Структура выглядела вот так:

```
internal/server/handlers
├── apispec.go
├── apispec_test.go
├── assets
│   ├── qualification_data.json
│   ├── veksel_terms_rub_v1.json
│   └── veksel_terms_usd_v1.json
├── auth.go
├── auth_test.go
├── chat.go
├── chat_test.go
├── common.go
├── common_test.go
```

То есть все домены смешаны, а также все привязывалось к реализации:

```

```

```go
// internal/server/handlers/
package handlers

// Auth содержит хендлеры для аутентификации.
type Auth struct {
	authClient *authclient.Client
}

// NewAuth создает новый экземпляр Auth хендлеров.
func NewAuth(authClient *authclient.Client) *Auth

// LoginByPhone вход и/или регистрация исключительно по телефону.
func (auth *Auth) LoginByPhone(w http.ResponseWriter, r *http.Request, _ api.LoginByPhoneParams)

// LoginApprove подтверждение входа/регистрации (финальный шаг).
func (auth *Auth) LoginApprove(w http.ResponseWriter, r *http.Request, _ api.LoginApproveParams)

// Logout выход из системы.
func (auth *Auth) Logout(w http.ResponseWriter, r *http.Request, _ api.LogoutParams)

// OtpVerify проверка OTP кода.
func (auth *Auth) OtpVerify(w http.ResponseWriter, r *http.Request, _ api.OtpVerifyParams)

// OtpRetry повторная отправка OTP.
func (auth *Auth) OtpRetry(w http.ResponseWriter, r *http.Request, _ api.OtpRetryParams)

// RefreshToken обновление токенов.
func (auth *Auth) RefreshToken(w http.ResponseWriter, r *http.Request, _ api.RefreshTokenParams)

// VerifyEmail -запускает процесс верификации почтового адреса.
func (auth *Auth) VerifyEmail(w http.ResponseWriter, r *http.Request, _ api.VerifyEmailParams)

func (auth *Auth) ApproveEmail(w http.ResponseWriter, r *http.Request, _ api.ApproveEmailParams)
```

С помощью рефакторинга сделали так:

```
internal/components/endpoints
├── broker
│   └── assets
├── calculators
│   └── assets
├── cards
├── currencies
├── dadata
├── showcase
│   ├── assets
│   └── testdata
└── unep
```

Домены разделены, а также уменьшена связность:

```go
package cards


type (
	PABEService interface {
		ListCards(ctx context.Context, token string) (api.CardsList, error)

		GetCard(ctx context.Context, token string, cardID uuid.UUID) (*api.Card, error)

		FindInvoice(ctx context.Context, token string, invoiceID string) (api.Card, error)

		IssueCard(
			ctx context.Context,
			token string,
			idempotencyKey uuid.UUID,
			request api.CardIssueRequest,
		) (*api.CardIssueResponse, error)

		CalculateCardQuote(
			ctx context.Context,
			token string,
			request api.CardQuoteRequest,
		) (*api.CardQuoteResponse, error)
		FindOperationByInvoiceID(ctx context.Context, token string, invoiceID string) (api.Operation, error)

		TopUpCard(
			ctx context.Context,
			token string,
			cardID uuid.UUID,
			idempotencyKey uuid.UUID,
			paymentMethod monads.Optional[string],
			request api.CardTopupRequest,
		) (*api.CardTopupResponse, error)

		FindSecrets(ctx context.Context, token string, cardID uuid.UUID) (api.CardSecrets, error)

		FindOperations(
			ctx context.Context,
			token string,
			cardID uuid.UUID,
			filter api.CardOperationsCursorPaginationRequest,
		) (api.CardOperationsResponse, error)
	}

	CardHandler struct {
		pabeClient PABEService
	}
)

func New(pabeClient PABEService) *CardHandler

// TopUpCard — POST /v1/cards/{card_id}/topup.
func (handler *CardHandler) TopUpCard(
	w http.ResponseWriter,
	r *http.Request,
	cardID uuid.UUID,
	params api.TopUpCardParams,
)

// ListCards — GET /v1/cards.
func (handler *CardHandler) ListCards(w http.ResponseWriter, r *http.Request, _ api.ListCardsParams)

// GetCard — GET /v1/cards/{card_id}.
func (handler *CardHandler) GetCard(w http.ResponseWriter, r *http.Request, cardID uuid.UUID, _ api.GetCardParams)

func (handler *CardHandler) GetCardInvoice(
	writer http.ResponseWriter,
	request *http.Request,
	invoiceID uuid.UUID,
	_ api.GetCardInvoiceParams,
)

func (handler *CardHandler) FindOperationByInvoice(
	writer http.ResponseWriter, request *http.Request,
	invoiceID uuid.UUID,
	_ api.FindOperationByInvoiceParams,
)

// IssueCard — POST /v1/cards/issue.
func (handler *CardHandler) IssueCard(w http.ResponseWriter, r *http.Request, params api.IssueCardParams)

// CalculateCardQuote — POST /v1/cards/quote.
func (handler *CardHandler) CalculateCardQuote(
	w http.ResponseWriter,
	r *http.Request,
	_ api.CalculateCardQuoteParams,
)

func (handler *CardHandler) FindCardSecrets(
	writer http.ResponseWriter,
	request *http.Request,
	cardID uuid.UUID,
	_ api.FindCardSecretsParams,
)

func (handler *CardHandler) FindCardOperations(
	writer http.ResponseWriter,
	request *http.Request,
	cardID uuid.UUID,
	_ api.FindCardOperationsParams,
)

func (handler *CardHandler) ReturnCallback(
	writer http.ResponseWriter,
	request *http.Request,
	label api.Label,
)
```

Еще один из примеров, которые можно привести был в том, что мы явно указали, что не стоит дробить один большой файл на несколько маленьких (в рамках одного домена), где будут содержаться функции, относящиеся к одному домену:

```
internal/components/endpoints
├── broker
│   └── assets
├── calculators
│   └── assets
├── cards
├── currencies
├── dadata
├── showcase
│   ├── assets
│   └── testdata
└── unep
```

```
internal/components/endpoints/cards
├── cards_get
├── cards_post
├── cards_delete
...
```

Решили сделать так, что все будет в одном файле и просто в редакторах разработчики будут использовать инструмент - `Outline Panel`, который присутствует везде.

После рефакторинга получилось вот так:

```
internal/components/endpoints/cards
├── cards.go
```

Следующий пример в рефакторинге конструктора сервера, проблема в том, что функция создания как и структура принимает очень много значений, что не очень хорошо, поэтому следует абстрагировать общим интерфейсом отдельные сущности, вот как было до:

```go
package server

type Server struct {
	*handlers.Auth
	*handlers.User
	*handlers.Chat
	*handlers.Reference
	*calculators.CalculatorHandler
	*cards.CardHandler
	*showcaseHandler.ShowcaseHandler
	*handlers.PABE
	*handlers.Permission
	*handlers.OperationsFilters
	*handlers.ProductOperations
	*handlers.Dadata
	*dadataHandler.DadataHandler
	*handlers.Gold
	*handlers.Documents
	*handlers.MediaHandler
	*handlers.PersonalData
	*broker.Handler
	*unep.UnepHandler
	*handlers.QualificationData
	*currenciesHandler.CurrencyHandler

	userClient       *user.Client
	authClient       *auth.Client
	a7ruClient       *a7ru.Client
	pabeClient       *pabe.Client
	chatClient       *chat.Client
	permissionClient *permission.Client
}

// NewServer создает новый экземпляр сервера.
func NewServer(
	userClient *user.Client,
	authClient *auth.Client,
	a7ruClient *a7ru.Client,
	pabeClient *pabe.Client,
	brokerPABEClient *brokerpabe.Client,
	cardsPABEClient *cardspabe.Client,
	dadataPABEClient *dadatapabe.Client,
	showcasePABEClient *showcasepabe.Client,
	unepClient *unep_client.Client,
	chatClient *chat.Client,
	permissionClient *permission.Client,
	feas config.FeasConfig,
	storage *objectstorage.Client,
	features *acl.AccessControlList,
	mediaPrefix string,
	showcasesFilter *hashset.HashSet[int64],
	currenciesClient *currencies.Client,
	consentsClient *consents.Client,
	env config.Env,
) *Server {
	// ...
}
```

После:

```go
package server

type Server struct {
	handler *server.Handler
	client *server.Client
}

// NewServer создает новый экземпляр сервера.
func NewServer(
	client *server.Client,
	handler *server.Handler,
	config config.Config,
	features *acl.AccessControlList,
	showcasesFilter *hashset.HashSet[int64],
) *Server {
	// ...
}
```

Здесь мы вынесли отдельно клиента как домен и хендлеры, получили более минималистичный код, который будет легко поддерживать и развивать, так как мы уменьшили связность.

## 3. Почему некоторые проектные решения нельзя отменить.

Принятые вами архитектурные решения вначале - могут закрепиться и остаться, так что их потом попросту нельзя будет изменить, поэтому важно изначально подходить к проектированию с особым вниманием, чтобы выстроить хорошие архитектурные границы, которые в будущем составят именно хороший архитекрутный дизайн приложения, с которым мы будем готовы жить и развивать проект с максимальной простотой.

## 4. Не переусердствуйте с интерьером.

Не стоит уходить в безрассудную духоту - которая возникает на почве названия переменных и функций, например. Мы прежде всего стремимся к написанию простых систем, значит основной нашей целью должна быть простая структура приложения в целом и минимальная связность между его модулями и т.д.

Если идеи рефакторинга основываются лишь на том, как же красиво назвать переменную или функцию - вы достигли лишь локального максимума, и если вы там застряли, то не так легко будет забраться на ту вершину, к которой необходимо стремится. Наша задача в правильной организации дизайна системы, а не ее "припудривание".

## 5. Качество кода (в значительной степени) не имеет отношения к самому коду.

Важно сперва думать о том - **"что делается"**, а не о том - "как это делается", когда мы начинаем думать о данных и их взаимосвязях - как самособой разумеющееся получаем чистый код. Если же поступать наоборот - начиная с низкого уровня, то в итоге простую систему будет построить сложнее, как и повышать качество кода. Сначала дизайн, затем код, а не наоборот.

## 6. Застревание в старом дизайне.

При возникновении вопросов к дизайну - важно изменить его так, чтобы он преобразился в простой! Не стоит застревать в старом дизайне, важно менять его - это окупится с лихвой в будущем, проект будет простым и лаконичным и его также будет легко поддерживать.

> **В любом проекте возможен дизайн, который сделает код, который вы пишете,
> красивым и простым**. Если выбранные проектные конструкции для этого не подходят,
> то просто измените их.

## 7. Что делает плохие тесты плохими?

Плохие тесты делает "плохими" - не знание функциональности, не понимание проекта и его контекста. Важно проверять, что тот или иной код делает то, что должен в рамках проекта. Важно тестировать его смысл, его поведение и сущность, его побочные эффекты. Проверяйте не правильность выполнения операторов и команд в функции как таковых, проверяйте поведение кода в системе.

> **Понимание системы -- это о чём-то большем, нежели просто покрытие кода
> тестами**. Речь идет об одном из фундаментальных положений программной
> инженерии: **думать о смысле, о предназначении программы, которое отличается от
> самой программы, от её кода**.
> Так что не проверяйте, что код делает то, что в нём непосредственно написано. 
> **Убедитесь, что код делает именно то, что должен делать в рамках всей системы в
> целом**.
