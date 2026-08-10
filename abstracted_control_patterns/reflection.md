# Абстрагируем управляющие паттерны

В данном уроке нам необходимо заняться рефакторингом кода. Необходимо найти части в программе, которые можно упросить управляющими конструкциями, максимально абстрагировав инструкции.

Можно рассмотреть на данном примере: https://gist.github.com/Maecenas/5878ceee890a797ee6c9ad033a0ae0f1

Здесь Raymond Hettinger, показывает как необходимо пользоваться всеми прелестями Python. Он в рамках [доклада](https://www.youtube.com/watch?v=wf-BqAjZb8M) абстрагирует конструкцию (`try-catch-except-finally`) простым `with`, а все необходимые методы абстрагиурет внутри класса, которые обрабатываются в нем на основе `properties`

## Примеры

*для каждого примера необходимо найти управляющую конструкцию(for, if/else, switch) и упростить ее (абстрагирование классом и методами, уменьшение цикломатической сложности, декомпозиция функций, полиморфизм, рекурсия, итератор)

### Транзакция

Было:

```go
func SaveUser(ctx context.Context, tx *sql.Tx, user dto.User) (UserID, error) {
	userId, err := sql.SaveUser(
		ctx,
		user,
	)

	if err != nil {
		logger.Err("some occusion on save user", err)
		tx.Rollback()

		return NullUserID, err
	}

	tx.Commit()

	return userId, nil
}
```

Стало:

```go
func WithTx(ctx context.Contex, func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
    if err != nil {
        return err
    }

    defer func() {
		tx.Rollback()
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

func SaveUser(ctx context.Context, tx *sql.Tx, user dto.User) (UserID, error) {
	var userId UserID

	err := WithTx(ctx, func(tx *sql.Tx) error {
		userId, err := sql.SaveUser(
			ctx,
			user,
		)

		if err != nil {
			return err
		}

		userId = UserID{
			id: userId,
		}

		return nil
	} )

	if err != nil {
		logger.Err("some occusion on save user", err)

		return NullUserID, err
	}

	return userId, nil
}

```

В данном примере мы инкапсулировали логику, которая оборачивает каждую операцию в транзакцию, тогда нам будет достаточно вызвать данную функцию, и мы отходим от способ самостоятельно контролировать открытие, закрытие и откат транзакции.

### Обработка файлов

Было:

```go
func ValidateFile(ctx context.Context, pathToFile string) (bool, error) {
	file, err := os.OpenFile(pathToFile)

	if err != nil {
		return false, err
	}

	fileData := make([]byte, FILE_SIZE)

	_, err := file.ReadBytes(fileData)

	if err != nil {
		file.Close()
		return false, err
	}

	var info Info

	if err := json.Unmarshal(fileData, &info); err != nil {
		file.Close()
		return false, err
	}

	// some validation...

	file.Close()

	return true, nil
}
```

Стало:

```go
func WithFile(path string, fn func(*os.File) error) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // спрятано ЗДЕСЬ, один раз, для ВСЕХ вызовов

    return fn(f)
}

func ValidateFile(ctx context.Context, pathToFile string) (bool, error) {
	err := WithFile(pathToFile, func(f *os.File) error {
	    data, err := io.ReadAll(f)
	    if err != nil {
		    return err
	    }

	    var info Info
		if err := json.Unmarshal(data, &info); err != nil {
			return false, err
		}

		// some validation...

	    return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
```

В данном примере мы вынесли обработку файла в отдельную функцию, что дает нам сосредоточиться именно на логике самого алгоритма и не переживать за не очищенные ресурсы.

### Инкапсуляция обработки цикла

Было:

```go
func ProcessEvents(ctx context.Context, events []Event) ([]Message, error) {
	result := make([]Message, 0, len(s))
    for _, event := range events {
        // processing event
        message := buildMessage(event)
        result = append(result, message)
    }
    return result, nil
}
```

Стало:

```go
func TryMap[T, U any](ctx context.Context, s []T, fn func(T) (U, error)) ([]U, error) {
    result := make([]U, 0, len(s))
    for _, v := range s {
        if err := ctx.Err(); err != nil {
            return nil, err
        }
        u, err := fn(v)
        if err != nil {
            return nil, err
        }
        result = append(result, u)
    }
    return result, nil
}

func ProcessEvents(ctx context.Context, events []Event) ([]Message, error) {
    messages, err := TryMap(ctx, events, processOne)
    // processOne - function for processing events
    if err != nil {
        return nil, err
    }
    return messages, nil
}
```

Таким образом мы изящно инкапсулировали логику обработки данных, и вынесли цикл в отдельную функцию, теперь у нас идет простой вызов.
