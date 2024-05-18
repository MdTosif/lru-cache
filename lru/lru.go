package lru

import (
	"errors"
	"fmt"
	"sync"
	"time"

	linkedlist "github.com/mdtosif/lru-go/lru/linked-list"
)

type entry struct {
	Key       string    // key of the cache
	Value     string    // value of the cache
	Timestamp time.Time // timestamp when it get added
}

type LRUCache struct {
	capacity int                                   // if the number of the cache item exceed capacity, it will delete the oldest elements
	cache    map[string]*linkedlist.Element[entry] // pointer of the linked list
	eviction *linkedlist.DoublyLiknedList[entry]   // list of the elements with the new to oldest accessed item
	mutex    sync.Mutex                            // mutex to use lock resource
}

// initialize the lru cache
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*linkedlist.Element[entry]),
		eviction: linkedlist.NewDll[entry](),
	}
}

// get the key value from cache
func (lru *LRUCache) Get(key string) (*entry, error) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// check if key is in the cache
	if element, found := lru.cache[key]; found {
		// check if it is expired, if yes delete it from cache
		if lru.isExpired(&element.Value) {
			lru.removeKey(element, key)
			return nil, errors.New("key is expired")
		}
		// move the key to front of linked list
		lru.eviction.MoveToFront(element)
		return &element.Value, nil
	}
	return nil, errors.New("key doesn't exist")
}

// get all the keys, value & timestamp cache
func (lru *LRUCache) GetAll() ([]entry, error) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	var entries []entry

	f := lru.eviction.Head

	// going through the linked list of cache and get the value
	for {
		if f == nil {
			break
		}
		// if it is expired remove it from cache & linked list
		if lru.isExpired(&f.Value) {
			lru.removeKey(f, f.Value.Key)
		} else {
			// if not expired return it
			entries = append(entries, f.Value)
		}
		f = f.Next

	}
	println(len(entries))
	return entries, nil

}

// set key and value in cache
func (lru *LRUCache) Set(key string, value string, expireTime time.Duration) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	expireTimestamp := time.Now().Add(expireTime)

	// if key exist update the value and timestamp
	if element, found := lru.cache[key]; found {
		lru.eviction.MoveToFront(element)
		element.Value.Value = value
		element.Value.Timestamp = expireTimestamp
	} else {
		// evicting the oldest cache in list if cache have more than capacity of the lru
		if len(lru.cache) >= lru.capacity {
			lru.evictOldest()
		}

		// move it in the front of the cache list
		lru.eviction.PushFront(&linkedlist.Element[entry]{Value: entry{Key: key, Value: value, Timestamp: expireTimestamp}})
		lru.cache[key] = lru.eviction.Head

	}
}

// delete the key from cache
func (lru *LRUCache) Delete(key string) error {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, found := lru.cache[key]; found {
		// check if it is expired, if yes delete it from cache
		if lru.isExpired(&element.Value) {
			lru.removeKey(element, key)
			return errors.New("key is expired")
		} else {
			lru.removeKey(element, key)
		}
	} else {
		return errors.New("key not found")

	}
	return nil
}

// evict the oldest cache
func (lru *LRUCache) evictOldest() {
	if lru.eviction == nil {
		return
	}
	element := lru.eviction.Tail
	delete(lru.cache, element.Value.Key)
	lru.eviction.Remove(element)
}

// check if cache is expired
func (lru *LRUCache) isExpired(e *entry) bool {
	return e.Timestamp.Before(time.Now())
}

// delete the key from lru
func (lru *LRUCache) removeKey(element *linkedlist.Element[entry], key string) {
	lru.eviction.Remove(element)
	delete(lru.cache, key)
	fmt.Println(key + " got deleted")
}
