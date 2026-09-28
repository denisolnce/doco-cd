package retrywindow

import (
	"errors"
	"testing"
	"time"

	"github.com/avast/retry-go/v5"
)

var (
	errTransient = errors.New("transient")
	errFatal     = errors.New("fatal")
)

func isTransient(err error) bool { return errors.Is(err, errTransient) }

func TestOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		window      time.Duration
		err         error
		minAttempts int
		maxAttempts int
	}{
		{"zero window never retries", 0, errTransient, 1, 1},
		{"fatal error never retries", time.Second, errFatal, 1, 1},
		{"transient error retries within window", 700 * time.Millisecond, errTransient, 2, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			attempts := 0
			start := time.Now()

			err := retry.New(Options(tt.window, isTransient)...).Do(func() error {
				attempts++

				return tt.err
			})

			if !errors.Is(err, tt.err) {
				t.Fatalf("Do() error = %v, want %v", err, tt.err)
			}

			if attempts < tt.minAttempts || attempts > tt.maxAttempts {
				t.Errorf("attempts = %d, want between %d and %d", attempts, tt.minAttempts, tt.maxAttempts)
			}

			if elapsed := time.Since(start); elapsed > tt.window+maxJitter+time.Second {
				t.Errorf("elapsed = %v, want at most window %v plus jitter", elapsed, tt.window)
			}
		})
	}
}

func TestOptions_StopsOnSuccess(t *testing.T) {
	t.Parallel()

	attempts := 0

	err := retry.New(Options(5*time.Second, isTransient)...).Do(func() error {
		attempts++
		if attempts < 3 {
			return errTransient
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Do() error = %v, want nil", err)
	}

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}
