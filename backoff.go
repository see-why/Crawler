package main

import (
	"math"
	"time"
)

// CalculateBackoffDelay calculates exponential backoff delay with overflow protection
func CalculateBackoffDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return 0
	}
	if attempt > 10 {
		attempt = 10
	}
	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(baseDelay) * multiplier)
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
