//nolint:revive
package memory

import (
	"os"
	"testing"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
)

// TestRunStress tests the main memory stress function
func TestRunStress(t *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan os.Signal, 1)

	// Start memory stress in a goroutine
	go func() {
		RunStress(50.0, paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	stopChan <- os.Interrupt
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressInternal tests the internal memory stress function
func TestRunStressInternal(t *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Start memory stress in a goroutine
	go func() {
		runStressInternal(50.0, paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressWithParameterUpdates tests memory stress with parameter updates
func TestRunStressWithParameterUpdates(t *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Send a parameter update
	update := types.ParameterUpdate{
		Value:     2.0, // Double the utilization
		Duration:  500 * time.Millisecond,
		Timestamp: time.Now(),
	}

	// Start memory stress in a goroutine
	go func() {
		runStressInternal(50.0, paramChan, stopChan)
	}()

	// Send parameter update
	time.Sleep(50 * time.Millisecond)
	paramChan <- update

	// Stop after a bit more time
	time.Sleep(600 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestGetTotalSystemMemory tests the system memory detection function
func TestGetTotalSystemMemory(t *testing.T) {
	totalMemory := getTotalSystemMemory()

	// The total memory should be positive
	if totalMemory <= 0 {
		t.Errorf("getTotalSystemMemory() returned non-positive value: %d", totalMemory)
	}

	// On most systems, memory should be at least 1GB (1073741824 bytes)
	if totalMemory < 1073741824 {
		t.Logf("Warning: detected memory is less than 1GB: %d bytes", totalMemory)
	}
}

// TestStressAdjustMemory tests the memory adjustment function
func TestStressAdjustMemory(t *testing.T) {
	ms := &Stress{}

	// Test allocating memory
	ms.adjustMemory(1024 * 1024) // 1MB
	if len(ms.allocatedMemory) == 0 {
		t.Error("Expected memory to be allocated")
	}

	// Test freeing memory
	ms.adjustMemory(512 * 1024) // 0.5MB, should free half
	currentBytes := 0
	for _, mem := range ms.allocatedMemory {
		currentBytes += cap(mem)
	}
	if currentBytes > 512*1024 {
		t.Errorf("Expected memory to be reduced to ~0.5MB, got ~%d bytes", currentBytes)
	}

	// Cleanup
	ms.cleanup()
	if ms.allocatedMemory != nil {
		t.Error("Expected memory to be cleaned up")
	}
}

// TestStressCompactMemory tests the memory compaction function
func TestStressCompactMemory(t *testing.T) {
	ms := &Stress{}

	// Add some memory allocations
	ms.allocatedMemory = make([][]byte, 3)
	ms.allocatedMemory[0] = make([]byte, 1024)
	ms.allocatedMemory[1] = make([]byte, 0) // Empty allocation
	ms.allocatedMemory[2] = make([]byte, 2048)

	// Compact memory
	ms.compactMemory()

	// Count non-empty allocations
	count := 0
	for _, mem := range ms.allocatedMemory {
		if cap(mem) > 0 {
			count++
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 non-empty allocations after compaction, got %d", count)
	}
}

// TestStressCleanup tests the memory cleanup function
func TestStressCleanup(t *testing.T) {
	ms := &Stress{}

	// Add some memory allocations
	ms.allocatedMemory = make([][]byte, 2)
	ms.allocatedMemory[0] = make([]byte, 1024)
	ms.allocatedMemory[1] = make([]byte, 2048)

	// Call cleanup
	ms.cleanup()

	// Verify cleanup happened
	if ms.allocatedMemory != nil {
		t.Error("Expected allocatedMemory to be nil after cleanup")
	}
}
