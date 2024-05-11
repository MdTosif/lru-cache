package lru

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type entry struct {
	Key       string // key of the cache
	Value     string // value of the cache
	Timestamp time.Time // timestamp when it get added
}

type LRUCache struct {
	capacity   int // if the number of the cache item exceed capacity, it will delete the oldest elements
	cache      map[string]*list.Element // pointer of the linked list
	eviction   *list.List    // list of the elements with the new to oldest accessed item
	expiration time.Duration // duration after the lru cache will get deleted
	mutex      sync.Mutex    // mutex to use lock resource
}

// initialize the lru cache
func NewLRUCache(capacity int, expiration time.Duration) *LRUCache {
	return &LRUCache{
		capacity:   capacity, 
		cache:      make(map[string]*list.Element),
		eviction:   list.New(),
		expiration: expiration,
	}
}

// get the key value from cache
func (lru *LRUCache) Get(key string) (*entry, error) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// check if key is in the cache
	if element, found := lru.cache[key]; found {
		// check if it is expired, if yes delete it from cache
		if lru.isExpired(element.Value.(*entry)) {
			lru.eviction.Remove(element)
			delete(lru.cache, key)
			return nil, errors.New("key is expired")
		}
		// move the key to front of linked list
		lru.eviction.MoveToFront(element)
		return element.Value.(*entry), nil
	}
	return nil, errors.New("key doesn't exist")
}

// get all the keys, value & timestamp cache
func (lru *LRUCache) GetAll() ([]*entry, error) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	var entries []*entry

	f := lru.eviction.Front()

	// going through the linked list of cache and get the value
	for {
		if f == nil {
			break
		}
		// if it is expired remove it from cache & linked list
		if lru.isExpired(f.Value.(*entry)) {
			lru.eviction.Remove(f)
			delete(lru.cache, f.Value.(*entry).Key)
		} else {
			// if not expired return it
			entries = append(entries, f.Value.(*entry))
		}
		f = f.Next()
	}
	return entries, nil

}

// set key and value in cache
func (lru *LRUCache) Set(key string, value string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// if key exist update the value and timestamp
	if element, found := lru.cache[key]; found {
		lru.eviction.MoveToFront(element)
		element.Value.(*entry).Value = value
		element.Value.(*entry).Timestamp = time.Now()
	} else {
		// evicting the oldest cache in list if cache have more than capacity of the lru
		if len(lru.cache) >= lru.capacity {
			lru.evictOldest()
		}
		// move it in the front of the cache list
		element := lru.eviction.PushFront(&entry{Key: key, Value: value, Timestamp: time.Now()})
		lru.cache[key] = element
	}
}

// delete the key from cache
func (lru *LRUCache) Delete(key string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, found := lru.cache[key]; found {
		lru.eviction.Remove(element)
		delete(lru.cache, key)
	}
}

// evict the oldest cache
func (lru *LRUCache) evictOldest() {
	if lru.eviction.Len() == 0 {
		return
	}
	element := lru.eviction.Back()
	delete(lru.cache, element.Value.(*entry).Key)
	lru.eviction.Remove(element)
}

// check if cache is expired
func (lru *LRUCache) isExpired(e *entry) bool {
	return time.Since(e.Timestamp) > lru.expiration
}
