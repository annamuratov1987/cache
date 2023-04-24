package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_Set(t *testing.T) {
	cache := New()
	cache.Set("test", "test", time.Second)
	if val, ok := cache.items["test"]; ok {
		if val.value != "test" {
			t.Error("Item value incorrect")
		}
	} else {
		t.Error("Item not set")
	}
}

func TestCache_Get(t *testing.T) {
	cache := New()
	cache.Set("test", 100, time.Second)

	val, err := cache.Get("test")

	if err != nil {
		t.Error(err)
	} else {
		if val != 100 {
			t.Error("Item value incorrect")
		}
	}

	time.Sleep(time.Second)
	_, err = cache.Get("test")
	if err == nil {
		t.Error("Getting item where over expiration time")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New()
	cache.Set("test", 100, time.Second)

	err := cache.Delete("test")
	if err != nil {
		t.Error(err)
	}
}

func TestCache_Cleaner(t *testing.T) {
	cache := New()

	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, 100, time.Millisecond*500)
	}

	for i := 11; i <= 20; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, 100, time.Millisecond*1500)
	}

	cache.StartCleaner(time.Second)

	time.Sleep(time.Millisecond * 1100)

	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.RLock()
		if _, ok := cache.items[key]; ok {
			t.Error("Item not cleaned in expiration time")
		}
		cache.RUnlock()
	}
	for i := 11; i <= 20; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.RLock()
		if _, ok := cache.items[key]; !ok {
			t.Error("Item cleaned before expiration time")
		}
		cache.RUnlock()
	}

	cache.StopCleaner()
}
