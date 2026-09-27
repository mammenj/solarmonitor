package storage

import (
	"slices"
	"testing"
)

// Helper function to extract all keys in order from the c.All() iterator
func collectOrderedKeys[K comparable, V any](c *Cache[K, V]) []K {
	var keys []K
	for k := range c.All() {
		keys = append(keys, k)
	}
	return keys
}

// TestInsertionOrder verifies that elements are iterated over in the exact order they were first inserted.
func TestInsertionOrder(t *testing.T) {
	c := NewCache[string, int]()

	inputs := []struct {
		key   string
		value int
	}{
		{"apple", 1},
		{"banana", 2},
		{"cherry", 3},
	}

	for _, input := range inputs {
		c.Set(input.key, input.value)
	}

	expectedOrder := []string{"apple", "banana", "cherry"}
	actualOrder := collectOrderedKeys(c)

	if !slices.Equal(actualOrder, expectedOrder) {
		t.Errorf("expected order %v, got %v", expectedOrder, actualOrder)
	}
}

// TestUpdateDoesNotChangeOrder verifies that updating the value of an existing key
// does not alter its position in the insertion sequence.
func TestUpdateDoesNotChangeOrder(t *testing.T) {
	c := NewCache[string, int]()

	c.Set("apple", 1)
	c.Set("banana", 2)
	c.Set("cherry", 3)

	// Update an existing key in the middle
	c.Set("banana", 20)

	// The order must remain unchanged despite the value change
	expectedOrder := []string{"apple", "banana", "cherry"}
	actualOrder := collectOrderedKeys(c)

	if !slices.Equal(actualOrder, expectedOrder) {
		t.Errorf("expected order %v after update, got %v", expectedOrder, actualOrder)
	}

	// Double check the value updated correctly
	val, ok := c.Get("banana")
	if !ok || val != 20 {
		t.Errorf("expected updated value 20 for key 'banana', got %v (ok: %t)", val, ok)
	}
}

// TestIteratorEarlyBreak verifies that the c.All() iterator cleanly exits when the loop breaks early.
func TestIteratorEarlyBreak(t *testing.T) {
	c := NewCache[string, int]()
	c.Set("A", 1)
	c.Set("B", 2)
	c.Set("C", 3)

	count := 0
	for range c.All() {
		count++
		if count == 2 {
			break // Break early to trigger standard yield evaluation path return
		}
	}

	if count != 2 {
		t.Errorf("expected iterator to process exactly 2 items before breaking, processed %d", count)
	}
}
