// Package main implements a comprehensive system stress simulator that generates
// realistic load patterns with seasonal variations and unpredictable anomalies.
//
//nolint:gosec
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/cpu"
	"github.com/phuonguno98/stress_simulator/internal/disk"
	"github.com/phuonguno98/stress_simulator/internal/memory"
	"github.com/phuonguno98/stress_simulator/internal/network"
	"github.com/phuonguno98/stress_simulator/internal/types"
	"github.com/phuonguno98/stress_simulator/internal/utils"
)

var (
	cpuUtilization    = flag.Float64("cpu", 50.0, "Average CPU utilization percentage (0-100)")
	ioWaitRatio       = flag.Float64("iowait", 20.0, "Average CPU I/O wait percentage (0-100)")
	memoryUtilization = flag.Float64("memory", 50.0, "Average memory utilization percentage (0-100)")
	diskUtilization   = flag.Float64("disk", 50.0, "Average disk utilization percentage (0-100)")
	networkRate       = flag.Float64("network", 10.0, "Average network rate in MB/s")
	targetIP          = flag.String("target", "", "Target IP address for network stress")
	forever           = flag.Bool("forever", true, "Run indefinitely with seasonal patterns (default true)")
)

func main() {
	flag.Parse()

	if *targetIP == "" {
		log.Fatal("Target IP is required. Use -target flag.")
	}

	fmt.Printf("Starting stress simulation with:\n")
	fmt.Printf("- CPU Utilization: %.2f%%\n", *cpuUtilization)
	fmt.Printf("- I/O Wait Ratio: %.2f%%\n", *ioWaitRatio)
	fmt.Printf("- Memory Utilization: %.2f%%\n", *memoryUtilization)
	fmt.Printf("- Disk Utilization: %.2f%%\n", *diskUtilization)
	fmt.Printf("- Network Rate: %.2f MB/s\n", *networkRate)
	fmt.Printf("- Target IP: %s\n", *targetIP)
	fmt.Printf("- Run Forever: %t\n", *forever)

	// Initialize shared payload buffer (1MB) to prevent double stress on CPU/Mem
	utils.InitSharedPayload(1024 * 1024)

	// Create stop channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	// Create channels for dynamic parameter updates
	cpuParams := make(chan types.ParameterUpdate, 10)
	memoryParams := make(chan types.ParameterUpdate, 10)
	diskParams := make(chan types.ParameterUpdate, 10)
	networkParams := make(chan types.ParameterUpdate, 10)

	// Start all stress components
	var wg sync.WaitGroup

	wg.Go(func() {
		cpu.RunStress(*cpuUtilization, *ioWaitRatio, cpuParams, stopChan)
	})

	wg.Go(func() {
		memory.RunStress(*memoryUtilization, memoryParams, stopChan)
	})

	wg.Go(func() {
		disk.RunStress(*diskUtilization, diskParams, stopChan)
	})

	wg.Go(func() {
		network.RunStress(*networkRate, *targetIP, networkParams, stopChan)
	})

	// If running forever, also start the seasonal pattern generator
	if *forever {
		wg.Go(func() {
			generateSeasonalPatterns(cpuParams, memoryParams, diskParams, networkParams, stopChan)
		})
	}

	// Wait for interrupt signal
	<-stopChan
	fmt.Println("\nStopping stress simulation...")

	// Close parameter update channels to signal components to stop
	close(cpuParams)
	close(memoryParams)
	close(diskParams)
	close(networkParams)

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("Stress simulation stopped.")
}

// generateSeasonalPatterns creates realistic seasonal load patterns and unpredictable anomalies
func generateSeasonalPatterns(
	cpuChan chan<- types.ParameterUpdate,
	memoryChan chan<- types.ParameterUpdate,
	diskChan chan<- types.ParameterUpdate,
	networkChan chan<- types.ParameterUpdate,
	stopChan <-chan os.Signal,
) {
	// Convert os.Signal channel to struct{} channel
	doneChan := make(chan struct{})
	go func() {
		<-stopChan
		close(doneChan)
	}()

	// Regular seasonal pattern ticker (checks every 5 minutes for smoother transitions)
	seasonalTicker := time.NewTicker(5 * time.Minute)
	defer seasonalTicker.Stop()

	// Initial time to next anomaly
	nextAnomalyDelay := generateNextAnomalyDelay()
	anomalyTimer := time.NewTimer(nextAnomalyDelay)
	defer anomalyTimer.Stop()

	var anomalyEnd time.Time

	for {
		select {
		case <-doneChan:
			return
		case <-seasonalTicker.C:
			// Skip seasonal updates if an anomaly is currently active
			if time.Now().Before(anomalyEnd) {
				continue
			}

			now := time.Now()
			// Generate seasonal adjustments using a smooth sine wave based on time of day
			// Maps 0-24h to 0-2PI
			hourFloat := float64(now.Hour()) + float64(now.Minute())/60.0

			// Base multiplier from time of day (sine wave peaking at midday)
			// sin( (h-6)/24 * 2pi ) gives lowest at 0:00, highest at 12:00
			timeOfDayFactor := math.Sin((hourFloat - 6.0) / 24.0 * 2.0 * math.Pi)

			// Normalize to a lower baseline: [0.3, 0.7] so it drops significantly outside business hours
			dailyMultiplier := 0.5 + (timeOfDayFactor * 0.2)

			// Business hours effect (7h-11h, 13h-17h)
			if isBusinessHours(now.Hour()) {
				// Boost during business hours (stronger boost to compensate for lower base)
				dailyMultiplier += 0.6 + rand.Float64()*0.2
			}

			// Weekday vs weekend effect
			currentWeekday := now.Weekday()
			isWeekday := currentWeekday >= time.Monday && currentWeekday <= time.Friday

			finalMultiplier := dailyMultiplier
			if !isWeekday {
				// Lower overall traffic on weekends
				finalMultiplier *= (0.6 + rand.Float64()*0.2)
			}

			// Apply seasonal adjustment to all components
			updateComponents(cpuChan, memoryChan, diskChan, networkChan, finalMultiplier, 5*time.Minute)

		case <-anomalyTimer.C:
			// Unpredictable anomalies: Random magnitude and duration
			var anomalyMultiplier float64
			if rand.Float64() < 0.6 {
				// Spike: 1.5x to 3.5x of average
				anomalyMultiplier = 1.5 + rand.Float64()*2.0
			} else {
				// Dip: 0.1x to 0.5x of average
				anomalyMultiplier = 0.1 + rand.Float64()*0.4
			}

			// Random duration: from 5 minutes to 4 hours
			durationMinutes := 5 + rand.Intn(235)
			anomalyDuration := time.Duration(durationMinutes) * time.Minute

			// Track anomaly end time so seasonal updates don't overwrite it
			anomalyEnd = time.Now().Add(anomalyDuration)

			type anomalyType string
			var aType anomalyType
			if anomalyMultiplier > 1.0 {
				aType = "SPIKE"
			} else {
				aType = "DIP"
			}

			fmt.Printf("[ANOMALY] %s detected: %.2fx multiplier for %v\n", aType, anomalyMultiplier, anomalyDuration)
			updateComponents(cpuChan, memoryChan, diskChan, networkChan, anomalyMultiplier, anomalyDuration)

			// Schedule next anomaly
			nextDelay := generateNextAnomalyDelay()
			fmt.Printf("[ANOMALY] Next anomaly expected in approximately %v\n", nextDelay)
			anomalyTimer.Reset(nextDelay)
		}
	}
}

// generateNextAnomalyDelay generates a random duration until the next anomaly.
// Uses an exponential distribution to simulate a Poisson process, ensuring
// true unpredictability (can happen back-to-back or take weeks).
func generateNextAnomalyDelay() time.Duration {
	// Average time between anomalies (e.g., 3 days)
	avgHoursBetweenAnomalies := 72.0
	// ExpFloat64 returns an exponentially distributed float64 in the range (0, +math.MaxFloat64]
	hours := rand.ExpFloat64() * avgHoursBetweenAnomalies

	// Add a small jitter (e.g., min 1 hour) to avoid instant back-to-back spam
	minHours := 1.0

	totalHours := minHours + hours
	return time.Duration(totalHours * float64(time.Hour))
}

const (
	morningShiftStart   = 7
	morningShiftEnd     = 11
	afternoonShiftStart = 13
	afternoonShiftEnd   = 17
)

// isBusinessHours checks if the given hour falls within defined business hours
func isBusinessHours(hour int) bool {
	return (hour >= morningShiftStart && hour < morningShiftEnd) ||
		(hour >= afternoonShiftStart && hour < afternoonShiftEnd)
}

// updateComponents sends parameter updates to all stress components
func updateComponents(
	cpuChan, memoryChan, diskChan, networkChan chan<- types.ParameterUpdate,
	multiplier float64,
	duration time.Duration,
) {
	update := types.ParameterUpdate{
		Value:     multiplier,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	// Send update to all components (with non-blocking send)
	select {
	case cpuChan <- update:
	default:
	}

	select {
	case memoryChan <- update:
	default:
	}

	select {
	case diskChan <- update:
	default:
	}

	select {
	case networkChan <- update:
	default:
	}
}
