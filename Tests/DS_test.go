/*
 *
 *     @file: DS_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/29 下午11:24
 *     @last modified: 2023/3/29 下午6:19
 *
 *
 *
 */

/*
 *
 *     @file: DS_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/28 下午3:59
 *     @last modified: 2023/3/28 下午3:59
 *
 *
 *
 */

package Tests

import (
	"GodDns/Util"
	"fmt"
	"math/rand"
	"testing"
)

func TestRemoveDuplicate(t *testing.T) {
	s := []string{"a", "b", "c", "a", "b", "c", "a", "b", "c"}
	Util.RemoveDuplicate(&s)
	ss := [][]string{{"a", "b", "c"}, {"a", "c", "b"}, {"b", "a", "c"}, {"b", "c", "a"}, {"c", "a", "b"}, {"c", "b", "a"}}
	// shuffle ss
	for i := range ss {
		j := rand.Intn(i + 1)
		ss[i], ss[j] = ss[j], ss[i]
	}

	found := false
	for _, strs := range ss {
		if len(s) != len(strs) {
			continue
		}
		match := true
		for i := range strs {
			if strs[i] != s[i] {
				match = false
				break
			}
		}
		if match {
			found = true
			break
		}
	}
	if !found {
		t.Error("RemoveDuplicate error")
	}
}

func TestPair(t *testing.T) {

	p := Util.NewPair[int, string](0, "")
	p.Set(1, "a")
	if *p.First != 1 || *p.Second != "a" {
		t.Error("Pair set error")
	}

	p.Clear()
	if *p.First != 0 || *p.Second != "" {
		t.Error("Pair clear error")
	}

	p2 := Util.NewPair[int, string](0, "")
	p2.Set(2, "b")
	Util.ExchangePairs(p, p2)

	if *p.First != 2 || *p.Second != "b" || *p2.First != 0 || *p2.Second != "" {
		t.Error("Pairs exchange error")
	}

	p3 := Util.MakePair[int, string]()
	if *p3.First != 0 || *p3.Second != "" {
		t.Error("Make error")
	}

	*p = Util.MakePair[int, string](3, "d")
	if *p.First != 3 || *p.Second != "d" {
		t.Error("Make error")
	}

	pDeepClone := p.Clone()
	fmt.Println("pDeepClone:", pDeepClone)
	fmt.Println("pDeepClone:", &pDeepClone)

	if *pDeepClone.First != 3 || *pDeepClone.Second != "d" || &pDeepClone == p {
		t.Error("Clone error")
	}

	pCopy := p
	if pCopy.First != p.First || pCopy.Second != p.Second {
		t.Error("Copy error")
	}

	a := 3
	b := "a"
	*p = Util.EmplacePair(&a, &b)
	a = 4
	b = "b"
	if &a != p.First || &b != p.Second {
		t.Error("EmplacePair error")
	}

}

func TestSet(t *testing.T) {
	sub := Util.NewSet[int]()
	sub.Add(1)
	sub.Add(2)
	sub.Add(3)

	sup := sub.Clone()

	fmt.Println("s is subset of sup:", sub.IsSubOf(sup))
	fmt.Println("s is proper subset of sup:", sub.IsProperSubOf(sup))
	sup.Add(4)
	fmt.Println("s is subset of sup:", sub.IsSubOf(sup))
	fmt.Println("s is proper subset of sup:", sub.IsProperSubOf(sup))

	fmt.Println("to slice:", sub.ToSlice())
	fmt.Println("items:", sub.Items())

	fmt.Println("sup contains sub:", sup.ContainsAll(sub.Items()...))
	fmt.Println("sub contains sup:", sub.ContainsAll(sup.Items()...))

	fmt.Println("diff:", sup.Diff(sub))
}
