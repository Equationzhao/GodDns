/*
 *     @Copyright
 *     @file: DS_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/22 上午6:29
 *     @last modified: 2023/3/22 上午6:21
 *
 *
 *
 */

package Util

import (
	"GodDns/Util"
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
	p := Util.Pair[int, string]{}
	p.Set(1, "a")
	if p.First != 1 || p.Second != "a" {
		t.Error("Pair set error")
	}

	p.Clear()
	if p.First != 0 || p.Second != "" {
		t.Error("Pair clear error")
	}

	p2 := Util.Pair[int, string]{}
	p2.Set(2, "b")
	Util.ExchangePairs(&p, &p2)

	if p.First != 2 || p.Second != "b" || p2.First != 0 || p2.Second != "" {
		t.Error("Pair set error")
	}

}
