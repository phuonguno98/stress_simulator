// Package cpu provides CPU stress simulation capabilities.
//
//nolint:gosec
package cpu

import (
	"math/rand"
	"os"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
)

// RunStress starts CPU stress simulation with given parameters
func RunStress(avgUtilization, avgIoWait float64, paramChan <-chan types.ParameterUpdate, stopChan <-chan os.Signal) {
	// Convert os.Signal channel to struct{} channel
	doneChan := make(chan struct{})
	go func() {
		<-stopChan
		close(doneChan)
	}()

	runStressInternal(avgUtilization, avgIoWait, paramChan, doneChan)
}

// runStressInternal does the actual stress work with struct{} channel
func runStressInternal(avgUtilization, avgIoWait float64, paramChan <-chan types.ParameterUpdate, stopChan <-chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Current values with base utilization
	currentUtil := avgUtilization
	currentIoWait := avgIoWait

	// Track temporary overrides
	var tempOverrideEnd time.Time
	var tempMultiplier float64

	// Track random walk state
	var utilVariation float64
	var iowaitVariation float64

	for {
		select {
		case <-stopChan:
			return
		case update := <-paramChan:
			// Apply temporary override
			tempMultiplier = update.Value
			tempOverrideEnd = time.Now().Add(update.Duration)

			// Update current values with temporary multiplier
			currentUtil = avgUtilization * tempMultiplier
			if currentUtil > 100 {
				currentUtil = 100
			} else if currentUtil < 0 {
				currentUtil = 0
			}

			currentIoWait = avgIoWait * tempMultiplier
			if currentIoWait > 100 {
				currentIoWait = 100
			} else if currentIoWait < 0 {
				currentIoWait = 0
			}
		case <-ticker.C:
			// Check if temporary override has expired
			if time.Now().After(tempOverrideEnd) && !tempOverrideEnd.IsZero() {
				// Reset to base values
				currentUtil = avgUtilization
				currentIoWait = avgIoWait
				tempOverrideEnd = time.Time{} // Clear the override time
			}

			// Random walk for smooth variation
			// Max variation of ±20%, but step changes are small (e.g., ±2% per tick)
			utilStep := (rand.Float64() - 0.5) * 0.04
			utilVariation += utilStep
			// Cap variation to ±20%
			if utilVariation > 0.2 {
				utilVariation = 0.2
			} else if utilVariation < -0.2 {
				utilVariation = -0.2
			}

			// Calculate current CPU utilization with smooth variation
			finalUtil := currentUtil + (currentUtil * utilVariation)
			if finalUtil > 100 {
				finalUtil = 100
			} else if finalUtil < 0 {
				finalUtil = 0
			}

			// Random walk for I/O wait variation
			iowaitStep := (rand.Float64() - 0.5) * 0.04
			iowaitVariation += iowaitStep
			if iowaitVariation > 0.2 {
				iowaitVariation = 0.2
			} else if iowaitVariation < -0.2 {
				iowaitVariation = -0.2
			}

			// Calculate current I/O wait with smooth variation
			finalIoWait := currentIoWait + (currentIoWait * iowaitVariation)
			if finalIoWait > 100 {
				finalIoWait = 100
			} else if finalIoWait < 0 {
				finalIoWait = 0
			}

			// Perform CPU intensive work based on utilization
			cpuWork(finalUtil)

			// Simulate I/O wait by sleeping periodically
			ioWait(finalIoWait)
		}
	}
}

// cpuWork performs CPU-intensive operations to simulate CPU utilization
func cpuWork(utilizationPercent float64) {
	if utilizationPercent <= 0 {
		return
	}

	// Calculate work and sleep times based on utilization
	workDuration := time.Duration(float64(100*time.Millisecond) * utilizationPercent / 100.0)
	sleepDuration := 100*time.Millisecond - workDuration

	startTime := time.Now()
	for time.Since(startTime) < workDuration {
		// Perform some CPU intensive work
		_ = performIntensiveCalculation()
	}

	// If we have more CPUs, we might need to account for parallelism differently
	// For now, just sleep for the remaining time to achieve desired utilization
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
	}
}

// ioWait simulates I/O wait by blocking operations
func ioWait(ioWaitPercent float64) {
	if ioWaitPercent <= 0 {
		return
	}

	// For true I/O wait simulation, we need actual blocked disk I/O.
	// time.Sleep only puts the thread to SLEEPING state, which contributes to idle time, not I/O wait.
	if rand.Float64() < ioWaitPercent/100.0 {
		blockTime := time.Duration(10+rand.Intn(20)) * time.Millisecond

		// Use a deterministic filename to prevent leaking multiple files into OS temp dir on crash
		fileName := "cpu_iowait_temp.dat"
		tmpFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_SYNC, 0644)
		if err == nil {
			// Order is important: Close first, then Remove
			defer func() {
				_ = tmpFile.Close()
				_ = os.Remove(fileName)
			}()

			data := make([]byte, 4096)

			startTime := time.Now()
			for time.Since(startTime) < blockTime {
				_, _ = tmpFile.Seek(0, 0) // Overwrite same block to prevent file growth
				_, _ = tmpFile.Write(data)
				_ = tmpFile.Sync() // This blocks and causes true system I/O wait
			}
		}
	}
}

// performIntensiveCalculation performs a CPU-intensive calculation
func performIntensiveCalculation() int {
	result := 0
	for i := range 1000 {
		result += i * i
	}
	return result
}
