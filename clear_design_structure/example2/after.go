package example2

import (
	"context"
	"errors"
	"fmt"
	"time"
)

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

func (s *Service) initNewIdentification(ctx context.Context, finmartClientID, finmartInsuranceID int64, provider string, person *domain.Person) error {
	clientID, err := s.storeClientInfo(ctx, finmartClientID, person)
	if err != nil {
		return err
	}

	identID, err := s.identificationRepo.CreateIdentification(ctx, clientID, provider, domain.IdentificationNew)
	if err != nil {
		if errors.Is(err, repository.ErrProviderNotFound) {
			return domaint_errors.ErrProviderNotFound
		}

		return fmt.Errorf("repo.CreateIdentification: %w", err)
	}

	return s.identificationRepo.CreateInsuranceId(ctx, identID, finmartInsuranceID)
}

func (s *Service) handleExistingIdentification(ctx context.Context, identID, finmartClientID, finmartInsuranceID int64, provider string, status domain.IdentificationStatus, person *domain.Person) error {
	switch status {
	case domain.IdentificationNew, domain.IdentificationInProgress:
		// кейс: клиент оформляет еще 1 страховой продукт, когда идентификация в статусе new/in_progress.
		// При отправке новых данных персоны - пользователь в БД не будет обновляться/вставляться новый,
		// т.е. создается только новый страховой продукт и линкуется к clientID.
		//
		// Логика временная, потенциально в будущем изменится (нужна информация от СК)
		return s.identificationRepo.CreateInsuranceId(ctx, identID, finmartInsuranceID)

	case domain.IdentificationIdentified:
		if err := s.identificationRepo.CreateInsuranceId(ctx, identID, finmartInsuranceID); err != nil {
			return err
		}

		return s.outboxRepo.Insert(ctx, finmartInsuranceID, domain.IdentificationIdentified)

	case domain.IdentificationNotIdentified, domain.IdentificationError:
		internalClientID, err := s.storeClientInfo(ctx, finmartClientID, person)
		if err != nil {
			return err
		}

		newIdentID, errCreate := s.identificationRepo.CreateIdentification(ctx, internalClientID, provider, domain.IdentificationNew)
		if errCreate != nil {
			return fmt.Errorf("repo.CreateIdentification: %w", errCreate)
		}

		return s.identificationRepo.CreateInsuranceId(ctx, newIdentID, finmartInsuranceID)

	default:
		return fmt.Errorf("unexpected identification status: %s", status)
	}
}

func (s *Service) validateAge(birthDate time.Time) error {
	const (
		minAge = 18
		maxAge = 86
	)

	now := time.Now()
	if now.Before(birthDate.AddDate(minAge, 0, 0)) || !now.Before(birthDate.AddDate(maxAge, 0, 0)) {
		return fmt.Errorf("age: must be between %d and %d", minAge, maxAge)
	}

	return nil
}

// storeClientInfo - сохранение данных клиента(персоны) в postgres + S3
// Возвращает внутренний clients.id (BIGSERIAL PK)
func (s *Service) storeClientInfo(ctx context.Context, finmartClientID int64, person *domain.Person) (int64, error) {
	const op = "service.identification.storeClientInfo"

	if person.PersonType == "" {
		return 0, fmt.Errorf("%s: person_type is empty", op)
	}

	clientID, err := s.identificationRepo.CreateClientWithPassport(ctx, finmartClientID, person)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if len(person.Documents) == 0 {
		return clientID, nil
	}

	// сохраняем документы (сканы и т.д.) персоны
	uploaded, err := s.storage.BatchCreateAtomic(ctx, person.Documents)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// линкуем в БД clientID и ссылку на доки в хранилище
	docsToInsert := make([]dto.DocumentCreate, 0, len(uploaded))
	for doc, s3Link := range uploaded {
		docsToInsert = append(docsToInsert, dto.DocumentCreate{
			TypeName: doc.Type.String(),
			Name:     doc.Name,
			S3Link:   s3Link,
		})
	}

	documentFileIds, err := s.documentsRepo.CreateBatchWithTypeName(ctx, docsToInsert)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.documentsRepo.SaveClientDocuments(ctx, documentFileIds, clientID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return clientID, nil
}
