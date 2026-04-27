package unnecessary_tests_methods

import (
	"context"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

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
					f.IsPinned = true
					f.DisplayOrder = 1
					f.Slides = []domain.Slide{
						fixtures.Slide(1, func(sl *domain.Slide) {
							sl.Buttons = []domain.Button{fixtures.Button(1)}
						}),
					}
				})
				feed2 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					f.DisplayOrder = 2
					f.Slides = []domain.Slide{fixtures.Slide(1)}
				})
				feed3 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					f.DisplayOrder = 3
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
		"empty_feed": {
			setup: func() {},
			verify: func(got []domain.FeedStory) {
				assert.Empty(s.T(), got)
			},
		},
		"pinned_first": {
			setup: func() {
				feed1 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					f.DisplayOrder = 1
				})
				feed2 := fixtures.FeedStoryInput(func(f *domain.FeedStoryInput) {
					f.IsPinned = true
					f.DisplayOrder = 2
				})

				require.NoError(s.T(), s.repo.CreateFeedStories(s.Ctx, []domain.FeedStoryInput{feed1, feed2}, allAudiences))
			},
			verify: func(got []domain.FeedStory) {
				require.Len(s.T(), got, 2)
				assert.True(s.T(), got[0].IsPinned)
				assert.False(s.T(), got[1].IsPinned)
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
