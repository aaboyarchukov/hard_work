# Clear design in code

В данном уроке необходимо разобрать свой рабочий код и его логический дизайн, а затем сопоставить спецификацию с реальной реализацией.
Суть заключается в реализации программной системы таким образом, чтобы при ее чтении все было прозрачно и понятно, а также получившиеся реализация должна следовать 1:1 с дизайном, но важно не закапываться в "декларативщине", а стремится к схожести дизайна.
То есть не стоит пренебрегать ясностью кода в угоду лишним абстракциям.

### Пример 1

#### Дизайн:

Нам дана функция оформления страховки на клиента, который идентифицирован по определенному страховому продукту. 

В спецификации к данной функции сервиса сказано следующее:

- необходимо проверять идентификацию клиента при оформлении продукта страховки
	- нет идентификации -> возвращаем ошибку идентификации
	- есть -> идем дальше
- необходимо валидировать параметры продукта, а также его состояние
	- валидировать минимальные и максимальный параметр указанной суммы
	- также валидировать процент выгодоприобретателей
- также при передаче иных лиц страхования, необходимо сохранять их персональные данные, передавать документы в хранилище на сохранение, если иных лиц нет, тогда не сохраняем данные и документы

#### До: [before](https://github.com/aaboyarchukov/hard_work/tree/main/clear_design_structure/example1/before.go)

```go
func (s *Service) CreateIsurance(ctx context.Context, insurance *dto.Insurance) (uuid.UUID, error) {
	var insuranceId uuid.UUID
	uploadedFiles := make([]uploadedDoc, 0)

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		logger.Info("start")

		if insurance == nil {
			err := fmt.Errorf("nil object")
			logger.Error("failed", "msg", err)

			return err
		}

		product, err := s.catalogRepo.Get(ctx, insurance.ProductId)
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info("product inactive")
			return ErrProductInactive
		}

		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		if product.ID <= 0 {
			logger.Error("failed", "msg", ErrProductNotFound)
			return ErrProductNotFound
		}

		insurance.ProviderId = int64(product.ProviderId)

		providerCode, err := s.providerRepo.GetCode(ctx, insurance.ProviderId)
		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		clientIdentification, err := s.identificationRepo.GetByClientAndProvider(ctx, insurance.ClientId, providerCode)
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info("no active identification (no rows)")
			return ErrIdentificationNotFound
		}

		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		const identifiedClient = "identified"
		if clientIdentification.Status != identifiedClient {
			logger.Info("client not identified")
			return ErrClientNotIdentified
		}

		if insurance.Sum > product.MaxSum || insurance.Sum < product.MinSum {
			logger.Info("sum of insurance out of limit")
			return ErrLimitRange
		}

		const (
			MaxShareSumLimit = 100.0
			MinShareSumLimit = 0
			eps              = 1e-9
		)

		var shareSum float64
		for _, b := range insurance.Beneficiary {
			if b.Share < MinShareSumLimit || b.Share > MaxShareSumLimit {
				logger.Error("failed", "msg", ErrInvalidShare)
				return ErrInvalidShare
			}
			shareSum += float64(b.Share)
		}

		if shareSum > MaxShareSumLimit+eps {
			logger.Error("failed", "msg", ErrHigherShare)
			return ErrHigherShare
		}

		requisiteId, err := s.requisitesRepo.Exist(ctx, insurance.Requisites.Bic)

		if errors.Is(err, pgx.ErrNoRows) {
			requisiteId, err = s.requisitesRepo.Create(ctx, *insurance.Requisites)
			if err != nil {
				logger.Error("failed", "msg", err)
				return err
			}
		}

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			logger.Error("failed", "msg", err)
			return err
		}

		insurance.RequisiteId = requisiteId

		insuranceStatusId, err := s.insuranceRepo.GetStatusId(ctx, NEW_STATUS)
		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}
		insurance.InsuranceStatusId = insuranceStatusId

		const defaultDuration = 5
		insurance.Duration = defaultDuration
		insurance.Currency = product.Currency

		// очень ужасно - надо исправить
		var personDocs []int64
		if insurance.InsuredPerson != nil {

			logger.Info("upload documents")

			for _, d := range insurance.InsuredPerson.Documents {
				key, errCreate := s.storageRepo.Create(
					ctx,
					uuid.New().String(),
					d.Name,
					d.File,
					detectContentType(d.File),
				)

				if errCreate != nil {
					logger.Error("upload failed", "msg", errCreate)

					return errCreate
				}

				uploadedFiles = append(uploadedFiles, uploadedDoc{
					Name:  d.Name,
					Type:  d.Type,
					S3Key: key,
				})
			}
			logger.Info("upload success")

			logger.Info("create insured person")
			insuredPersonID, err := s.personRepo.CreatePerson(ctx, &domain.Person{
				Name:                insurance.InsuredPerson.Name,
				Surname:             insurance.InsuredPerson.Surname,
				Patronymic:          insurance.InsuredPerson.Patronymic,
				PersonType:          insurance.InsuredPerson.PersonType,
				BirthDate:           insurance.InsuredPerson.BirthDate,
				Phone:               insurance.InsuredPerson.Phone,
				Email:               insurance.InsuredPerson.Email,
				RegistrationAddress: insurance.InsuredPerson.RegistrationAddress,
				ActualAddress:       insurance.InsuredPerson.ActualAddress,
				PostalAddress:       insurance.InsuredPerson.PostalAddress,

				CitizenshipCountryCode: insurance.InsuredPerson.CitizenshipCountryCode,
				MigrationCardNumber:    insurance.InsuredPerson.MigrationCardNumber,
				ResidencePermitNumber:  insurance.InsuredPerson.ResidencePermitNumber,
			})
			if err != nil {
				logger.Error("create insured person", "msg", err)
				return err
			}
			logger.Info("save person success")

			// save passport
			_, err = s.documentsRepo.SavePersonPassport(ctx, insurance.InsuredPerson.Passport, insuredPersonID)
			if err != nil {
				logger.Error("save passport failed", "msg", err)
				return err
			}
			logger.Info("save passport success")

			resultDocuments := make([]dto.Document, 0, len(uploadedFiles))
			for _, u := range uploadedFiles {
				documentTypeId, errGet := s.documentsRepo.GetDocumentTypeId(ctx, u.Type)
				if errGet != nil {
					logger.Error("get document type failed", "msg", errGet)
					return errGet
				}

				resultDocuments = append(resultDocuments, dto.Document{
					Name:       u.Name,
					TypeID:     int32(documentTypeId),
					S3Link:     u.S3Key,
					CreatedAt:  time.Now(),
					ModifiedAt: time.Now(),
				})
			}

			personDocs, err = s.documentsRepo.Create(ctx, resultDocuments)
			if err != nil {
				logger.Error("create documents failed", "msg", err)
				return err
			}
			logger.Info("create documents success")

			err = s.documentsRepo.SavePersonDocuments(ctx, personDocs, insuredPersonID)
			if err != nil {
				logger.Error("create documents failed", "msg", err)
				return err
			}
			logger.Info("create documents success")

			insurance.InsuredPersonId = insuredPersonID
		}

		providerId, err := s.providerRepo.GetId(ctx, providerCode)
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("get provider id failed", "msg", err)
			return ErrProviderNotFound
		}

		if err != nil {
			logger.Error("get provider id failed", "msg", err)
			return err
		}

		logger.Info("get provider id success")
		insurance.ProviderId = providerId

		statusId, err := s.identificationRepo.GetStatusId(ctx, NEW_STATUS)
		if err != nil {
			logger.Error("get status id failed", "msg", err)
			return err
		}
		logger.Info("get status id success")
		insurance.InsuranceStatusId = int64(statusId)

		insuranceId, err = s.insuranceRepo.Create(ctx, *insurance)
		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		beneficiariesId, err := s.beneficiariesRepo.CreateMany(ctx, insurance.Beneficiary)
		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		err = s.beneficiariesRepo.ConnecBeneficiariesToInsurance(ctx, beneficiariesId, insuranceId)
		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		clientDocs, err := s.documentsRepo.GetClientDocIdsById(ctx, insurance.ClientId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			logger.Error("failed", "msg", err)
			return err
		}

		err = s.documentsRepo.SaveInsuranceDocuments(ctx, clientDocs, insuranceId)
		if err != nil {
			logger.Error("failed", "msg", err)
			return err
		}

		// очень ужасно - надо исправить
		if insurance.InsuredPerson != nil {
			err = s.documentsRepo.SavePersonDocuments(ctx, personDocs, insurance.ClientId)
			if err != nil {
				logger.Error("failed", "msg", err)
				return err
			}
		}

		logger.Info("success")
		return nil
	})

	if err != nil {
		s.rollbackS3Files(ctx, uploadedFiles)
		return uuid.Nil, err
	}

	return insuranceId, nil
}
```

#### После: [after](https://github.com/aaboyarchukov/hard_work/tree/main/clear_design_structure/example1/after.go)

Теперь функционал следует одному дизайну, также разделено все по слоям с разделением ответственности и также с уменьшением связности, для дальнейшего более простого масштабирования.

```go

func (s *Service) CreateIsurance(ctx context.Context, req CreateInsuranceReq) (uuid.UUID, error) {
	var (
		insuranceID  uuid.UUID
		uploadedDocs []UploadedDoc
	)

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		product, identification, err := s.loadAndValidateBaseData(txCtx, req)
		if err != nil {
			return err
		}

		sum, beneficiaries, err := s.validateFinancialData(req, product)
		if err != nil {
			return err
		}

		requisiteID, err := s.getOrCreateRequisites(txCtx, req.Requisites)
		if err != nil {
			return err
		}

		insuredPersonID, insuredPersonDocIDs, err := s.processOptionalInsuredPerson(txCtx, req.InsuredPerson, &uploadedDocs)
		if err != nil {
			return err
		}

		insurance := s.buildInsurance(
			req,
			product,
			identification,
			requisiteID,
			insuredPersonID,
			sum,
		)

		insuranceID, err = s.persistInsurance(
			txCtx,
			insurance,
			beneficiaries,
			req.ClientID,
			insuredPersonDocIDs,
		)

		return err
	})

	if err != nil {
		s.rollbackS3Files(ctx, uploadedDocs)
		return uuid.Nil, err
	}

	return insuranceID, nil
}

```

### Пример 2

#### Дизайн:

Необходимо реализовать ручку, которая идентифицирует в страховой существующего клиента:
- необходимо проверить наличие идентификации клиента по определенному продукту
	- если идентификация есть, возвращаем ошибку о том, что клиент уже идентифицирован по данному продукту
	- если идентификации нет -> идем дальше
- сохраняем данные клиента и отправляем их в страховую, для создания идентификации
- после возвращаем ответ со статусом вновь созданной идентификацией, а также идет асинхронная работа по обновлению статуса идентификации
#### До: [before](https://github.com/aaboyarchukov/hard_work/tree/main/clear_design_structure/example2/before.go)

```go
func (s *Service) CreateClientIdentification(ctx context.Context, clientID int64, providerCode string, person *domain.Person) error {
	uploadedFiles := make([]uploadedDoc, 0, len(person.Documents))
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		exist, err := s.identificationRepo.ExistsActiveByClientAndProvider(ctx, clientID, providerCode)
		if exist {
			return ErrIdentificationAlreadyExists
		}

		if err != nil {
			return err
		}

		logger.Info("upload documents")

		existClient, err := s.personRepo.ExistClient(ctx, clientID)
		if err != nil {
			return err
		}

		// !existClient
		// этот ужас надо переделать
		if !existClient {
			for _, d := range person.Documents {
				key, errCreate := s.storageRepo.Create(
					ctx,
					uuid.New().String(),
					d.Name,
					d.File,
					detectContentType(d.File),
				)

				if errCreate != nil {
					return errCreate
				}

				uploadedFiles = append(uploadedFiles, uploadedDoc{
					Name:  d.Name,
					Type:  d.Type,
					S3Key: key,
				})
			}
		}

		// этот ужас надо переделать
		if !existClient {
			_, err = s.personRepo.CreateClient(ctx, clientID, person)
			if err != nil {
				return err
			}
		}

		// этот ужас надо переделать
		if !existClient {
			_, err = s.documentsRepo.SaveClientPassport(ctx, person.Passport, clientID)
			if err != nil {
				return err
			}
		}

		providerId, err := s.providerRepo.GetId(ctx, providerCode)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProviderNotFound
		}

		if err != nil {
			return err
		}

		logger.Info("get provider id success")

		statusId, err := s.identificationRepo.GetStatusId(ctx, NEW_STATUS)
		if err != nil {
			return err
		}

		// этот ужас надо переделать
		if !existClient {
			resultDocuments := make([]dto.Document, 0, len(uploadedFiles))
			for _, u := range uploadedFiles {
				documentTypeId, errGet := s.documentsRepo.GetDocumentTypeId(ctx, u.Type)
				if errGet != nil {
					logger.Error("get document type failed", "msg", errGet)
					return errGet
				}

				resultDocuments = append(resultDocuments, dto.Document{
					Name:       u.Name,
					TypeID:     int32(documentTypeId),
					S3Link:     u.S3Key,
					CreatedAt:  time.Now(),
					ModifiedAt: time.Now(),
				})
			}

			var documentFileIds []int64
			documentFileIds, err = s.documentsRepo.Create(ctx, resultDocuments)
			if err != nil {
				return err
			}

			err = s.documentsRepo.SaveClientDocuments(ctx, documentFileIds, clientID)
			if err != nil {
				logger.Error("save person docs", "msg", err)
				return err
			}
		}

		err = s.identificationRepo.Save(ctx, dto.Identification{
			ClientId:           clientID,
			ProviderId:         int32(providerId),
			StatusId:           int32(statusId),
			IdentificationDate: time.Now(),
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.rollbackS3Files(ctx, uploadedFiles)
		return err
	}

	return nil
}
```
#### После: [after](https://github.com/aaboyarchukov/hard_work/tree/main/clear_design_structure/example2/after.go)

```go
func (s *Service) InitIdentification(ctx context.Context, finmartInsuranceID, finmartClientID int64, provider string, person *domain.Person) error {
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.validateAge(person.BirthDate); err != nil {
			return err
		}

		check, err := s.identificationRepo.CheckIdentification(txCtx, finmartClientID, provider)
		if err != nil {
			return fmt.Errorf("repo.CheckIdentification: %w", err)
		}

		// если у клиента совсем нет идентификации -> иницируем с 0
		if !check.Exists {
			return s.initNewIdentification(txCtx, finmartClientID, finmartInsuranceID, provider, person)
		}

		// идентификация уже есть -> логика ветвится от статуса
		return s.handleExistingIdentification(txCtx, check.IdentificationID, finmartClientID, finmartInsuranceID, provider, check.Status, person)
	})

	if errors.Is(err, repository.ErrFinmartInsuranceIdDuplicate) {
		return ErrFinmartInsuranceIdAlreadyUsed
	}

	return err
}
```

### Пример 3

#### Дизайн:

Необходимо реализовать ручку, которая по данным клиента составляет заявку на отказ по страховке (раннее закрытие, наступление страхового случая):
- необходимо проверять наличие страховки, офрмленной на клиента
	- нет страховки -> возвращаем ошибку
	- есть -> идем дальше
- валидировать заявку по ее типу:
	- если тип валиден -> передаем заявку в страховую и идем дальше
	- если тип невалиден -> возвращаем ошибку
- в конце возвращаем успешный ответ, который сигнализирует о том, что заявка успешно отправлена, за кулисами происходит асинхронная работа, в которой нам необходимо получать и обновлять статус по заявке со страховой

#### До: [before](https://github.com/aaboyarchukov/hard_work/tree/main/clear_design_structure/example3/before.go)

```go
func (s *Service) CreateApplication(ctx context.Context, application *dto.Application) (uuid.UUID, error) {
	var applicationID uuid.UUID
	uploadedFiles := make([]uploadedDoc, 0, len(application.Documents))

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		insurance, err := s.insuranceRepo.GetUserInsurance(ctx, application.InsuranceId)
		if err != nil {
			return err
		}
		application.ProductId = insurance.ProductId
		application.CustomerId = insurance.CustomerId

		identificationClientId, err := s.identificationRepo.GetClientId(ctx, insurance.CustomerId)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInsuranceNotFound
		}

		if err != nil {
			return err
		}

		if application.ClientId != identificationClientId {
			return ErrIdentificationNotFound
		}

		if insurance.UUID == uuid.Nil {
			return ErrInsuranceNotFound
		}

		appTypeId, err := s.applicationRepo.GetApplicationTypeId(ctx, application.ApplicationType)
		if err != nil {
			return err
		}
		application.TypeId = int64(appTypeId)

		applicationID = uuid.New()

		for _, d := range application.Documents {
			key, errCreate := s.storageRepo.Create(
				ctx,
				uuid.New().String(),
				d.Name,
				d.File,
				detectContentType(d.File),
			)

			if errCreate != nil {
				return errCreate
			}

			uploadedFiles = append(uploadedFiles, uploadedDoc{
				Name:  d.Name,
				Type:  d.Type,
				S3Key: key,
			})
		}

		if err = s.applicationRepo.Create(txCtx, *application, applicationID); err != nil {
			return err
		}

		resultDocuments := make([]dto.Document, 0, len(uploadedFiles))
		for _, u := range uploadedFiles {
			documentTypeId, errGet := s.documentsRepo.GetDocumentTypeId(ctx, u.Type)
			if errGet != nil {
				return errGet
			}

			resultDocuments = append(resultDocuments, dto.Document{
				InsuranceID:   &insurance.UUID,
				ApplicationID: &applicationID,
				ProductID:     &insurance.ProductId,
				Name:          u.Name,
				TypeID:        int32(documentTypeId),
				S3Link:        u.S3Key,
				CreatedAt:     time.Now(),
				ModifiedAt:    time.Now(),
			})
		}

		appDocs, err := s.documentsRepo.Create(txCtx, resultDocuments)
		if err != nil {
			return err
		}

		err = s.documentsRepo.SaveApplicationDocuments(txCtx, appDocs, applicationID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.rollbackS3Files(ctx, uploadedFiles)
		return uuid.Nil, err
	}

	return applicationID, nil
}
```
#### После: [after](https://github.com/aaboyarchukov/hard_work/tree/main/clear_design_structure/example3/after.go)

```go
func (s *Service) CreateApplication(ctx context.Context, req CreateApplication) (uuid.UUID, error) {
	var (
		applicationID uuid.UUID
		uploadedDocs  []UploadedDoc
	)

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {

		insurance, appTypeId, err := validateAndGetBaseData(txCtx, req)
		if err != nil {
			return err
		}

		application := BuildApplication(insurance, appTypeId)

		applicationID, err := s.applicationRepo.Create(txCtx, application)
		if err != nil {
			return err
		}

		if err := s.attachDocuments(txCtx, req.ClientId, applicationID, req.Files, uploadedDocs); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.rollbackS3Files(ctx, uploadedFiles)
		return uuid.Nil, err
	}

	return applicationID, nil
}
```
