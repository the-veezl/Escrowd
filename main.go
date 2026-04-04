package main

import (
	"escrowd/internal/crypto"
	"escrowd/internal/escrow"
	"escrowd/internal/store"
	"fmt"
	"log"
)

func main() {
	// open the database
	db, err := store.New("./data")
	if err != nil {
		log.Fatal("could not open database:", err)
	}
	defer db.Close()

	// create a new deal
	secret := crypto.GenerateSecret()
	deal := escrow.New("alice", "bob", 10, secret)
	fmt.Println("Created deal:", deal.ID)

	// save it to the database
	err = db.Save(deal)
	if err != nil {
		log.Fatal("could not save deal:", err)
	}
	fmt.Println("Deal saved to database")

	// retrieve it back by ID
	found, err := db.Get(deal.ID)
	if err != nil {
		log.Fatal("could not get deal:", err)
	}
	fmt.Println("Retrieved deal:", found.ID)
	fmt.Println("Sender:        ", found.Sender)
	fmt.Println("Receiver:      ", found.Receiver)
	fmt.Println("Amount:        ", found.Amount)
	fmt.Println("Status:        ", found.Status)
	fmt.Println("Expires at:    ", found.ExpiresAt.Format("2006-01-02 15:04:05"))

	err = escrow.Claim(&deal, secret)
	if err != nil {
		log.Fatal("claim failed:", err)
	}

	err = db.Save(deal)
	if err != nil {
		log.Fatal("could not save claimed deal:", err)
	}

	found2, err := db.Get(deal.ID)
	if err != nil {
		log.Fatal("could not get deal:", err)
	}
	fmt.Println("Status after claim:", found2.Status)
}
