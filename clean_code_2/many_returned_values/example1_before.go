package manyreturnedvalues

import "context"

func GetPersonData(ctx context.Context) (string, string, string, int) {
	info, err := api.GetPerson(ctx)
	if err != nil {
		return "", "", "", 0
	}

	return info.Name, info.Surname, info.Patronymic, info.Age
}
