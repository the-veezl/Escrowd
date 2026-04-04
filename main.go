package main

import (
	"escrowd/internal/crypto"
	"escrowd/internal/escrow"
	"fmt"
)

func main() {
	secret := crypto.GenerateSecret()

	deal := escrow.New("alice", "bob", 10, secret)

	fmt.Println("ID:         ", deal.ID)
	fmt.Println("Sender:     ", deal.Sender)
	fmt.Println("Receiver:   ", deal.Receiver)
	fmt.Println("Amount:     ", deal.Amount)
	fmt.Println("Status:     ", deal.Status)
	fmt.Println("Created at: ", deal.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("Expires at: ", deal.ExpiresAt.Format("2006-01-02 15:04:05"))

	err := escrow.Claim(&deal, secret)
	if err != nil {
		fmt.Println("Claim failed:", err)
	} else {
		fmt.Println("Claimed! New status:", deal.Status)
	}

	fmt.Println("Is expired?", escrow.IsExpired(deal))
}
