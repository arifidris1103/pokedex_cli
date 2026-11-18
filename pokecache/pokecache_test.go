package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	c := NewCache(time.Minute)
	c.Add("https://example.com", []byte("data"))

	got, ok := c.Get("https://example.com")
	if !ok || string(got) != "data" {
		t.Fatalf("got %q, %v", got, ok)
	}
	if _, ok := c.Get("https://missing.com"); ok {
		t.Fatal("expected a miss for a key that was never added")
	}
}

func TestReapLoop(t *testing.T) {
	const interval = 10 * time.Millisecond
	c := NewCache(interval)
	c.Add("key", []byte("data"))

	if _, ok := c.Get("key"); !ok {
		t.Fatal("entry should exist right after Add")
	}
	time.Sleep(interval * 5)
	if _, ok := c.Get("key"); ok {
		t.Fatal("entry should have been reaped")
	}
}
