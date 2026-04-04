# escrowd

A lightweight, cryptographic escrow CLI written in Go.

escrowd lets two parties trade digital goods safely without trusting each other.
Alice locks funds with a secret. Bob delivers the goods. Alice reveals the secret.
Bob claims. No middleman. No fees. No accounts.

## The problem it solves

Every day thousands of people trade digital goods on Discord, Telegram and Reddit.
Game skins, gift cards, freelance work, digital services. Most trades are small —
under $50. PayPal charges too much. Crypto is too complex. The current solution
is "trust me bro" — and it fails 1 in 5 times.

escrowd is a cryptographic handshake that makes small trades safe.

## How it works

Alice locks 10 USDC with a secret → system generates a hash of the secret
Bob sees the locked deal → delivers the goods
Alice reveals the secret → Bob claims → funds release
If Alice never reveals → deal auto-expires → Bob gets proof of non-delivery

## Installation
```bash
git clone https://github.com/xbuyan/Escrowd.git
cd Escrowd
go build -o escrowd .
```

## Usage

### Lock a deal
```bash
./escrowd lock <sender> <receiver> <amount>
```
Example:
```bash
./escrowd lock alice bob 10
```
Output:
Escrow locked successfully
ID:         3f2a1b4c-8e7d-4f6a-9b2c-1d3e5f7a9b0c
Sender:     alice
Receiver:   bob
Amount:     10
Expires at: 2026-04-06 16:36:05
SECRET (share this with receiver to claim):
32f4a98ad1e2fd8fb037e6d80118f94b

### Check status
```bash
./escrowd status <id>
```

### Claim a deal
```bash
./escrowd claim <id> <secret>
```

### Refund a deal
```bash
./escrowd refund <id>
```

## Architecture

escrowd/
├── cmd/escrowd/      # CLI commands (lock, claim, refund, status)
├── internal/
│   ├── crypto/       # Hash-lock primitives (SHA-256, secret generation)
│   ├── escrow/       # State machine (locked → claimed / refunded)
│   └── store/        # BadgerDB persistence layer
└── main.go           # Entry point

## Security model

- Secrets are never stored — only their SHA-256 hash
- A wrong secret produces a completely different hash — one character off fails
- Deals auto-expire after 48 hours protecting both parties
- BadgerDB stores all deals locally — no external server needed

## Tech stack

- **Language:** Go (stdlib only for core logic)
- **Database:** BadgerDB (embedded, no server required)
- **Cryptography:** crypto/sha256, crypto/rand (Go standard library)
- **IDs:** UUID v4 (github.com/google/uuid)

## Roadmap

- [ ] Discord bot integration
- [ ] Telegram bot integration  
- [ ] Dispute resolution system
- [ ] Stripe payments for premium dispute resolution
- [ ] Reputation system across platforms

## License

MIT