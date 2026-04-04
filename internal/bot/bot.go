package bot

import (
	"escrowd/internal/crypto"
	"escrowd/internal/escrow"
	"escrowd/internal/store"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var db *store.Store

func Start() {
	var err error
	db, err = store.New("./data")
	if err != nil {
		fmt.Println("could not open database:", err)
		return
	}
	defer db.Close()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("DISCORD_TOKEN environment variable not set")
		return
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("could not create bot session:", err)
		return
	}

	session.AddHandler(messageHandler)
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentDirectMessages

	err = session.Open()
	if err != nil {
		fmt.Println("could not connect to Discord:", err)
		return
	}
	defer session.Close()

	fmt.Println("escrowd bot is running. Press CTRL+C to stop.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("bot stopped")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, "!escrow") {
		return
	}

	parts := strings.Fields(m.Content)
	if len(parts) < 2 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow lock/claim/refund/status")
		return
	}

	command := parts[1]

	switch command {
	case "lock":
		handleLock(s, m, parts)
	case "claim":
		handleClaim(s, m, parts)
	case "refund":
		handleRefund(s, m, parts)
	case "status":
		handleStatus(s, m, parts)
	default:
		s.ChannelMessageSend(m.ChannelID, "unknown command: "+command)
	}
}

func handleLock(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow lock <receiver> <amount>")
		return
	}

	receiver := parts[2]
	amountStr := parts[3]

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "amount must be a number")
		return
	}

	sender := m.Author.Username
	secret := crypto.GenerateSecret()
	deal := escrow.New(sender, receiver, amount, secret)

	err = db.Save(deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save deal")
		return
	}

	// post deal info publicly in the channel
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Escrow locked!\nID: `%s`\nFrom: %s\nTo: %s\nAmount: %d\nExpires: %s\n\nSecret sent to your DMs %s",
		deal.ID, deal.Sender, deal.Receiver, deal.Amount,
		deal.ExpiresAt.Format("2006-01-02 15:04:05"), sender,
	))

	// send secret privately to sender
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		fmt.Println("could not create DM:", err)
		return
	}
	s.ChannelMessageSend(dm.ID, fmt.Sprintf(
		"Your secret for escrow `%s`:\n`%s`\n\nKeep this private. Share it with %s only after they deliver.",
		deal.ID, secret, receiver,
	))
}

func handleClaim(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow claim <id> <secret>")
		return
	}

	id := parts[2]
	secret := parts[3]

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	err = escrow.Claim(&deal, secret)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "claim failed: "+err.Error())
		return
	}

	err = db.Save(deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save deal")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Escrow claimed!\nID: `%s`\nStatus: %s",
		deal.ID, deal.Status,
	))
}

func handleRefund(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 3 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow refund <id>")
		return
	}

	id := parts[2]

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	err = escrow.Refund(&deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "refund failed: "+err.Error())
		return
	}

	err = db.Save(deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save deal")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Escrow refunded!\nID: `%s`\nStatus: %s",
		deal.ID, deal.Status,
	))
}

func handleStatus(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 3 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow status <id>")
		return
	}

	id := parts[2]

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"ID: `%s`\nFrom: %s\nTo: %s\nAmount: %d\nStatus: %s\nExpires: %s\nExpired: %v",
		deal.ID, deal.Sender, deal.Receiver, deal.Amount,
		deal.Status, deal.ExpiresAt.Format("2006-01-02 15:04:05"),
		escrow.IsExpired(deal),
	))
}
