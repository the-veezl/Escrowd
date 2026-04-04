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
)

type Escrow struct {
	ID        string
	Sender    string
	Receiver  string
	Amount    int
	Status    Status
	HashLock  string
	CreatedAt time.Time
	ExpiresAt time.Time
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

// the refund function gets called only on a locked deal and once its executed,
// one cannot revert since a state machine only moves forward.
func Refund(deal *Escrow) error {
	if deal.Status != StatusLocked {
		return errors.New("escrow is not locked")
	}
	deal.Status = StatusRefunded
	return nil
}

func Claim(deal *Escrow, guess string) error {
	if deal.Status != StatusLocked {
		return errors.New("escrow is not locked")
	}
	if !crypto.CheckSecret(deal.HashLock, guess) {
		return errors.New("wrong secret")
	}
	deal.Status = StatusClaimed
	return nil

}
func IsExpired(deal Escrow) bool {
	return time.Now().After(deal.ExpiresAt)
}
