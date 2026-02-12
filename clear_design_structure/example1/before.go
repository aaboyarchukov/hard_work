package main

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type Requisites struct {
	Bic         string
	BankName    string
	Account     string
	CorrAccount string
}

type Passport struct {
	Series         string
	Number         string
	IssuedBy       string
	IssueDate      time.Time
	DepartmentCode string
}

type Document struct {
	ID   int64
	Name string
	Type string
	File []byte
}

type Person struct {
	Name                   string
	PersonType             string
	Patronymic             *string
	Surname                string
	BirthDate              time.Time
	Phone                  string
	Email                  string
	RegistrationAddress    string
	ActualAddress          string
	PostalAddress          string
	Passport               Passport
	CitizenshipCountryCode *string
	MigrationCardNumber    *string
	ResidencePermitNumber  *string
	Documents              []Document
}

type Beneficiary struct {
	Name       string
	Surname    string
	Patronymic *string
	BirthDate  time.Time
	Share      float64
	Relation   string
}

type CreateInsuranceReq struct {
	ClientID      int64
	ProductID     int64
	InsuranceSum  decimal.Decimal
	Requisites    Requisites
	InsuredPerson *Person
	Beneficiaries []Beneficiary
}

func (s *Service) CreateIsurance(ctx context.Context, req CreateInsuranceReq) (uuid.UUID, error) {
	var insuranceId uuid.UUID
	uploadedFiles := make([]uploadedDoc, 0)

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		product, err := s.catalogRepo.GetActive(txCtx, req.ProductID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductInactive
		}

		if err != nil {
			return err
		}

		clientIdentification, err := s.identificationRepo.GetByClientAndProvider(txCtx, req.ClientID, product.ProviderCode)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrIdentificationNotFound
		}

		if err != nil {
			return err
		}

		if clientIdentification.Status != catalog.IDENT_IDENTIFIED {
			return ErrClientNotIdentified
		}

		sum, err := insurances.DecimalToFloat64(req.InsuranceSum)
		if err != nil {
			return err
		}

		if !product.ValidInsuranceSum(sum) {
			return ErrLimitRange
		}

		beneficiaries := AggregateBeneficiaries(req.Beneficiaries)

		if !users.ValidShareSum(beneficiaries) {
			return ErrHigherShare
		}

		requisites := catalog.Requisites{
			Bic:         req.Requisites.Bic,
			BankName:    req.Requisites.BankName,
			Account:     req.Requisites.Account,
			CorrAccount: req.Requisites.CorrAccount,
		}
		requisiteId, err := s.requisitesRepo.GetOrCreate(txCtx, requisites)

		if err != nil {
			return err
		}

		var insuredPersonDocIds []int64
		var insuredPersonId *int64
		if req.InsuredPerson != nil {
			personResult, err := s.processPerson(txCtx, req.InsuredPerson, uploadedFiles)
			if err != nil {
				return err
			}

			insuredPersonId = personResult.PersonId
			insuredPersonDocIds = personResult.PersonDocIds
		}

		const defaultDuration = 5

		insurance := insurances.Insurance{
			Id:              uuid.New(),
			RequisiteID:     requisiteId,
			InsuredPersonId: insuredPersonId,
			Status:          insurances.INS_NEW,
			Currency:        product.Currency,
			Duration:        defaultDuration,
			ClientID:        req.ClientID,
			CustomerId:      clientIdentification.Id,
			Sum:             sum,
		}

		beneficiariesId, err := s.beneficiariesRepo.CreateBatch(txCtx, beneficiaries)
		if err != nil {
			return err
		}

		insuranceId, err = s.insuranceRepo.Create(txCtx, insurance)
		if err != nil {
			return err
		}

		err = s.beneficiariesRepo.ConnecBeneficiariesToInsurance(txCtx, beneficiariesId, insuranceId)
		if err != nil {
			return err
		}

		clientDocs, err := s.documentsRepo.GetClientDocIdsById(txCtx, req.ClientID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		err = s.documentsRepo.SaveInsuranceDocuments(txCtx, clientDocs, insuranceId)
		if err != nil {
			return err
		}

		if req.InsuredPerson != nil {
			err = s.documentsRepo.SaveInsuranceDocuments(txCtx, insuredPersonDocIds, insuranceId)
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

type PersonResult struct {
	PersonId      *int64
	PersonDocIds  []int64
	UploadedFiles []uploadedDoc
}

func (s *Service) processPerson(ctx context.Context, person *Person, uploadedFiles []uploadedDoc) (PersonResult, error) {

	insuredPerson := BuildPerson(*person)

	if !insuredPerson.IsAgeAllowed(time.Now()) {
		return PersonResult{}, ErrPersonData
	}

	documents := DocumentsToRepoContract(person.Documents)
	failedFiles, err := s.storageRepo.CreateBatch(ctx, documents)

	for _, file := range failedFiles {
		if file.Error != nil {
			return PersonResult{}, file.Error
		}

		uploadedFiles = append(uploadedFiles, uploadedDoc{
			S3Key:   file.S3Link,
			DocType: file.FileType,
		})
	}

	personID, err := s.personRepo.CreatePerson(ctx, &users.Person{
		Name:                   insuredPerson.Name,
		Surname:                insuredPerson.Surname,
		Patronymic:             insuredPerson.Patronymic,
		BirthDate:              insuredPerson.BirthDate,
		Phone:                  insuredPerson.Phone,
		Email:                  insuredPerson.Email,
		RegistrationAddress:    insuredPerson.RegistrationAddress,
		ActualAddress:          insuredPerson.ActualAddress,
		PostalAddress:          insuredPerson.PostalAddress,
		CitizenshipCountryCode: insuredPerson.CitizenshipCountryCode,
		MigrationCardNumber:    insuredPerson.MigrationCardNumber,
		ResidencePermitNumber:  insuredPerson.ResidencePermitNumber,
	})
	if err != nil {
		return PersonResult{}, err
	}

	_, err = s.documentsRepo.SavePersonPassport(
		ctx,
		&insuredPerson.Passport,
		personID,
	)
	if err != nil {
		return PersonResult{}, err
	}

	resultDocuments := make([]dto.Document, 0, len(uploadedFiles))
	for _, u := range uploadedFiles {
		documentTypeId, err := s.documentsRepo.GetDocumentTypeId(ctx, u.DocType)
		if err != nil {
			return PersonResult{}, err
		}

		resultDocuments = append(resultDocuments, dto.Document{
			Name:       u.Name,
			TypeID:     int32(documentTypeId),
			S3Link:     u.S3Key,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		})
	}

	docIDs, err := s.documentsRepo.Create(ctx, resultDocuments)
	if err != nil {
		return PersonResult{}, err
	}

	if err := s.documentsRepo.SavePersonDocuments(ctx, docIDs, personID); err != nil {
		return PersonResult{}, err
	}

	return PersonResult{
		PersonId:      &personID,
		PersonDocIds:  docIDs,
		UploadedFiles: uploadedFiles,
	}, nil
}

func BuildPerson(person Person) users.Person {
	passport := BuildPassport(person.Passport)
	documents := AggregateDocuments(person.Documents)

	return users.Person{
		Name:                   person.Name,
		Patronymic:             person.Patronymic,
		Surname:                person.Surname,
		BirthDate:              person.BirthDate,
		Phone:                  person.Phone,
		Email:                  person.Email,
		RegistrationAddress:    person.RegistrationAddress,
		ActualAddress:          person.ActualAddress,
		PostalAddress:          person.PostalAddress,
		Passport:               passport,
		CitizenshipCountryCode: person.CitizenshipCountryCode,
		MigrationCardNumber:    person.MigrationCardNumber,
		ResidencePermitNumber:  person.ResidencePermitNumber,
		Documents:              documents,
	}
}

func DocumentsToRepoContract(documents []Document) map[repository.FileID]repository.StorageFileInput {
	result := make(map[repository.FileID]repository.StorageFileInput, len(documents))
	for _, doc := range documents {
		content := bytes.NewReader(doc.File)
		fileId := uuid.New()
		result[fileId] = repository.StorageFileInput{
			Id:          fileId,
			Name:        doc.Name,
			File:        content,
			Size:        content.Size(),
			ContentType: detectContentType(doc.File),
		}
	}

	return result
}

func AggregateDocuments(files []Document) []documents.Document {
	result := make([]documents.Document, 0, len(files))
	for _, file := range files {
		result = append(result, documents.Document{
			Name: file.Name,
			Type: file.Type,
			File: file.File,
		})
	}

	return result
}

func BuildPassport(passport Passport) documents.Passport {
	return documents.Passport{
		Series:         passport.Series,
		Number:         passport.Number,
		IssuedBy:       passport.IssuedBy,
		IssueDate:      passport.IssueDate,
		DepartmentCode: passport.DepartmentCode,
	}
}
func AggregateBeneficiaries(beneficiaries []Beneficiary) []users.Beneficiary {
	result := make([]users.Beneficiary, 0, len(beneficiaries))
	for _, beneficiary := range beneficiaries {
		result = append(result, users.Beneficiary{
			Name:       beneficiary.Name,
			Surname:    beneficiary.Surname,
			Patronymic: beneficiary.Patronymic,
			BirthDate:  beneficiary.BirthDate,
			Share:      beneficiary.Share,
			Relation:   beneficiary.Relation,
		})
	}

	return result
}
