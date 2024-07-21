package cache_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/sorushp/cache/pkg/cache"
)

func BenchmarkCache_Set(b *testing.B) {
	c := cache.NewCache[string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key"+strconv.Itoa(i), "value", 5*time.Second)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := cache.NewCache[string]()
	for i := 0; i < 1000; i++ {
		c.Set("key"+strconv.Itoa(i), "value", 5*time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key" + strconv.Itoa(i%1000))
	}
}

func BenchmarkCache_SetAndEvict(b *testing.B) {
	c := cache.NewCache[string]()
	for i := 0; i < cache.MaxCacheSize; i++ {
		key := "key" + strconv.Itoa(i)
		c.Set(key, "value", 5*time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(cache.MaxCacheSize+i)
		c.Set(key, "value", 5*time.Second)
	}
}
