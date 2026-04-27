package manyarguments

import "time"

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
