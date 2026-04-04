package main

import (
	"escrowd/internal/crypto"
	"escrowd/internal/escrow"
	"fmt"
)

func main() {
	secret := crypto.GenerateSecret()
	fmt.Println("Secret:", secret)

	deal := escrow.New("escrow-001", "alice", "bob", 10, secret)
	fmt.Println("Status after creation:", deal.Status)

	err := escrow.Claim(&deal, "wrong-secret")
	if err != nil {
		fmt.Println("Claim failed:", err)
	}

	err = escrow.Claim(&deal, secret)
	if err != nil {
		fmt.Println("Claim failed:", err)
	} else {
		fmt.Println("Claim succeeded! Status:", deal.Status)
	}
	// Task 1 - try to claim again after already claimed
	err = escrow.Claim(&deal, secret)
	if err != nil {
		fmt.Println("Claim again failed:", err)
	}
	// Task 2 - create a second deal and refund it
	secret2 := crypto.GenerateSecret()
	deal2 := escrow.New("escrow-002", "charlie", "diana", 25, secret2)
	fmt.Println("\nDeal2 status after creation:", deal2.Status)

	err = escrow.Refund(&deal2)
	if err != nil {
		fmt.Println("Refund failed:", err)
	} else {
		fmt.Println("Refund succeeded! Status:", deal2.Status)
	}

	// now try to claim it after refunding
	err = escrow.Claim(&deal2, secret2)
	if err != nil {
		fmt.Println("Claim after refund failed:", err)
	}
}
