package cache

import (
	"strconv"
	"sync"
	"testing"
)

func TestLRUEviction(t *testing.T) {

	lru := NewLRUCache(2)

	lru.Set("a", "va")
	lru.Set("b", "vb")

	expected := "va"
	val, err := lru.Get("a")
	if err != nil {
		t.Error(err)
	}
	if val != expected {
		t.Errorf("expected %v received %v", expected, val)
	}

	expected = "vb"
	val, err = lru.Get("b")
	if err != nil {
		t.Error(err)
	}
	if val != expected {
		t.Errorf("expected %v received %v", expected, val)
	}

	lru.Set("c", "vc")
	expected = "vc"
	val, err = lru.Get("c")
	if err != nil {
		t.Error(err)
	}
	if val != expected {
		t.Errorf("expected %v received %v", expected, val)
	}

	val, err = lru.Get("a")
	if err == nil {
		t.Error(err)
	}

}

func TestLRUConcurrency(t *testing.T) {
	lru := NewLRUCache(26)
	var wg sync.WaitGroup
	actions := 100000

	for i := 0; i < actions; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string("key_" + strconv.Itoa(i%26))
			value := "value_" + strconv.Itoa(i)
			lru.Set(key, value)
			if _, err := lru.Get(key); err != nil {
				t.Errorf("Expected key %v to exist", key)
			}
		}(i)
	}

	wg.Wait()

}
