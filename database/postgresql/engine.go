package postgresql

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PostgresFinancialEngine coordinates transactional operations, double-entry validation,
// partition routing, and Row-Level Security checks for PostgreSQL.
type PostgresFinancialEngine struct {
	mu           sync.RWMutex
	users        map[string]*User
	accounts     map[string]*Account
	journals     map[string]*JournalEntry
	postings     map[string][]*Posting // entry_id -> postings
	orders       map[string]*Order
	trades       map[string]*Trade
	allocations  map[string]*CustodyAllocation // isin -> allocation
	fees         map[string]*TransactionFee
	taxLots      map[string]*TaxLotDisposal
	partitions   map[string]bool // registered partition table names
}

// NewPostgresFinancialEngine initializes an institutional Postgres financial engine.
func NewPostgresFinancialEngine() *PostgresFinancialEngine {
	engine := &PostgresFinancialEngine{
		users:       make(map[string]*User),
		accounts:    make(map[string]*Account),
		journals:    make(map[string]*JournalEntry),
		postings:    make(map[string][]*Posting),
		orders:      make(map[string]*Order),
		trades:      make(map[string]*Trade),
		allocations: make(map[string]*CustodyAllocation),
		fees:        make(map[string]*TransactionFee),
		taxLots:     make(map[string]*TaxLotDisposal),
		partitions:  make(map[string]bool),
	}

	// Register default baseline partitions
	engine.partitions["ledger.journal_entries_y2026m09"] = true
	engine.partitions["trading.trades_y2026m09"] = true

	return engine
}

// RegisterUser adds a user with entity boundary metadata.
func (e *PostgresFinancialEngine) RegisterUser(u *User) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if u.UserID == "" {
		return fmt.Errorf("user_id cannot be empty")
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = u.CreatedAt
	}
	e.users[u.UserID] = u
	return nil
}

// CreateAccount registers a new ledger account.
func (e *PostgresFinancialEngine) CreateAccount(acc *Account) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if acc.Balance < 0 {
		return ErrNegativeBalance
	}
	if acc.LockedBalance < 0 || acc.LockedBalance > acc.Balance {
		return ErrLockedExceedsBalance
	}
	if acc.CreatedAt.IsZero() {
		acc.CreatedAt = time.Now().UTC()
	}

	e.accounts[acc.AccountID] = acc
	return nil
}

// GetAccountRLS retrieves an account respecting Row-Level Security policies.
func (e *PostgresFinancialEngine) GetAccountRLS(ctx context.Context, rls RLSContext, accountID string) (*Account, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	acc, exists := e.accounts[accountID]
	if !exists {
		return nil, ErrAccountNotFound
	}

	if !rls.BypassRLS && rls.ActiveEntity != "" && acc.UserID != "" {
		u, uExists := e.users[acc.UserID]
		if uExists && u.EntityType != rls.ActiveEntity {
			return nil, ErrTenantAccessViolation
		}
	}

	// Return a copy
	accCopy := *acc
	return &accCopy, nil
}

// LockBalance locks funds in an account (e.g. for pending buy orders).
func (e *PostgresFinancialEngine) LockBalance(accountID string, amount int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	acc, exists := e.accounts[accountID]
	if !exists {
		return ErrAccountNotFound
	}

	if amount <= 0 {
		return fmt.Errorf("lock amount must be strictly positive")
	}

	if acc.AvailableBalance() < amount {
		return ErrInsufficientAvailableFunds
	}

	acc.LockedBalance += amount
	return nil
}

// UnlockBalance unlocks funds in an account.
func (e *PostgresFinancialEngine) UnlockBalance(accountID string, amount int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	acc, exists := e.accounts[accountID]
	if !exists {
		return ErrAccountNotFound
	}

	if amount <= 0 {
		return fmt.Errorf("unlock amount must be strictly positive")
	}

	if acc.LockedBalance < amount {
		return fmt.Errorf("cannot unlock more than currently locked balance (%d < %d)", acc.LockedBalance, amount)
	}

	acc.LockedBalance -= amount
	return nil
}

// RecordJournalEntry validates and posts an immutable double-entry transaction.
// Invariant: sum(postings.amount) == 0; all resulting account balances >= 0.
func (e *PostgresFinancialEngine) RecordJournalEntry(entry *JournalEntry, postings []*Posting) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(postings) < 2 {
		return fmt.Errorf("journal entry must contain at least 2 postings for double-entry")
	}

	// 1. Verify postings sum to zero (Debits + Credits == 0)
	var sum int64
	for _, p := range postings {
		if p.Amount == 0 {
			return ErrZeroAmountPosting
		}
		sum += p.Amount
	}
	if sum != 0 {
		return ErrDoubleEntryUnbalanced
	}

	// 2. Validate all accounts exist and will not violate non-negative balance constraints
	type updateProposal struct {
		acc       *Account
		newBal    int64
		newLocked int64
	}
	proposals := make([]updateProposal, len(postings))

	for i, p := range postings {
		acc, exists := e.accounts[p.AccountID]
		if !exists {
			return fmt.Errorf("%w: %s", ErrAccountNotFound, p.AccountID)
		}

		newBal := acc.Balance + p.Amount
		if newBal < 0 {
			return fmt.Errorf("%w for account %s: would be %d", ErrNegativeBalance, p.AccountID, newBal)
		}

		newLocked := acc.LockedBalance
		// If balance decreased (debit) and locked balance exceeds new balance, adjust locked if unlocked
		if newBal < newLocked {
			return fmt.Errorf("%w for account %s", ErrLockedExceedsBalance, p.AccountID)
		}

		proposals[i] = updateProposal{
			acc:       acc,
			newBal:    newBal,
			newLocked: newLocked,
		}
	}

	// 3. Commit postings atomically
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	// Ensure partition exists or auto-route
	partitionName := e.ResolvePartition("ledger.journal_entries", entry.CreatedAt)
	e.partitions[partitionName] = true

	for i, p := range postings {
		prop := proposals[i]
		prop.acc.Balance = prop.newBal
		prop.acc.LockedBalance = prop.newLocked

		p.EntryID = entry.EntryID
		p.EntryCreatedAt = entry.CreatedAt
		p.BalanceAfter = prop.newBal
	}

	e.journals[entry.EntryID] = entry
	e.postings[entry.EntryID] = postings

	return nil
}

// AttemptUpdateJournalEntry simulates an UPDATE attempt on ledger.journal_entries,
// which must be blocked by the immutability trigger.
func (e *PostgresFinancialEngine) AttemptUpdateJournalEntry(entryID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, exists := e.journals[entryID]; exists {
		return ErrImmutableRecordViolation
	}
	return fmt.Errorf("entry not found")
}

// AttemptDeleteTrade simulates a DELETE attempt on trading.trades,
// which must be blocked by the immutability trigger.
func (e *PostgresFinancialEngine) AttemptDeleteTrade(tradeID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, exists := e.trades[tradeID]; exists {
		return ErrImmutableRecordViolation
	}
	return fmt.Errorf("trade not found")
}

// RecordTrade stores an executed trade, updating partition routing and order states.
func (e *PostgresFinancialEngine) RecordTrade(trade *Trade) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if trade.FractionalUnits <= 0 {
		return ErrInvalidQuantity
	}
	if trade.PricePerUnit <= 0 {
		return ErrInvalidPrice
	}

	if trade.ExecutedAt.IsZero() {
		trade.ExecutedAt = time.Now().UTC()
	}

	partitionName := e.ResolvePartition("trading.trades", trade.ExecutedAt)
	e.partitions[partitionName] = true

	e.trades[trade.TradeID] = trade
	return nil
}

// ResolvePartition returns the declarative range partition table name for a timestamp.
func (e *PostgresFinancialEngine) ResolvePartition(baseTable string, t time.Time) string {
	utc := t.UTC()
	return fmt.Sprintf("%s_y%04dm%02d", baseTable, utc.Year(), int(utc.Month()))
}

// HasPartition checks if a partition table has been registered or routed to.
func (e *PostgresFinancialEngine) HasPartition(partitionName string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.partitions[partitionName]
}

// UpdateCustodyAllocation verifies 1:1 physical backing and updates allocations.
// Invariant: tokens_minted <= physical_shares_held.
func (e *PostgresFinancialEngine) UpdateCustodyAllocation(alloc *CustodyAllocation) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if alloc.TokensMinted > alloc.PhysicalSharesHeld {
		alloc.ReserveStatus = ReserveDiscrepancy
		return ErrCustodyOverissuance
	}

	alloc.ReserveStatus = ReserveBalanced
	alloc.LastReconciledAt = time.Now().UTC()
	e.allocations[alloc.ISIN] = alloc
	return nil
}

// GetCustodyAllocation retrieves custody backing for an ISIN.
func (e *PostgresFinancialEngine) GetCustodyAllocation(isin string) (*CustodyAllocation, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	alloc, exists := e.allocations[isin]
	if !exists {
		return nil, fmt.Errorf("no custody allocation found for ISIN %s", isin)
	}
	allocCopy := *alloc
	return &allocCopy, nil
}

// CalculateAndRecordTaxLot assesses capital gains under Section 111A/112A.
func (e *PostgresFinancialEngine) CalculateAndRecordTaxLot(disposal *TaxLotDisposal) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if disposal.HoldingPeriodDays < 365 {
		disposal.GainType = GainSTCG // Short-Term Capital Gains
	} else {
		disposal.GainType = GainLTCG // Long-Term Capital Gains
	}

	disposal.RealizedCapitalGain = disposal.SaleProceeds - disposal.CostBasis
	if disposal.AssessedAt.IsZero() {
		disposal.AssessedAt = time.Now().UTC()
	}

	e.taxLots[disposal.DisposalID] = disposal
	return nil
}

// RecordTransactionFee verifies fee breakdown sum invariant.
func (e *PostgresFinancialEngine) RecordTransactionFee(fee *TransactionFee) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	expectedTotal := fee.TreasuryPortionINR + fee.CoreSGFPortionINR + fee.IPFPortionINR
	if fee.TotalFeeINR != expectedTotal {
		return fmt.Errorf("fee split sum discrepancy: total %d != treasury(%d) + sgf(%d) + ipf(%d)",
			fee.TotalFeeINR, fee.TreasuryPortionINR, fee.CoreSGFPortionINR, fee.IPFPortionINR)
	}

	if fee.AssessedAt.IsZero() {
		fee.AssessedAt = time.Now().UTC()
	}

	e.fees[fee.FeeID] = fee
	return nil
}
