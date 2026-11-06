# Антипаттерн "Интерфейс должен быть проще реализации"

Многие считают, что интерфейс, должен быть компактнее реализации - но это полнейший антипаттерн, поскольку интерфейс представляет из себя формальную спецификацию для какой-то части программы - то он должен быть максимально глубоко продуман и написан, чтобы взглянув на него - было понятно, что система может делать и как это использовать

Почему стоит иметь более сложную спецификацию? В виду следующих причин:

- неточности и погрешности;
- побочные эффекты (призрачное состояние)

Неточности могут определять гибкость программы, например, если мы в спецификации определили, что температура может быть в пределе от 50 до 100 градусов, то в этом пределе мы можем сделать реализацию, например, которая нам даст 70 градусов, соответсвенно, **чем шире погрешность, тем гибче может быть программа.**

Побочные эффекты проявляются там, где появляются ненужные зависимости, например, в функции объявляются ненужные переменные, или проводятся дополнительные вычисления (нарушение SRP), где нигде не регламентированы.

Так как пока нереально применять формальной спецификации (не мейнстрим), то достаточно придерживаться следующего правила, которое избавляет от большинства проблем - "используйте более точные типы данных".

## Призрачное состояние

В данном примере в рамках функции блокировки пользователей существует призрачное состояние за счет удаление сессий - такое действие должно разноситься отдельно от основной функции

```go
func (s *Service) BlockUsers(ctx context.Context, initiatorRole domain.Role, initiatorID domain.UserID, usersID []domain.UserID) (map[domain.UserID]error, error) {
	userRoles, err := s.UserRepository.Roles(ctx, usersID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	allowedIDs, excludedIDs := s.filterUsersByPermissions(initiatorRole, &initiatorID, usersID, userRoles)

	blockedUsers, err := s.UserRepository.Block(ctx, allowedIDs)
	if err != nil {
		return nil, fmt.Errorf("block users: %w", err)
	}

	// здесь удаляем сессии пользователей
	for _, userID := range allowedIDs {
		if blockedUsers[userID] != nil {
			continue
		}

		if err := s.SessionRepository.DeleteUserSessions(ctx, userID.String()); err != nil {
			blockedUsers[userID] = fmt.Errorf("%w:%s", domain_errors.ErrDeleteSession, err.Error())
		}
	}

	for id := range excludedIDs {
		blockedUsers[id] = domain_errors.ErrInsufficientPermissions
	}

	return blockedUsers, nil
}
```

В этом примере при создании пользователя - отправляем письмо (причем его отправка необязательна, что уже вносит неодназначности), появляется сообщение `emailMsg`, которое носит в себе призрачное состояние как и весь кусок отправки сообщения на почту.

```go
func (s *Service) CreateUser(ctx context.Context, initiatorRole domain.Role, user *domain.User) (*domain.User, error) {
	if err := s.validateCreate(user, initiatorRole); err != nil {
		return nil, fmt.Errorf("validateCreate: %w", err)
	}

	createdUser, err := s.UserRepository.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("s.UserRepository.CreateUser: %w", err)
	}

	emailMsg := []byte("<someDomen>/api/v1/auth/recovery_pwd" + "?email=" + user.Email)

	logger.Debug("email body", "msg", string(emailMsg))

	if err := s.EmailClient.Send(ctx, createdUser.Email, s.NewUserEmailSubject, emailMsg); err != nil {
		logger.Error("send email to client", "email", createdUser.Email)


		return createdUser, nil
	}

	return createdUser, nil
}
```

В этом примере понаставлены проверки сумм определенным условиям, что создает грязное состояние функции, так как по сигнатуре непонятно сравнение, нужно выделить в отдельный метод валидации и валидировать суммы на другом уровне.

```go
func (handler *Handler) GetBrokerRevenueCalculator(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetBrokerRevenueCalculatorParams,
) {
	ctx := r.Context()

	if letserve.ExtractToken(r) == "" {
		httpresponse.ErrUnauthorized.Write(ctx, w)

		return
	}

	marketParams := generateRevenueMarketParams()
	period := marketParams.Period

	rate, err := decimal.NewFromString(marketParams.Rate)
	if err != nil {
		httpresponse.ErrInvalidRequest.WithMessage("Неверный формат rate").WithField("rate").Write(ctx, w)

		return
	}

	amount, err := decimal.NewFromString(params.Amount)
	if err != nil {
		httpresponse.ErrInvalidRequest.WithMessage("Неверный формат amount").WithField("amount").Write(ctx, w)

		return
	}

	if amount.LessThan(decimal.NewFromInt(int64(marketParams.MinAmount))) {
		httpresponse.ErrInvalidRequest.WithMessage("amount не может быть меньше min_amount").WithField("amount").Write(ctx, w)

		return
	}

	if amount.GreaterThan(decimal.NewFromInt(int64(marketParams.MaxAmount))) {
		httpresponse.ErrInvalidRequest.WithMessage("amount не может быть больше max_amount").WithField("amount").Write(ctx, w)

		return
	}

	httpresponse.JSON(ctx, w, http.StatusOK, api.BrokerRevenueCalculatorResponse{
		PlanRevenue: calculateBrokerRevenuePlan(amount, rate, period),
	})
}
```

## Сужение логики кода

В данном примере в сигнатуре функции в виде `request` параметра указан сырой срез байт, также в типе возвращаемый функцией, что дает неточность и неопределенность в состоянии, куда лучше было бы выделить отдельный тип - например сформированный на основе строгой спецификации -> `api.CardIssueRequest, api.CardIssueResponse`

```go
IssueCard(
	ctx context.Context,
	token string,
	idempotencyKey uuid.UUID,
	request []byte)
([]byte, error)
```

При настройке прослойки сервиса - считывается корневой файл, для построения `RBAC`, но при таком хардкодинге пути мы как раз сужаем логику (а что если файл поменяет название? что если при разворачивании сервиса его путь внутри будет другой?).

```go
openapiPath := "./api/openapi/v1.yml"
if _, err := os.Stat(openapiPath); os.IsNotExist(err) {
	openapiPath = "../../api/openapi/v1.yml"
}

rbac, err := middleware.NewRBAC(openapiPath)
if err != nil {
	return nil, fmt.Errorf("RBAC initialization: %w", err)
}
```

При смене пароля - валидируется его длина, причем валидируется на конкретном значении, что очень сужает логику. Куда логичнее было не использовать конкретную длину, а сравнивать по целой составляющей надежности пароля -> например можно задать границы энтропии в рамках которой валидировать надежность паспорта.

```go
func (s *Service) ChangePassword(ctx context.Context, email, oldPwd, newPwd string) error {
	// ...

	// проверяем новый пароль
	if err = s.validatePassword(newPwd); err != nil {
		return err
	}

	// ...

	return nil
}

func (s *Service) validatePassword(password string) error {
	if len(password) < 18 {
		return errors.New("password is not valid - wrong len")
	}

	return nil
}
```

## Интерфейс сложнее реализации

Например, когда у нас происходят сложные бизнесовые процессы, где должны быть учтены определенные ограничения, которые не всегда отобразишь в сигнатуре функции. Тогда они должны быть указаны в комментариях или в `api` спецификации.

В данном примере непонятны условия при поиске платежки, при выставлении счета и пополнения карты, хотя там будет скрываться целое множество подводных камней.

Также типы размыты и неясны.

```go
PABEService interface {
		FindInvoice(ctx context.Context, token string, invoiceID string) (api.Card, error)
		CalculateCardQuote(
			ctx context.Context,
			token string,
			request []byte,
		) ([]byte, error)
		FindOperationByInvoiceID(ctx context.Context, token string, invoiceID string) ([]byte, error)
		TopUpCard(
			ctx context.Context,
			token string,
			cardID uuid.UUID,
			idempotencyKey uuid.UUID,
			paymentMethod string,
			request []byte,
		) ([]byte, error)
	}

```

Функция декодирования имени файла, интерфейс до ужаса простой, тем самым мы загоняем себя в целое множество неизвестных состояний (например, имя файла может быть в разной кодировке). При составлении интерфейса стоит выделить эти ограничения, например определить сет кодировок, которые будут обрабатываться.

```go
charmap.Charmap{
		charmap.Windows1251,
		charmap.KOI8R,
		charmap.CodePage866,
		charmap.Windows1252,
	}
```

И заменить входной параметр на тип имени файла, который будет валидироваться на определенную кодировку, тогда интерфейс усложниться, но сделает разработку сервиса намного проще, так как мы определили границы.

```go
func DecodeFilename(in string) string
```

Обработка файлов: в данном случае опять не определены типы, как у выходных данных, так и не заданы ограничения для обрабатываемого файла (какие типы файлов мы сможем обрабатывать ? :)

```go
func ExtractFiles(file []byte) ([]byte, error)
```

Ну и конечно в любом примере - всегда стоит выделять отдельные типы, ведь тогда код становится более явным, ясным и понятным. И с ними можно сделать интерфейсы более наглядными. Интерфейс стоит усложнять за счет выделения типов и определения состояний, в рамках которых выполняется функция - таким образом мы задаем границы системе и в таком системе труднее всего допустить ошибки.
