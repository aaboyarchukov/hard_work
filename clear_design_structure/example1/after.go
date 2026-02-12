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

// Requisites вынести в отдельный пакет модели данных с которыми работаем
// не используем domain, так как данные модели нужны только для перекладывания данных, без логики
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

func (s *Service) loadAndValidateBaseData(ctx context.Context, req CreateInsuranceReq) (*catalog.Product, *identification_domain.Identification, error) {
	product, err := s.catalogRepo.GetActive(ctx, req.ProductID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, ErrProductInactive
	}
	if err != nil {
		return nil, nil, err
	}

	identification, err := s.identificationRepo.GetByClientAndProvider(ctx, req.ClientID, product.ProviderCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, ErrIdentificationNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	if identification.Status != identification_domain.IdentifiedStatus {
		return nil, nil, ErrClientNotIdentified
	}

	return &product, &identification, nil
}

func (s *Service) validateFinancialData(req CreateInsuranceReq, product *catalog.Product) (float64, []users.Beneficiary, error) {
	sum, err := insurances.DecimalToFloat64(req.InsuranceSum)
	if err != nil {
		return 0, nil, err
	}

	if !product.ValidInsuranceSum(sum) {
		return 0, nil, ErrLimitRange
	}

	beneficiaries := AggregateBeneficiaries(req.Beneficiaries)
	if !users.ValidShareSum(beneficiaries) {
		return 0, nil, ErrHigherShare
	}

	return sum, beneficiaries, nil
}

func (s *Service) getOrCreateRequisites(ctx context.Context, req Requisites) (int64, error) {
	requisites := catalog.Requisites{
		Bic:         req.Bic,
		BankName:    req.BankName,
		Account:     req.Account,
		CorrAccount: req.CorrAccount,
	}

	return s.requisitesRepo.GetOrCreate(ctx, requisites)
}

func (s *Service) processOptionalInsuredPerson(ctx context.Context, person *Person, uploadedFiles *[]UploadedDoc) (personID *int64, personDocIds []int64, err error) {
	if person == nil {
		return nil, nil, nil
	}

	result, err := s.processPerson(ctx, person, *uploadedFiles)
	if err != nil {
		return nil, nil, err
	}

	*uploadedFiles = append(*uploadedFiles, result.UploadedFiles...)

	return result.PersonId, result.PersonDocIds, nil
}

func (s *Service) buildInsurance(req CreateInsuranceReq, product *catalog.Product, identification *identification_domain.Identification, requisiteID int64, insuredPersonID *int64, sum float64) insurances.Insurance {
	const defaultDuration = 5

	return insurances.Insurance{
		Id:              uuid.New(),
		RequisiteID:     requisiteID,
		InsuredPersonId: insuredPersonID,
		Status:          insurances.NewStatus,
		Currency:        product.Currency,
		Duration:        defaultDuration,
		ClientID:        req.ClientID,
		CustomerId:      identification.Id,
		Sum:             sum,
		ContractNumber:  utils.RandomNumericString(),
		ProductId:       product.Id,
		ProviderID:      int64(identification.ProviderId),
	}
}

func (s *Service) persistInsurance(ctx context.Context, insurance insurances.Insurance, beneficiaries []users.Beneficiary, clientID int64, insuredPersonDocIDs []int64) (uuid.UUID, error) {
	beneficiaryIDs, err := s.beneficiariesRepo.CreateBatch(ctx, beneficiaries)
	if err != nil {
		return uuid.Nil, err
	}

	insuranceID, err := s.insuranceRepo.Create(ctx, insurance)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.beneficiariesRepo.
		ConnectBeneficiariesToInsurance(ctx, beneficiaryIDs, insuranceID); err != nil {
		return uuid.Nil, err
	}

	if err := s.attachDocuments(ctx, clientID, insuredPersonDocIDs, insuranceID); err != nil {
		return uuid.Nil, err
	}

	return insuranceID, nil
}

func (s *Service) attachDocuments(ctx context.Context, clientID int64, insuredPersonDocIDs []int64, insuranceID uuid.UUID) error {
	clientDocs, err := s.documentsRepo.GetClientDocIdsById(ctx, clientID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if err := s.documentsRepo.SaveInsuranceDocuments(ctx, clientDocs, insuranceID); err != nil {
		return err
	}

	return s.documentsRepo.SaveInsuranceDocuments(ctx, insuredPersonDocIDs, insuranceID)
}

type PersonResult struct {
	PersonId      *int64
	PersonDocIds  []int64
	UploadedFiles []UploadedDoc
}

func (s *Service) processPerson(ctx context.Context, person *Person, uploadedFiles []UploadedDoc) (PersonResult, error) {
	// build domain.Person
	insuredPerson := BuildPerson(*person)

	// valid person
	if !insuredPerson.IsAgeAllowed(time.Now()) {
		return PersonResult{}, ErrPersonData
	}

	// upload docs
	documents := DocumentsToRepoContract(person.Documents)
	failedFiles := s.storageRepo.CreateBatch(ctx, documents)

	for _, file := range failedFiles {
		if file.Error != nil {
			return PersonResult{}, file.Error
		}

		uploadedFiles = append(uploadedFiles, UploadedDoc{
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
		resultDocuments = append(resultDocuments, dto.Document{
			Name:       u.Name,
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

func AggregateDocuments(files []Document) []documents_domain.Document {
	result := make([]documents_domain.Document, 0, len(files))
	for _, file := range files {
		result = append(result, documents_domain.Document{
			Name: file.Name,
			Type: file.Type,
			File: file.File,
		})
	}

	return result
}

func BuildPassport(passport Passport) documents_domain.Passport {
	return documents_domain.Passport{
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
			// Share:      beneficiary.Share,
			Relation: beneficiary.Relation,
		})
	}

	return result
}
