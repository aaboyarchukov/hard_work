package example2

// [1, 2, 3, 5, 6] 0
func BinarySearchLeft(array []int, target int) int {
	left, right := 0, len(array)-1

	for left < right {
		middle := (left + right) / 2

		if target > array[middle] {
			left = middle + 1
		}

		if target < array[middle] {
			right = middle - 1
		}
	}

	return left
}
