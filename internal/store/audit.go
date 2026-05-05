package store

import (
	"context"
	"fmt"
	"time"
)

type AuditEntry struct {
	ID        string
	EscrowID  string
	Event     string
	ActorID   string
	ActorName string
	Detail    string
	Timestamp time.Time
}

func (a *AuditStore) Record(escrowID string, event string, actorID string, actorName string, detail string) error {
	id := fmt.Sprintf("audit-%d", time.Now().UnixNano())
	_, err := a.pool.Exec(context.Background(), `
		INSERT INTO audit_log (id, escrow_id, event, actor_id, actor_name, detail)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, escrowID, event, actorID, actorName, detail)
	return err
}

func (a *AuditStore) GetByEscrow(escrowID string) ([]AuditEntry, error) {
	rows, err := a.pool.Query(context.Background(), `
		SELECT id, escrow_id, event, actor_id, actor_name, detail, created_at
		FROM audit_log
		WHERE escrow_id = $1
		ORDER BY created_at ASC
	`, escrowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.EscrowID, &e.Event,
			&e.ActorID, &e.ActorName, &e.Detail, &e.Timestamp); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}
