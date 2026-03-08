package example2

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
