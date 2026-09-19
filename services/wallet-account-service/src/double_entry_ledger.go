package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type AccountType string

const (
	AccUserAvailable AccountType = "USER_AVAILABLE"
	AccUserLocked    AccountType = "USER_LOCKED"
	AccExchangeFee   AccountType = "EXCHANGE_FEE"
	AccClearingHouse AccountType = "CLEARING_HOUSE"
	AccTaxWithheld   AccountType = "TAX_WITHHELD"
)

type JournalPosting struct {
	AccountID   string
	Type        AccountType
	DebitPaise  uint64
	CreditPaise uint64
}

type JournalTransaction struct {
	TxID        string
	Description string
	Postings    []JournalPosting
	CreatedAt   time.Time
}

type DoubleEntryLedger struct {
	mu       sync.Mutex
	balances map[string]int64 // accountID -> net balance
	history  []JournalTransaction
}

func NewDoubleEntryLedger() *DoubleEntryLedger {
	return &DoubleEntryLedger{
		balances: make(map[string]int64),
		history:  make([]JournalTransaction, 0),
	}
}

// PostTransaction executes a multi-leg entry verifying sum(debits) == sum(credits)
func (l *DoubleEntryLedger) PostTransaction(txID, desc string, postings []JournalPosting) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var totalDebit, totalCredit uint64
	for _, p := range postings {
		totalDebit += p.DebitPaise
		totalCredit += p.CreditPaise
	}

	if totalDebit != totalCredit {
		return fmt.Errorf("ledger imbalance: debits (%d) != credits (%d)", totalDebit, totalCredit)
	}

	// Verify no negative balance for user available accounts
	for _, p := range postings {
		if p.Type == AccUserAvailable && p.DebitPaise > 0 {
			current := l.balances[p.AccountID]
			if current < int64(p.DebitPaise) {
				return errors.New("insufficient funds for account " + p.AccountID)
			}
		}
	}

	// Commit entries atomically
	for _, p := range postings {
		l.balances[p.AccountID] += int64(p.CreditPaise) - int64(p.DebitPaise)
	}

	l.history = append(l.history, JournalTransaction{
		TxID:        txID,
		Description: desc,
		Postings:    postings,
		CreatedAt:   time.Now().UTC(),
	})

	return nil
}
