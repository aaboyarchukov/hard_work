package example2

import "testing"

func TestBinarySearchLeft(t *testing.T) {
	type binarySearchLeftCase struct {
		Name       string
		InputArray []int
		Target     int
		Output     int
	}

	cases := []binarySearchLeftCase{
		{
			"empty array",
			[]int{},
			1,
			0,
		},
		{
			"one element array -> append",
			[]int{1},
			2,
			0,
		},
		{
			"one element array -> switch",
			[]int{1},
			0,
			0,
		},
		{
			"odd elements array -> switch",
			[]int{1, 3, 5, 7, 9},
			4,
			2,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(subT *testing.T) {
			resultIndx := BinarySearchLeft(testCase.InputArray, testCase.Target)

			if resultIndx != testCase.Output {
				subT.Fatalf("FAILED: %s, wanted: %v, got: %d", testCase.Name, testCase.Output, resultIndx)
			}
		})
	}
}
