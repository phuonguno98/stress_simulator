package utils

import (
	"testing"
	"time"
)

// TestInitSharedPayload tests the initialization of the shared payload
func TestInitSharedPayload(t *testing.T) {
	// Initialize with different sizes
	InitSharedPayload(1024) // 1KB
	if len(SharedPayload) != 1024 {
		t.Errorf("Expected SharedPayload to have length 1024, got %d", len(SharedPayload))
	}

	InitSharedPayload(2048) // 2KB
	if len(SharedPayload) != 2048 {
		t.Errorf("Expected SharedPayload to have length 2048, got %d", len(SharedPayload))
	}

	// Test with zero size
	InitSharedPayload(0)
	if len(SharedPayload) != 0 {
		t.Errorf("Expected SharedPayload to have length 0, got %d", len(SharedPayload))
	}
}

// TestGetPayload tests the payload retrieval function
func TestGetPayload(t *testing.T) {
	// Initialize the shared payload first
	InitSharedPayload(1000)

	// Test getting payloads of different sizes
	payload1 := GetPayload(100)
	if len(payload1) != 100 {
		t.Errorf("Expected payload of length 100, got %d", len(payload1))
	}

	payload2 := GetPayload(500)
	if len(payload2) != 500 {
		t.Errorf("Expected payload of length 500, got %d", len(payload2))
	}

	// Test getting a payload larger than the shared payload
	payload3 := GetPayload(2000)
	if len(payload3) != 1000 { // Should return the full shared payload
		t.Errorf("Expected payload of length 1000 (full shared payload), got %d", len(payload3))
	}

	// Test with zero size
	payload4 := GetPayload(0)
	if len(payload4) != 0 {
		t.Errorf("Expected payload of length 0, got %d", len(payload4))
	}

	// Test with negative size (should return empty slice)
	payload5 := GetPayload(-10)
	if len(payload5) != 0 {
		t.Errorf("Expected payload of length 0 for negative input, got %d", len(payload5))
	}
}

// TestGetPayloadWithoutInitialization tests GetPayload when shared payload is not initialized
func TestGetPayloadWithoutInitialization(t *testing.T) {
	// Save original payload
	originalPayload := SharedPayload

	// Reset shared payload to nil to simulate uninitialized state
	SharedPayload = nil

	// This should use the fallback mechanism
	payload := GetPayload(100)
	if len(payload) != 100 {
		t.Errorf("Expected payload of length 100 when SharedPayload is uninitialized, got %d", len(payload))
	}

	// Verify the shared payload remains nil (no reinitialization)
	if SharedPayload != nil {
		t.Errorf("Expected shared payload to remain uninitialized, but it was reinitialized")
	}

	// Restore original payload
	SharedPayload = originalPayload
}

// TestGetPayloadRandomness tests that the payload retrieval is random
func TestGetPayloadRandomness(t *testing.T) {
	InitSharedPayload(1000)

	// Get two payloads of the same size
	payload1 := GetPayload(100)
	payload2 := GetPayload(100)

	// They might be the same by chance, but with a large enough payload,
	// the probability of them being identical should be very low
	// Just make sure both have the correct length
	if len(payload1) != 100 || len(payload2) != 100 {
		t.Errorf("Expected both payloads to have length 100")
	}

	// Check that they're not identical
	if string(payload1) == string(payload2) {
		t.Errorf("Payloads are identical, expected different random payloads")
	}
}

// TestInitSharedPayloadRandomness tests that the initialized payload is random
func TestInitSharedPayloadRandomness(t *testing.T) {
	// Initialize payload twice with a delay to get different seeds
	InitSharedPayload(1000)
	payload1 := make([]byte, len(SharedPayload))
	copy(payload1, SharedPayload)

	// Wait a bit to ensure different seed
	time.Sleep(10 * time.Millisecond)
	InitSharedPayload(1000)
	payload2 := make([]byte, len(SharedPayload))
	copy(payload2, SharedPayload)

	// The payloads might be similar by chance, but we at least verify they have the right length
	if len(payload1) != 1000 || len(payload2) != 1000 {
		t.Errorf("Expected both payloads to have length 1000")
	}

	// Check that they're not identical
	if string(payload1) == string(payload2) {
		t.Errorf("Payloads are identical, expected different random payloads")
	}
}
