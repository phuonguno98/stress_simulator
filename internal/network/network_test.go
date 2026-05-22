//nolint:revive
package network

import (
	"os"
	"testing"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
)

// TestRunStress tests the main network stress function
func TestRunStress(_ *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan os.Signal, 1)

	// Start network stress in a goroutine
	go func() {
		RunStress(10.0, "127.0.0.1", paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	stopChan <- os.Interrupt
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressInternal tests the internal network stress function
func TestRunStressInternal(_ *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Start network stress in a goroutine
	go func() {
		runStressInternal(10.0, "127.0.0.1", paramChan, stopChan)
	}()

	// Send a stop signal after a short time
	time.Sleep(100 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestRunStressWithParameterUpdates tests network stress with parameter updates
func TestRunStressWithParameterUpdates(_ *testing.T) {
	// Create channels for testing
	paramChan := make(chan types.ParameterUpdate, 10)
	stopChan := make(chan struct{}, 1)

	// Send a parameter update
	update := types.ParameterUpdate{
		Value:     2.0, // Double the rate
		Duration:  500 * time.Millisecond,
		Timestamp: time.Now(),
	}

	// Start network stress in a goroutine
	go func() {
		runStressInternal(10.0, "127.0.0.1", paramChan, stopChan)
	}()

	// Send parameter update
	time.Sleep(50 * time.Millisecond)
	paramChan <- update

	// Stop after a bit more time
	time.Sleep(600 * time.Millisecond)
	close(stopChan)
	time.Sleep(50 * time.Millisecond) // Allow time for graceful shutdown
}

// TestPerformNetworkOperations tests the network operations function
func TestPerformNetworkOperations(t *testing.T) {
	// Test with different rate values
	performNetworkOperations(0.0, "127.0.0.1")
	performNetworkOperations(5.0, "127.0.0.1")
	performNetworkOperations(10.0, "127.0.0.1")

	// Test with multiple targets
	performNetworkOperations(5.0, "127.0.0.1,192.168.1.1,10.0.0.1")

	// Test with empty target (should handle gracefully)
	performNetworkOperations(5.0, "")
}

// TestPerformHTTPRequest tests the HTTP request function
func TestPerformHTTPRequest(t *testing.T) {
	// This test will attempt to make HTTP requests to localhost
	// It should not panic even if the server is not available
	performHTTPRequest("127.0.0.1:9999", 1024) // Use a port that's likely closed
}

// TestPerformTCPConnection tests the TCP connection function
func TestPerformTCPConnection(t *testing.T) {
	// This test will attempt to make TCP connections to localhost
	// It should not panic even if the connection fails
	performTCPConnection("127.0.0.1", 1024)
}

// TestPerformUDPPacket tests the UDP packet function
func TestPerformUDPPacket(t *testing.T) {
	// This test will attempt to send UDP packets to localhost
	// It should not panic even if the connection fails
	performUDPPacket("127.0.0.1", 1024)
}

// TestStartListener tests the listener function
func TestStartListener(t *testing.T) {
	// Create a stop channel
	stopChan := make(chan struct{})

	// Start the listener in a goroutine
	go startListener(stopChan)

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop the listener
	close(stopChan)

	// Give it time to shut down
	time.Sleep(50 * time.Millisecond)
}
