package lru

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type entry struct {
	Key       string
	Value     string
	Timestamp time.Time
}

type LRUCache struct {
	capacity   int // if the number of the cache item exceed capacity, it will delete the oldest elements
	cache      map[string]*list.Element
	eviction   *list.List    // list of the elements with the new to oldest accessed item
	expiration time.Duration // duration after the lru cache will get deleted
	mutex      sync.Mutex    //
}

func NewLRUCache(capacity int, expiration time.Duration) *LRUCache {
	return &LRUCache{
		capacity:   capacity,
		cache:      make(map[string]*list.Element),
		eviction:   list.New(),
		expiration: expiration,
	}
}

func (lru *LRUCache) Get(key string) (*entry, error) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, found := lru.cache[key]; found {
		if lru.isExpired(element.Value.(*entry)) {
			lru.eviction.Remove(element)
			delete(lru.cache, key)
			return nil, errors.New("key is expired")
		}
		lru.eviction.MoveToFront(element)
		return element.Value.(*entry), nil
	}
	return nil, errors.New("key doesn't exist")
}

func (lru *LRUCache) GetAll() ([]*entry, error) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	var entries []*entry

	f := lru.eviction.Front()

	for {
		if f == nil {
			break
		}
		if lru.isExpired(f.Value.(*entry)) {
			lru.eviction.Remove(f)
			delete(lru.cache, f.Value.(*entry).Key)
		} else {

			entries = append(entries, f.Value.(*entry))
		}
		f = f.Next()
	}
	return entries, nil

}

func (lru *LRUCache) Set(key string, value string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, found := lru.cache[key]; found {
		lru.eviction.MoveToFront(element)
		element.Value.(*entry).Value = value
		element.Value.(*entry).Timestamp = time.Now()
	} else {
		if len(lru.cache) >= lru.capacity {
			lru.evictOldest()
		}
		element := lru.eviction.PushFront(&entry{Key: key, Value: value, Timestamp: time.Now()})
		lru.cache[key] = element
	}
}

func (lru *LRUCache) Delete(key string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, found := lru.cache[key]; found {
		lru.eviction.Remove(element)
		delete(lru.cache, key)
	}
}

func (lru *LRUCache) evictOldest() {
	if lru.eviction.Len() == 0 {
		return
	}
	element := lru.eviction.Back()
	delete(lru.cache, element.Value.(*entry).Key)
	lru.eviction.Remove(element)
}

func (lru *LRUCache) isExpired(e *entry) bool {
	return time.Since(e.Timestamp) > lru.expiration
}
