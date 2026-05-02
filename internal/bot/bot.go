package bot

import (
	"escrowd/internal/crypto"
	"escrowd/internal/escrow"
	"escrowd/internal/payment"
	"escrowd/internal/store"
	"escrowd/internal/validator"
	"escrowd/internal/watcher"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"escrowd/internal/ratelimit"
	"time"

	"escrowd/internal/audit"
	"escrowd/internal/backup"

	"github.com/bwmarrin/discordgo"
)

var db *store.Store
var limiter *ratelimit.Limiter
var auditLog *audit.Log

func Start() {
	var err error
	db, err = store.New("./data")
	if err != nil {
		fmt.Println("could not open database:", err)
		return
	}
	defer db.Close()

	watcher.Start(db)
	limiter = ratelimit.New(10, time.Hour)
	auditLog = audit.New(db.AuditDB)
	backup.StartScheduled("./data", "./backups", 24*time.Hour)

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

	if !limiter.Allow(m.Author.ID) {
		s.ChannelMessageSend(m.ChannelID, "you have reached the limit of 10 escrow operations per hour — try again later")
		return
	}

	if !strings.HasPrefix(m.Content, "!escrow") {
		return
	}

	parts := strings.Fields(m.Content)
	if len(parts) < 2 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow lock/claim/refund/status/dispute/evidence/resolve")
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
	case "dispute":
		handleDispute(s, m, parts)
	case "evidence":
		handleEvidence(s, m, parts)
	case "resolve":
		handleResolve(s, m, parts)
	case "history":
		handleHistory(s, m, parts)
	case "forget":
		handleForget(s, m, parts)
	case "backup":
		handleBackup(s, m, parts)
	case "paid":
		handlePaid(s, m, parts)
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

	if err := validator.ValidateName(receiver); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid receiver: "+err.Error())
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "amount must be a number")
		return
	}

	if err := validator.ValidateAmount(amount); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid amount: "+err.Error())
		return
	}

	senderID := m.Author.ID
	senderName := m.Author.Username

	if err := validator.ValidateName(senderName); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid sender name: "+err.Error())
		return
	}

	secret := crypto.GenerateSecret()
	deal := escrow.New(senderID, senderName, receiver, receiver, amount, secret)

	err = db.Save(deal)
	auditLog.Record(deal.ID, audit.EventLocked, senderID, senderName,
		fmt.Sprintf("locked %d for %s", deal.Amount, deal.ReceiverName))
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save deal")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Escrow locked!\nID: `%s`\nFrom: %s\nTo: %s\nAmount: %d\nExpires: %s\n\nSecret sent to your DMs %s",
		deal.ID, deal.SenderName, deal.ReceiverName, deal.Amount,
		deal.ExpiresAt.Format("2006-01-02 15:04:05"), senderName,
	))

	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		fmt.Println("could not create DM:", err)
		return
	}
	s.ChannelMessageSend(dm.ID, fmt.Sprintf(
		"Your secret for escrow `%s`:\n`%s`\n\nKeep this private. Share it with %s only after they deliver.",
		deal.ID, secret, deal.ReceiverName,
	))
	crypto.ZeroString(&secret)
}

func handleClaim(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow claim <id> <secret>")
		return
	}

	id := parts[2]
	secret := parts[3]

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	if secret == "" {
		s.ChannelMessageSend(m.ChannelID, "secret cannot be empty")
		return
	}

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
	auditLog.Record(deal.ID, audit.EventClaimed, m.Author.ID, m.Author.Username,
		"escrow claimed with correct secret")
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

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	if deal.SenderID != m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "only the sender can refund this escrow")
		return
	}

	err = escrow.Refund(&deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "refund failed: "+err.Error())
		return
	}

	err = db.Save(deal)
	auditLog.Record(deal.ID, audit.EventRefunded, m.Author.ID, m.Author.Username,
		"escrow refunded by sender")
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

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	verified := "✅ verified"
	if !escrow.VerifySignature(deal) {
		verified = "⚠️ TAMPERED — signature mismatch"
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"ID: `%s`\nFrom: %s\nTo: %s\nAmount: %d\nStatus: %s\nExpires: %s\nExpired: %v\nIntegrity: %s",
		deal.ID, deal.SenderName, deal.ReceiverName, deal.Amount,
		deal.Status, deal.ExpiresAt.Format("2006-01-02 15:04:05"),
		escrow.IsExpired(deal), verified,
	))
}
func handleDispute(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow dispute <id> <reason>")
		return
	}

	id := parts[2]
	reason := strings.Join(parts[3:], " ")

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	if err := validator.ValidateReason(reason); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid reason: "+err.Error())
		return
	}

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	if deal.SenderID != m.Author.ID && deal.ReceiverID != m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "only the sender or receiver can dispute this escrow")
		return
	}

	err = escrow.RaiseDispute(&deal, m.Author.ID, m.Author.Username, reason)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "dispute failed: "+err.Error())
		return
	}

	err = db.Save(deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save dispute")
		return
	}

	auditLog.Record(deal.ID, audit.EventDisputed,
		m.Author.ID, m.Author.Username,
		"dispute raised: "+reason)

	// generate payment link for turbo dispute
	reference := fmt.Sprintf("dispute-%s", deal.Dispute.ID)
	payURL, err := payment.InitializePayment(
		m.Author.Username+"@escrowd.app",
		6000, // KES 60 in kobo
		reference,
		map[string]string{
			"escrow_id":  deal.ID,
			"dispute_id": deal.Dispute.ID,
			"raised_by":  m.Author.Username,
		},
	)

	freeOption := "Free: auto-resolved in 24 hours"
	fastOption := "Fast option unavailable right now"
	if err == nil {
		fastOption = fmt.Sprintf("Fast: pay KES 60 for 15-min resolution\n%s", payURL)
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Dispute raised!\nID: `%s`\nDispute ID: `%s`\nRaised by: %s\nReason: %s\n\n"+
			"The escrow is now frozen.\n\n"+
			"Resolution options:\n• %s\n• %s\n\n"+
			"Submit evidence with:\n`!escrow evidence %s <link-to-proof>`",
		deal.ID, deal.Dispute.ID, m.Author.Username, reason,
		freeOption, fastOption, deal.ID,
	))
}
func handleEvidence(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow evidence <id> <link>")
		return
	}

	id := parts[2]
	link := parts[3]

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	if err := validator.ValidateLink(link); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid link: "+err.Error())
		return
	}

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	if deal.SenderID != m.Author.ID && deal.ReceiverID != m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "only the sender or receiver can submit evidence")
		return
	}

	err = escrow.AddEvidence(&deal, m.Author.ID, m.Author.Username, link)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "evidence failed: "+err.Error())
		return
	}

	err = db.Save(deal)
	auditLog.Record(deal.ID, audit.EventEvidence, m.Author.ID, m.Author.Username,
		"evidence submitted: "+link)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save evidence")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Evidence recorded!\nDispute: `%s`\nSubmitted by: %s\nLink: %s\nTotal evidence: %d piece(s)",
		deal.Dispute.ID, m.Author.Username, link, len(deal.Dispute.Evidence),
	))
}

func handleResolve(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow resolve <id> <refund|release>")
		return
	}

	id := parts[2]
	resolution := parts[3]

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	if resolution != "refund" && resolution != "release" {
		s.ChannelMessageSend(m.ChannelID, "resolution must be either 'refund' or 'release'")
		return
	}

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	// only the bot owner can resolve disputes
	// the bot owner is identified by their Discord username
	// replace "klucianob" with your actual Discord username
	if m.Author.Username != "klucianob_95373" {
		s.ChannelMessageSend(m.ChannelID, "only an escrowd admin can resolve disputes")
		return
	}

	err = escrow.ResolveDispute(&deal, resolution)
	auditLog.Record(deal.ID, audit.EventResolved, m.Author.ID, m.Author.Username,
		"dispute resolved: "+resolution)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "resolve failed: "+err.Error())
		return
	}

	err = db.Save(deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save resolution")
		return
	}

	outcome := "funds released to receiver"
	if resolution == "refund" {
		outcome = "funds returned to sender"
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Dispute resolved!\nID: `%s`\nResolution: %s\nOutcome: %s\nFinal status: %s",
		deal.ID, resolution, outcome, deal.Status,
	))
}
func handleHistory(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 3 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow history <id>")
		return
	}

	id := parts[2]

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	entries, err := auditLog.GetByEscrow(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not retrieve history")
		return
	}

	if len(entries) == 0 {
		s.ChannelMessageSend(m.ChannelID, "no history found for this escrow")
		return
	}

	msg := fmt.Sprintf("Audit trail for `%s`:\n", id)
	for _, e := range entries {
		msg += fmt.Sprintf("• %s — %s by %s at %s\n",
			e.Event, e.Detail, e.ActorName,
			e.Timestamp.Format("2006-01-02 15:04:05"))
	}

	s.ChannelMessageSend(m.ChannelID, msg)
}
func handleForget(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not open DM")
		return
	}

	count, err := db.DeleteUserData(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(dm.ID, "could not process data deletion request")
		return
	}

	auditLog.Record(
		"system",
		audit.EventType("USER_DATA_DELETED"),
		m.Author.ID,
		m.Author.Username,
		fmt.Sprintf("user requested data deletion — %d deals anonymized", count),
	)

	s.ChannelMessageSend(dm.ID, fmt.Sprintf(
		"Your data deletion request has been processed.\n\n"+
			"• %d deal(s) have been anonymized\n"+
			"• Your user ID and username have been replaced with 'deleted-user'\n"+
			"• Financial records are retained for legal compliance\n"+
			"• Audit logs retain event timestamps but not your identity\n\n"+
			"This action is irreversible.",
		count,
	))

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"%s your data deletion request has been processed. Check your DMs for details.",
		m.Author.Mention(),
	))
}
func handleBackup(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if m.Author.Username != "klucianob_95373" {
		s.ChannelMessageSend(m.ChannelID, "only an escrowd admin can trigger backups — your username is: "+m.Author.Username)
		return
	}

	filename, err := backup.Create("./data", "./backups")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "backup failed: "+err.Error())
		return
	}

	auditLog.Record("system", audit.EventType("BACKUP_CREATED"),
		m.Author.ID, m.Author.Username, "manual backup: "+filename)

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Backup created successfully\nFile: `%s`\nBoth databases backed up and compressed.",
		filename,
	))
}
func handlePaid(s *discordgo.Session, m *discordgo.MessageCreate, parts []string) {
	if len(parts) < 4 {
		s.ChannelMessageSend(m.ChannelID, "usage: !escrow paid <id> <paystack-reference>")
		return
	}

	id := parts[2]
	reference := parts[3]

	if err := validator.ValidateID(id); err != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid ID: "+err.Error())
		return
	}

	deal, err := db.Get(id)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "deal not found: "+id)
		return
	}

	if deal.Status != escrow.StatusDisputed {
		s.ChannelMessageSend(m.ChannelID, "no active dispute on this escrow")
		return
	}

	if deal.SenderID != m.Author.ID && deal.ReceiverID != m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "only the sender or receiver can upgrade this dispute")
		return
	}

	paid, err := payment.VerifyPayment(reference)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not verify payment: "+err.Error())
		return
	}

	if !paid {
		s.ChannelMessageSend(m.ChannelID, "payment not confirmed — please complete payment first")
		return
	}

	deal.Dispute.Priority = true
	deal.Dispute.PayReference = reference

	err = db.Save(deal)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "could not save deal")
		return
	}

	auditLog.Record(deal.ID, audit.EventType("DISPUTE_UPGRADED"),
		m.Author.ID, m.Author.Username,
		"dispute upgraded to priority via payment: "+reference)

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Payment confirmed! Dispute `%s` upgraded to priority.\nAn admin will review and resolve within 15 minutes.",
		deal.Dispute.ID,
	))
}
