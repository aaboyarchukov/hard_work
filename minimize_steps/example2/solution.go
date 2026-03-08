package example2

func BinarySearchLeft(array []int, target int) int {
	left, right := 0, len(array)

	for left < right {
		middle := (left + right) / 2

		if target > array[middle] {
			left = middle + 1
		}

		if target <= array[middle] {
			right = middle
		}

	}

	return left
}
