package example3

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

type Document struct {
	ID   int64
	Name string
	Type string
	File []byte
}

type CreateApplication struct {
	InsuranceId     uuid.UUID
	ClientId        int64
	ApplicationType ApplicationType
	Files           []Document
}

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

type AppTypeId = int32

func (s *Service) validateAndGetBaseData(ctx context.Context, req CreateApplication) (dto.Insurance, AppTypeId, error) {
	insurance, err := s.insuranceRepo.GetUserInsurance(ctx, req.InsuranceId)
	if err != nil {
		return dto.Insurance{}, 0, err
	}

	identificationClientId, err := s.identificationRepo.GetClientId(ctx, insurance.CustomerId)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Insurance{}, 0, ErrInsuranceNotFound
	}

	if err != nil {
		return err
	}

	if req.ClientId != identificationClientId {
		return dto.Insurance{}, 0, ErrIdentificationNotFound
	}

	appTypeId, err := s.applicationRepo.GetApplicationTypeId(ctx, req.ApplicationType)
	if err != nil {
		return dto.Insurance{}, 0, err
	}

	return insurance, appTypeId, nil
}

func (s *Service) attachDocuments(ctx context.Context, clientID int64, applicationID uuid.UUID, applicationDocs []Document, uploadedFiles []Document) error {
	clientDocs, err := s.documentsRepo.GetClientDocIdsById(ctx, clientID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	documents := DocumentsToRepoContract(applicationDocs)
	failedFiles := s.storageRepo.CreateBatch(ctx, documents)

	for _, file := range failedFiles {
		if file.Error != nil {
			return ErrAttachDocument
		}

		uploadedFiles = append(uploadedFiles, Document{
			S3Key:   file.S3Link,
			DocType: file.FileType,
		})
	}

	if err := s.documentsRepo.SaveApplicationDocuments(ctx, applicationID, clientDocs); err != nil {
		return err
	}

	if err := s.documentsRepo.SaveApplicationDocuments(ctx, applicationID, uploadedFiles); err != nil {
		return err
	}

	return nil
}
