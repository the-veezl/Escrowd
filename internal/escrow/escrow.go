package escrow

import (
	"errors"
	"escrowd/internal/crypto"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusLocked   Status = "locked"
	StatusClaimed  Status = "claimed"
	StatusRefunded Status = "refunded"
	StatusDisputed Status = "disputed"
	StatusResolved Status = "resolved"
)

type Dispute struct {
	ID         string
	Reason     string
	Evidence   []string
	RaisedBy   string
	RaisedAt   time.Time
	ResolvedAt time.Time
	Resolution string
}

type Escrow struct {
	ID        string
	Sender    string
	Receiver  string
	Amount    int
	Status    Status
	HashLock  string
	CreatedAt time.Time
	ExpiresAt time.Time
	Dispute   *Dispute
}

func New(sender string, receiver string, amount int, secret string) Escrow {
	now := time.Now()
	return Escrow{
		ID:        uuid.NewString(),
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Status:    StatusLocked,
		HashLock:  crypto.HashSecret(secret),
		CreatedAt: now,
		ExpiresAt: now.Add(48 * time.Hour),
	}
}

func Claim(deal *Escrow, guess string) error {
	if deal.Status == StatusDisputed {
		return errors.New("escrow is under dispute — cannot claim")
	}
	if deal.Status != StatusLocked {
		return errors.New("escrow is not locked")
	}
	if !crypto.CheckSecret(deal.HashLock, guess) {
		return errors.New("wrong secret")
	}
	deal.Status = StatusClaimed
	return nil
}

// the refund function gets called only on a locked deal and once its executed,
// one cannot revert since a state machine only moves forward.
func Refund(deal *Escrow) error {
	if deal.Status == StatusDisputed {
		return errors.New("escrow is under dispute — cannot refund directly")
	}
	if deal.Status != StatusLocked {
		return errors.New("escrow is not locked")
	}
	deal.Status = StatusRefunded
	return nil
}

func RaiseDispute(deal *Escrow, raisedBy string, reason string) error {
	if deal.Status != StatusLocked {
		return errors.New("can only dispute a locked escrow")
	}
	if deal.Dispute != nil {
		return errors.New("dispute already exists for this escrow")
	}
	deal.Dispute = &Dispute{
		ID:       "d-" + uuid.NewString()[:8],
		Reason:   reason,
		RaisedBy: raisedBy,
		RaisedAt: time.Now(),
		Evidence: []string{},
	}
	deal.Status = StatusDisputed
	return nil
}

func AddEvidence(deal *Escrow, submittedBy string, link string) error {
	if deal.Status != StatusDisputed {
		return errors.New("no active dispute on this escrow")
	}
	entry := submittedBy + ": " + link
	deal.Dispute.Evidence = append(deal.Dispute.Evidence, entry)
	return nil
}

func ResolveDispute(deal *Escrow, resolution string) error {
	if deal.Status != StatusDisputed {
		return errors.New("no active dispute on this escrow")
	}
	deal.Dispute.Resolution = resolution
	deal.Dispute.ResolvedAt = time.Now()
	deal.Status = StatusResolved

	if resolution == "refund" {
		deal.Status = StatusRefunded
	} else if resolution == "release" {
		deal.Status = StatusClaimed
	}
	return nil
}

func IsExpired(deal Escrow) bool {
	return time.Now().After(deal.ExpiresAt)
}
