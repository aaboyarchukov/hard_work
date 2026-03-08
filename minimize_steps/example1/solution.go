package example1

func StrictlyMonotonousSequence(array []int) []int {
	if len(array) <= 1 {
		return array
	}
	return []int{1}
}
