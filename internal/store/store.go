package store

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"escrowd/internal/escrow"
	"os"

	badger "github.com/dgraph-io/badger/v4"
)

type Store struct {
	db      *badger.DB
	AuditDB *badger.DB
}

func New(path string) (*Store, error) {
	key, err := hex.DecodeString(os.Getenv("ESCROWD_DB_KEY"))
	if err != nil || len(key) != 32 {
		return nil, errors.New("ESCROWD_DB_KEY must be a valid 32-byte hex string")
	}

	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	opts.EncryptionKey = key
	opts.IndexCacheSize = 100 << 20
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	auditOpts := badger.DefaultOptions(path + "-audit")
	auditOpts.Logger = nil
	auditOpts.EncryptionKey = key
	auditOpts.IndexCacheSize = 100 << 20
	auditDB, err := badger.Open(auditOpts)
	if err != nil {
		return nil, err
	}

	return &Store{db: db, AuditDB: auditDB}, nil
}

func (s *Store) Close() {
	s.db.Close()
	s.AuditDB.Close()
}

func (s *Store) Save(deal escrow.Escrow) error {
	data, err := json.Marshal(deal)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(deal.ID), data)
	})
}

func (s *Store) Get(id string) (escrow.Escrow, error) {
	var deal escrow.Escrow
	err := s.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &deal)
		})
	})
	return deal, err
}

func (s *Store) Delete(id string) error {
	return s.db.Update(func(tx *badger.Txn) error {
		return tx.Delete([]byte(id))
	})
}

func (s *Store) ListIDs() ([]string, error) {
	var ids []string
	err := s.db.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			ids = append(ids, string(item.Key()))
		}
		return nil
	})
	return ids, err
}
func (s *Store) DeleteUserData(userID string) (int, error) {
	ids, err := s.ListIDs()
	if err != nil {
		return 0, err
	}

	anonymized := 0
	for _, id := range ids {
		deal, err := s.Get(id)
		if err != nil {
			continue
		}

		changed := false

		if deal.SenderID == userID {
			deal.SenderID = "deleted-user"
			deal.SenderName = "deleted-user"
			changed = true
		}

		if deal.ReceiverID == userID {
			deal.ReceiverID = "deleted-user"
			deal.ReceiverName = "deleted-user"
			changed = true
		}

		if deal.Dispute != nil && deal.Dispute.RaisedByID == userID {
			deal.Dispute.RaisedByID = "deleted-user"
			deal.Dispute.RaisedByName = "deleted-user"
			changed = true
		}

		if changed {
			err = s.Save(deal)
			if err != nil {
				continue
			}
			anonymized++
		}
	}

	return anonymized, nil
}
