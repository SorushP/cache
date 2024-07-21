// Package cache provides a concurrent LRU cache.
//
// To use this package, call NewCache.
package cache

import (
	"container/list"
	"hash/maphash"
	"sync"
	"time"
)

// MaxCacheSize defines the maximum cache size. It should satisfy
// 1 <= MaxCacheSize <= 10000, but for some test cases it should
// satisfy more (check them).
const MaxCacheSize = 5

// NewCache makes and returns a new instance of cache.
func NewCache[V any]() *lruCache[V] {
	return &lruCache[V]{
		seed: maphash.MakeSeed(),
	}
}

// lruCache is the LRU cache structure. It's not exported and should be used by
// NewCache function.
type lruCache[V any] struct {
	m    sync.Mutex
	size uint64                   // size holds the cache size. It should always satisfy size <= MaxCacheSize.
	seed maphash.Seed             // seed is used for seeding hash algorithm. It must be read-only after initialization.
	lru  list.List                // lru holds the LRU metadata as a doubly linked list.
	data [MaxCacheSize]*record[V] // data holds the cache records as an array. Each entry holds a simple linked list.
}

type record[V any] struct {
	key   string
	val   V
	below *record[V]
	t     *time.Timer // t is just uses to kill the TTL process when the key will be reset
	e     *list.Element
}

// Get gets the value corresponding to a requested key. It returns false if
// the key does not exist.
func (c *lruCache[V]) Get(key string) (v V, found bool) {
	i := c.hash(key)

	c.m.Lock()
	defer c.m.Unlock()

	if r, above, root := c.lookup(i, key); r != nil {
		if r != root {
			// Bring r to the top
			above.below = r.below
			r.below = root
			c.data[i] = r
		}
		c.lru.MoveToFront(r.e)
		return r.val, true
	}
	return
}

// Set sets a new (key, value, ttl) record in the cache. If key exists, value
// and ttl will be reset for the key. If the size of cache exceeded from
// MaxCacheSize, it evicts an entry based on LRU policy.
func (c *lruCache[V]) Set(key string, value V, ttl time.Duration) {
	i := c.hash(key)

	c.m.Lock()
	defer c.m.Unlock()

	switch r, above, root := c.lookup(i, key); {
	case r != nil: // reset
		r.t.Stop()
		r.val = value
		r.t = time.AfterFunc(ttl, func() {
			c.deleteWithLock(i, key)
		})

		// Bring r to the top
		if r != root {
			above.below = r.below
			r.below = root
			c.data[i] = r
		}

		c.lru.MoveToFront(r.e)
	case c.size == MaxCacheSize: // evict and set
		er := c.lru.Back().Value.(*record[V])
		if er == root {
			root = root.below
		}
		c.delete(c.hash(er.key), er.key)
		fallthrough
	default: // set
		r = &record[V]{
			key:   key,
			val:   value,
			below: root,
			t: time.AfterFunc(ttl, func() {
				c.deleteWithLock(i, key)
			}),
		}
		r.e = c.lru.PushFront(r)
		c.data[i] = r // Bring r to the top
		c.size++
	}
}

// lookup finds the corresponding records for a key within i'th entry in c.data
// array. It returns the record itself (if any), the one above that (if any),
// and the root record (if any). It must be executed within a transaction
func (c *lruCache[V]) lookup(i uint64, key string) (rec, above, root *record[V]) {
	c.mustBeLocked()

	root = c.data[i]
	for rec = root; rec != nil && rec.key != key; {
		rec, above = rec.below, rec
	}

	return
}

func (c *lruCache[V]) deleteWithLock(idx uint64, key string) {
	c.m.Lock()
	defer c.m.Unlock()
	c.delete(idx, key)
}

// delete deletes a key from cache. It receives both key and idx and they must
// be consistent. This method must be executed within a cache transaction.
func (c *lruCache[V]) delete(idx uint64, key string) {
	c.mustBeLocked()

	r, above, root := c.lookup(idx, key)
	if r == nil {
		return
	}

	r.t.Stop() // is safe to be executed during either TTL or normal deletion.

	if r == root {
		c.data[idx] = r.below
	} else {
		above.below = r.below
	}

	c.lru.Remove(r.e)
	c.size--
}

// hash implements hash function by using hash/maphash strandard library.
func (c *lruCache[V]) hash(s string) uint64 {
	return maphash.String(c.seed, s) % MaxCacheSize
}

// mustBeLocked ensures that the caller are running withing a cache transaction
// (using c.m capability), without acquiring the lock by itself. It's good for
// safe development.
func (c *lruCache[V]) mustBeLocked() {
	if c.m.TryLock() {
		c.m.Unlock()
		panic("the code must be executed within a cache transaction")
	}
}
