package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	limiter := NewRateLimiter()
	key := "rate-test"
	window := 20 * time.Millisecond

	if !limiter.Allow(key, 2, window) {
		t.Fatal("expected first request to be allowed")
	}
	if !limiter.Allow(key, 2, window) {
		t.Fatal("expected second request to be allowed")
	}
	if limiter.Allow(key, 2, window) {
		t.Fatal("expected third request to be rate limited")
	}
	time.Sleep(window + 5*time.Millisecond)
	if !limiter.Allow(key, 2, window) {
		t.Fatal("expected request to be allowed after window reset")
	}
}
