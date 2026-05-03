package stellar

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/stellar/go-stellar-sdk/clients/horizonclient"
	"github.com/stellar/go-stellar-sdk/keypair"
	"github.com/stellar/go-stellar-sdk/network"
	"github.com/stellar/go-stellar-sdk/txnbuild"
)

const (
	TestnetHorizon = "https://horizon-testnet.stellar.org"
	TestnetNetwork = network.TestNetworkPassphrase
)

type EscrowWallet struct {
	PublicKey string
	SecretKey string
	EscrowID  string
}

func client() *horizonclient.Client {
	return &horizonclient.Client{
		HorizonURL: TestnetHorizon,
		HTTP:       http.DefaultClient,
	}
}

func isTestnet() bool {
	return os.Getenv("STELLAR_NETWORK") != "mainnet"
}

func networkPassphrase() string {
	if isTestnet() {
		return TestnetNetwork
	}
	return network.PublicNetworkPassphrase
}

// GenerateEscrowWallet creates a new Stellar keypair for a specific deal
func GenerateEscrowWallet(escrowID string) (*EscrowWallet, error) {
	if !isTestnet() {
		return nil, errors.New("mainnet not enabled — set STELLAR_NETWORK=mainnet explicitly")
	}

	kp, err := keypair.Random()
	if err != nil {
		return nil, fmt.Errorf("could not generate keypair: %w", err)
	}

	return &EscrowWallet{
		PublicKey: kp.Address(),
		SecretKey: kp.Seed(),
		EscrowID:  escrowID,
	}, nil
}

// FundTestnetWallet uses Stellar's Friendbot to fund a testnet wallet
func FundTestnetWallet(publicKey string) error {
	if !isTestnet() {
		return errors.New("Friendbot only works on testnet")
	}

	resp, err := http.Get(
		fmt.Sprintf("https://friendbot.stellar.org?addr=%s", publicKey),
	)
	if err != nil {
		return fmt.Errorf("friendbot request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("friendbot returned status: %d", resp.StatusCode)
	}

	return nil
}

// GetBalance returns the XLM balance of a Stellar account
func GetBalance(publicKey string) (string, error) {
	c := client()

	account, err := c.AccountDetail(horizonclient.AccountRequest{
		AccountID: publicKey,
	})
	if err != nil {
		return "", fmt.Errorf("could not get account: %w", err)
	}

	for _, balance := range account.Balances {
		if balance.Asset.Type == "native" {
			return balance.Balance, nil
		}
	}

	return "0", nil
}

// SendPayment sends XLM from one account to another
func SendPayment(fromSecret string, toPublic string, amount string, memo string) (string, error) {
	if !isTestnet() {
		return "", errors.New("mainnet payments not enabled")
	}

	c := client()

	kp, err := keypair.ParseFull(fromSecret)
	if err != nil {
		return "", fmt.Errorf("invalid secret key: %w", err)
	}

	sourceAccount, err := c.AccountDetail(horizonclient.AccountRequest{
		AccountID: kp.Address(),
	})
	if err != nil {
		return "", fmt.Errorf("could not load source account: %w", err)
	}

	// use the provided amount directly rather than full balance
	// caller is responsible for leaving enough for fees and reserve
	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{
			&txnbuild.Payment{
				Destination: toPublic,
				Amount:      amount,
				Asset:       txnbuild.NativeAsset{},
			},
		},
		BaseFee:       txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{TimeBounds: txnbuild.NewInfiniteTimeout()},
		Memo:          txnbuild.MemoText(memo),
	})
	if err != nil {
		return "", fmt.Errorf("could not build transaction: %w", err)
	}

	tx, err = tx.Sign(networkPassphrase(), kp)
	if err != nil {
		return "", fmt.Errorf("could not sign transaction: %w", err)
	}

	resp, err := c.SubmitTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("could not submit transaction: %w", err)
	}

	return resp.Hash, nil
}

// GetMasterPublicKey returns the master wallet public key from env
func GetMasterPublicKey() (string, error) {
	secret := os.Getenv("STELLAR_MASTER_SECRET")
	if secret == "" {
		return "", errors.New("STELLAR_MASTER_SECRET not set")
	}

	kp, err := keypair.ParseFull(secret)
	if err != nil {
		return "", fmt.Errorf("invalid master secret: %w", err)
	}

	return kp.Address(), nil
}

// ValidateNetwork confirms we are on the right network before any operation
func ValidateNetwork() error {
	network := os.Getenv("STELLAR_NETWORK")
	if network == "" {
		return errors.New("STELLAR_NETWORK not set — use 'testnet' or 'mainnet'")
	}
	if network != "testnet" && network != "mainnet" {
		return fmt.Errorf("invalid STELLAR_NETWORK value: %s — must be 'testnet' or 'mainnet'", network)
	}
	if network == "mainnet" {
		fmt.Println("WARNING: running on Stellar mainnet — real funds at risk")
	}
	return nil
}

// GenerateMasterKeypair generates a new master keypair for first-time setup
func GenerateMasterKeypair() (publicKey string, secretKey string, err error) {
	kp, err := keypair.Random()
	if err != nil {
		return "", "", fmt.Errorf("could not generate master keypair: %w", err)
	}
	return kp.Address(), kp.Seed(), nil
}

// NetworkInfo returns the SDK version details for the current connection
func NetworkInfo() string {
	net := os.Getenv("STELLAR_NETWORK")
	return fmt.Sprintf("Stellar SDK v0.5.0 | Network: %s | Horizon: %s", net, TestnetHorizon)
}
