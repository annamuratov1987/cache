package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	DefaultExpiration = time.Minute
)

type cache[T any] struct {
	sync.RWMutex
	items  map[string]Item[T]
	ticker *time.Ticker
}

type Item[T any] struct {
	value     T
	createdAt time.Time
	expireOn  time.Time
}

func New[T any]() *cache[T] {
	return &cache[T]{
		items: make(map[string]Item[T]),
	}
}

func (c *cache[T]) Set(key string, value T, expiration time.Duration) {
	item := Item[T]{
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

func (c cache[T]) Get(key string) (val T, err error) {
	c.RLock()
	defer c.RUnlock()

	el, ok := c.items[key]
	if !ok {
		return val, errors.New("Item not found")
	}

	if el.expireOn.Compare(time.Now()) == -1 {
		return val, errors.New("Item is expired")
	}

	return el.value, nil
}

func (c *cache[T]) Delete(key string) error {
	if _, ok := c.items[key]; ok {
		c.Lock()
		delete(c.items, key)
		c.Unlock()

		return nil
	}

	return errors.New("Item not found")
}

func (c *cache[T]) clean() {
	c.Lock()
	defer c.Unlock()

	for key, item := range c.items {
		if item.expireOn.Compare(time.Now()) < 1 {
			delete(c.items, key)
		}
	}
}

func (c *cache[T]) StartCleaner(cleanupInterval time.Duration) {
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

func (c *cache[T]) StopCleaner() {
	c.ticker.Stop()
}
