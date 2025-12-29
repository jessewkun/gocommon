package localcache

import (
	"testing"
	"time"
)

type typedUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Tags []int  `json:"tags,omitempty"`
}

func TestTypedCache_SetGetStruct(t *testing.T) {
	cache, err := NewDefaultTypedBigCache[typedUser]()
	if err != nil {
		t.Fatalf("failed to create typed cache: %v", err)
	}
	defer cache.Close()

	user := typedUser{
		ID:   1,
		Name: "Alice",
		Tags: []int{1, 2, 3},
	}

	if err := cache.Set("user:1", user); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	got, ok := cache.Get("user:1")
	if !ok {
		t.Fatalf("expected value to exist")
	}

	if got.ID != user.ID || got.Name != user.Name {
		t.Fatalf("expected %v, got %v", user, got)
	}

	if len(got.Tags) != len(user.Tags) {
		t.Fatalf("expected tags length %d, got %d", len(user.Tags), len(got.Tags))
	}
}

func TestTypedCache_SetWithTTL(t *testing.T) {
	cache, err := NewDefaultTypedBigCache[typedUser]()
	if err != nil {
		t.Fatalf("failed to create typed cache: %v", err)
	}
	defer cache.Close()

	user := typedUser{ID: 2, Name: "Bob"}

	if err := cache.SetWithTTL("user:2", user, 50*time.Millisecond); err != nil {
		t.Fatalf("failed to set cache with ttl: %v", err)
	}

	if _, ok := cache.Get("user:2"); !ok {
		t.Fatalf("expected value to exist")
	}

	time.Sleep(60 * time.Millisecond)

	if _, ok := cache.Get("user:2"); ok {
		t.Fatalf("expected value to be expired")
	}
}
