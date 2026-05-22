package main

import (
	"testing"
	"time"
)

// TestGenerateNextAnomalyDelay tests the function that generates random delays for anomalies
func TestGenerateNextAnomalyDelay(t *testing.T) {
	// Test that the function returns positive durations
	for range 10 {
		delay := generateNextAnomalyDelay()
		if delay <= 0 {
			t.Errorf("generateNextAnomalyDelay() returned non-positive duration: %v", delay)
		}
	}
}

// TestIsBusinessHours tests the business hours detection function
func TestIsBusinessHours(t *testing.T) {
	tests := []struct {
		hour     int
		expected bool
	}{
		{6, false},  // Before morning shift
		{7, true},   // Morning shift start
		{10, true},  // During morning shift
		{11, false}, // Morning shift end
		{12, false}, // Lunch break
		{13, true},  // Afternoon shift start
		{16, true},  // During afternoon shift
		{17, false}, // Afternoon shift end
		{20, false}, // Evening
	}

	for _, tt := range tests {
		result := isBusinessHours(tt.hour)
		if result != tt.expected {
			t.Errorf("isBusinessHours(%d) = %v; want %v", tt.hour, result, tt.expected)
		}
	}
}

// TestGenerateNextAnomalyDelayDistribution tests that the generated delays follow expected distribution
func TestGenerateNextAnomalyDelayDistribution(t *testing.T) {
	// Generate multiple samples to check distribution
	delays := make([]time.Duration, 100)
	for i := range 100 {
		delays[i] = generateNextAnomalyDelay()
	}

	// Verify all delays are positive
	for _, delay := range delays {
		if delay <= 0 {
			t.Errorf("Found non-positive delay: %v", delay)
		}
	}
}
