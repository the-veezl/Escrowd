package watcher

import (
	"escrowd/internal/escrow"
	"escrowd/internal/store"
	"fmt"
	"time"
)

func Start(db *store.Store) {
	go func() {
		fmt.Println("expiry watcher started — checking every hour")
		for {
			runCheck(db)
			time.Sleep(1 * time.Hour)
		}
	}()
}

func runCheck(db *store.Store) {
	ids, err := db.ListIDs()
	if err != nil {
		fmt.Println("watcher error listing deals:", err)
		return
	}

	expired := 0
	for _, id := range ids {
		deal, err := db.Get(id)
		if err != nil {
			continue
		}

		if deal.Status == escrow.StatusLocked && escrow.IsExpired(deal) {
			err = escrow.Refund(&deal)
			if err != nil {
				continue
			}
			err = db.Save(deal)
			if err != nil {
				continue
			}
			fmt.Printf("auto-refunded expired deal: %s (sender: %s)\n", deal.ID, deal.SenderName)
			expired++
		}

		if deal.Status == escrow.StatusDisputed && escrow.IsExpired(deal) {
			err = escrow.ResolveDispute(&deal, "refund")
			if err != nil {
				continue
			}
			err = db.Save(deal)
			if err != nil {
				continue
			}
			fmt.Printf("auto-resolved expired dispute: %s (sender: %s)\n", deal.ID, deal.SenderName)
			expired++
		}
	}

	if expired > 0 {
		fmt.Printf("watcher: auto-refunded %d expired deal(s)\n", expired)
	}
}
