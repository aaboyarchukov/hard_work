package example1

func StrictlyMonotonousSequence(array []int) []int {
	if len(array) <= 1 {
		return []int{}
	}
	return []int{1}
}
