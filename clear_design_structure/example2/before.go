package example2

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

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
