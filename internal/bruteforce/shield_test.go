package bruteforce

import (
	"testing"
	"time"
)

func TestRecordFailure_ExponentialBackoff(t *testing.T) {
	s := New()

	delays := []float64{1, 2, 4, 8, 16, 32}

	for i, expected := range delays {
		delay, locked := s.RecordFailure("deal-001")
		if locked {
			t.Errorf("expected not locked on attempt %d", i+1)
		}
		if delay != expected {
			t.Errorf("attempt %d: expected delay %.0f, got %.0f", i+1, expected, delay)
		}
	}
}

func TestRecordFailure_LocksAfterMaxAttempts(t *testing.T) {
	s := New()

	for i := 0; i < MaxAttempts-1; i++ {
		s.RecordFailure("deal-002")
	}

	_, locked := s.RecordFailure("deal-002")
	if !locked {
		t.Errorf("expected deal to be locked after %d attempts", MaxAttempts)
	}
}

func TestIsLocked_AfterMaxAttempts(t *testing.T) {
	s := New()

	for i := 0; i < MaxAttempts; i++ {
		s.RecordFailure("deal-003")
	}

	locked, remaining := s.IsLocked("deal-003")
	if !locked {
		t.Errorf("expected deal to be locked")
	}
	if remaining <= 0 {
		t.Errorf("expected positive remaining lockout time")
	}
}

func TestIsLocked_NewDeal(t *testing.T) {
	s := New()

	locked, _ := s.IsLocked("deal-004")
	if locked {
		t.Errorf("expected new deal to not be locked")
	}
}

func TestRecordSuccess_ClearsAttempts(t *testing.T) {
	s := New()

	s.RecordFailure("deal-005")
	s.RecordFailure("deal-005")
	s.RecordSuccess("deal-005")

	locked, _ := s.IsLocked("deal-005")
	if locked {
		t.Errorf("expected deal to be unlocked after success")
	}

	remaining := s.AttemptsRemaining("deal-005")
	if remaining != MaxAttempts {
		t.Errorf("expected full attempts after success, got %d", remaining)
	}
}

func TestAttemptsRemaining_DecreasesWithFailures(t *testing.T) {
	s := New()

	if s.AttemptsRemaining("deal-006") != MaxAttempts {
		t.Errorf("expected %d attempts remaining for new deal", MaxAttempts)
	}

	s.RecordFailure("deal-006")
	if s.AttemptsRemaining("deal-006") != MaxAttempts-1 {
		t.Errorf("expected %d attempts remaining after 1 failure", MaxAttempts-1)
	}
}

func TestMustWait_AfterFailure(t *testing.T) {
	s := New()

	s.RecordFailure("deal-007")

	mustWait, remaining := s.MustWait("deal-007")
	if !mustWait {
		t.Errorf("expected to must wait after failure")
	}
	if remaining <= 0 {
		t.Errorf("expected positive remaining wait time")
	}
}

func TestDifferentDeals_IndependentTracking(t *testing.T) {
	s := New()

	for i := 0; i < MaxAttempts; i++ {
		s.RecordFailure("deal-008")
	}

	locked, _ := s.IsLocked("deal-009")
	if locked {
		t.Errorf("expected deal-009 to be unaffected by deal-008 failures")
	}
}

func TestFormatDelay_Seconds(t *testing.T) {
	result := FormatDelay(30)
	if result != "30 seconds" {
		t.Errorf("expected '30 seconds', got %s", result)
	}
}

func TestFormatDelay_Minutes(t *testing.T) {
	result := FormatDelay(120)
	if result != "2 minutes" {
		t.Errorf("expected '2 minutes', got %s", result)
	}
}

func TestFormatDelay_Hours(t *testing.T) {
	result := FormatDelay(3600)
	if result != "1.0 hours" {
		t.Errorf("expected '1.0 hours', got %s", result)
	}
}

func TestMustWait_NoWaitForFreshDeal(t *testing.T) {
	s := New()
	mustWait, _ := s.MustWait("deal-010")
	if mustWait {
		t.Errorf("expected no wait for fresh deal")
	}
}

func TestLockout_ExpiresAfterPeriod(t *testing.T) {
	s := New()

	for i := 0; i < MaxAttempts; i++ {
		s.RecordFailure("deal-011")
	}

	// manually expire the lockout for testing
	s.mu.Lock()
	s.attempts["deal-011"].lockedUntil = time.Now().Add(-1 * time.Second)
	s.mu.Unlock()

	locked, _ := s.IsLocked("deal-011")
	if locked {
		t.Errorf("expected lockout to expire after period")
	}
}
