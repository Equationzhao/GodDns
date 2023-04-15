package core

import (
	"sync"

	"GodDns/util/json"
	"github.com/charmbracelet/glamour"
	"github.com/go-resty/resty/v2"
	"github.com/panjf2000/ants/v2"
)

const DEFAULTGOPOOLSIZE = 100

var MainGoroutinePool *ants.Pool

func init() {
	MainGoroutinePool, _ = ants.NewPool(DEFAULTGOPOOLSIZE, ants.WithNonblocking(true))
}

// MainClientPool is a global ClientPool
var MainClientPool *sync.Pool

const DEFAULTPOOLSIZE = 20

func init() {
	MainClientPool = &sync.Pool{
		New: func() any {
			r := resty.New()
			r.JSONUnmarshal = json.Unmarshal
			r.JSONMarshal = json.Marshal
			return r
		},
	}

	for i := 0; i < DEFAULTPOOLSIZE; i++ {
		MainClientPool.Put(MainClientPool.New())
	}
}

var (
	mdRenderer sync.Once
	renderer   *glamour.TermRenderer
)

func GetMDRenderer() *glamour.TermRenderer {
	mdRenderer.Do(
		func() {
			renderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithEmoji(),
			)
		},
	)
	return renderer
}
