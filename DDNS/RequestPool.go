/*
 *
 *     @file: RequestPool.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/30 下午11:29
 *     @last modified: 2023/3/30 下午10:51
 *
 *
 *
 */

package DDNS

import (
	"GodDns/Util"
	"sync"
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

var MainPoolMap ClientPoolMap

const DEFAULTPOOLSIZE = 16
const MAXPOOLSIZE = 10240
const DEFAULTMAPSIZE = 16
const MAXMAPSIZE = 10240
const DEFAULTPOOLADDSTEP = 4

func init() {
	mpm := make(nonLockCpm, DEFAULTMAPSIZE)
	MainPoolMap = ClientPoolMap(Util.EmplacePair(&mpm, &sync.RWMutex{}))
}

const (
	engaged = false
	idle    = true
)

type Client Util.Pair[resty.Client, atomic.Bool]

func NewClient(prototype resty.Client) Client {
	b := new(atomic.Bool)
	b.Store(idle) // set to idle
	return Client(Util.EmplacePair(&prototype, b))
}

func (c *Client) Engage() {
	c.Second.Store(engaged)
}

func (c *Client) Valid() bool {
	return c.First != nil && c.Second != nil
}

func (c *Client) Invalid() bool {
	return c.First != nil || c.Second != nil
}

func (c *Client) Release() {
	c.Second.Store(idle)
}

type ClientPool struct {
	Prototype resty.Client // Should not be modified
	pool      []Client
	m         sync.RWMutex
}

func (c *ClientPool) Available() int {
	c.m.RLock()
	defer c.m.RUnlock()
	var count int
	for _, v := range c.pool {
		if v.Second.Load() {
			count++
		}
	}
	return count
}

// findIdle find an idle Client
// not thread safe
// if found, return the Client and true
// if no available Client, return nil and false
func (c *ClientPool) findIdle() (res Client, ok bool) {
	for _, v := range c.pool {
		if v.Second.Load() {
			return v, true
		}
	}
	return Client{}, false
}

// Init the pool
// shouldn't be called after the pool is used
func (c *ClientPool) Init(capacity int) {
	newPool := make([]Client, 0, capacity)
	for i := 0; i < capacity; i++ {
		newPool = append(newPool, NewClient(c.Prototype))
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.pool = newPool

}

func (c *ClientPool) Size() int {
	c.m.RLock()
	defer c.m.RUnlock()
	return len(c.pool)
}

func (c *ClientPool) Len() int {
	c.m.RLock()
	defer c.m.RUnlock()
	return len(c.pool)
}

func (c *ClientPool) Add() {
	c.m.Lock()
	defer c.m.Unlock()
	for i := 0; i < DEFAULTPOOLADDSTEP; i++ {
		c.add()
	}
}

// add a Client to pool and return the Client ptr
// if reach the max size already, return nil
func (c *ClientPool) add() (Client, bool) {
	var res Client
	ok := len(c.pool) < MAXPOOLSIZE-1
	if ok {
		res = NewClient(c.Prototype)
		c.pool = append(c.pool, res)
	}
	return res, ok
}

// TryGet try to get a client from pool
// if pool is empty, return nil and false
// if no available client, return Pair(nil,nil) and false
func (c *ClientPool) TryGet() (Client, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	if len(c.pool) != 0 {
		r, ok := c.findIdle()
		if ok {
			r.Second.Store(engaged)
			return r, ok
		}

		return r, ok
	} else {
		return Client{}, false
	}
}

// Get a client from pool
// try to get an idle client, if no idle client, add new clients to pool
// if failed to add any clients, return nil
func (c *ClientPool) Get() Client {
	c.m.Lock()
	defer c.m.Unlock()

	r, ok := c.findIdle()
	if ok {
		r.Second.Store(engaged)
		return r
	} else {
		var res Client
		for i := 0; i < DEFAULTPOOLADDSTEP; i++ {
			r, ok := c.add()
			if ok {
				if !res.Valid() {
					res = r
				}
				continue
			} else {
				break
			}
		}
		return res
	}
}

func NewClientPool(prototype resty.Client) *ClientPool {
	return &ClientPool{Prototype: prototype, pool: make([]Client, 0, DEFAULTPOOLSIZE)}
}

type nonLockCpm map[string]*ClientPool
type ClientPoolMap Util.Pair[nonLockCpm, sync.RWMutex]

func (m *ClientPoolMap) ForEach(f func(string, *ClientPool) bool) {
	m.Second.RLock()
	defer m.Second.RUnlock()
	for k, v := range *m.First {
		if f(k, v) {
			return
		}
	}
}

func (m *ClientPoolMap) Size() int {
	m.Second.RLock()
	defer m.Second.RUnlock()
	return len(*m.First)
}

func (m *ClientPoolMap) Get(name string) (pool *ClientPool, ok bool) {
	m.Second.RLock()
	defer m.Second.RUnlock()
	pool, ok = (*m.First)[name]
	return
}

func (m *ClientPoolMap) GetOrCreate(name string, generatePrototype func() (resty.Client, error)) (*ClientPool, error) {
	m.Second.RLock()
	defer m.Second.RUnlock()
	pool, ok := (*m.First)[name]
	if ok {
		return pool, nil
	} else {
		c, err := generatePrototype()
		if err != nil {
			return nil, err
		}
		pool = NewClientPool(c)
		pool.Init(DEFAULTPOOLSIZE)
		(*m.First)[name] = pool
		return pool, nil
	}
}

func (m *ClientPoolMap) Add(name string, pool *ClientPool) (ok bool) {
	m.Second.Lock()
	defer m.Second.Unlock()
	if len(*m.First) >= MAXMAPSIZE {
		return false
	}
	(*m.First)[name] = pool
	return true
}

func (m *ClientPoolMap) Remove(name string) *ClientPoolMap {
	m.Second.Lock()
	defer m.Second.Unlock()
	delete(*m.First, name)
	return m
}
