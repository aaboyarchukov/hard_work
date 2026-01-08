package linkedlist

import (
	"errors"
	"testing"
)

func TestAdditionLL(t *testing.T) {
	tests := []struct {
		name    string
		inputL1 *LinkedList
		inputL2 *LinkedList
		err     error
		wantL3  *LinkedList
	}{
		{"Test1: ", GetLinkedList([]int{}), GetLinkedList([]int{}), nil, GetLinkedList([]int{})},
		{"Test2: ", GetLinkedList([]int{22, 3, 2, 45, 6}), GetLinkedList([]int{10, 11, 1, 2, 3}), nil, GetLinkedList([]int{32, 14, 3, 47, 9})},
		{"Test3: ", GetLinkedList([]int{22, 3, 2, 45, 6}), GetLinkedList([]int{10, 11}), errors.New("different lengths"), GetLinkedList([]int{})},
		{"Test4: ", GetLinkedList([]int{22}), GetLinkedList([]int{10}), nil, GetLinkedList([]int{32})},
		{"Test5: ", GetLinkedList([]int{22}), GetLinkedList([]int{10, 11}), errors.New("different lengths"), GetLinkedList([]int{})},
	}

	for _, tempTest := range tests {
		test := tempTest
		resultL, err := GetAdditionalLists(test.inputL1, test.inputL2)
		if !EqualLists(resultL, test.wantL3) && err != test.err {
			t.Errorf("failed %s: additional lists", test.name)
		}
	}
}
