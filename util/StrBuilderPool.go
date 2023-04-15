package Util

import (
	"strings"
	"sync"
)

const bufferSize = 1024

const (
	StrBuilderPoolEnable = true
)

var StrBuilderPool = sync.Pool{
	New: func() any {
		a := strings.Builder{}
		a.Grow(bufferSize)
		return &a
	},
}
