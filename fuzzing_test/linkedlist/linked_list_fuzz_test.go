package linkedlist

import (
	"encoding/binary"
	"testing"
)

func bytesToInts(b []byte, maxN int) []int {
	n := len(b) / 2
	if n > maxN {
		n = maxN
	}
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		u := binary.LittleEndian.Uint16(b[i*2 : i*2+2])
		out = append(out, int(int16(u)))
	}
	return out
}

func intsToBytes(ints []int) []byte {
	out := make([]byte, 2*len(ints))
	for i, v := range ints {
		binary.LittleEndian.PutUint16(out[i*2:i*2+2], uint16(int16(v)))
	}
	return out
}

func listToSlice(l *LinkedList, limit int) []int {
	out := make([]int, 0)
	steps := 0
	for n := l.head; n != nil; n = n.next {
		out = append(out, n.value)
		steps++
		if steps > limit {
			break
		}
	}
	return out
}
func FuzzGetLinkedList(f *testing.F) {
	f.Add(intsToBytes([]int{}))
	f.Add(intsToBytes([]int{1, 2, 3}))
	f.Add(intsToBytes([]int{0, 0, 0, 0, 0}))
	f.Add(intsToBytes([]int{1, 2, 3, 3, 2, 1, 1, 2, 3}))

	f.Fuzz(func(t *testing.T, values []byte) {
		ints := bytesToInts(values, 100)
		if len(values) > 3000 {
			t.Skip()
		}

		list := GetLinkedList(ints)

		if actualLen, expectedLen := list.Count(), len(ints); actualLen != expectedLen {
			t.Fatalf("Count mismatch: got=%d want=%d", actualLen, expectedLen)
		}

		if len(ints) == 0 {
			if list.head != nil || list.tail != nil {
				t.Fatalf("empty list must have nil head/tail")
			}
			return
		}

		if list.head == nil || list.tail == nil {
			t.Fatalf("non-empty list must have non-nil head/tail")
		}

		if list.tail.next != nil {
			t.Fatalf("tail.next must be nil")
		}

		steps := 0
		for n := list.head; n != nil; n = n.next {
			steps++
			if steps > len(ints)+1 {
				t.Fatalf("possible cycle detected, steps=%d len=%d", steps, len(ints))
			}
		}
	})
}
func FuzzAddInTail(f *testing.F) {
	f.Add(intsToBytes([]int{}))
	f.Add(intsToBytes([]int{1}))
	f.Add(intsToBytes([]int{1, 2, 3}))
	f.Add(intsToBytes([]int{0, 0, 0, 0, 0}))
	f.Add(intsToBytes([]int{1, 2, 3, 3, 2, 1, 1, 2, 3}))
	f.Add(intsToBytes([]int{-1, 0, 1, -1, 32767, -32768}))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 3000 {
			t.Skip()
		}

		ints := bytesToInts(data, 200)

		var l LinkedList
		for _, v := range ints {
			l.AddInTail(Node{value: v})
		}

		if got, want := l.Count(), len(ints); got != want {
			t.Fatalf("Count mismatch: got=%d want=%d ints=%v", got, want, ints)
		}

		if len(ints) == 0 {
			if l.head != nil || l.tail != nil {
				t.Fatalf("empty list must have nil head/tail")
			}
			return
		}
		if l.head == nil || l.tail == nil {
			t.Fatalf("non-empty list must have non-nil head/tail")
		}
		if l.tail.next != nil {
			t.Fatalf("tail.next must be nil")
		}
		if l.head.value != ints[0] {
			t.Fatalf("head.value mismatch: got=%d want=%d", l.head.value, ints[0])
		}
		if l.tail.value != ints[len(ints)-1] {
			t.Fatalf("tail.value mismatch: got=%d want=%d", l.tail.value, ints[len(ints)-1])
		}

		gotSlice := listToSlice(&l, len(ints)+1)
		if len(gotSlice) != len(ints) {
			t.Fatalf("traversal len mismatch (cycle or broken links?): got=%d want=%d got=%v want=%v",
				len(gotSlice), len(ints), gotSlice, ints)
		}
		for i := range ints {
			if gotSlice[i] != ints[i] {
				t.Fatalf("order mismatch at %d: got=%v want=%v", i, gotSlice, ints)
			}
		}
	})
}

func FuzzLinkedList_Find(f *testing.F) {
	f.Add(intsToBytes([]int{}), 1)
	f.Add(intsToBytes([]int{1, 2, 3}), 2)
	f.Add(intsToBytes([]int{5, 5, 5}), 5)
	f.Add(intsToBytes([]int{-1, 0, 1}), -1)

	f.Fuzz(func(t *testing.T, data []byte, target int) {
		if len(data) > 3000 {
			t.Skip()
		}
		ints := bytesToInts(data, 300)
		list := GetLinkedList(ints)

		if len(ints) == 0 {
			if list.head != nil || list.tail != nil {
				t.Fatalf("empty list must have nil head/tail")
			}
		} else {
			if list.head == nil || list.tail == nil {
				t.Fatalf("non-empty list must have non-nil head/tail")
			}
			if list.tail.next != nil {
				t.Fatalf("tail.next must be nil")
			}
			if got, want := list.Count(), len(ints); got != want {
				t.Fatalf("Count mismatch: got=%d want=%d", got, want)
			}
		}

		wantExists := false
		for _, v := range ints {
			if v == target {
				wantExists = true
				break
			}
		}

		gotNode, err := list.Find(target)
		if wantExists {
			if err != nil {
				t.Fatalf("Find should succeed, got err=%v (target=%d ints=%v)", err, target, ints)
			}
			if gotNode.value != target {
				t.Fatalf("Find returned wrong value: got=%d want=%d", gotNode.value, target)
			}
		} else {
			if err == nil {
				t.Fatalf("Find should fail, but err=nil (target=%d ints=%v gotNode=%v)", target, ints, gotNode)
			}
			if gotNode.value != -1 {
				t.Fatalf("Find miss should return sentinel -1, got=%d", gotNode.value)
			}
		}
	})
}

func FuzzLinkedList_FindAll(f *testing.F) {
	f.Add(intsToBytes([]int{}), 1)
	f.Add(intsToBytes([]int{1, 2, 3}), 2)
	f.Add(intsToBytes([]int{7, 7, 7, 1}), 7)
	f.Add(intsToBytes([]int{-1, 0, -1, 0}), 0)

	f.Fuzz(func(t *testing.T, data []byte, target int) {
		if len(data) > 3000 {
			t.Skip()
		}
		ints := bytesToInts(data, 300)
		list := GetLinkedList(ints)

		if len(ints) == 0 {
			if list.head != nil || list.tail != nil {
				t.Fatalf("empty list must have nil head/tail")
			}
		} else {
			if list.head == nil || list.tail == nil {
				t.Fatalf("non-empty list must have non-nil head/tail")
			}
			if list.tail.next != nil {
				t.Fatalf("tail.next must be nil")
			}
		}

		wantCount := 0
		for _, v := range ints {
			if v == target {
				wantCount++
			}
		}

		nodes := list.FindAll(target)
		if got, want := len(nodes), wantCount; got != want {
			t.Fatalf("FindAll count mismatch: got=%d want=%d target=%d ints=%v", got, want, target, ints)
		}
		for i := range nodes {
			if nodes[i].value != target {
				t.Fatalf("FindAll wrong element at %d: got=%d want=%d", i, nodes[i].value, target)
			}
		}
	})
}

func FuzzLinkedList_Delete(f *testing.F) {
	f.Add(intsToBytes([]int{}), 1, false)
	f.Add(intsToBytes([]int{1, 2, 3}), 2, false)
	f.Add(intsToBytes([]int{1, 2, 2, 3}), 2, true)
	f.Add(intsToBytes([]int{5}), 5, false)
	f.Add(intsToBytes([]int{7, 7, 7}), 7, false)
	f.Add(intsToBytes([]int{-1, 0, -1, 0}), -1, true)

	f.Fuzz(func(t *testing.T, data []byte, target int, all bool) {
		if len(data) > 3000 {
			t.Skip()
		}
		ints := bytesToInts(data, 300)
		list := GetLinkedList(ints)

		list.Delete(target, all)

		want := make([]int, 0, len(ints))
		deleted := false
		for _, v := range ints {
			if v == target {
				if all {
					continue
				}
				if !deleted {
					deleted = true
					continue
				}
			}
			want = append(want, v)
		}

		got := make([]int, 0, len(ints))
		steps := 0
		for n := list.head; n != nil; n = n.next {
			got = append(got, n.value)
			steps++
			if steps > len(ints)+5 {
				t.Fatalf("possible cycle detected after Delete: steps=%d", steps)
			}
		}

		if len(got) != len(want) {
			t.Fatalf("Delete len mismatch: got=%d want=%d got=%v want=%v target=%d all=%v ints=%v",
				len(got), len(want), got, want, target, all, ints)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Delete mismatch at %d: got=%v want=%v target=%d all=%v ints=%v",
					i, got, want, target, all, ints)
			}
		}

		if len(want) == 0 {
			if list.head != nil || list.tail != nil {
				t.Fatalf("after Delete resulting empty list must have nil head/tail")
			}
		} else {
			if list.head == nil || list.tail == nil {
				t.Fatalf("after Delete non-empty list must have non-nil head/tail")
			}
			if list.tail.next != nil {
				t.Fatalf("after Delete tail.next must be nil")
			}
			if gotCount, wantCount := list.Count(), len(want); gotCount != wantCount {
				t.Fatalf("after Delete Count mismatch: got=%d want=%d", gotCount, wantCount)
			}
		}
	})
}

func FuzzLinkedList_Insert_AfterExistingNode(f *testing.F) {
	f.Add(intsToBytes([]int{1}), 0, 9)
	f.Add(intsToBytes([]int{1, 2, 3}), 1, 99)
	f.Add(intsToBytes([]int{5, 6, 7, 8}), 3, -1)

	f.Fuzz(func(t *testing.T, data []byte, afterIndex int, addVal int) {
		if len(data) > 3000 {
			t.Skip()
		}
		ints := bytesToInts(data, 300)
		if len(ints) == 0 {
			t.Skip()
		}

		if afterIndex < 0 {
			afterIndex = -afterIndex
		}
		afterIndex %= len(ints)

		list := GetLinkedList(ints)

		var after *Node
		i := 0
		for n := list.head; n != nil; n = n.next {
			if i == afterIndex {
				after = n
				break
			}
			i++
		}
		if after == nil {
			t.Skip()
		}

		list.Insert(after, Node{value: addVal})

		want := make([]int, 0, len(ints)+1)
		want = append(want, ints[:afterIndex+1]...)
		want = append(want, addVal)
		want = append(want, ints[afterIndex+1:]...)

		got := make([]int, 0, len(want))
		steps := 0
		for n := list.head; n != nil; n = n.next {
			got = append(got, n.value)
			steps++
			if steps > len(want)+5 {
				t.Fatalf("possible cycle detected after Insert: steps=%d", steps)
			}
		}

		if len(got) != len(want) {
			t.Fatalf("Insert len mismatch: got=%d want=%d got=%v want=%v afterIndex=%d addVal=%d ints=%v",
				len(got), len(want), got, want, afterIndex, addVal, ints)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Insert mismatch at %d: got=%v want=%v afterIndex=%d addVal=%d ints=%v",
					i, got, want, afterIndex, addVal, ints)
			}
		}

		if list.head == nil || list.tail == nil {
			t.Fatalf("after Insert list must have non-nil head/tail")
		}
		if list.tail.next != nil {
			t.Fatalf("after Insert tail.next must be nil")
		}
		if gotCount, wantCount := list.Count(), len(want); gotCount != wantCount {
			t.Fatalf("after Insert Count mismatch: got=%d want=%d", gotCount, wantCount)
		}
	})
}

func FuzzLinkedList_InsertFirst(f *testing.F) {
	f.Add(intsToBytes([]int{}), 10)
	f.Add(intsToBytes([]int{1, 2, 3}), 0)
	f.Add(intsToBytes([]int{-1, 0, 1}), -32768)

	f.Fuzz(func(t *testing.T, data []byte, firstVal int) {
		if len(data) > 3000 {
			t.Skip()
		}
		ints := bytesToInts(data, 300)
		list := GetLinkedList(ints)

		list.InsertFirst(Node{value: firstVal})

		want := make([]int, 0, len(ints)+1)
		want = append(want, firstVal)
		want = append(want, ints...)

		got := make([]int, 0, len(want))
		steps := 0
		for n := list.head; n != nil; n = n.next {
			got = append(got, n.value)
			steps++
			if steps > len(want)+5 {
				t.Fatalf("possible cycle detected after InsertFirst: steps=%d", steps)
			}
		}

		if len(got) != len(want) {
			t.Fatalf("InsertFirst len mismatch: got=%d want=%d got=%v want=%v firstVal=%d ints=%v",
				len(got), len(want), got, want, firstVal, ints)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("InsertFirst mismatch at %d: got=%v want=%v firstVal=%d ints=%v",
					i, got, want, firstVal, ints)
			}
		}

		if list.head == nil || list.tail == nil {
			t.Fatalf("after InsertFirst list must have non-nil head/tail")
		}
		if list.tail.next != nil {
			t.Fatalf("after InsertFirst tail.next must be nil")
		}
		if gotCount, wantCount := list.Count(), len(want); gotCount != wantCount {
			t.Fatalf("after InsertFirst Count mismatch: got=%d want=%d", gotCount, wantCount)
		}
	})
}
