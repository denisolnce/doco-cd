// Package retrywindow builds retry-go options that keep retrying a transient
// failure until a wall-clock window since the first attempt has elapsed.
package retrywindow

import (
	"time"

	"github.com/avast/retry-go/v5"
)

const (
	// Default is the window used when the user sets nothing.
	Default = 15 * time.Second

	initialDelay = 250 * time.Millisecond
	maxDelay     = 10 * time.Second
	maxJitter    = 500 * time.Millisecond
)

// Options returns options for one retry-go Do call: exponential backoff with
// jitter, retrying errors accepted by retryIf until window elapses. The window
// starts when Options is called, so build a fresh set per operation. A window
// of 0 disables retries.
func Options(window time.Duration, retryIf retry.RetryIfFunc) []retry.Option {
	deadline := time.Now().Add(window)
	backoff := retry.CombineDelay(retry.BackOffDelay, retry.RandomDelay)

	return []retry.Option{
		retry.UntilSucceeded(),
		retry.Delay(initialDelay),
		retry.MaxDelay(maxDelay),
		retry.MaxJitter(maxJitter),
		retry.LastErrorOnly(true),
		retry.DelayType(func(n uint, err error, cfg retry.DelayContext) time.Duration {
			// Never sleep past the deadline, so the last attempt lands right at it.
			return min(backoff(n, err, cfg), time.Until(deadline))
		}),
		retry.RetryIf(func(err error) bool {
			return time.Now().Before(deadline) && retryIf(err)
		}),
	}
}
