package escrow

import (
	"errors"
	"escrowd/internal/crypto"
)

type Status string

const (
	StatusLocked   Status = "locked"
	StatusClaimed  Status = "claimed"
	StatusRefunded Status = "refunded"
)

type Escrow struct {
	ID       string
	Sender   string
	Receiver string
	Amount   int
	Status   Status
	HashLock string
}

func New(id string, sender string, receiver string, amount int, secret string) Escrow {

	return Escrow{

		ID:       id,
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
		Status:   StatusLocked,
		HashLock: crypto.HashSecret(secret),
	}
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

// the refund function gets called only on a locked deal and once its executed,
// one cannot revert since a state machine only moves forward.
func Refund(deal *Escrow) error {

	if deal.Status != StatusLocked {
		return errors.New("esscrow is not locked")
	}
	deal.Status = StatusRefunded
	return nil
}
