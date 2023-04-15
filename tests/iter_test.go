package tests

import (
	"fmt"
	"testing"

	"GodDns/util"
)

func TestNewIter(t *testing.T) {
	i := util.NewIter[int](nil)
	// i.NotLast() panic
	if i.Valid() == true {
		t.Errorf("i.Valid() should be false, but got %v", i.Valid())
	}

	i = util.NewIter(new([]int))
	i.NotLast()
	if i.Valid() == true {
		t.Errorf("i.Valid() should be false, but got %v", i.Valid())
	}

	s := []int{1, 2, 3, 4, 5}
	i = util.NewIter(&s)
	for i.NotLast() {
		t.Log(i.Next())
	}

	if i.Valid() == true {
		t.Errorf("i.Valid() should be false, but got %v", i.Valid())
	}

	for i.NotFirst() {
		t.Log(i.Prev())
	}

	for i.NotLast() {
		*i.GetRaw()++
		i.Next()
	}

	fmt.Println(s) // changed
	for i.NotFirst() {
		i.Prev()
	}

	for i.NotLast() {
		a := i.Get()
		a++
		_ = a
		i.Next()
	}
	fmt.Println(s) // no change

	i = util.NewCopyIter(s)
	for i.NotLast() {
		*i.GetRaw()++
		i.Next()
	}

	fmt.Println(s) // no change
}
