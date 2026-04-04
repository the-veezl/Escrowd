package main

import (
	"escrowd/internal/crypto"
	"fmt"
)

type Escrow struct {
	ID        string
	Sender    string
	Receiver  string
	Amount    int
	Status    string
	CreatedAt string
}

func createEscrow(id string, sender string, receiver string, amount int) Escrow {
	deal := Escrow{

		ID:       id,
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
		Status:   "Locked",
	}
	return deal
}
func printEscrow(deal Escrow) {

	fmt.Println("------ Escrow-----")
	fmt.Println("Deal ID:", deal.ID)
	fmt.Println("From:", deal.Sender)
	fmt.Println("To:", deal.Receiver)
	fmt.Println("Amount:", deal.Amount)
	fmt.Println("Status:", deal.Status)
	fmt.Println()
}
func totalAmount(deal1 Escrow, deal2 Escrow) int {

	total := deal1.Amount + deal2.Amount
	return total
}
func checkSecret(hash string, guess string) bool {
	hashedGuess := crypto.Hash(guess)
	return hashedGuess == hash
}

func main() {

	deal1 := createEscrow("escrow-001", "alice", "Bob", 10)
	deal2 := createEscrow("escrow-002", "Charlie", "Diana", 25)
	printEscrow(deal1)
	printEscrow(deal2)

	total := totalAmount(deal1, deal2)

	fmt.Printf("Total locked: %d\n", total)

	secret := crypto.GenerateSecret()
	hash := crypto.Hash(secret)

	fmt.Println("Secret:  ", secret)
	fmt.Println("Hash:     ", hash)

	guess := crypto.Hash(secret)
	fmt.Println("Match?    ", guess == hash)

	wrongGuess := crypto.Hash("wrong-secret")
	fmt.Println("Wrong?   ", wrongGuess == hash)
	// Task 1 - hash the same word twice
	hash1 := crypto.Hash("hello")
	hash2 := crypto.Hash("hello")
	fmt.Println("Hash 1:", hash1)
	fmt.Println("Hash 2:", hash2)
	fmt.Println("Are they the same?", hash1 == hash2)

	fmt.Println()

	// Task 2 - generate three secrets
	secret1 := crypto.GenerateSecret()
	secret2 := crypto.GenerateSecret()
	secret3 := crypto.GenerateSecret()
	fmt.Println("Secret 1:", secret1)
	fmt.Println("Secret 2:", secret2)
	fmt.Println("Secret 3:", secret3)

	fmt.Println()
	fmt.Println("--- Task 3: checkSecret ---")
	mySecret := crypto.GenerateSecret()
	myHash := crypto.Hash(mySecret)

	fmt.Println("Correct guess:", checkSecret(myHash, mySecret))
	fmt.Println("Wrong guess:  ", checkSecret(myHash, "wrong-secret"))

}
