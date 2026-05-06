# Нарушение SRP

## Нарушение SRP
---

**Пример 1:**

Было:
Здесь мы в строке и вычисляем нужное значение по индексу и валидируем его

```go
func ImportEquipmentFromEXCEL(file File) {
	// ....
	for ind_row := 0; ind_row < len(rows); ind_row++ {
		if rows[ind_row][0] == "" || rows[ind_row][0] == " " {
			continue
		}
		// ...
	}
	// ...
}
```

Физическая строка:
```go
if rows[ind_row][0] == "" || rows[ind_row][0] == " "
```

Стало:
Вынесли в отдельную функцию, которая проверяет валидность строки, а вычисляем саму строку раньше

```go
func ImportEquipmentFromEXCEL(file File) {
	// ....
	for ind_row := 0; ind_row < len(rows); ind_row++ {
		target_row := rows[ind_row]
		if validRow(target_row) {
			continue
		}
		// ...
	}
	// ...
}
```

Физическая строка:
```go
target_row := rows[ind_row]
// ...

if validRow(target_row)
```

**Пример 2:**

Было:
В строке мы складываем значение двух ошибок и сериализуем
```go
return ctx.JSON(err_get_winner.Error() + err_get_winner_id.Error())
```

Физическая строка:
```go
return ctx.JSON(err_get_winner.Error() + err_get_winner_id.Error())
```

Стало:
Правильнее будет ловить общую ошибку выше и сериализовать только ее.

```go
return ctx.JSON(err)
```

Физическая строка:
```go
return ctx.JSON(err_get_winner.Error() + err_get_winner_id.Error())
```

**Пример 3:**

Было:
При формировании информации о компании, мы в строке соединяем составляющие данных о директоре, что нарушет SRP

```go
return KP_info{
		CompanyInfo: models.Companies{
			// ...
			Director_data: companyResultInfo[0].Data.Management.Name + " " +
				companyResultInfo[0].Data.Management.Post,
		},
	}, nil
```

Физическая строка:
```go
Director_data: companyResultInfo[0].Data.Management.Name + " " +
				companyResultInfo[0].Data.Management.Post,
```

Стало:
Применяем паттерн фабрика.
```go
return KP_info{
		CompanyInfo: models.Companies{
			// ...
			Director_data: FabricDirectorInfo(companyResultInfo),
		},
	}, nil
```

Физическая строка:
```go
Director_data: FabricDirectorInfo(companyResultInfo),
```

**Пример 4:**

Было:
То же самое с менеджером
```go
return KP_info{
		CompanyInfo: models.Companies{
			// ...
			Manager_data: companyResultInfo[0].Data.Management.Name + " " +
				companyResultInfo[0].Data.Management.Post,
		},
	}, nil
```

Физическая строка:
```go
Manager_data: companyResultInfo[0].Data.Management.Name + " " +
				companyResultInfo[0].Data.Management.Post,
```

Стало:
Применяем паттерн фабрика.
```go
return KP_info{
		CompanyInfo: models.Companies{
			// ...
			Manager_data: FabricManagerInfo(companyResultInfo),
		},
	}, nil
```

Физическая строка:
```go
Manager_data: FabricManagerInfo(companyResultInfo),
```

**Пример 5:**

Было:
В данном случае, в строке и вычисляется длина и разделяется строка.

```go
if len(strings.SplitN(itemWithoutSpaces, "=", 2)) != 2 {
	resultMap[itemWithoutSpaces] = ""
	continue
}
```

Физическая строка:
```go
len(strings.SplitN(itemWithoutSpaces, "=", 2)) != 2
```

Стало:

```go
const (
	minPairLen = 2
	splitLen = 2
)

var pair []string = strings.SplitN(itemWithoutSpaces, "=", splitLen)
if len(pair) != minPairLen {
	resultMap[itemWithoutSpaces] = ""
	continue
}
```

Физическая строка:
```go
len(pair) != minPairLen
```

**Пример 6:**

Было:
Складывание строк и присваивание
```go
contact := models.Contact{
		FIO:     req_contact.Surname + " " + req_contact.Name + " " + req_contact.Patronymic,
	}
```

Физическая строка:
```go
FIO:     req_contact.Surname + " " + req_contact.Name + " " + req_contact.Patronymic,
```


Стало:
Вынесли в отдельную фабричную функцию
```go
contact := models.Contact{
		FIO:     FabricUserFIO(userInfo),
	}
```

Физическая строка:
```go
FIO:     FabricUserFIO(userInfo),
```

