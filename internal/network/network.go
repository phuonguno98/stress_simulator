// Package network provides network stress simulation capabilities.
//
//nolint:gosec
package network

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
	"github.com/phuonguno98/stress_simulator/internal/utils"
)

// RunStress starts network stress simulation with given parameters
func RunStress(avgRateMB float64, targetIP string, paramChan <-chan types.ParameterUpdate, stopChan <-chan os.Signal) {
	// Convert os.Signal channel to struct{} channel
	doneChan := make(chan struct{})
	go func() {
		<-stopChan
		close(doneChan)
	}()

	runStressInternal(avgRateMB, targetIP, paramChan, doneChan)
}

// runStressInternal does the actual stress work with struct{} channel
func runStressInternal(avgRateMB float64, targetIP string, paramChan <-chan types.ParameterUpdate, stopChan <-chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Start a listener to accept incoming connections and simulate IN traffic
	go startListener(stopChan)

	// Current rate value
	currentRate := avgRateMB
	// Track temporary overrides
	var tempOverrideEnd time.Time
	var tempMultiplier float64

	// Track random walk state
	var rateVariation float64

	for {
		select {
		case <-stopChan:
			return
		case update := <-paramChan:
			// Apply temporary override
			tempMultiplier = update.Value
			tempOverrideEnd = time.Now().Add(update.Duration)

			// Update current value with temporary multiplier
			currentRate = avgRateMB * tempMultiplier
			if currentRate < 0 {
				currentRate = 0
			}
		case <-ticker.C:
			// Check if temporary override has expired
			if time.Now().After(tempOverrideEnd) && !tempOverrideEnd.IsZero() {
				// Reset to base value
				currentRate = avgRateMB
				tempOverrideEnd = time.Time{} // Clear the override time
			}

			// Random walk for smooth network variation
			rateStep := (rand.Float64() - 0.5) * 0.04
			rateVariation += rateStep
			if rateVariation > 0.2 {
				rateVariation = 0.2
			} else if rateVariation < -0.2 {
				rateVariation = -0.2
			}

			// Calculate current network rate with smooth variation
			finalRate := currentRate + (currentRate * rateVariation)
			if finalRate < 0 {
				finalRate = 0
			}

			// Perform network operations based on rate
			performNetworkOperations(finalRate, targetIP)
		}
	}
}

// startListener starts a simple TCP server to consume IN traffic
func startListener(stopChan <-chan struct{}) {
	port := "8080" // default port

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Failed to start network listener on port %s: %v\n", port, err)
		return
	}

	// Close listener when stopping
	go func() {
		<-stopChan
		_ = l.Close()
	}()

	fmt.Printf("Started network listener on port %s for IN traffic\n", port)

	for {
		conn, err := l.Accept()
		if err != nil {
			// Expected when l.Close() is called
			return
		}

		// Handle connection in a goroutine to keep accepting
		go func(c net.Conn) {
			defer func() { _ = c.Close() }()
			// Read incoming data and discard it to simulate IN traffic processing
			_, _ = io.Copy(io.Discard, c)
		}(conn)
	}
}

// performNetworkOperations performs various network operations to simulate traffic
func performNetworkOperations(rateMB float64, targetIP string) {
	// Support comma-separated list of target IPs in the targetIP flag
	allTargets := make([]string, 0)
	for target := range strings.SplitSeq(targetIP, ",") {
		trimmed := strings.TrimSpace(target)
		if trimmed != "" {
			allTargets = append(allTargets, trimmed)
		}
	}
	if len(allTargets) == 0 {
		return // Should not happen, main checks this
	}

	// Calculate target bytes to send in this interval (100ms interval)
	targetBytes := int(rateMB * 1024 * 1024 / 10)
	if targetBytes <= 0 {
		return
	}

	chunkSize := 1024 * 1024 // 1MB chunks maximum to spread connections
	bytesSent := 0

	// Perform network operations until we reach the target byte rate
	for bytesSent < targetBytes {
		currentChunk := min(targetBytes-bytesSent, chunkSize)

		// Select a random target from the available targets
		selectedTarget := allTargets[rand.Intn(len(allTargets))]

		// Randomly select network operation type
		opType := rand.Intn(3) // 0: HTTP request, 1: TCP connection, 2: UDP packet

		switch opType {
		case 0: // HTTP request to target
			performHTTPRequest(selectedTarget, currentChunk)
		case 1: // TCP connection to target
			performTCPConnection(selectedTarget, currentChunk)
		case 2: // UDP packet to target
			performUDPPacket(selectedTarget, currentChunk)
		}

		bytesSent += currentChunk
	}
}

// performHTTPRequest performs HTTP requests to the target
func performHTTPRequest(targetIP string, size int) {
	url := fmt.Sprintf("http://%s/", targetIP)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create a request with payload from shared buffer to simulate real traffic
	data := utils.GetPayload(size)
	payload := bytes.NewReader(data)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "StressSimulator/1.0")

	resp, err := client.Do(req)
	if err == nil {
		// Read response body to consume it
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
}

// performTCPConnection makes TCP connections to the target
func performTCPConnection(targetIP string, size int) {
	// Try to connect to common ports on the target
	ports := []string{"22", "80", "443", "8080", "3306", "6379"} // default ports

	port := ports[rand.Intn(len(ports))]

	address := net.JoinHostPort(targetIP, port)

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		// Connection failed, which is expected for stress testing
		return
	}
	defer func() { _ = conn.Close() }()

	// Send data to simulate real traffic
	data := utils.GetPayload(size)
	_, _ = conn.Write(data)

	// Read response if any
	buffer := make([]byte, 128)
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _ = conn.Read(buffer)
}

// performUDPPacket sends UDP packets to the target
func performUDPPacket(targetIP string, size int) {
	// Try to send UDP to common ports
	ports := []string{"53", "67", "68", "123", "161", "162", "500", "4500"} // default UDP ports

	port := ports[rand.Intn(len(ports))]

	address := net.JoinHostPort(targetIP, port)

	conn, err := net.DialTimeout("udp", address, 2*time.Second)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	// Send UDP packet (chunking if needed since UDP has max size)
	maxUDPSize := 65507
	bytesSent := 0
	for bytesSent < size {
		chunkSize := min(size-bytesSent, maxUDPSize)
		data := utils.GetPayload(chunkSize)
		_, err = conn.Write(data)
		if err != nil {
			break
		}
		bytesSent += chunkSize
	}
}
