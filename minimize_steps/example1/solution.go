package example1

import (
	"minimize_steps/example2"
	"slices"
)

func StrictlyMonotonousSequence(array []int) []int {
	size := len(array)
	if size <= 1 {
		return array
	}

	collection := make([]int, 0, size)
	parents := make([]int, size)

	for indx := range array {
		target := array[indx]
		collection_pos := example2.BinarySearchLeft(collection, target)

		if collection_pos == len(collection) {
			collection = append(collection, target)
		} else {
			collection[collection_pos] = target
		}

		if collection_pos > 0 {
			parents[collection_pos] = collection[collection_pos-1]
		} else {
			parents[collection_pos] = -1
		}
	}

	collection_size := len(collection)
	result := make([]int, 0, collection_size)

	result = append(result, collection[collection_size-1])
	for indx := collection_size - 1; indx > 0; indx-- {
		result = append(result, parents[indx])
	}

	slices.Reverse(result)

	return result
}
