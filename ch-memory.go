package zch

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Item struct {
	Object     interface{}
	Expiration int64
}

func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Memory struct {
	*memory
}

type memory struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	janitor           *janitor
}

func (c *memory) Set(k string, x interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
	c.mu.Unlock()
}

func (c *memory) set(k string, x interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
}

func (c *memory) SetDefault(k string, x interface{}) {
	c.Set(k, x, DefaultExpiration)
}

func (c *memory) SetNX(k string, x interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.get(k)
	if found {
		return fmt.Errorf("item %s already exists", k)
	}
	c.set(k, x, d)
	return nil
}

func (c *memory) Replace(k string, x interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.get(k)
	if !found {
		return fmt.Errorf("item %s doesn't exist", k)
	}
	c.set(k, x, d)
	return nil
}

func (c *memory) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Object, true
}

func (c *memory) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, time.Time{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, time.Time{}, false
		}
		return item.Object, time.Unix(0, item.Expiration), true
	}
	return item.Object, time.Time{}, true
}

func (c *memory) get(k string) (interface{}, bool) {
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Object, true
}

func (c *memory) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, k)
}

func (c *memory) DeleteExpired() {
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			c.Delete(k)
		}
	}
}

func (c *memory) Items() map[string]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]Item, len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.Expiration > 0 {
			if now > v.Expiration {
				continue
			}
		}
		m[k] = v
	}
	return m
}

func (c *memory) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := len(c.items)
	return n
}

func (c *memory) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *memory) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *Memory) {
	c.janitor.stop <- true
}

func runJanitor(c *memory, ci time.Duration) {
	j := &janitor{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

func newMemory(de time.Duration, m map[string]Item) *memory {
	if de == 0 {
		de = -1
	}
	c := &memory{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newMemoryWithJanitor(de time.Duration, ci time.Duration, m map[string]Item) *Memory {
	c := newMemory(de, m)
	C := &Memory{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

func NewMemory(defaultExpiration, cleanupInterval time.Duration) *Memory {
	items := make(map[string]Item)
	return newMemoryWithJanitor(defaultExpiration, cleanupInterval, items)
}

func NewMemoryFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *Memory {
	return newMemoryWithJanitor(defaultExpiration, cleanupInterval, items)
}
