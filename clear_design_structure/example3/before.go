package example3

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

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
