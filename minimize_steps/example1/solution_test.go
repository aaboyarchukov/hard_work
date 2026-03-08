package example1

import (
	"slices"
	"testing"
)

type Case struct {
	Name   string
	Input  []int
	Output []int
}

func TestStrictlyMonotonousSequence(t *testing.T) {
	cases := []Case{
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
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(tt *testing.T) {
			if slices.Compare(StrictlyMonotonousSequence(testCase.Input), testCase.Output) != 0 {
				tt.Fatalf("FAILED: %s", testCase.Name)
			}
		})
	}
}
