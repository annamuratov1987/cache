package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	DefaultExpiration = time.Minute
)

type cache struct {
	sync.RWMutex
	items  map[string]Item
	ticker *time.Ticker
}

type Item struct {
	value     interface{}
	createdAt time.Time
	expireOn  time.Time
}

func New() *cache {
	return &cache{
		items: make(map[string]Item),
	}
}

func (c *cache) Set(key string, value interface{}, expiration time.Duration) {
	item := Item{
		value:     value,
		createdAt: time.Now(),
		expireOn:  time.Now().Add(DefaultExpiration),
	}

	if expiration > 0 {
		item.expireOn = time.Now().Add(expiration)
	}

	c.Lock()
	c.items[key] = item
	c.Unlock()
}

func (c cache) Get(key string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	el, ok := c.items[key]
	if !ok {
		return nil, errors.New("Item not found")
	}

	if el.expireOn.Compare(time.Now()) == -1 {
		return nil, errors.New("Item is expired")
	}

	return el.value, nil
}

func (c *cache) Delete(key string) error {
	if _, ok := c.items[key]; ok {
		c.Lock()
		delete(c.items, key)
		c.Unlock()

		return nil
	}

	return errors.New("Item not found")
}

func (c *cache) clean() {
	c.Lock()
	defer c.Unlock()

	for key, item := range c.items {
		if item.expireOn.Compare(time.Now()) < 1 {
			delete(c.items, key)
		}
	}
}

func (c *cache) StartCleaner(cleanupInterval time.Duration) {
	c.ticker = time.NewTicker(cleanupInterval)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.clean()
			}
		}
	}()
}

func (c *cache) StopCleaner() {
	c.ticker.Stop()
}
