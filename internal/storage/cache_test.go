package storage

import (
	"reflect"
	"sync"
	"testing"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache[string, int]()

	cache.Set("apple", 1)
	cache.Set("banana", 2)

	// Test existing keys
	if val, ok := cache.Get("apple"); !ok || val != 1 {
		t.Errorf("Expected apple to be 1, got %v (exists: %v)", val, ok)
	}

	// Test updating an existing key
	cache.Set("apple", 10)
	if val, ok := cache.Get("apple"); !ok || val != 10 {
		t.Errorf("Expected updated apple to be 10, got %v", val)
	}

	// Test non-existent key
	if _, ok := cache.Get("orange"); ok {
		t.Errorf("Expected orange to not exist")
	}
}

func TestCache_AllValuesOrder(t *testing.T) {
	cache := NewCache[string, string]()

	cache.Set("k1", "v1")
	cache.Set("k2", "v2")
	cache.Set("k3", "v3")
	cache.Set("k1", "v1_updated") // Update shouldn't change insertion order index

	expected := []string{"v1_updated", "v2", "v3"}
	actual := cache.AllValues()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected order %v, got %v", expected, actual)
	}
}

func TestCache_Del(t *testing.T) {
	cache := NewCache[int, string]()

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	cache.Del(2)

	// Verify key is gone from map lookup
	if _, ok := cache.Get(2); ok {
		t.Errorf("Expected key 2 to be deleted")
	}

	// Verify order slice removed the item and preserved remaining order
	expectedValues := []string{"one", "three"}
	if actualValues := cache.AllValues(); !reflect.DeepEqual(actualValues, expectedValues) {
		t.Errorf("Expected remaining values %v, got %v", expectedValues, actualValues)
	}

	// Verify length updated
	if cache.Len() != 2 {
		t.Errorf("Expected length 2, got %d", cache.Len())
	}
}

func TestCache_Concurrency(t *testing.T) {
	cache := NewCache[int, int]()
	var wg sync.WaitGroup
	workers := 100

	// Concurrent writes
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			cache.Set(val, val*10)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			_, _ = cache.Get(val)
			_ = cache.AllValues()
			_ = cache.Len()
		}(i)
	}

	wg.Wait()

	if cache.Len() != workers {
		t.Errorf("Expected final size to match worker count %d, got %d", workers, cache.Len())
	}
}
