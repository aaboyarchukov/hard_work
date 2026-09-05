# Неочевидные проектные ошибки

## 1. Отладка, логи, проверки

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
