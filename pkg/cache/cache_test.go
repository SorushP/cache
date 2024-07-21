package cache

import (
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	c := NewCache[string]()

	c.Set("key1", "value1", 5*time.Second)
	c.Set("key2", "value2", 5*time.Second)

	c.m.Lock()
	defer c.m.Unlock()

	if c.size != 2 {
		t.Errorf("failed for cache size, expected 2, got %v", c.size)
	}

	// Looking up existing key
	r, _, _ := c.lookup(c.hash("key1"), "key1")
	if r == nil || r.val != "value1" {
		t.Errorf("lookup failed for 'key1', expected 'value1', got '%v'", r)
	}

	// Looking up non-existing key
	r, _, _ = c.lookup(c.hash("weird"), "weird")
	if r != nil {
		t.Errorf("lookup should have failed for 'weird', got '%v'", r)
	}
}

func TestDelete(t *testing.T) {
	c := NewCache[string]()

	c.Set("key1", "value1", 5*time.Second)
	c.Set("key2", "value2", 5*time.Second)
	c.Set("key3", "value3", 5*time.Second)

	// Deleting an existing key
	c.m.Lock()
	c.delete(c.hash("key2"), "key2")
	if c.size != 2 {
		t.Errorf("failed for cache size, expected 2, got %v", c.size)
	}
	c.m.Unlock()

	if _, ok := c.Get("key2"); ok {
		t.Errorf("expected 'key2' to be deleted")
	}

	// Verifying internal state after deletion
	c.m.Lock()
	r, _, _ := c.lookup(c.hash("key2"), "key2")
	c.m.Unlock()
	if r != nil {
		t.Errorf("lookup should not find 'key2', got '%v'", r)
	}
}
