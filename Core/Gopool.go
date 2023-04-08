package Core

import (
	"GodDns/Util/Json"
	"github.com/go-resty/resty/v2"
	"github.com/panjf2000/ants/v2"
	"sync"
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
			r.JSONUnmarshal = Json.Unmarshal
			r.JSONMarshal = Json.Marshal
			return r
		},
	}

	for i := 0; i < DEFAULTPOOLSIZE; i++ {
		MainClientPool.Put(MainClientPool.New())
	}

}
