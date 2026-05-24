package chainofmethods

import (
	"context"
	"time"

	"git.dip.pics/dip/platform/go/logger.git"
)

var (
	emptyFile = []byte{}
)

func Generate(ctx context.Context, targetTemplate domain.Template, img map[string][]byte) []byte {
	file, err := getFillTemplate(
		ctx,
		targetTemplate.Version,
		targetTemplate.Code,
		targetTemplate.StartDate,
		s.templateRepo,
	)
	if err != nil {
		return nil
	}

	return file
}

func getFillTemplate(ctx context.Context, version string, code string, date time.Time, repo Repository) ([]byte, error) {
	template, err := strategy.GetTemplate(
		ctx,
		version,
		code,
		date,
		repo,
	)

	if template == nil {
		logger.Error("template not found or deleted", err)
		return emptyFile, ErrTemplateNotFound
	}

	return getDocument(ctx, template, repo)
}

func getDocument(ctx context.Context, template Template, repo Repository) ([]byte, error) {
	file, err := repo.Get(ctx, template.FilePath)
	if err != nil {
		return emptyFile, ErrInvalidDataFields
	}

	return validFields(ctx, template, file)

}

func validFields(ctx context.Context, file []byte, template Template) ([]byte, error) {
	equalFields, errEqual := comparator.CompareStructFields(template, file)

	if errEqual != nil {
		logger.Error("error with compare files from s3", "err", errEqual)
		return emptyFile, ErrInvalidDataFields
	}

	return fillAndConvert(ctx, file, template)
}

func fillAndConvert(ctx context.Context, file []byte, template Template) ([]byte, error) {
	filledTemplate, err := aggregator.FillTemplateWithData(file, template)
	if err != nil {
		logger.Error("failed to fill template", err)
		return emptyFile, ErrInvalidDataFields
	}

	logger.Info("template filled with data successfully")

	fileBase64, err := office.Convert(filledTemplate, "pdf", template.FileName)
	if err != nil {
		logger.Error("failed to convert to PDF", err)
		return emptyFile, ErrConvertTemplate
	}

	logger.Info("PDF generated successfully")
	return fileBase64, nil
}
