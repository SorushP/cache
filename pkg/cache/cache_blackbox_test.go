package cache_test

import (
	"math/rand/v2"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sorushp/cache/pkg/cache"
)

func TestCache_SetAndGet(t *testing.T) {
	c := cache.NewCache[string]()

	c.Set("key1", "value1", 5*time.Second)
	c.Set("key2", "value2", 5*time.Second)

	if val, ok := c.Get("key1"); !ok || val != "value1" {
		t.Errorf("expected 'value1', got '%v'", val)
	}

	if val, ok := c.Get("key2"); !ok || val != "value2" {
		t.Errorf("expected 'value2', got '%v'", val)
	}

	if _, ok := c.Get("key3"); ok {
		t.Error("expected 'key3' to be absent")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := cache.NewCache[string]()

	c.Set("key1", "value1", 2*time.Second)
	time.Sleep(3 * time.Second)

	if _, ok := c.Get("key1"); ok {
		t.Error("expected 'key1' to have expired")
	}
}

// For this test, cache.MaxCacheSize >= 2 is mandatory
func TestCache_LRU(t *testing.T) {
	c := cache.NewCache[string]()

	for i := 1; i <= cache.MaxCacheSize+1; i++ { // The last iteration (i == maxCacheSize+1) should evict "key1"
		n := strconv.Itoa(i)
		c.Set("key"+n, "value"+n, 5*time.Second)
	}

	if _, ok := c.Get("key1"); ok {
		t.Error("expected 'key1' to be evicted")
	}

	// Access "key2" to make it recently used
	if val, ok := c.Get("key2"); !ok || val != "value2" {
		t.Errorf("expected 'value2', got '%v'", val)
	}

	n := strconv.Itoa(cache.MaxCacheSize + 2)
	c.Set("key"+n, "value"+n, 5*time.Second) // This should evict "key3"

	if _, ok := c.Get("key3"); ok {
		t.Error("expected 'key3' to be evicted")
	}
}

func TestCache_ConcurrentSetAndGet(t *testing.T) {
	c := cache.NewCache[string]()

	var wg sync.WaitGroup
	numGoroutines := cache.MaxCacheSize

	// Concurrently setting values
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func() {
			defer wg.Done()
			n := strconv.Itoa(i)
			c.Set("key"+n, "value"+n, 100*time.Millisecond)
		}()
	}

	wg.Wait()

	// Concurrently getting values
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func() {
			defer wg.Done()
			n := strconv.Itoa(i)
			if val, ok := c.Get("key" + n); !ok || val != "value"+n {
				t.Errorf("expected 'value%s', got '%s'", n, val)
			}
		}()
	}

	wg.Wait()
}

func TestCache_ConcurrentSetAndExpire(t *testing.T) {
	c := cache.NewCache[string]()

	var wg sync.WaitGroup
	numGoroutines := cache.MaxCacheSize
	numIterations := 5

	tOffset := 100 * time.Millisecond

	// Concurrently setting values with varying TTL
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func() {
			defer wg.Done()
			for j := range numIterations {
				n := strconv.Itoa(i*numIterations + j)
				ttl := time.Duration(rand.IntN(50)) + tOffset
				c.Set("key"+n, "value"+n, ttl)
			}
		}()
	}

	wg.Wait()

	// Wait for all TTLs to expire
	time.Sleep(2 * tOffset)

	// Concurrently getting values to check expiration
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func() {
			defer wg.Done()
			for j := range numIterations {
				key := "key" + strconv.Itoa(i*numIterations+j)
				if _, ok := c.Get(key); ok {
					t.Errorf("expected 'key%d' to be expired", i*numIterations+j)
				}
			}
		}()
	}

	wg.Wait()
}
