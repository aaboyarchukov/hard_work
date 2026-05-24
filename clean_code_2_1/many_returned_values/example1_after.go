package manyreturnedvalues

import (
	"context"
	"fmt"
)

func GetPersonData(ctx context.Context) (domain.Person, error) {
	info, err := api.GetPerson(ctx)
	if err != nil {
		return domain.Person{}, fmt.Errorf("error with api")
	}

	return ScratchPersonData(ctx, info)
}
