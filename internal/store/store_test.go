package store

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func setupTestDB(t *testing.T, db string) Store {
	t.Helper()

	store, err := NewStore(db)
	if err != nil {
		log.Printf("%v", err.Error())
		return nil
	}
	return store
}

func TestShortenAndLookup(t *testing.T) {
	store := setupTestDB(t, ":memory:")
	defer store.Close()

	longURL := "https://example.com"

	id, err := store.Shorten(longURL)
	if err != nil {
		t.Fatalf("failed to shorten URL: %v", err)
	}

	if id != 1 {
		t.Errorf("expected %v, got %v", 1, id)
	}

	retrievedURL, err := store.Lookup(id)
	if err != nil {
		t.Fatalf("failed to lookup URL: %v", err)
	}

	if retrievedURL != longURL {
		t.Errorf("expected %v, got %v", longURL, retrievedURL)
	}
}

func TestLookupNonexistentCode(t *testing.T) {
	service := setupTestDB(t, ":memory:")
	defer service.Close()

	_, err := service.Lookup(2)
	if err == nil || err.Error() != "short URL not found" {
		t.Errorf("expected 'short URL not found' error, got %v", err)
	}
}

func TestMixedConcurrentAccess(t *testing.T) {
	service := setupTestDB(t, "file:test.db?mode=rwc")

	defer os.Remove("test.db")
	defer service.Close()

	var wg sync.WaitGroup
	numOperations := 2000
	longURLBase := "http://example.com/page"

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			longURL := fmt.Sprintf("%s%d", longURLBase, i)
			shortCode := fmt.Sprintf("code%d", i)

			id, err := service.Shorten(longURL)
			if err != nil {
				t.Errorf("Failed to shorten URL: %v", err)
				return
			}

			time.Sleep(time.Millisecond * 10)

			retrievedURL, err := service.Lookup(id)
			if err != nil {
				t.Errorf("Failed to lookup URL for code %s: %v", shortCode, err)
				return
			}

			if retrievedURL != longURL {
				t.Errorf("Expected URL %s, but got %s for code %s", longURL, retrievedURL, shortCode)
			}
		}(i)
	}

	wg.Wait()
}
