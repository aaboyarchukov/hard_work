package example1

func StrictlyMonotonousSequence(array []int) []int {
	size := len(array)
	if size <= 1 {
		return array
	}

	result := make([]int, 0, size)
	return result
}
