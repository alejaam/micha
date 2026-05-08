package appshared

import (
	"sync"
	"time"
)

var (
	timeOffset     time.Duration
	timeOffsetMu   sync.RWMutex
	simulationMode bool
)

// SetSimulationMode enables or disables time simulation.
func SetSimulationMode(enabled bool) {
	timeOffsetMu.Lock()
	defer timeOffsetMu.Unlock()
	simulationMode = enabled
}

// AddTimeOffset adds the given duration to the global time offset.
func AddTimeOffset(d time.Duration) {
	timeOffsetMu.Lock()
	defer timeOffsetMu.Unlock()
	timeOffset += d
}

// ResetTimeOffset clears the global time offset.
func ResetTimeOffset() {
	timeOffsetMu.Lock()
	defer timeOffsetMu.Unlock()
	timeOffset = 0
}

// Now returns the current time, potentially adjusted by a global offset if simulation mode is active.
func Now() time.Time {
	timeOffsetMu.RLock()
	defer timeOffsetMu.RUnlock()

	if !simulationMode {
		return time.Now()
	}

	return time.Now().Add(timeOffset)
}
