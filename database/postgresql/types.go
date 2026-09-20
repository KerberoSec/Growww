package postgresql

import (
	"errors"
	"time"
)

// Standard institutional error invariants.
var (
	ErrAccountNotFound           = errors.New("account not found")
	ErrNegativeBalance           = errors.New("balance cannot be negative")
	ErrLockedExceedsBalance      = errors.New("locked balance cannot exceed total balance")
	ErrDoubleEntryUnbalanced     = errors.New("double-entry journal postings must balance to exactly zero")
	ErrInsufficientAvailableFunds= errors.New("insufficient available balance")
	ErrZeroAmountPosting         = errors.New("posting amount cannot be zero")
	ErrImmutableRecordViolation  = errors.New("mutation or deletion forbidden on immutable financial record")
	ErrCustodyOverissuance       = errors.New("minted tokens exceed physical shares held (1:1 custody violation)")
	ErrTenantAccessViolation     = errors.New("cross-tenant access violation under row-level security policy")
	ErrInvalidOrderStatus        = errors.New("invalid order status transition")
	ErrInvalidQuantity           = errors.New("fractional quantity must be strictly positive")
	ErrInvalidPrice              = errors.New("price must be strictly positive")
	ErrPartitionRangeNotFound    = errors.New("no matching partition range found for execution timestamp")
)

type EntityType string

const (
	EntityDomesticRegulated EntityType = "DOMESTIC_RE"
	EntityGiftCityGateway   EntityType = "GIFT_CITY_GW"
)

type UserStatus string

const (
	UserStatusPendingKYC UserStatus = "PENDING_KYC"
	UserStatusActive     UserStatus = "ACTIVE"
	UserStatusFrozen     UserStatus = "FROZEN"
	UserStatusClosed     UserStatus = "CLOSED"
)

type AccountType string

const (
	AccountUserWallet       AccountType = "USER_WALLET"
	AccountSettlementEscrow AccountType = "SETTLEMENT_ESCROW"
	AccountFeeRevenue       AccountType = "FEE_REVENUE"
	AccountCustodyReserve   AccountType = "CUSTODY_RESERVE"
	AccountTreasury         AccountType = "TREASURY"
	AccountCoreSGF          AccountType = "CORE_SGF"
	AccountIPF              AccountType = "IPF"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusActive          OrderStatus = "ACTIVE"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

type SettlementStatus string

const (
	SettlementPending SettlementStatus = "PENDING"
	SettlementSettled SettlementStatus = "SETTLED_DVP"
	SettlementFailed  SettlementStatus = "FAILED"
)

type ReserveStatus string

const (
	ReserveBalanced    ReserveStatus = "BALANCED"
	ReserveDiscrepancy ReserveStatus = "DISCREPANCY"
	ReserveRebalancing ReserveStatus = "REBALANCING"
)

type GainType string

const (
	GainSTCG GainType = "STCG"
	GainLTCG GainType = "LTCG"
)

// User represents an investor profile in identity.users.
type User struct {
	UserID            string     `json:"user_id"`
	EntityType        EntityType `json:"entity_type"`
	BlockchainAddress string     `json:"blockchain_address"`
	Status            UserStatus `json:"status"`
	KYCLevel          string     `json:"kyc_level"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Account represents a double-entry ledger account.
type Account struct {
	AccountID     string      `json:"account_id"`
	UserID        string      `json:"user_id"`
	Currency      string      `json:"currency"`
	AccountType   AccountType `json:"account_type"`
	Balance       int64       `json:"balance"`        // scaled by 10,000 (4 decimals, e.g. 10000 = 1.0000 INR)
	LockedBalance int64       `json:"locked_balance"` // scaled by 10,000
	CreatedAt     time.Time   `json:"created_at"`
}

// AvailableBalance returns balance - locked_balance.
func (a *Account) AvailableBalance() int64 {
	return a.Balance - a.LockedBalance
}

// JournalEntry is the header record for an immutable transaction.
type JournalEntry struct {
	EntryID       string    `json:"entry_id"`
	ReferenceID   string    `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// Posting represents an atomic leg of a double-entry journal entry.
type Posting struct {
	PostingID      string    `json:"posting_id"`
	EntryID        string    `json:"entry_id"`
	EntryCreatedAt time.Time `json:"entry_created_at"`
	AccountID      string    `json:"account_id"`
	Amount         int64     `json:"amount"`        // Positive: credit, Negative: debit (scaled by 10,000)
	BalanceAfter   int64     `json:"balance_after"` // Scaled by 10,000
}

// Order represents an order in trading.orders.
type Order struct {
	OrderID        string      `json:"order_id"`
	UserID         string      `json:"user_id"`
	ISIN           string      `json:"isin"`
	Symbol         string      `json:"symbol"`
	Side           OrderSide   `json:"side"`
	OrderType      string      `json:"order_type"`
	Price          int64       `json:"price"`           // Scaled by 10,000
	Quantity       int64       `json:"quantity"`        // Scaled by 1,000,000 (6 decimals)
	FilledQuantity int64       `json:"filled_quantity"` // Scaled by 1,000,000
	Status         OrderStatus `json:"status"`
	TimeInForce    string      `json:"time_in_force"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// Trade represents an executed trade in trading.trades.
type Trade struct {
	TradeID          string           `json:"trade_id"`
	BuyerOrderID     string           `json:"buyer_order_id"`
	SellerOrderID    string           `json:"seller_order_id"`
	ISIN             string           `json:"isin"`
	Symbol           string           `json:"symbol"`
	TokenAddress     string           `json:"token_address"`
	FractionalUnits  int64            `json:"fractional_units"` // Scaled by 1,000,000 (6 decimals)
	PricePerUnit     int64            `json:"price_per_unit"`   // Scaled by 10,000 (4 decimals)
	GrossAmount      int64            `json:"gross_amount"`     // Scaled by 10,000
	FeeAmount        int64            `json:"fee_amount"`       // Scaled by 10,000
	SettlementStatus SettlementStatus `json:"settlement_status"`
	OnChainTxHash    string           `json:"on_chain_tx_hash"`
	OnChainBlockNum  uint64           `json:"on_chain_block_num"`
	ExecutedAt       time.Time        `json:"executed_at"`
}

// CustodyAllocation represents physical demat share backing in custody.allocations.
type CustodyAllocation struct {
	AllocationID       string        `json:"allocation_id"`
	ISIN               string        `json:"isin"`
	Depository         string        `json:"depository"` // NSDL, CDSL
	DematAccountNo     string        `json:"demat_account_no"`
	PhysicalSharesHeld int64         `json:"physical_shares_held"` // Scaled by 1,000,000
	TokensMinted       int64         `json:"tokens_minted"`        // Scaled by 1,000,000
	ReserveStatus      ReserveStatus `json:"reserve_status"`
	MerkleRootHash     string        `json:"merkle_root_hash"`
	LastReconciledAt   time.Time     `json:"last_reconciled_at"`
}

// TransactionFee represents fee breakdown in fees.transaction_fees.
type TransactionFee struct {
	FeeID              string    `json:"fee_id"`
	UserID             string    `json:"user_id"`
	TradeID            string    `json:"trade_id"`
	ISIN               string    `json:"isin"`
	TurnoverAmount     int64     `json:"turnover_amount"`     // Scaled by 10,000
	FeeRatePPM         int64     `json:"fee_rate_ppm"`        // Parts per million (0 = 0.00%)
	TotalFeeINR        int64     `json:"total_fee_inr"`       // Scaled by 10,000
	TreasuryPortionINR int64     `json:"treasury_portion_inr"`// Scaled by 10,000
	CoreSGFPortionINR  int64     `json:"core_sgf_portion_inr"`// Scaled by 10,000
	IPFPortionINR      int64     `json:"ipf_portion_inr"`     // Scaled by 10,000
	AssessedAt         time.Time `json:"assessed_at"`
}

// TaxLotDisposal represents Section 111A/112A FIFO tax compliance tracking.
type TaxLotDisposal struct {
	DisposalID          string    `json:"disposal_id"`
	UserID              string    `json:"user_id"`
	TradeID             string    `json:"trade_id"`
	ISIN                string    `json:"isin"`
	CostBasis           int64     `json:"cost_basis"`            // Scaled by 10,000
	SaleProceeds        int64     `json:"sale_proceeds"`         // Scaled by 10,000
	RealizedCapitalGain int64     `json:"realized_capital_gain"` // Scaled by 10,000
	HoldingPeriodDays   int       `json:"holding_period_days"`
	GainType            GainType  `json:"gain_type"` // STCG (< 365 days) or LTCG (>= 365 days)
	AssessedAt          time.Time `json:"assessed_at"`
}

// RLSContext defines the current tenant execution context.
type RLSContext struct {
	ActiveEntity EntityType
	BypassRLS    bool
}
