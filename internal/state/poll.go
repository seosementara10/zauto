package state

import "time"

// DefaultPollInterval is used when engine.poll_sec is 0 or unset.
const DefaultPollInterval = 400 * time.Millisecond

// PollIntervalFromSec returns poll duration from config seconds (0 → DefaultPollInterval).
func PollIntervalFromSec(sec float64) time.Duration {
	if sec > 0 {
		return time.Duration(sec * float64(time.Second))
	}
	return DefaultPollInterval
}
