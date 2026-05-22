// Package memory provides memory stress simulation capabilities.
//
//nolint:gosec
package memory

import (
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
	"github.com/phuonguno98/stress_simulator/internal/utils"
)

// Stress holds the state for memory stress operations
type Stress struct {
	allocatedMemory [][]byte
}

// RunStress starts memory stress simulation with given parameters
func RunStress(avgUtilization float64, paramChan <-chan types.ParameterUpdate, stopChan <-chan os.Signal) {
	// Convert os.Signal channel to struct{} channel
	doneChan := make(chan struct{})
	go func() {
		<-stopChan
		close(doneChan)
	}()

	runStressInternal(avgUtilization, paramChan, doneChan)
}

// runStressInternal does the actual stress work with struct{} channel
func runStressInternal(avgUtilization float64, paramChan <-chan types.ParameterUpdate, stopChan <-chan struct{}) {
	ms := &Stress{}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	totalMemory := getTotalSystemMemory()

	// Current utilization value
	currentUtil := avgUtilization
	// Track temporary overrides
	var tempOverrideEnd time.Time
	var tempMultiplier float64

	// Track random walk state
	var utilVariation float64

	for {
		select {
		case <-stopChan:
			// Clean up allocated memory
			ms.cleanup()
			return
		case update := <-paramChan:
			// Apply temporary override
			tempMultiplier = update.Value
			tempOverrideEnd = time.Now().Add(update.Duration)

			// Update current value with temporary multiplier
			currentUtil = avgUtilization * tempMultiplier
			if currentUtil > 100 {
				currentUtil = 100
			} else if currentUtil < 0 {
				currentUtil = 0
			}
		case <-ticker.C:
			// Check if temporary override has expired
			if time.Now().After(tempOverrideEnd) && !tempOverrideEnd.IsZero() {
				// Reset to base value
				currentUtil = avgUtilization
				tempOverrideEnd = time.Time{} // Clear the override time
			}

			// Random walk for smooth memory variation
			utilStep := (rand.Float64() - 0.5) * 0.04
			utilVariation += utilStep
			if utilVariation > 0.2 {
				utilVariation = 0.2
			} else if utilVariation < -0.2 {
				utilVariation = -0.2
			}

			// Calculate current memory utilization with smooth variation
			finalUtil := currentUtil + (currentUtil * utilVariation)
			if finalUtil > 100 {
				finalUtil = 100
			} else if finalUtil < 0 {
				finalUtil = 0
			}

			// Allocate memory based on current utilization
			targetBytes := int(float64(totalMemory) * finalUtil / 100.0)

			// Adjust memory allocation
			ms.adjustMemory(targetBytes)

			// Force garbage collection periodically to simulate real memory pressure
			if rand.Float64() < 0.1 { // 10% chance every second
				runtime.GC()
			}
		}
	}
}

// getTotalSystemMemory attempts to detect the total system memory
func getTotalSystemMemory() int {
	// Attempt to read from /proc/meminfo (Linux)
	data, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		lines := strings.SplitSeq(string(data), "\n")
		for line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if memKB, err := strconv.Atoi(parts[1]); err == nil {
						return memKB * 1024 // Convert KB to bytes
					}
				}
			}
		}
	}

	// For now, we'll use a reasonable estimate (8GB) fallback
	return 8 * 1024 * 1024 * 1024 // 8GB in bytes
}

// adjustMemory adjusts the amount of allocated memory
func (ms *Stress) adjustMemory(targetBytes int) {
	// Current allocated memory
	currentBytes := 0
	for _, mem := range ms.allocatedMemory {
		currentBytes += cap(mem)
	}

	// If we need more memory
	if targetBytes > currentBytes {
		bytesToAllocate := targetBytes - currentBytes
		chunkSize := 1024 * 1024 // 1MB chunks
		numChunks := bytesToAllocate / chunkSize
		if bytesToAllocate%chunkSize > 0 {
			numChunks++
		}

		// Allocate memory in chunks
		for i := 0; i < numChunks && len(ms.allocatedMemory) < 10000; i++ {
			size := chunkSize
			if i == numChunks-1 && bytesToAllocate%chunkSize > 0 {
				size = bytesToAllocate % chunkSize
			}

			// Create memory chunk
			chunk := make([]byte, size)

			// Fill a small portion with random data from payload to prevent over-optimization
			// Not filling the entire chunk to save CPU
			fillSize := min(100, size)
			if fillSize > 0 {
				copy(chunk[:fillSize], utils.GetPayload(fillSize))
			}

			ms.allocatedMemory = append(ms.allocatedMemory, chunk)
		}
	} else if targetBytes < currentBytes {
		// Need to free memory
		bytesToFree := currentBytes - targetBytes
		for bytesToFree > 0 && len(ms.allocatedMemory) > 0 {
			lastIdx := len(ms.allocatedMemory) - 1
			lastChunk := ms.allocatedMemory[lastIdx]
			chunkSize := cap(lastChunk)

			if chunkSize <= bytesToFree {
				// Free entire chunk
				ms.allocatedMemory = ms.allocatedMemory[:lastIdx]
				bytesToFree -= chunkSize
			} else {
				// Partially shrink the last chunk
				newSize := chunkSize - bytesToFree
				newChunk := make([]byte, newSize)
				copy(newChunk, lastChunk[:newSize])

				ms.allocatedMemory[lastIdx] = newChunk
				bytesToFree = 0
			}
		}
	}

	// Keep memory allocations in check
	if len(ms.allocatedMemory) > 10000 {
		// Periodically compact memory
		ms.compactMemory()
	}
}

// compactMemory compacts memory allocations
func (ms *Stress) compactMemory() {
	// This is a simplified compaction - in practice you might want more sophisticated logic
	newAllocations := make([][]byte, 0, len(ms.allocatedMemory))

	for _, mem := range ms.allocatedMemory {
		if cap(mem) > 0 {
			newAllocations = append(newAllocations, mem)
		}
	}

	ms.allocatedMemory = newAllocations
}

// cleanup frees all allocated memory
func (ms *Stress) cleanup() {
	ms.allocatedMemory = nil
	runtime.GC() // Force garbage collection
}
