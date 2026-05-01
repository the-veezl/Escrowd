package store

import (
	"encoding/json"
	"escrowd/internal/escrow"

	badger "github.com/dgraph-io/badger/v4"
)

type Store struct {
	db      *badger.DB
	AuditDB *badger.DB
}

func New(path string) (*Store, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	auditOpts := badger.DefaultOptions(path + "-audit")
	auditOpts.Logger = nil
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
