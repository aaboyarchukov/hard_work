package example1

func BinarySearchLeft(array []int, target int) int {
	left, right := 0, len(array)-1

	for left < right {
		middle := (left + right) / 2

		if array[middle] >= target {
			return middle
		}

		if array[middle] < target {
			left = middle + 1
		}
	}

	return -1
}

func StrictlyMonotonousSequence(array []int) []int {
	size := len(array)
	if size <= 1 {
		return array
	}

	result := make([]int, 0, size)
	collection := make([]int, 0, size)
	for indx := range array {
		target := array[indx]
		collection_pos := BinarySearchLeft(collection, target)

		if collection_pos == -1 {
			collection = append(collection, target)
		}

	}
	return result
}
