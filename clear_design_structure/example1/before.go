package example1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) CreateIsurance(ctx context.Context, insurance *dto.Insurance) (uuid.UUID, error) {
	var insuranceId uuid.UUID
	uploadedFiles := make([]uploadedDoc, 0)

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {

		if insurance == nil {
			err := fmt.Errorf("nil object")

			return err
		}

		product, err := s.catalogRepo.Get(ctx, insurance.ProductId)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductInactive
		}

		if err != nil {
			return err
		}

		if product.ID <= 0 {
			return ErrProductNotFound
		}

		insurance.ProviderId = int64(product.ProviderId)

		providerCode, err := s.providerRepo.GetCode(ctx, insurance.ProviderId)
		if err != nil {
			return err
		}

		clientIdentification, err := s.identificationRepo.GetByClientAndProvider(ctx, insurance.ClientId, providerCode)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrIdentificationNotFound
		}

		if err != nil {
			return err
		}

		const identifiedClient = "identified"
		if clientIdentification.Status != identifiedClient {
			return ErrClientNotIdentified
		}

		if insurance.Sum > product.MaxSum || insurance.Sum < product.MinSum {
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
				return ErrInvalidShare
			}
			shareSum += float64(b.Share)
		}

		if shareSum > MaxShareSumLimit+eps {
			return ErrHigherShare
		}

		requisiteId, err := s.requisitesRepo.Exist(ctx, insurance.Requisites.Bic)

		if errors.Is(err, pgx.ErrNoRows) {
			requisiteId, err = s.requisitesRepo.Create(ctx, *insurance.Requisites)
			if err != nil {
				return err
			}
		}

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		insurance.RequisiteId = requisiteId

		insuranceStatusId, err := s.insuranceRepo.GetStatusId(ctx, NEW_STATUS)
		if err != nil {
			return err
		}
		insurance.InsuranceStatusId = insuranceStatusId

		const defaultDuration = 5
		insurance.Duration = defaultDuration
		insurance.Currency = product.Currency

		// очень ужасно - надо исправить
		var personDocs []int64
		if insurance.InsuredPerson != nil {

			for _, d := range insurance.InsuredPerson.Documents {
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

			var insuredPersonID int64
			insuredPersonID, err = s.personRepo.CreatePerson(ctx, &domain.Person{
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
				return err
			}

			// save passport
			_, err = s.documentsRepo.SavePersonPassport(ctx, insurance.InsuredPerson.Passport, insuredPersonID)
			if err != nil {
				return err
			}

			resultDocuments := make([]dto.Document, 0, len(uploadedFiles))
			for _, u := range uploadedFiles {
				documentTypeId, errGet := s.documentsRepo.GetDocumentTypeId(ctx, u.Type)
				if errGet != nil {
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
				return err
			}

			err = s.documentsRepo.SavePersonDocuments(ctx, personDocs, insuredPersonID)
			if err != nil {
				return err
			}

			insurance.InsuredPersonId = insuredPersonID
		}

		providerId, err := s.providerRepo.GetId(ctx, providerCode)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProviderNotFound
		}

		if err != nil {
			return err
		}

		insurance.ProviderId = providerId

		statusId, err := s.identificationRepo.GetStatusId(ctx, NEW_STATUS)
		if err != nil {
			return err
		}
		insurance.InsuranceStatusId = int64(statusId)

		insuranceId, err = s.insuranceRepo.Create(ctx, *insurance)
		if err != nil {
			return err
		}

		beneficiariesId, err := s.beneficiariesRepo.CreateMany(ctx, insurance.Beneficiary)
		if err != nil {
			return err
		}

		err = s.beneficiariesRepo.ConnecBeneficiariesToInsurance(ctx, beneficiariesId, insuranceId)
		if err != nil {
			return err
		}

		clientDocs, err := s.documentsRepo.GetClientDocIdsById(ctx, insurance.ClientId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		err = s.documentsRepo.SaveInsuranceDocuments(ctx, clientDocs, insuranceId)
		if err != nil {
			return err
		}

		// очень ужасно - надо исправить
		if insurance.InsuredPerson != nil {
			err = s.documentsRepo.SavePersonDocuments(ctx, personDocs, insurance.ClientId)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.rollbackS3Files(ctx, uploadedFiles)
		return uuid.Nil, err
	}

	return insuranceId, nil
}
