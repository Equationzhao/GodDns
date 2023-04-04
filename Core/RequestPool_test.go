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

func BenchmarkClientPool(b *testing.B) {
	r := resty.New()

	pool := NewClientPool(*r)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			{
				get := pool.Get()
				_, err := get.First.R().Get("http://ident.me")
				if err != nil {
					return
				}
				get.Release()

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
