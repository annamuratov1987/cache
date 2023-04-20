package cache

import (
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
	cache.Set("test-one", 100, time.Millisecond*500)
	cache.Set("test-two", 100, time.Millisecond*1500)

	cache.StartCleaner(time.Second)

	time.Sleep(time.Millisecond * 1100)

	if _, ok := cache.items["test-one"]; ok {
		t.Error("Item not cleaned in expiration time")
	}

	if _, ok := cache.items["test-two"]; !ok {
		t.Error("Item cleaned before expiration time")
	}

	cache.StopCleaner()
}
