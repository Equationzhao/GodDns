/*
 *
 *     @file: Iter_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/29 下午11:24
 *     @last modified: 2023/3/29 下午6:16
 *
 *
 *
 */

/*
 *
 *     @file: Iter_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/29 下午4:39
 *     @last modified: 2023/3/29 下午4:39
 *
 *
 *
 */

package Tests

import (
	"GodDns/Util"
	"fmt"
	"testing"
)

func TestNewIter(t *testing.T) {
	i := Util.NewIter[int](nil)
	// i.NotLast() panic
	if i.Valid() == true {
		t.Error(fmt.Sprintf("i.Valid() should be false, but got %v", i.Valid()))
	}

	i = Util.NewIter[int](new([]int))
	i.NotLast()
	if i.Valid() == true {
		t.Error(fmt.Sprintf("i.Valid() should be false, but got %v", i.Valid()))
	}

	s := []int{1, 2, 3, 4, 5}
	i = Util.NewIter[int](&s)
	for i.NotLast() {
		t.Log(i.Next())
	}

	if i.Valid() == true {
		t.Error(fmt.Sprintf("i.Valid() should be false, but got %v", i.Valid()))
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
		i.Next()
	}
	fmt.Println(s) // no change

	i = Util.NewCopyIter[int](s)
	for i.NotLast() {
		*i.GetRaw()++
		i.Next()
	}

	fmt.Println(s) // no change

}
