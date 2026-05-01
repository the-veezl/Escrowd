package audit

import (
	"encoding/json"
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

type EventType string

const (
	EventLocked   EventType = "ESCROW_LOCKED"
	EventClaimed  EventType = "ESCROW_CLAIMED"
	EventRefunded EventType = "ESCROW_REFUNDED"
	EventDisputed EventType = "DISPUTE_RAISED"
	EventEvidence EventType = "EVIDENCE_SUBMITTED"
	EventResolved EventType = "DISPUTE_RESOLVED"
	EventExpired  EventType = "ESCROW_EXPIRED"
)

type Entry struct {
	ID        string
	EscrowID  string
	Event     EventType
	ActorID   string
	ActorName string
	Detail    string
	Timestamp time.Time
}

type Log struct {
	db *badger.DB
}

func New(db *badger.DB) *Log {
	return &Log{db: db}
}

func (l *Log) Record(escrowID string, event EventType, actorID string, actorName string, detail string) error {
	entry := Entry{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		EscrowID:  escrowID,
		Event:     event,
		ActorID:   actorID,
		ActorName: actorName,
		Detail:    detail,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return l.db.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(entry.ID), data)
	})
}

func (l *Log) GetByEscrow(escrowID string) ([]Entry, error) {
	var entries []Entry

	err := l.db.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := tx.NewIterator(opts)
		defer it.Close()

		prefix := []byte("audit-")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var entry Entry
				if err := json.Unmarshal(val, &entry); err != nil {
					return err
				}
				if entry.EscrowID == escrowID {
					entries = append(entries, entry)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return entries, err
}
