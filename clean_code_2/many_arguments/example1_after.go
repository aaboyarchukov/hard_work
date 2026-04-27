package manyarguments

import (
	"time"

	"github.com/google/uuid"
)

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
