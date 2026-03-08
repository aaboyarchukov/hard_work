package example1

import "minimize_steps/example2"

func StrictlyMonotonousSequence(array []int) []int {
	size := len(array)
	if size <= 1 {
		return array
	}

	result := make([]int, 0, size)
	collection := make([]int, 0, size)
	for indx := range array {
		target := array[indx]
		collection_pos := example2.BinarySearchLeft(collection, target)

		if collection_pos == -1 {
			collection = append(collection, target)
		}

	}
	return result
}
