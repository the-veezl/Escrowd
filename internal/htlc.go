package stellar

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	horizon           = "https://horizon-testnet.stellar.org"
	networkPassphrase = "Test SDF Network : September 2015"
	timeoutHours      = 48
)

type HTLC struct {
	ID        string
	Sender    string
	Receiver  string
	Amount    string
	HashX     string
	ExpiresAt time.Time
	Funded    bool
	Claimed   bool
	TxHash    string
}

func CreateHTLC(sender, receiver, amount, secret string) (*HTLC, error) {
	hashX := sha256.Sum256([]byte(secret))
	hashXStr := hex.EncodeToString(hashX[:])

	expiresAt := time.Now().Add(time.Duration(timeoutHours) * time.Hour)

	return &HTLC{
		ID:        fmt.Sprintf("htlc_%d", time.Now().UnixNano()),
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		HashX:     hashXStr,
		ExpiresAt: expiresAt,
		Funded:    false,
		Claimed:   false,
	}, nil
}
func SubmitLockTransaction(htlc *HTLC) error {
	fmt.Printf("Submitting lock transaction for %s XLM from %s to escrow\n", htlc.Amount, htlc.Sender)
	fmt.Printf("HashX: %s\n", htlc.HashX)
	fmt.Printf("Expires at: %s\n", htlc.ExpiresAt.Format(time.RFC3339))
	fmt.Println("This will call Stellar Horizon API with ClaimableBalance operation")
	return nil
}

/*type Transaction struct {
    XDR string `json:"xdr"`
}*/

func main() {

	fmt.Println("=== HTLC Operation 1: Create Locked Payment ===")
	fmt.Println("This will lock funds with a hash condition")
	fmt.Println("Timeout: 48 hours")

}
