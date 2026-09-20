package postgresql

import (
	"context"
	"testing"
	"time"
)

func TestPostgresDoubleEntryBalancing(t *testing.T) {
	engine := NewPostgresFinancialEngine()

	// Register accounts
	acc1 := &Account{AccountID: "acc-user-1", Balance: 100000, LockedBalance: 0} // 10.0000 INR
	acc2 := &Account{AccountID: "acc-escrow", Balance: 500000, LockedBalance: 0}  // 50.0000 INR
	if err := engine.CreateAccount(acc1); err != nil {
		t.Fatalf("failed to create acc1: %v", err)
	}
	if err := engine.CreateAccount(acc2); err != nil {
		t.Fatalf("failed to create acc2: %v", err)
	}

	// 1. Balanced entry (Debit acc1 20000, Credit acc2 20000)
	entry1 := &JournalEntry{
		EntryID:       "je-1",
		ReferenceID:   "ref-trade-1",
		ReferenceType: "TRADE",
		Description:   "Buy settlement",
		CreatedAt:     time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC),
	}
	postings1 := []*Posting{
		{AccountID: "acc-user-1", Amount: -20000},
		{AccountID: "acc-escrow", Amount: 20000},
	}

	if err := engine.RecordJournalEntry(entry1, postings1); err != nil {
		t.Fatalf("expected valid balanced journal entry, got err: %v", err)
	}

	// Verify post-balances
	acc1Check, _ := engine.GetAccountRLS(context.Background(), RLSContext{BypassRLS: true}, "acc-user-1")
	if acc1Check.Balance != 80000 {
		t.Errorf("expected acc1 balance 80000, got %d", acc1Check.Balance)
	}
	acc2Check, _ := engine.GetAccountRLS(context.Background(), RLSContext{BypassRLS: true}, "acc-escrow")
	if acc2Check.Balance != 520000 {
		t.Errorf("expected acc2 balance 520000, got %d", acc2Check.Balance)
	}

	// 2. Unbalanced entry
	entry2 := &JournalEntry{
		EntryID:       "je-2",
		ReferenceID:   "ref-trade-2",
		ReferenceType: "TRADE",
		Description:   "Unbalanced entry",
	}
	postings2 := []*Posting{
		{AccountID: "acc-user-1", Amount: -5000},
		{AccountID: "acc-escrow", Amount: 4000}, // off by 1000!
	}
	if err := engine.RecordJournalEntry(entry2, postings2); err != ErrDoubleEntryUnbalanced {
		t.Fatalf("expected ErrDoubleEntryUnbalanced, got: %v", err)
	}

	// 3. Overdraft rejection
	entry3 := &JournalEntry{
		EntryID:       "je-3",
		ReferenceID:   "ref-trade-3",
		ReferenceType: "TRADE",
		Description:   "Overdraft test",
	}
	postings3 := []*Posting{
		{AccountID: "acc-user-1", Amount: -100000}, // acc1 only has 80000
		{AccountID: "acc-escrow", Amount: 100000},
	}
	if err := engine.RecordJournalEntry(entry3, postings3); err == nil {
		t.Fatal("expected error on overdraft, got nil")
	}
}

func TestLockedBalanceInvariant(t *testing.T) {
	engine := NewPostgresFinancialEngine()
	acc := &Account{AccountID: "acc-lock", Balance: 50000, LockedBalance: 0}
	_ = engine.CreateAccount(acc)

	// Lock 20000
	if err := engine.LockBalance("acc-lock", 20000); err != nil {
		t.Fatalf("unexpected error locking: %v", err)
	}

	// Try to lock another 40000 (available is only 30000)
	if err := engine.LockBalance("acc-lock", 40000); err != ErrInsufficientAvailableFunds {
		t.Fatalf("expected ErrInsufficientAvailableFunds, got: %v", err)
	}

	// Unlock 10000
	if err := engine.UnlockBalance("acc-lock", 10000); err != nil {
		t.Fatalf("unexpected error unlocking: %v", err)
	}

	// Unlocking more than locked balance fails
	if err := engine.UnlockBalance("acc-lock", 20000); err == nil {
		t.Fatal("expected error unlocking more than locked balance, got nil")
	}
}

func TestCustodyOneToOneBacking(t *testing.T) {
	engine := NewPostgresFinancialEngine()

	// 1. Valid allocation (shares >= tokens)
	validAlloc := &CustodyAllocation{
		AllocationID:       "alloc-1",
		ISIN:               "INE002A01018",
		Depository:         "NSDL",
		DematAccountNo:     "1208160012345678",
		PhysicalSharesHeld: 100000000, // 100.000000 shares
		TokensMinted:       100000000, // 100.000000 tokens
	}
	if err := engine.UpdateCustodyAllocation(validAlloc); err != nil {
		t.Fatalf("expected valid custody allocation, got %v", err)
	}

	alloc, _ := engine.GetCustodyAllocation("INE002A01018")
	if alloc.ReserveStatus != ReserveBalanced {
		t.Errorf("expected ReserveBalanced, got %v", alloc.ReserveStatus)
	}

	// 2. Overissuance violation (tokens > physical shares)
	invalidAlloc := &CustodyAllocation{
		AllocationID:       "alloc-2",
		ISIN:               "INE009A01021",
		Depository:         "CDSL",
		DematAccountNo:     "1208160087654321",
		PhysicalSharesHeld: 50000000,  // 50 shares
		TokensMinted:       60000000,  // 60 tokens (Overissuance!)
	}
	if err := engine.UpdateCustodyAllocation(invalidAlloc); err != ErrCustodyOverissuance {
		t.Fatalf("expected ErrCustodyOverissuance, got %v", err)
	}
}

func TestImmutabilityAuditTriggers(t *testing.T) {
	engine := NewPostgresFinancialEngine()

	trade := &Trade{
		TradeID:          "trade-immut-1",
		BuyerOrderID:     "ord-b1",
		SellerOrderID:    "ord-s1",
		ISIN:             "INE002A01018",
		Symbol:           "RELIANCE",
		TokenAddress:     "0x1234567890123456789012345678901234567890",
		FractionalUnits:  1000000,
		PricePerUnit:     25000000,
		GrossAmount:      25000000,
		SettlementStatus: SettlementSettled,
		ExecutedAt:       time.Now().UTC(),
	}

	if err := engine.RecordTrade(trade); err != nil {
		t.Fatalf("failed to record trade: %v", err)
	}

	// Simulate DELETE attempt on trading.trades
	if err := engine.AttemptDeleteTrade("trade-immut-1"); err != ErrImmutableRecordViolation {
		t.Fatalf("expected ErrImmutableRecordViolation on trade delete, got %v", err)
	}

	// Simulate UPDATE attempt on journal entry
	entry := &JournalEntry{EntryID: "je-immut-1", ReferenceID: "ref-1", ReferenceType: "DVP"}
	acc1 := &Account{AccountID: "acc-i1", Balance: 50000}
	acc2 := &Account{AccountID: "acc-i2", Balance: 50000}
	_ = engine.CreateAccount(acc1)
	_ = engine.CreateAccount(acc2)
	_ = engine.RecordJournalEntry(entry, []*Posting{
		{AccountID: "acc-i1", Amount: -10000},
		{AccountID: "acc-i2", Amount: 10000},
	})

	if err := engine.AttemptUpdateJournalEntry("je-immut-1"); err != ErrImmutableRecordViolation {
		t.Fatalf("expected ErrImmutableRecordViolation on journal update, got %v", err)
	}
}

func TestRowLevelSecurityIsolation(t *testing.T) {
	engine := NewPostgresFinancialEngine()

	// Register domestic user
	userDom := &User{
		UserID:            "user-domestic-1",
		EntityType:        EntityDomesticRegulated,
		BlockchainAddress: "0x1111111111111111111111111111111111111111",
		Status:            UserStatusActive,
	}
	_ = engine.RegisterUser(userDom)

	accDom := &Account{
		AccountID: "acc-domestic",
		UserID:    "user-domestic-1",
		Balance:   100000,
	}
	_ = engine.CreateAccount(accDom)

	// Register GIFT City user
	userGift := &User{
		UserID:            "user-gift-1",
		EntityType:        EntityGiftCityGateway,
		BlockchainAddress: "0x2222222222222222222222222222222222222222",
		Status:            UserStatusActive,
	}
	_ = engine.RegisterUser(userGift)

	accGift := &Account{
		AccountID: "acc-gift",
		UserID:    "user-gift-1",
		Balance:   200000,
	}
	_ = engine.CreateAccount(accGift)

	// Query with DOMESTIC context: can see accDom, cannot see accGift
	ctxDom := RLSContext{ActiveEntity: EntityDomesticRegulated}
	if _, err := engine.GetAccountRLS(context.Background(), ctxDom, "acc-domestic"); err != nil {
		t.Errorf("domestic context should be allowed to view acc-domestic: %v", err)
	}
	if _, err := engine.GetAccountRLS(context.Background(), ctxDom, "acc-gift"); err != ErrTenantAccessViolation {
		t.Errorf("domestic context accessing GIFT City account must return ErrTenantAccessViolation, got %v", err)
	}

	// Query with GIFT CITY context: can see accGift, cannot see accDom
	ctxGift := RLSContext{ActiveEntity: EntityGiftCityGateway}
	if _, err := engine.GetAccountRLS(context.Background(), ctxGift, "acc-gift"); err != nil {
		t.Errorf("gift context should be allowed to view acc-gift: %v", err)
	}
	if _, err := engine.GetAccountRLS(context.Background(), ctxGift, "acc-domestic"); err != ErrTenantAccessViolation {
		t.Errorf("gift context accessing domestic account must return ErrTenantAccessViolation, got %v", err)
	}
}

func TestPartitionRouting(t *testing.T) {
	engine := NewPostgresFinancialEngine()

	date1 := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	partitionName := engine.ResolvePartition("trading.trades", date1)
	if partitionName != "trading.trades_y2026m09" {
		t.Errorf("expected trading.trades_y2026m09, got %s", partitionName)
	}

	trade := &Trade{
		TradeID:         "trade-p1",
		ISIN:            "INE002A01018",
		FractionalUnits: 1000000,
		PricePerUnit:    10000,
		GrossAmount:     10000,
		ExecutedAt:      date1,
	}
	if err := engine.RecordTrade(trade); err != nil {
		t.Fatalf("failed to record trade: %v", err)
	}

	if !engine.HasPartition("trading.trades_y2026m09") {
		t.Errorf("expected partition trading.trades_y2026m09 to be registered")
	}
}

func TestFeesAndTaxCompliance(t *testing.T) {
	engine := NewPostgresFinancialEngine()

	// 1. Fee breakdown sum invariant
	fee := &TransactionFee{
		FeeID:              "fee-1",
		TradeID:            "trade-1",
		ISIN:               "INE002A01018",
		TurnoverAmount:     1000000,
		TotalFeeINR:        0, // 0.00% at launch
		TreasuryPortionINR: 0,
		CoreSGFPortionINR:  0,
		IPFPortionINR:      0,
	}
	if err := engine.RecordTransactionFee(fee); err != nil {
		t.Fatalf("expected valid zero fee, got %v", err)
	}

	invalidFee := &TransactionFee{
		FeeID:              "fee-2",
		TotalFeeINR:        1000,
		TreasuryPortionINR: 500,
		CoreSGFPortionINR:  300,
		IPFPortionINR:      100, // Sums to 900 != 1000
	}
	if err := engine.RecordTransactionFee(invalidFee); err == nil {
		t.Fatal("expected error on mismatched fee sum, got nil")
	}

	// 2. Tax Lot STCG (<365 days) vs LTCG (>=365 days)
	stcgLot := &TaxLotDisposal{
		DisposalID:        "tax-1",
		HoldingPeriodDays: 180,
		CostBasis:         100000,
		SaleProceeds:      150000,
	}
	_ = engine.CalculateAndRecordTaxLot(stcgLot)
	if stcgLot.GainType != GainSTCG || stcgLot.RealizedCapitalGain != 50000 {
		t.Errorf("expected STCG with gain 50000, got %v with gain %d", stcgLot.GainType, stcgLot.RealizedCapitalGain)
	}

	ltcgLot := &TaxLotDisposal{
		DisposalID:        "tax-2",
		HoldingPeriodDays: 400,
		CostBasis:         100000,
		SaleProceeds:      220000,
	}
	_ = engine.CalculateAndRecordTaxLot(ltcgLot)
	if ltcgLot.GainType != GainLTCG || ltcgLot.RealizedCapitalGain != 120000 {
		t.Errorf("expected LTCG with gain 120000, got %v with gain %d", ltcgLot.GainType, ltcgLot.RealizedCapitalGain)
	}
}
