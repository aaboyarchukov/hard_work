package example1

import (
	"slices"
	"testing"
)

func TestStrictlyMonotonousSequence(t *testing.T) {
	type strictlyMonotonousSequenceCase struct {
		Name   string
		Input  []int
		Output []int
	}

	cases := []strictlyMonotonousSequenceCase{
		{
			"empty array",
			[]int{},
			[]int{},
		},
		{
			"one element array",
			[]int{1},
			[]int{1},
		},
		{
			"odd elements array",
			[]int{7, 1, 2, 3, 0, 4, 5, 6, 5},
			[]int{1, 2, 3, 4, 5, 6},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(subT *testing.T) {
			result := StrictlyMonotonousSequence(testCase.Input)
			if slices.Compare(result, testCase.Output) != 0 {
				subT.Fatalf("FAILED: %s, wanted: %v, got: %v", testCase.Name, testCase.Output, result)
			}
		})
	}
}
