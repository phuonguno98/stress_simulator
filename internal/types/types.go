// Package types defines shared types used across the stress simulator.
package types

import "time"

// ParameterUpdate represents a dynamic parameter update
type ParameterUpdate struct {
	Value     float64
	Duration  time.Duration
	Timestamp time.Time
}
