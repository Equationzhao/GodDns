/*
 *     @Copyright
 *     @file: DS.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/27 下午11:19
 *     @last modified: 2023/3/27 下午10:49
 *
 *
 *
 */

package Util

import (
	"fmt"
	"strings"
)

// Pair is a struct that contains two variables
type Pair[T, U any] struct {
	First  T
	Second U
}

// NewPair return a new Pair Ptr
func NewPair[T, U any](first T, second U) *Pair[T, U] {
	return &Pair[T, U]{
		First:  first,
		Second: second,
	}
}

// MakePair return a new Pair
func MakePair[T, U any](in ...any) Pair[T, U] {
	if len(in) == 0 {
		return Pair[T, U]{}
	}

	if len(in) != 2 {
		panic("invalid len")
	}

	// panic if input
	return Pair[T, U]{
		First:  in[0].(T),
		Second: in[1].(U),
	}
}

func (p *Pair[T, U]) Set(first T, second U) {
	p.First = first
	p.Second = second
}

func ExchangePairs[T, U any](a, b *Pair[T, U]) {
	*a, *b = *b, *a
}

// Clear the pair
func (p *Pair[T, U]) Clear() {
	*p = Pair[T, U]{}
}

// --------------------------------------------------------------//

type emptyType = struct{}

var empty emptyType

type Set[T comparable] struct {
	m map[T]emptyType
}

func (s *Set[T]) Pop() (v T, ok bool) {
	for t := range s.m {
		delete(s.m, t)
		return t, true
	}
	return v, false
}

func (s *Set[T]) Clone() *Set[T] {
	cloned := Set[T]{m: make(map[T]emptyType, len(s.m))}
	for t := range s.m {
		cloned.m[t] = empty
	}
	return &cloned
}

func (s *Set[T]) String() string {
	v := make([]string, 0, len(s.m))
	for t, _ := range s.m {
		v = append(v, fmt.Sprint(t))
	}
	return "set[" + strings.Join(v, " ") + "]"
}

// NewSet create a new set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		m: make(map[T]emptyType),
	}
}

// Add elements
func (s *Set[T]) Add(val ...T) {
	for _, v := range val {
		s.m[v] = empty
	}
}

// Remove elements
func (s *Set[T]) Remove(val ...T) {
	for _, v := range val {
		delete(s.m, v)
	}
}

// Contains check if set contains element
func (s *Set[T]) Contains(val T) bool {
	_, ok := s.m[val]
	return ok
}

func (s *Set[T]) ContainsAll(val ...T) bool {
	for _, v := range val {
		if !s.Contains(v) {
			return false
		}
	}
	return true
}

// Len return the length of set
func (s *Set[T]) Len() int {
	return len(s.m)
}

// Clear  all elements
func (s *Set[T]) Clear() {
	s.m = make(map[T]emptyType)
}

// Items return all items in Slice
func (s *Set[T]) Items() []T {
	item := make([]T, 0, len(s.m))
	for k := range s.m {
		item = append(item, k)
	}
	return item
}

// ToSlice return all items in Slice
func (s *Set[T]) ToSlice() []T {
	return s.Items()
}

// Diff return the difference of two sets
func (s *Set[T]) Diff(other *Set[T]) *Set[T] {
	diff := make(map[T]emptyType, len(s.m))
	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			diff[k] = empty
		}
	}
	return &Set[T]{m: diff}
}

// Equals check if two sets are equal
func (s *Set[T]) Equals(other *Set[T]) bool {
	if other == nil || s.Len() != other.Len() {
		return false
	}

	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			return false
		}
	}
	return true
}

// IsSubOf check if s is a subset of other
func (s *Set[T]) IsSubOf(other *Set[T]) bool {
	if other == nil {
		return false
	}

	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			return false
		}
	}
	return true
}

// IsSuperOf check if s is a super set of other
func (s *Set[T]) IsSuperOf(other *Set[T]) bool {
	if other == nil {
		return false
	}

	for k := range other.m {
		if _, ok := s.m[k]; !ok {
			return false
		}
	}
	return true
}

// IsProperSubOf check if s is a true subset of other
func (s *Set[T]) IsProperSubOf(other *Set[T]) bool {
	if other == nil || s.Len() >= other.Len() {
		return false
	}

	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			return false
		}
	}
	return true
}

// IsProperSuperOf check if s is a true super set of other
func (s *Set[T]) IsProperSuperOf(other *Set[T]) bool {
	if other == nil || s.Len() <= other.Len() {
		return false
	}

	for k := range other.m {
		if _, ok := s.m[k]; !ok {
			return false
		}
	}
	return true
}

// RemoveDuplicate remove duplicate elements in slice
func RemoveDuplicate[T comparable](slice *[]T) {
	set := NewSet[T]()
	set.Add(*slice...)
	sliceTemp := set.Items()
	*slice = sliceTemp
}
