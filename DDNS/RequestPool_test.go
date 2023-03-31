/*
 *
 *     @file: RequestPool_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/31 下午3:16
 *     @last modified: 2023/3/30 下午10:55
 *
 *
 *
 */

package DDNS

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	r := resty.New()

	pool := NewClientPool(*r)
	c := atomic.Int32{}
	for i := 0; i < 5000; i++ {
		go func() {
			get := pool.Get()
			response, err := get.First.R().Get("http://ident.me")
			if err != nil {
				return
			}
			get.Release()
			_ = response
			c.Add(1)
		}()
	}

	time.Sleep(10 * time.Second)
	fmt.Println(pool.Len(), "  ", pool.Available())
	fmt.Println(c.Load())
}

func TestNoPool(t *testing.T) {
	c := atomic.Int32{}

	for i := 0; i < 5000; i++ {
		go func() {
			get := resty.New()
			response, err := get.R().Get("http://ident.me")
			if err != nil {
				return
			}
			_ = response
			c.Add(1)

		}()
	}
	time.Sleep(10 * time.Second)

	fmt.Println(c.Load())
}
