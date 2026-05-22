//nolint:revive
package cpu

import (
	"os"
	"testing"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
)

// TestRunStress tests the main CPU stress function with different parameters
func TestRunStress(_ *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan os.Signal, 1)

	// Start CPU stress in a goroutine
	go func() {
		RunStress(50.0, 20.0, paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	stopChan <- os.Interrupt
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressInternal tests the internal CPU stress function
func TestRunStressInternal(_ *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Start CPU stress in a goroutine
	go func() {
		runStressInternal(50.0, 20.0, paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressWithParameterUpdates tests CPU stress with parameter updates
func TestRunStressWithParameterUpdates(_ *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Send a parameter update
	update := types.ParameterUpdate{
		Value:     2.0, // Double the utilization
		Duration:  200 * time.Millisecond,
		Timestamp: time.Now(),
	}

	// Start CPU stress in a goroutine
	go func() {
		runStressInternal(50.0, 20.0, paramChan, stopChan)
	}()

	// Send parameter update
	time.Sleep(50 * time.Millisecond)
	paramChan <- update

	// Stop after a bit more time
	time.Sleep(300 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestCpuWork tests the CPU work function with different utilization levels
func TestCpuWork(_ *testing.T) {
	// Test with 0% utilization (should return immediately)
	cpuWork(0.0)

	// Test with 50% utilization
	cpuWork(50.0)

	// Test with 100% utilization
	cpuWork(100.0)

	// Test with negative utilization (should return immediately)
	cpuWork(-10.0)
}

// TestIoWait tests the I/O wait simulation function
func TestIoWait(t *testing.T) {
	// Test with 0% I/O wait (should return immediately)
	ioWait(0.0)

	// Test with 50% I/O wait
	ioWait(50.0)

	// Test with 100% I/O wait
	ioWait(100.0)

	// Test with negative I/O wait (should return immediately)
	ioWait(-10.0)
}

// TestPerformIntensiveCalculation tests the CPU-intensive calculation function
func TestPerformIntensiveCalculation(t *testing.T) {
	result := performIntensiveCalculation()

	// The result should be a specific calculated value
	expected := 0
	for i := range 1000 {
		expected += i * i
	}

	if result != expected {
		t.Errorf("performIntensiveCalculation() = %d; want %d", result, expected)
	}
}
