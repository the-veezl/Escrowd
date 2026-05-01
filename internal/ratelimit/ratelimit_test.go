package ratelimit

import (
	"testing"
	"time"
)

func TestAllow_UnderLimit(t *testing.T) {
	l := New(10, time.Hour)

	for i := 0; i < 10; i++ {
		if !l.Allow("user1") {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}
}

func TestAllow_OverLimit(t *testing.T) {
	l := New(10, time.Hour)

	for i := 0; i < 10; i++ {
		l.Allow("user1")
	}

	if l.Allow("user1") {
		t.Errorf("expected 11th request to be blocked")
	}
}

func TestAllow_DifferentUsers(t *testing.T) {
	l := New(2, time.Hour)

	l.Allow("user1")
	l.Allow("user1")

	if l.Allow("user2") != true {
		t.Errorf("expected user2 to be allowed independently of user1")
	}
}

func TestAllow_WindowResets(t *testing.T) {
	l := New(2, 100*time.Millisecond)

	l.Allow("user1")
	l.Allow("user1")

	if l.Allow("user1") {
		t.Errorf("expected user1 to be blocked before window resets")
	}

	time.Sleep(150 * time.Millisecond)

	if !l.Allow("user1") {
		t.Errorf("expected user1 to be allowed after window resets")
	}
}
