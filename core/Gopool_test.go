package core

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

func TestPool(t *testing.T) {
	c := atomic.Int32{}
	for i := 0; i < 5000; i++ {
		go func() {
			get := MainClientPool.Get().(*resty.Client)
			defer MainClientPool.Put(get)
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

func BenchmarkClientPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			{
				get := MainClientPool.Get().(*resty.Client)
				response, err := get.R().Get("http://ident.me")
				if err != nil {
					return
				}
				_ = response
				MainClientPool.Put(get)
			}
		}
	})
}

func BenchmarkNopool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			{
				get := resty.New()
				_, err := get.R().Get("http://ident.me")
				if err != nil {
					return
				}
			}
		}
	})
}
