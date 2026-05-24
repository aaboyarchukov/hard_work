# Ясный код 2

В данном уроке нам необходимо проследить нарушение нескольких принципов "Ясного кода" на уровне реализации, а именно:

1.1. Методы, которые используются только в тестах (подумайте, как от них избавиться).

Пример 1: 

*Проблема:*

[unnecessary_methods_before.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/unnecessary_tests_methods/example1_before.go)

В данном примере мы использовали анонимную функцию с сфрфм sql запросом в пакете с тестами для того, чтобы вставить тестовые данные в базу. Эта функция использовалась только в пакете с тестами, что не очень хорошо.


```go
func createFeedStories(ctx context.Context, feeds []domain.FeedStoryInput, audience domain.Audience) {
	// ... row sql request
}

func (s *FeedIntegrationSuite) basic() {
	allAudiences := []domain.Audience{domain.AudienceAll}

	cases := map[string]struct {
		setup  func()
		verify func(got []domain.FeedStory)
	}{
		"happy_path": {
			setup: func() {
				feed1 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					// ...
				})
				feed2 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					// ...
				})
				feed3 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					// ...
				})

				require.NoError(s.T(), createFeedStories(s.Ctx, []domain.FeedStoryInput{feed1, feed2, feed3}, allAudiences))
			},
			verify: func(got []domain.FeedStory) {
				require.Len(s.T(), got, 3)
				assert.True(s.T(), got[0].IsPinned)
				require.Len(s.T(), got[0].Slides, 1)
				require.Len(s.T(), got[0].Slides[0].Buttons, 1)
			},
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.CleanDB(transactionalTables...)
			tc.setup()

			got, err := s.svc.FeedStories(s.Ctx, allAudiences)
			require.NoError(s.T(), err)
			tc.verify(got)
		})
	}
}
```

*Решение:*

[unnecessary_methods_after.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/unnecessary_tests_methods/example1_after.go)

Мы в репозитории реализуем **CRUD** операции, связанные с взаимодействием с доменной сущностью, таким образом мы будем использовать контракт репозитория (`s.repo.CreateFeedStories`) напрямую и у нас не будет необходимости в реализации дополнительных анонимных методах, которые будут использоваться только в тестах.

```go
func (s *FeedIntegrationSuite) basic() {
	allAudiences := []domain.Audience{domain.AudienceAll}

	cases := map[string]struct {
		setup  func()
		verify func(got []domain.FeedStory)
	}{
		"happy_path": {
			setup: func() {
				feed1 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					// ...
				})
				feed2 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					// ...
				})
				feed3 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					// ...
				})

				require.NoError(s.T(), s.repo.CreateFeedStories(s.Ctx, []domain.FeedStoryInput{feed1, feed2, feed3}, allAudiences))
			},
			verify: func(got []domain.FeedStory) {
				require.Len(s.T(), got, 3)
				assert.True(s.T(), got[0].IsPinned)
				require.Len(s.T(), got[0].Slides, 1)
				require.Len(s.T(), got[0].Slides[0].Buttons, 1)
			},
		},
		// ...
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.CleanDB(transactionalTables...)
			tc.setup()

			got, err := s.svc.FeedStories(s.Ctx, allAudiences)
			require.NoError(s.T(), err)
			tc.verify(got)
		})
	}
}

```

1.2. Цепочки методов. Метод вызывает другой метод, который вызывает другой метод, который вызывает другой метод, который вызывает другой метод... и далее и далее.

Пример 1:

*Проблема:*

[chain_of_methods_before.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/chain_of_methods/example1_before.go)

Подобный пример из вызова цепочки взаимосвязанных методов ведет за собой несколько проблем:
- сложность тестирования (для тестирования метода N надо также тестировать со всем зависимостями все N...1 методы)
- выявление проблем и ошибок (при дебаге сложно понять в какой момент с данными произошла ошибка)
- высокая связность
- нарушение [[SRP|SRP]]

```go
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
```

*Решение:*

[chain_of_methods_after.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/chain_of_methods/example1_after.go)

Мы применили принцип информационного эксперта и сделали единый орекстратор, который вызывает функции обработки и валидации шаблона и последующей генерацией

```go
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
```

1.3. У метода слишком большой список параметров.

Пример 1:

*Проблема:*

[many_arguments_before.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/many_arguments/example1_before.go)

Нам необходимы фикстуры для предоставления тестовых данных в тестах, для этого нам нужны методы для построения моделей, так мы получаем метод, в которой передаем кучу параметров. Таким образом, если нам понадобится передать еще одни параметры, тогда нам необходимо добавлять новые аргументы в функцию, что не очень хорошо:

```go
func FeedStoryInput(slides []domain.Slide, buttons []domain.Buttons, order []int, status string, audience domain.Audience) domain.FeedStoryInput {
	const (
		defaultAudience     = "all"
		defaultStatus       = domain.Published
		defaultDisplayOrder = 1
		defaultDateOffset   = 24 * time.Hour
	)

	yesterday := time.Now().Add(-defaultDateOffset)

	f := domain.FeedStoryInput{}
	
	if slides != nil {
		f.slides = slides
	}
	
	// ...

	return f
}
```


*Решение:*

[many_arguments_after.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/many_arguments/example1_after.go)

Многие языки справляются с этим с помощью опциональных аргументов и переменного их количества, в Go с этим можно справится с помощью паттерна функциональных опций, когда мы передаем переменное количество опций - функций, в каждую из которых передается объект, который мы хотим обогатить, и затем остается только в цикле применить каждую из опций:

```go
func FeedStoryInput(opts ...func(*domain.FeedStoryInput)) domain.FeedStoryInput {
	const (
		defaultAudience     = "all"
		defaultStatus       = domain.Published
		defaultDisplayOrder = 1
		defaultDateOffset   = 24 * time.Hour
	)

	yesterday := time.Now().Add(-defaultDateOffset)

	f := domain.FeedStoryInput{
		FeedStory: domain.FeedStory{
			ID:       uuid.New(),
			IsPinned: false,
			Preview:  Preview(),
		},
		DisplayOrder: defaultDisplayOrder,
		Audience:     defaultAudience,
		TypeCode:     domain.Feed,
		StatusCode:   defaultStatus,
		StartDate:    &yesterday,
		EndDate:      nil,
	}

	for _, opt := range opts {
		opt(&f)
	}

	return f
}
```

1.4. Странные решения. Когда несколько методов используются для решения одной и той же проблемы, создавая несогласованность.

Пример 1:

*Проблема:*

[inconsistency_before.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/inconsistency/example1_before.go)

В хранилище использовались две функции по удалению документов при их загрузке "пачками", и соответственно использовались они в разных местах по разному, что приводит к отсутствию консистентности и единоначалия:

```go
func (r *Repository) UploadAtomicBatch(
	ctx context.Context,
	files map[repository.FileID]repository.StorageFileInput,
) map[repository.S3Link]repository.StorageFileOutput {
	// ...
}

func (r *Repository) BatchCreateAtomic(ctx context.Context, docs []*domain.UploadDocument) (map[*domain.UploadDocument]string, error) {
	// ...
}
```

*Решение:*

[inconsistency_after.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/inconsistency/example1_after.go)

Убрать одну из функций и привести работу с загрузкой документов в хранилище "пачками" к единообразию:

```go
func (r *Repository) BatchCreateAtomic(ctx context.Context, docs []*domain.UploadDocument) (map[*domain.UploadDocument]string, error) {
	const op = "minio.BatchCreateAtomic"

	if len(docs) == 0 {
		return make(map[*domain.UploadDocument]string), nil
	}

	var (
		results = make(map[*domain.UploadDocument]string, len(docs))
		errs    = make(map[*domain.UploadDocument]error)
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, doc := range docs {
		wg.Go(func() {
			link, err := r.Create(ctx, repository.StorageFileInput{
				Id:          uuid.New(),
				Name:        doc.Name,
				File:        bytes.NewReader(doc.File),
				Size:        int64(len(doc.File)),
				ContentType: utils.DetectContentType(doc.File),
			})

			mu.Lock()
			defer mu.Unlock()

			results[doc] = link
			if err != nil {
				errs[doc] = err
			}
		})
	}

	wg.Wait()

	if len(errs) > 0 {
		keys := make(map[string]struct{}, len(results))
		for _, link := range results {
			keys[link] = struct{}{}
		}
		r.RollbackUploaded(ctx, keys)

		return nil, fmt.Errorf("%s: %d/%d failed: %v", op, len(errs), len(docs), errs)
	}

	return results, nil
}

```

1.5. Чрезмерный результат. Метод возвращает больше данных, чем нужно вызывающему его компоненту.

Пример 1:

*Проблема:*

[many_returned_values_before.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/many_returned_values/example1_before.go)

В данном примере мы получаем данные со стороннего api по клиенту, затем возвращаем эти данные. Проблема такого подхода в том, что непонятно по семантике, что возвращает метод; также большое количество возвращаемых значений это также плохо, по причине использования, то есть в некоторых случаях, возвращаемые значения могут быть не использованы и останутся только "висеть"

```go
func GetPersonData(ctx context.Context) (string, string, string, int) {
	info, err := api.GetPerson(ctx)
	if err != nil {
		return "", "", "", 0
	}
	
	return info.Name, info.Surname, info.Patronymic, info.Age
}
```

*Решение:*

[many_returned_values_after.go](https://github.com/aaboyarchukov/hard_work/blob/main/clean_code_2/many_returned_values/example1_after.go)

Такие примеры решаются следующими вещами:
- разделение на несколько методов, которые ответственны за свои данные
- определение отдельных типов данных/структур в виде возвращаемых значений
- ну и нормальное проектирование иерархии типов, грамотное разделение на доменные области :)

```go
func GetPersonData(ctx context.Context) (domain.Person, error) {
	info, err := api.GetPerson(ctx)
	if err != nil {
		return domain.Person{}, fmt.Errorf("error with api")
	}
	
	return ScratchPersonData(ctx, info)
}
```