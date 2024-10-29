package url_converter

import (
	"fmt"
	"testing"
)

func TestEncodeDecodeBase62(t *testing.T) {
	InitBase62Array("shuffle-key")
	id := int64(123456)
	expectedShortCode := "pQn"
	shortCode := base62Encode(id)

	if shortCode != expectedShortCode {
		t.Fatalf("expected %v, got %v", expectedShortCode, shortCode)
	}

	decodedID := base62Decode(shortCode)

	if decodedID != id {
		t.Fatalf("expected %v, got %v", id, decodedID)
	}
}

func TestEncodeDecodeID(t *testing.T) {
	id := int64(1)
	xorSecretKey := int64(15489079) // Large prime number
	expectedShortCode := "oyAVB"

	// Encode the ID
	shortCode := EncodeID(id, xorSecretKey)
	if shortCode != expectedShortCode {
		t.Fatalf("expected %v, got %v", expectedShortCode, shortCode)
	}

	decodedID := DecodeShortCode(shortCode, xorSecretKey)

	if decodedID != id {
		t.Fatalf("expected %v, got %v", id, decodedID)
	}
	fmt.Printf("Decoded ID: %d\n", decodedID)
}
