package escrow

import (
	"testing"
	"time"
)

func TestNew_CreatesLockedEscrow(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	if deal.Status != StatusLocked {
		t.Errorf("expected status locked, got %s", deal.Status)
	}
	if deal.Sender != "alice" {
		t.Errorf("expected sender alice, got %s", deal.Sender)
	}
	if deal.Receiver != "bob" {
		t.Errorf("expected receiver bob, got %s", deal.Receiver)
	}
	if deal.Amount != 10 {
		t.Errorf("expected amount 10, got %d", deal.Amount)
	}
	if deal.ID == "" {
		t.Errorf("expected non-empty ID")
	}
	if deal.HashLock == "" {
		t.Errorf("expected non-empty hash lock")
	}
}

func TestNew_ExpiryIsInFuture(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	if !deal.ExpiresAt.After(time.Now()) {
		t.Errorf("expected expiry to be in the future")
	}
}

func TestNew_UniqueIDs(t *testing.T) {
	deal1 := New("alice", "bob", 10, "secret123")
	deal2 := New("alice", "bob", 10, "secret123")

	if deal1.ID == deal2.ID {
		t.Errorf("expected unique IDs, got duplicate: %s", deal1.ID)
	}
}

func TestClaim_CorrectSecret(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	err := Claim(&deal, "secret123")

	if err != nil {
		t.Errorf("expected claim to succeed, got error: %s", err)
	}
	if deal.Status != StatusClaimed {
		t.Errorf("expected status claimed, got %s", deal.Status)
	}
}

func TestClaim_WrongSecret(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	err := Claim(&deal, "wrongsecret")

	if err == nil {
		t.Errorf("expected claim to fail with wrong secret")
	}
	if deal.Status != StatusLocked {
		t.Errorf("expected status to remain locked, got %s", deal.Status)
	}
}

func TestClaim_AlreadyClaimed(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	Claim(&deal, "secret123")

	err := Claim(&deal, "secret123")

	if err == nil {
		t.Errorf("expected second claim to fail")
	}
}

func TestClaim_AfterRefund(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	Refund(&deal)

	err := Claim(&deal, "secret123")

	if err == nil {
		t.Errorf("expected claim to fail after refund")
	}
}

func TestRefund_LockedDeal(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	err := Refund(&deal)

	if err != nil {
		t.Errorf("expected refund to succeed, got error: %s", err)
	}
	if deal.Status != StatusRefunded {
		t.Errorf("expected status refunded, got %s", deal.Status)
	}
}

func TestRefund_AlreadyClaimed(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	Claim(&deal, "secret123")

	err := Refund(&deal)

	if err == nil {
		t.Errorf("expected refund to fail on claimed deal")
	}
}

func TestRefund_AlreadyRefunded(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	Refund(&deal)

	err := Refund(&deal)

	if err == nil {
		t.Errorf("expected second refund to fail")
	}
}

func TestRaiseDispute_LockedDeal(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	err := RaiseDispute(&deal, "alice", "bob never delivered")

	if err != nil {
		t.Errorf("expected dispute to succeed, got error: %s", err)
	}
	if deal.Status != StatusDisputed {
		t.Errorf("expected status disputed, got %s", deal.Status)
	}
	if deal.Dispute == nil {
		t.Errorf("expected dispute to be set")
	}
	if deal.Dispute.Reason != "bob never delivered" {
		t.Errorf("expected reason to match, got %s", deal.Dispute.Reason)
	}
}

func TestRaiseDispute_AlreadyClaimed(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	Claim(&deal, "secret123")

	err := RaiseDispute(&deal, "alice", "i changed my mind")

	if err == nil {
		t.Errorf("expected dispute to fail on claimed deal")
	}
}

func TestRaiseDispute_Duplicate(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	RaiseDispute(&deal, "alice", "first dispute")

	err := RaiseDispute(&deal, "alice", "second dispute")

	if err == nil {
		t.Errorf("expected second dispute to fail")
	}
}

func TestAddEvidence_ActiveDispute(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	RaiseDispute(&deal, "alice", "bob never delivered")

	err := AddEvidence(&deal, "alice", "https://proof.com/screenshot")

	if err != nil {
		t.Errorf("expected evidence to be added, got error: %s", err)
	}
	if len(deal.Dispute.Evidence) != 1 {
		t.Errorf("expected 1 piece of evidence, got %d", len(deal.Dispute.Evidence))
	}
}

func TestAddEvidence_NoDispute(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	err := AddEvidence(&deal, "alice", "https://proof.com/screenshot")

	if err == nil {
		t.Errorf("expected evidence to fail with no active dispute")
	}
}

func TestResolveDispute_Refund(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	RaiseDispute(&deal, "alice", "bob never delivered")

	err := ResolveDispute(&deal, "refund")

	if err != nil {
		t.Errorf("expected resolve to succeed, got error: %s", err)
	}
	if deal.Status != StatusRefunded {
		t.Errorf("expected status refunded, got %s", deal.Status)
	}
}

func TestResolveDispute_Release(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	RaiseDispute(&deal, "alice", "bob never delivered")

	err := ResolveDispute(&deal, "release")

	if err != nil {
		t.Errorf("expected resolve to succeed, got error: %s", err)
	}
	if deal.Status != StatusClaimed {
		t.Errorf("expected status claimed, got %s", deal.Status)
	}
}

func TestResolveDispute_NoDispute(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	err := ResolveDispute(&deal, "refund")

	if err == nil {
		t.Errorf("expected resolve to fail with no active dispute")
	}
}

func TestIsExpired_NotExpired(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")

	if IsExpired(deal) {
		t.Errorf("expected deal to not be expired immediately after creation")
	}
}

func TestIsExpired_Expired(t *testing.T) {
	deal := New("alice", "bob", 10, "secret123")
	deal.ExpiresAt = time.Now().Add(-1 * time.Hour)

	if !IsExpired(deal) {
		t.Errorf("expected deal to be expired when ExpiresAt is in the past")
	}
}
