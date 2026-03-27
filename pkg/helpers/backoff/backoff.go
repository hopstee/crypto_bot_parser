package backoff

import (
	"math/rand"
	"time"
)

func Backoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if baseDelay <= 0 {
		baseDelay = time.Second * 3
	}

	if attempt > 10 {
		attempt = 10
	}

	delay := baseDelay * time.Duration(1<<attempt)
	delay = min(delay, maxDelay)

	if delay <= 0 {
		delay = baseDelay
	}

	minVal := delay / 2
	return minVal + time.Duration(rand.Int63n(int64(minVal)))
}
