// Package DDNS
// Define ClientPool and  ClientPoolMap to manage ClientPool
package DDNS

import (
	"GodDns/Util/Collections"
	"sync"
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

var MainPoolMap ClientPoolMap

const _DEFAULTPOOLSIZE = 16
const _MAXPOOLSIZE = 10240
const _DEFAULTMAPSIZE = 16
const _MAXMAPSIZE = 10240
const _DEFAULTPOOLADDSTEP = 4

func init() {
	mpm := make(nonLockCpm, _DEFAULTMAPSIZE)
	MainPoolMap = ClientPoolMap(Collections.EmplacePair(&mpm, &sync.RWMutex{}))
}

const (
	engaged = false
	idle    = true
)

// Client is a pair of resty.Client and atomic.Bool
// after using the resty.Client, call Release() to set the Bool to idle
type Client Collections.Pair[resty.Client, atomic.Bool]

func NewClient(prototype resty.Client) Client {
	b := new(atomic.Bool)
	b.Store(idle) // set to idle
	return Client(Collections.EmplacePair(&prototype, b))
}

// Engage set the Bool to engaged
// which means the Client is in use
func (c *Client) Engage() {
	c.Second.Store(engaged)
}

// Valid check if the Client is valid
func (c *Client) Valid() bool {
	return c.First != nil && c.Second != nil
}

// Invalid check if the Client is invalid
func (c *Client) Invalid() bool {
	return c.First != nil || c.Second != nil
}

// Release set the Bool to idle
// which means the Client is not in use
func (c *Client) Release() {
	c.Second.Store(idle)
}

type ClientPool struct {
	Prototype resty.Client // Should not be modified
	pool      []Client
	m         sync.RWMutex
}

// Available count the available Client
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

// Add num=_DEFAULTPOOLADDSTEP Clients to pool
func (c *ClientPool) Add() {
	c.m.Lock()
	defer c.m.Unlock()
	for i := 0; i < _DEFAULTPOOLADDSTEP; i++ {
		c.add()
	}
}

// add a Client to pool and return the Client ptr
// if reach the max size already, return Client contains nil and false
func (c *ClientPool) add() (Client, bool) {
	var res Client
	ok := len(c.pool) < _MAXPOOLSIZE-1
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

		for i := 0; i < _DEFAULTPOOLADDSTEP; i++ {
			r, ok := c.add()
			if !ok {
				break
			}
			if !res.Valid() {
				res = r
			}
		}
		return res
	}
}

// NewClientPool create a new ClientPool with prototype
func NewClientPool(prototype resty.Client) *ClientPool {
	return &ClientPool{Prototype: prototype, pool: make([]Client, 0, _DEFAULTPOOLSIZE)}
}

type nonLockCpm map[string]*ClientPool

// ClientPoolMap is a map of ClientPool with a RWMutex
type ClientPoolMap Collections.Pair[nonLockCpm, sync.RWMutex]

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

// Get a ClientPool from map
// if not exist, return nil and false
func (m *ClientPoolMap) Get(name string) (pool *ClientPool, ok bool) {
	m.Second.RLock()
	defer m.Second.RUnlock()
	pool, ok = (*m.First)[name]
	if ok {
		return pool, true
	}
	return nil, false
}

// GetOrCreate Get a ClientPool from map, if not exist, create one using generatePrototype func
// generatePrototype should return a new Client and nil if success, or nil and error if failed
func (m *ClientPoolMap) GetOrCreate(name string, generatePrototype func() (resty.Client, error)) (*ClientPool, error) {
	m.Second.Lock()
	defer m.Second.Unlock()
	pool, ok := (*m.First)[name]
	if ok {
		return pool, nil
	} else {
		c, err := generatePrototype()
		if err != nil {
			return nil, err
		}
		pool = NewClientPool(c)
		pool.Init(_DEFAULTPOOLSIZE)
		(*m.First)[name] = pool
		return pool, nil
	}
}

// Add a ClientPool to map
// if the map is full, return false
func (m *ClientPoolMap) Add(name string, pool *ClientPool) (ok bool) {
	m.Second.Lock()
	defer m.Second.Unlock()
	if len(*m.First) >= _MAXMAPSIZE {
		return false
	}
	(*m.First)[name] = pool
	return true
}

// Remove a ClientPool from map
func (m *ClientPoolMap) Remove(name string) *ClientPoolMap {
	m.Second.Lock()
	defer m.Second.Unlock()
	delete(*m.First, name)
	return m
}
