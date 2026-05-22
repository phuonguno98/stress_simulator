//nolint:revive
package disk

import (
	"os"
	"testing"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
)

// TestRunStress tests the main disk stress function
func TestRunStress(t *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan os.Signal, 1)

	// Start disk stress in a goroutine
	go func() {
		RunStress(50.0, paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	stopChan <- os.Interrupt
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressInternal tests the internal disk stress function
func TestRunStressInternal(t *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Start disk stress in a goroutine
	go func() {
		runStressInternal(50.0, paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressWithParameterUpdates tests disk stress with parameter updates
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

	// Start disk stress in a goroutine
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

// TestPerformDiskOperations tests the disk operations function
func TestPerformDiskOperations(t *testing.T) {
	fileHandles := make(map[string]*os.File)

	// Test with 0% utilization (should still perform minimum operations)
	performDiskOperations(0.0, fileHandles, "test_temp")

	// Test with 50% utilization
	performDiskOperations(50.0, fileHandles, "test_temp")

	// Test with 100% utilization
	performDiskOperations(100.0, fileHandles, "test_temp")

	// Close all file handles
	closeAllFiles(fileHandles)
}

// TestEnforceDiskLimit tests the disk limit enforcement function
func TestEnforceDiskLimit(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := "test_disk_limit_temp"
	err := os.MkdirAll(tempDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	fileHandles := make(map[string]*os.File)

	// Test with a high limit (should not trigger cleanup)
	enforceDiskLimit(tempDir, fileHandles, 1000000000) // 1GB limit

	// Test with a low limit (should trigger cleanup)
	enforceDiskLimit(tempDir, fileHandles, 100) // 100 bytes limit
}

// TestCloseAllFiles tests the function that closes all file handles
func TestCloseAllFiles(t *testing.T) {
	// Create some dummy file handles
	fileHandles := make(map[string]*os.File)

	// In a real scenario, these would be actual file handles
	// For testing purposes, we'll use nil handles since we can't easily create real files
	fileHandles["test1"] = nil
	fileHandles["test2"] = nil

	// This should not panic
	closeAllFiles(fileHandles)

	// After closing, the map should still exist but handles are closed
	if len(fileHandles) != 2 {
		t.Errorf("Expected 2 entries in fileHandles map, got %d", len(fileHandles))
	}
}

// TestPerformReadOperation tests the read operation function
func TestPerformReadOperation(t *testing.T) {
	fileHandles := make(map[string]*os.File)

	// This should not panic even with empty map
	performReadOperation(fileHandles)

	// Add a nil file handle
	fileHandles["test"] = nil
	performReadOperation(fileHandles)
}

// TestPerformWriteOperation tests the write operation function
func TestPerformWriteOperation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := "test_write_temp"
	err := os.MkdirAll(tempDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	fileHandles := make(map[string]*os.File)

	// This should not panic
	performWriteOperation(fileHandles, tempDir)
}

// TestPerformRandomIO tests the random I/O operation function
func TestPerformRandomIO(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := "test_random_io_temp"
	err := os.MkdirAll(tempDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	fileHandles := make(map[string]*os.File)

	// This should not panic
	performRandomIO(fileHandles, tempDir)
}

// TestPerformFileOps tests the file operations function
func TestPerformFileOps(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := "test_file_ops_temp"
	err := os.MkdirAll(tempDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// This should not panic
	performFileOps(tempDir)
}
