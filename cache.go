package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	DefaultExpiration      = time.Minute * 60
	CleanerGoroutinesCount = 5
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
		item.expireOn = item.createdAt.Add(expiration)
	}

	c.Lock()
	c.items[key] = item
	c.Unlock()
}

func (c cache) Get(key string) (val interface{}, err error) {
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

func (c *cache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.items[key]; ok {
		delete(c.items, key)

		return nil
	}

	return errors.New("Item not found")
}

func (c *cache) DeleteIsExpired(key string) error {
	c.Lock()
	defer c.Unlock()

	if item, ok := c.items[key]; ok {
		if item.expireOn.Compare(time.Now()) < 1 {
			delete(c.items, key)

			return nil
		}

		return errors.New("Item not expired")
	}

	return errors.New("Item not found")
}

func (c *cache) clean() {
	itemKeys := make(chan string, len(c.items))
	for key := range c.items {
		itemKeys <- key
	}
	close(itemKeys)

	for i := 0; i < CleanerGoroutinesCount; i++ {
		go func() {
			for key := range itemKeys {
				c.DeleteIsExpired(key)
			}
		}()
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
