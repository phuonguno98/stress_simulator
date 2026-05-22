// Package utils provides utility functions used across the stress simulator.
//
//nolint:gosec
package utils

import (
	"math/rand"
	"time"
)

// SharedPayload stores a shared random payload buffer used across the application
var SharedPayload []byte

// InitSharedPayload initializes the shared random payload buffer.
// Call this once at application startup.
func InitSharedPayload(size int) {
	// Use local rand generator so it doesn't block global rand
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	SharedPayload = make([]byte, size)
	for i := range size {
		SharedPayload[i] = byte(rng.Intn(256))
	}
}

// GetPayload returns a slice of random bytes of the requested size.
// It retrieves it from the SharedPayload by picking a random offset.
// This is extremely fast and allocates no new memory for the data itself
// (only a slice header).
func GetPayload(size int) []byte {
	if size <= 0 {
		return []byte{} // Return empty slice for non-positive sizes
	}

	if len(SharedPayload) == 0 {
		// Fallback if not initialized
		data := make([]byte, size)
		for i := range size {
			data[i] = byte(rand.Intn(256))
		}
		return data
	}

	if size >= len(SharedPayload) {
		return SharedPayload
	}

	maxOffset := len(SharedPayload) - size
	offset := rand.Intn(maxOffset + 1)
	return SharedPayload[offset : offset+size]
}
