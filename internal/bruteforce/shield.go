package bruteforce

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	MaxAttempts   = 7
	LockoutPeriod = time.Hour
	BaseDelay     = time.Second
)

type attempt struct {
	count       int
	lastTry     time.Time
	lockedUntil time.Time
}

type Shield struct {
	mu       sync.Mutex
	attempts map[string]*attempt
}

func New() *Shield {
	return &Shield{
		attempts: make(map[string]*attempt),
	}
}

func (s *Shield) RecordFailure(dealID string) (waitSeconds float64, locked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.attempts[dealID]
	if !exists {
		a = &attempt{}
		s.attempts[dealID] = a
	}

	a.count++
	a.lastTry = time.Now()

	if a.count >= MaxAttempts {
		a.lockedUntil = time.Now().Add(LockoutPeriod)
		return 0, true
	}

	// exponential backoff — doubles with each failure
	delay := math.Pow(2, float64(a.count-1))
	return delay, false
}

func (s *Shield) IsLocked(dealID string) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.attempts[dealID]
	if !exists {
		return false, 0
	}

	if time.Now().Before(a.lockedUntil) {
		remaining := time.Until(a.lockedUntil)
		return true, remaining
	}

	return false, 0
}

func (s *Shield) MustWait(dealID string) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.attempts[dealID]
	if !exists {
		return false, 0
	}

	if a.count == 0 {
		return false, 0
	}

	delay := time.Duration(math.Pow(2, float64(a.count-1))) * BaseDelay
	nextAllowed := a.lastTry.Add(delay)

	if time.Now().Before(nextAllowed) {
		remaining := time.Until(nextAllowed)
		return true, remaining
	}

	return false, 0
}

func (s *Shield) RecordSuccess(dealID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.attempts, dealID)
}

func (s *Shield) AttemptsRemaining(dealID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.attempts[dealID]
	if !exists {
		return MaxAttempts
	}

	remaining := MaxAttempts - a.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func FormatDelay(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0f seconds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%.0f minutes", seconds/60)
	}
	return fmt.Sprintf("%.1f hours", seconds/3600)
}
