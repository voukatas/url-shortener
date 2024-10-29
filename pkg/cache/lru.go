package cache

import (
	"errors"
	"fmt"
	"sync"
)

type LRUCacheItem struct {
	key      string
	value    string
	previous *LRUCacheItem
	next     *LRUCacheItem
}

type LRUCache struct {
	store    map[string]*LRUCacheItem
	head     *LRUCacheItem
	tail     *LRUCacheItem
	capacity int
	lock     sync.Mutex
	//logger   logger.Logger
}

func NewLRUCacheItem(key string, value string) *LRUCacheItem {
	return &LRUCacheItem{
		key:      key,
		value:    value,
		previous: nil,
		next:     nil,
	}
}

func NewLRUCache(capacity int) Cache {
	return &LRUCache{
		store:    make(map[string]*LRUCacheItem, capacity),
		head:     nil,
		tail:     nil,
		capacity: capacity,
		//logger:   logger,
	}
}

func (lru *LRUCache) removeItemFromQ(item *LRUCacheItem) {
	if item.previous != nil {
		item.previous.next = item.next
	} else {
		lru.head = item.next
	}

	if item.next != nil {
		item.next.previous = item.previous
	} else {
		lru.tail = item.previous
	}

	item.previous = nil
	item.next = nil
}

func (lru *LRUCache) addItemToFrontOfQ(item *LRUCacheItem) {
	item.previous = nil
	if lru.head == nil {
		lru.head = item
		lru.tail = item
		return
	}

	item.next = lru.head
	lru.head.previous = item
	lru.head = item
}

func (lru *LRUCache) moveToFrontOfQ(item *LRUCacheItem) {
	if lru.head == item {
		return
	}
	lru.removeItemFromQ(item)
	lru.addItemToFrontOfQ(item)
}

// Set
func (lru *LRUCache) Set(key string, value string) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	item, exists := lru.store[key]
	if exists {
		item.value = value
		lru.moveToFrontOfQ(item)
		return
	}

	if len(lru.store) >= lru.capacity {
		// evict tail
		fmt.Println("item for eviction", lru.tail.key)
		delete(lru.store, lru.tail.key)
		lru.removeItemFromQ(lru.tail)
	}

	item = NewLRUCacheItem(key, value)
	lru.addItemToFrontOfQ(item)
	lru.store[key] = item
}

// Get
func (lru *LRUCache) Get(key string) (string, error) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	item, exists := lru.store[key]
	if !exists {
		return "", errors.New("Key not found")
	}

	return item.value, nil
}

func (lru *LRUCache) PrintLRU() {
	if lru.head != nil {
		fmt.Println("cache head", lru.head.key)
	} else {
		fmt.Println("cache head is nil")
	}
	if lru.tail != nil {
		fmt.Println("cache tail", lru.tail.key)
	} else {
		fmt.Println("cache tail is nil")
	}
	fmt.Println("LRU Q contents: ")
	currentCacheItem := lru.head
	for currentCacheItem != nil {
		fmt.Println(currentCacheItem.key)
		currentCacheItem = currentCacheItem.next
	}

	fmt.Printf("LRU store contents: %v\n", lru.store)
}
