package chainofmethods

import (
	"context"
	"time"

	"git.dip.pics/dip/platform/go/logger.git"
)

var (
	emptyFile = []byte{}
)

func (s *Service) GenerateDocument(ctx context.Context, targetTemplate domain.Template, img map[string][]byte) ([]byte, error) {
	ctx, err := s.txManager.Begin(ctx)
	defer s.txManager.CommitOrRollback(ctx, &err)

	if err != nil {
		logger.Error("error with txmanager", err)
		return emptyFile, ErrTemplateNotFound
	}

	template, err := strategy.GetTemplate(
		ctx,
		targetTemplate.Version,
		targetTemplate.Code,
		targetTemplate.StartDate,
		s.templateRepo,
	)

	if template == nil {
		logger.Error("template not found or deleted", err)
		return emptyFile, ErrTemplateNotFound
	}

	if err != nil || template.Code == "" {
		logger.Error("template not found or deleted", err)
		return emptyFile, ErrTemplateNotFound
	}

	if template.StartDate.After(time.Now()) {
		logger.Error("template is not active", err)
		return emptyFile, ErrTemplateNotFound
	}

	logger.Info("template selected successfully")

	file, err := s.documentRepo.Get(ctx, template.FilePath)
	if err != nil {
		logger.Error("failed to get file from S3", err)
		return emptyFile, ErrInvalidDataFields
	}

	equalFields, errEqual := comparator.CompareStructFields(targetTemplate, file)

	if errEqual != nil {
		logger.Error("error with compare files from s3", "err", errEqual)
		return emptyFile, ErrInvalidDataFields
	}

	equalImages := comparator.CompareImgFields(template.Img, img)

	equalImagesExtends, errEqual := comparator.CompareImagesIxtends(img, file)

	if errEqual != nil {
		logger.Error("error with compare imgs in file from s3", "err", errEqual)
		return emptyFile, ErrInvalidDataFields
	}

	if !equalFields || !equalImagesExtends || !equalImages {
		logger.Error("data fields do not match template requirements")
		return emptyFile, ErrInvalidDataFields
	}

	logger.Info("template data validation passed successfully")

	filledTemplate, err := aggregator.FillTemplateWithData(file, targetTemplate.Data, img, template.Img)
	if err != nil {
		logger.Error("failed to fill template", err)
		return emptyFile, ErrInvalidDataFields
	}

	logger.Info("template filled with data successfully")

	fileBase64, err := s.office.Convert(filledTemplate, "pdf", template.FileName)
	if err != nil {
		logger.Error("failed to convert to PDF", err)
		return emptyFile, ErrConvertTemplate
	}

	logger.Info("PDF generated successfully")
	return fileBase64, template.FileName, nil
}
