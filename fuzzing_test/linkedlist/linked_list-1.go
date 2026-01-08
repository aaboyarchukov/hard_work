package linkedlist

import (
	"errors"
	_ "os"
	_ "reflect"
)

type Node struct {
	next  *Node
	value int
}

type LinkedList struct {
	head *Node
	tail *Node
}

func (l *LinkedList) AddInTail(item Node) {
	if l.head == nil {
		l.head = &item
	} else {
		l.tail.next = &item
	}
	l.tail = &item
}

// task 5
// t = O(n), where n = len(list)
func (l *LinkedList) Count() int {
	var count int
	tempNode := l.head
	for tempNode != nil {
		count++
		tempNode = tempNode.next
	}
	return count
}

func (l *LinkedList) Find(n int) (Node, error) {
	tempNode := l.head
	for tempNode != nil {
		if tempNode.value == n {
			return *tempNode, nil
		}
		tempNode = tempNode.next
	}
	return Node{value: -1, next: nil}, errors.New("node is not finding")
}

// task 4
// t = O(n), where n = len(list)
func (l *LinkedList) FindAll(n int) []Node {
	var nodes []Node
	tempNode := l.head
	for tempNode != nil {
		if tempNode.value == n {
			nodes = append(nodes, *tempNode)
		}
		tempNode = tempNode.next
	}
	return nodes
}

// task 1
// t = O(n), where n = len(list)
// task 2
// t = O(n), where n = len(list)
func (l *LinkedList) Delete(n int, all bool) {
	if l.head == nil {
		return
	}

	tempNode := l.head
	var prev *Node

	if l.Count() == 1 && tempNode.value == n {
		l.Clean()
		return
	}

	for tempNode != nil {
		deleted := false
		if tempNode.value == n && tempNode == l.head {
			l.head = tempNode.next
			deleted = true
		} else if tempNode.value == n && tempNode == l.tail {
			prev.next = nil
			l.tail = prev
			deleted = true
		} else if tempNode.value == n {
			prev.next = tempNode.next
			deleted = true
		}
		if !all && deleted {
			return
		}
		if !deleted {
			prev = tempNode
		}
		tempNode = tempNode.next
	}
}

// task 6
// t = O(n), where n = len(list)
func (l *LinkedList) Insert(after *Node, add Node) {
	if l.head == nil {
		l.InsertFirst(add)
		return
	}
	tempNode := l.head
	// if node will not exists, then we have to finding it first
	for tempNode.value != after.value {
		tempNode = tempNode.next
	}
	if tempNode == l.tail {
		l.AddInTail(add)
	} else {
		nextNode := tempNode.next
		tempNode.next = &add
		add.next = nextNode
	}

}

func (l *LinkedList) InsertFirst(first Node) {
	if l.head == nil {
		l.tail = &first
	} else {
		first.next = l.head
	}
	l.head = &first

}

// task 3
// t = O(1)
func (l *LinkedList) Clean() {
	l.head = nil
	l.tail = nil
}

// func PrintLL(LL *LinkedList) {
// 	temp := LL.head
// 	for temp != nil {
// 		fmt.Printf("%d ", temp.value)
// 		temp = temp.next
// 	}
// 	fmt.Println()
// }

func GetLinkedList(values []int) *LinkedList {
	var resultLL LinkedList // resulting linked list
	for _, value := range values {
		resultLL.AddInTail(Node{
			value: value,
		})
	}
	return &resultLL
}

func EqualLists(l1 *LinkedList, l2 *LinkedList) bool {
	// equals len and elements
	if l1.head == nil &&
		l2.head == nil {
		return true
	}

	if l1.head.value != l2.head.value {
		return false
	}
	if l1.tail.value != l2.tail.value {
		return false
	}

	countL1, countL2 := l1.Count(), l2.Count()
	if countL1 == countL2 {
		tempL1, tempL2 := l1.head, l2.head
		for tempL1 != nil && tempL2 != nil {
			if tempL1.value != tempL2.value {
				return false
			}
			tempL1 = tempL1.next
			tempL2 = tempL2.next
		}

		return true
	}

	return false
}
