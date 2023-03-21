/*
 *     @Copyright
 *     @file: DS.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/22 上午6:29
 *     @last modified: 2023/3/22 上午6:21
 *
 *
 *
 */

package Util

// Pair is a struct that contains two variables
type Pair[T, U any] struct {
	First  T
	Second U
}

func (receiver *Pair[T, U]) Set(first T, second U) {
	receiver.First = first
	receiver.Second = second
}

func ExchangePairs[T, U any](a, b *Pair[T, U]) {
	*a, *b = *b, *a
}

// Clear the pair
func (receiver *Pair[T, U]) Clear() {
	*receiver = Pair[T, U]{}
}

// --------------------------------------------------------------//

type emptyType = struct{}

var empty emptyType

type Set[T comparable] struct {
	m map[T]emptyType
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

// Len return the length of set
func (s *Set[T]) Len() int {
	return len(s.m)
}

// Clear  all elements
func (s *Set[T]) Clear() {
	s.m = make(map[T]emptyType)
}

// Items return all items in set
func (s *Set[T]) Items() []T {
	item := make([]T, 0)
	for k := range s.m {
		item = append(item, k)
	}
	return item
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

// IsTrueSubOf check if s is a true subset of other
func (s *Set[T]) IsTrueSubOf(other *Set[T]) bool {
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

// IsTrueSuperOf check if s is a true super set of other
func (s *Set[T]) IsTrueSuperOf(other *Set[T]) bool {
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
