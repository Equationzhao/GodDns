package Util

import "fmt"

type Iter[T any] struct {
	slice  *[]T
	length int // current index
}

// Len return length of slice
func (i *Iter[T]) Len() int {
	return len(*i.slice)
}

func NewCopyIter[T any](slice []T) *Iter[T] {
	copied := make([]T, 0,len(slice))
	copy(copied,slice)
	return &Iter[T]{
		slice:  &copied,
		length: 0,
	}
}

// Valid return true if current element is valid
// if slice is nil, return false
// if slice is empty, return false
func (i *Iter[T]) Valid() bool {
	return i.slice != nil && i.length >= 0 && i.length < len(*i.slice)
}

// NewIter return a new iter
// start from 0
func NewIter[T any](slice *[]T) *Iter[T] {
	return &Iter[T]{
		slice:  slice,
		length: 0,
	}
}

// NotFirst return true if current element is not the first element
// regardless of whether the slice is nil/empty or not
func (i *Iter[T]) NotFirst() bool {
	return i.length > 0
}

// NotLast return true if current element is not the last element
// if slice is nil, panic
// if slice is empty, return false
func (i *Iter[T]) NotLast() bool {
	return i.length < len(*i.slice)
}

// Next return current element and iterate to next element
// if no next element, return false and stop at the last element
func (i *Iter[T]) Next() (elem T) {
	i.length++
	return (*i.slice)[i.length-1]
}

// Prev return current element and move to prev element
// if no prev element, return false and stop at the first element
func (i *Iter[T]) Prev() (elem T) {
	i.length--
	return (*i.slice)[i.length]
}

// Get return current element
func (i *Iter[T]) Get() T {
	return (*i.slice)[i.length]
}

type ErrOutOfRange struct {
	length int
	index  int
}

func (e ErrOutOfRange) Error() string {
	return fmt.Sprintf("index out of range [%d] with length %d", e.length, e.index)
}

func (i *Iter[T]) TryGet() (T, error) {
	if i.length < len(*i.slice) {
		return (*i.slice)[i.length], nil
	} else {
		var t T
		return t, ErrOutOfRange{length: len(*i.slice), index: i.length}
	}
}

// GetRaw return current element's pointer
func (i *Iter[T]) GetRaw() *T {
	return &(*i.slice)[i.length]
}

func (i *Iter[T]) TryGetRaw() (*T, error) {
	if i.length < len(*i.slice) {
		return &(*i.slice)[i.length], nil
	} else {
		var t T
		return &t, ErrOutOfRange{length: len(*i.slice), index: i.length}
	}
}
