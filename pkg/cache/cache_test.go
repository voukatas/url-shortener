package cache

import "testing"

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache(3)
	key, value := "testKey", "testValue"
	cache.Set(key, value)

	if v, err := cache.Get(key); err != nil || v != value {
		t.Errorf(`Cache.Set("%s", "%s") = %s; want %s`, key, value, v, value)
	}
}
