package linkedlist

import (
	"errors"
)

// task 8
// Additional list
// t = O(n), where n - len(list), mem = O(n), where n - len(list)
func GetAdditionalLists(l1 *LinkedList, l2 *LinkedList) (*LinkedList, error) {
	L1Count, L2Count := l1.Count(), l2.Count()
	resultValues := make([]int, 0, L1Count)
	if L1Count == L2Count {
		tempNodeL1, tempNodeL2 := l1.head, l2.head
		for tempNodeL1 != nil && tempNodeL2 != nil {
			resultValues = append(resultValues, tempNodeL1.value+tempNodeL2.value)
			tempNodeL1 = tempNodeL1.next
			tempNodeL2 = tempNodeL2.next
		}
		return GetLinkedList(resultValues), nil
	}

	return GetLinkedList(resultValues), errors.New("different lengths")
}
