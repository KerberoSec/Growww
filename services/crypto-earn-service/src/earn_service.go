package src

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// StakingProductType represents flexible vs fixed terms
type StakingProductType string

const (
	ProductFlexible StakingProductType = "FLEXIBLE"
	ProductFixed30  StakingProductType = "FIXED_30"
	ProductFixed60  StakingProductType = "FIXED_60"
	ProductFixed90  StakingProductType = "FIXED_90"
)

const (
	SecondsPerYear   = 365 * 86400
	BpsDenominator   = 10000
	RiskReserveCutBp = 1000 // 10% of gross yield
)

// EarnProduct defines an investment tier
type EarnProduct struct {
	ProductID         string             `json:"product_id"`
	AssetSymbol       string             `json:"asset_symbol"`
	ProductType       StakingProductType `json:"product_type"`
	DurationDays      int                `json:"duration_days"`
	APRBps            int64              `json:"apr_bps"` // e.g., 600 = 6.00%
	MinDepositE8      int64              `json:"min_deposit_e8"`
	MaxCapacityE8     int64              `json:"max_capacity_e8"`
	TotalDepositedE8  int64              `json:"total_deposited_e8"`
	Active            bool               `json:"active"`
}

// UserSubscription represents an active stake
type UserSubscription struct {
	SubscriptionID    string    `json:"subscription_id"`
	UserID            string    `json:"user_id"`
	ProductID         string    `json:"product_id"`
	PrincipalE8       int64     `json:"principal_e8"`
	StartTime         time.Time `json:"start_time"`
	LastHarvestTime   time.Time `json:"last_harvest_time"`
	MaturityTime      time.Time `json:"maturity_time"`
	AccruedInterestE8 int64     `json:"accrued_interest_e8"`
	Status            string    `json:"status"` // ACTIVE, REDEEMED
}

// EarnService coordinates wealth management products and yield distribution
type EarnService struct {
	mu                   sync.RWMutex
	products             map[string]*EarnProduct
	subscriptions        map[string]*UserSubscription
	userSubIDs           map[string][]string // userID -> []subscriptionID
	accumulatedReserveE8 map[string]int64    // symbol -> total reserve fund
}

func NewEarnService() *EarnService {
	return &EarnService{
		products:             make(map[string]*EarnProduct),
		subscriptions:        make(map[string]*UserSubscription),
		userSubIDs:           make(map[string][]string),
		accumulatedReserveE8: make(map[string]int64),
	}
}

// CreateProduct registers a new savings/staking offering
func (s *EarnService) CreateProduct(productID, symbol string, pType StakingProductType, aprBps int64, minDepositE8, maxCapacityE8 int64) (*EarnProduct, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if productID == "" || symbol == "" {
		return nil, errors.New("invalid product parameters")
	}
	if aprBps <= 0 || aprBps > 5000 {
		return nil, errors.New("apr out of bounds (1 - 5000 bps)")
	}

	days := 0
	switch pType {
	case ProductFixed30:
		days = 30
	case ProductFixed60:
		days = 60
	case ProductFixed90:
		days = 90
	case ProductFlexible:
		days = 0
	default:
		return nil, fmt.Errorf("unsupported product type: %s", pType)
	}

	prod := &EarnProduct{
		ProductID:        productID,
		AssetSymbol:      symbol,
		ProductType:      pType,
		DurationDays:     days,
		APRBps:           aprBps,
		MinDepositE8:     minDepositE8,
		MaxCapacityE8:    maxCapacityE8,
		TotalDepositedE8: 0,
		Active:           true,
	}

	s.products[productID] = prod
	return prod, nil
}

// Subscribe deposits assets into an Earn product
func (s *EarnService) Subscribe(userID, productID string, amountE8 int64, now time.Time) (*UserSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prod, exists := s.products[productID]
	if !exists || !prod.Active {
		return nil, errors.New("product unavailable")
	}
	if amountE8 < prod.MinDepositE8 {
		return nil, fmt.Errorf("deposit %d below minimum %d", amountE8, prod.MinDepositE8)
	}
	if prod.TotalDepositedE8+amountE8 > prod.MaxCapacityE8 {
		return nil, errors.New("product capacity reached")
	}

	prod.TotalDepositedE8 += amountE8
	subID := fmt.Sprintf("sub_%s_%d", userID, now.UnixNano())

	maturity := now
	if prod.DurationDays > 0 {
		maturity = now.Add(time.Duration(prod.DurationDays) * 24 * time.Hour)
	}

	sub := &UserSubscription{
		SubscriptionID:    subID,
		UserID:            userID,
		ProductID:         productID,
		PrincipalE8:       amountE8,
		StartTime:         now,
		LastHarvestTime:   now,
		MaturityTime:      maturity,
		AccruedInterestE8: 0,
		Status:            "ACTIVE",
	}

	s.subscriptions[subID] = sub
	s.userSubIDs[userID] = append(s.userSubIDs[userID], subID)
	return sub, nil
}

// CalculateAccruedYield computes gross yield, 10% reserve cut, and net yield
func (s *EarnService) CalculateAccruedYield(subID string, now time.Time) (grossE8, netE8, reserveE8 int64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sub, exists := s.subscriptions[subID]
	if !exists || sub.Status != "ACTIVE" {
		return 0, 0, 0, errors.New("subscription not active")
	}

	prod := s.products[sub.ProductID]
	elapsedSec := now.Sub(sub.LastHarvestTime).Seconds()
	if elapsedSec <= 0 {
		return 0, 0, 0, nil
	}

	// gross = principal * aprBps * elapsedSec / (BpsDenominator * SecondsPerYear)
	gross := float64(sub.PrincipalE8) * float64(prod.APRBps) * elapsedSec / float64(BpsDenominator*SecondsPerYear)
	grossE8 = int64(gross)
	reserveE8 = (grossE8 * RiskReserveCutBp) / BpsDenominator
	netE8 = grossE8 - reserveE8
	return grossE8, netE8, reserveE8, nil
}

// Redeem completes flexible instant withdrawal or fixed-term maturity
func (s *EarnService) Redeem(subID, userID string, now time.Time) (principalReturned, yieldPaid int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, exists := s.subscriptions[subID]
	if !exists || sub.Status != "ACTIVE" {
		return 0, 0, errors.New("subscription not active")
	}
	if sub.UserID != userID {
		return 0, 0, errors.New("unauthorized subscriber")
	}

	prod := s.products[sub.ProductID]
	principal := sub.PrincipalE8
	netYield := int64(0)

	if prod.DurationDays > 0 && now.Before(sub.MaturityTime) {
		// Early fixed redemption penalty: forfeit accrued interest
		netYield = 0
	} else {
		// Eligible for accrued yield
		elapsedSec := now.Sub(sub.LastHarvestTime).Seconds()
		if elapsedSec > 0 {
			gross := float64(sub.PrincipalE8) * float64(prod.APRBps) * elapsedSec / float64(BpsDenominator*SecondsPerYear)
			grossE8 := int64(gross)
			reserveCut := (grossE8 * RiskReserveCutBp) / BpsDenominator
			netYield = grossE8 - reserveCut
			s.accumulatedReserveE8[prod.AssetSymbol] += reserveCut
		}
	}

	sub.Status = "REDEEMED"
	prod.TotalDepositedE8 -= principal

	return principal, netYield, nil
}

// GetReserveBalance returns accumulated insurance reserve for an asset
func (s *EarnService) GetReserveBalance(symbol string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accumulatedReserveE8[symbol]
}
