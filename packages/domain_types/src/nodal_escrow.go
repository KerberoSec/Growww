package src

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrMakerCheckerConflict  = errors.New("nodal_escrow: checker cannot be the same principal as maker")
	ErrBatchAlreadyApproved  = errors.New("nodal_escrow: payout batch is already approved or executed")
	ErrInsufficientNodalFund = errors.New("nodal_escrow: insufficient nodal escrow funds after statutory haircut")
	ErrReconciliationMismatch = errors.New("nodal_escrow: bank balance does not match internal nodal ledger")
)

type BatchStatus string

const (
	BatchPendingApproval BatchStatus = "PENDING_APPROVAL"
	BatchApproved        BatchStatus = "APPROVED"
	BatchExecuted        BatchStatus = "EXECUTED"
	BatchRejected        BatchStatus = "REJECTED"
)

// PayoutTransfer represents an individual beneficiary payout instruction.
type PayoutTransfer struct {
	TransferID       string `json:"transfer_id"`
	BeneficiaryPAN   string `json:"beneficiary_pan"`
	BankAccountNum   string `json:"bank_account_num"`
	IFSC             string `json:"ifsc"`
	AmountINR_Paise  uint64 `json:"amount_inr_paise"`
}

// NodalPayoutBatch represents a maker-checker authorized bank payout batch.
type NodalPayoutBatch struct {
	BatchID          string           `json:"batch_id"`
	MakerID          string           `json:"maker_id"`
	CheckerID        string           `json:"checker_id"`
	Status           BatchStatus      `json:"status"`
	TotalAmountPaise uint64           `json:"total_amount_paise"`
	Transfers        []PayoutTransfer `json:"transfers"`
	CreatedAt        time.Time        `json:"created_at"`
	ApprovedAt       time.Time        `json:"approved_at"`
}

// NodalEscrowManager governs the RBI-compliant nodal escrow settlement account.
type NodalEscrowManager struct {
	mu                   sync.RWMutex
	escrowBankBalance    uint64 // in Paise (1 INR = 100 Paise)
	statutoryHaircutBps  uint32 // Basis points, e.g. 200 bps = 2%
	batches              map[string]*NodalPayoutBatch
}

// NewNodalEscrowManager creates a new nodal escrow manager.
func NewNodalEscrowManager(initialBankBalance uint64, statutoryHaircutBps uint32) *NodalEscrowManager {
	return &NodalEscrowManager{
		escrowBankBalance:   initialBankBalance,
		statutoryHaircutBps: statutoryHaircutBps,
		batches:             make(map[string]*NodalPayoutBatch),
	}
}

// CreatePayoutBatch initiates a new batch under PENDING_APPROVAL by the maker.
func (m *NodalEscrowManager) CreatePayoutBatch(batchID, makerID string, transfers []PayoutTransfer) (*NodalPayoutBatch, error) {
	if batchID == "" || makerID == "" || len(transfers) == 0 {
		return nil, errors.New("invalid batch parameters")
	}

	var total uint64
	for _, t := range transfers {
		if t.AmountINR_Paise == 0 || t.BeneficiaryPAN == "" {
			return nil, errors.New("invalid transfer entry")
		}
		total += t.AmountINR_Paise
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check available liquidity after statutory haircut
	haircut := (m.escrowBankBalance * uint64(m.statutoryHaircutBps)) / 10000
	availableLiquidity := m.escrowBankBalance - haircut
	if total > availableLiquidity {
		return nil, fmt.Errorf("%w: requested %d, available %d after haircut %d", ErrInsufficientNodalFund, total, availableLiquidity, haircut)
	}

	batch := &NodalPayoutBatch{
		BatchID:          batchID,
		MakerID:          makerID,
		Status:           BatchPendingApproval,
		TotalAmountPaise: total,
		Transfers:        transfers,
		CreatedAt:        time.Now().UTC(),
	}

	m.batches[batchID] = batch
	return batch, nil
}

// ApprovePayoutBatch enforces maker-checker segregation of duties.
func (m *NodalEscrowManager) ApprovePayoutBatch(batchID, checkerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	batch, exists := m.batches[batchID]
	if !exists {
		return errors.New("batch not found")
	}

	if batch.Status != BatchPendingApproval {
		return ErrBatchAlreadyApproved
	}

	// Dual-control Maker-Checker requirement
	if batch.MakerID == checkerID {
		return fmt.Errorf("%w: checker '%s' cannot match maker '%s'", ErrMakerCheckerConflict, checkerID, batch.MakerID)
	}

	batch.CheckerID = checkerID
	batch.Status = BatchApproved
	batch.ApprovedAt = time.Now().UTC()
	return nil
}

// ExecuteBatch marks the batch as executed and decrements bank balance.
func (m *NodalEscrowManager) ExecuteBatch(batchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	batch, exists := m.batches[batchID]
	if !exists {
		return errors.New("batch not found")
	}

	if batch.Status != BatchApproved {
		return errors.New("cannot execute batch that is not approved")
	}

	if m.escrowBankBalance < batch.TotalAmountPaise {
		return ErrInsufficientNodalFund
	}

	m.escrowBankBalance -= batch.TotalAmountPaise
	batch.Status = BatchExecuted
	return nil
}

// ReconcileBankReport verifies external bank ledger against internal balance.
func (m *NodalEscrowManager) ReconcileBankReport(reportedBankBalance uint64) (bool, uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.escrowBankBalance != reportedBankBalance {
		var diff uint64
		if m.escrowBankBalance > reportedBankBalance {
			diff = m.escrowBankBalance - reportedBankBalance
		} else {
			diff = reportedBankBalance - m.escrowBankBalance
		}
		return false, diff, fmt.Errorf("%w: internal=%d reported=%d diff=%d", ErrReconciliationMismatch, m.escrowBankBalance, reportedBankBalance, diff)
	}

	return true, 0, nil
}
