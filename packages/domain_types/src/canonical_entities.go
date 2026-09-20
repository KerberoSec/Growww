package src

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// Value Objects
// ---------------------------------------------------------------------------

// Money represents an exact monetary amount without floating-point errors
type Money struct {
	Currency string `json:"currency"` // "INR", "USD", "USDT"
	Units    int64  `json:"units"`
	Nanos    int32  `json:"nanos"` // -999,999,999 to +999,999,999
}

func NewMoney(currency string, units int64, nanos int32) (*Money, error) {
	if currency == "" {
		return nil, errors.New("currency is required")
	}
	if nanos < -999999999 || nanos > 999999999 {
		return nil, errors.New("nanos out of range (-999,999,999 to 999,999,999)")
	}
	return &Money{
		Currency: strings.ToUpper(currency),
		Units:    units,
		Nanos:    nanos,
	}, nil
}

func (m *Money) Add(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, fmt.Errorf("cannot add different currencies %s and %s", m.Currency, other.Currency)
	}

	totalNanos := int64(m.Nanos) + int64(other.Nanos)
	unitsCarry := totalNanos / 1000000000
	remNanos := int32(totalNanos % 1000000000)

	return &Money{
		Currency: m.Currency,
		Units:    m.Units + other.Units + unitsCarry,
		Nanos:    remNanos,
	}, nil
}

// FractionalShare provides 6-decimal fixed point precision for equity shares
type FractionalShare struct {
	MicroShares int64 `json:"micro_shares"` // 1.000000 share = 1,000,000 micro-shares
}

func FromFloatShares(shares float64) FractionalShare {
	return FractionalShare{MicroShares: int64(shares * 1000000)}
}

func (f FractionalShare) ToFloat() float64 {
	return float64(f.MicroShares) / 1000000.0
}

func (f FractionalShare) Format() string {
	whole := f.MicroShares / 1000000
	fraction := f.MicroShares % 1000000
	if fraction < 0 {
		fraction = -fraction
	}
	return fmt.Sprintf("%d.%06d", whole, fraction)
}

// ISIN represents a 12-character International Securities Identification Number
var isinRegex = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)

type ISIN struct {
	Value string `json:"value"`
}

func NewISIN(val string) (*ISIN, error) {
	clean := strings.ToUpper(strings.TrimSpace(val))
	if len(clean) != 12 || !isinRegex.MatchString(clean) {
		return nil, fmt.Errorf("invalid ISIN format: '%s' (expected 2 letters, 9 alphanumeric, 1 digit)", val)
	}
	return &ISIN{Value: clean}, nil
}

func (i *ISIN) CountryCode() string {
	return i.Value[:2]
}

// ---------------------------------------------------------------------------
// State Machines
// ---------------------------------------------------------------------------

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "PENDING"
	OrderStatusRouted          OrderStatus = "ROUTED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

func ValidateOrderTransition(current, next OrderStatus) error {
	switch current {
	case OrderStatusPending:
		if next == OrderStatusRouted || next == OrderStatusRejected || next == OrderStatusCancelled {
			return nil
		}
	case OrderStatusRouted:
		if next == OrderStatusPartiallyFilled || next == OrderStatusFilled || next == OrderStatusCancelled {
			return nil
		}
	case OrderStatusPartiallyFilled:
		if next == OrderStatusPartiallyFilled || next == OrderStatusFilled || next == OrderStatusCancelled {
			return nil
		}
	case OrderStatusFilled, OrderStatusCancelled, OrderStatusRejected:
		return fmt.Errorf("terminal order state '%s' cannot transition to '%s'", current, next)
	}
	return fmt.Errorf("illegal order state transition from '%s' to '%s'", current, next)
}

type DvPStatus string

const (
	DvPStatusCreated      DvPStatus = "CREATED"
	DvPStatusEscrowLocked DvPStatus = "ESCROW_LOCKED"
	DvPStatusCommitted    DvPStatus = "COMMITTED"
	DvPStatusSettled      DvPStatus = "SETTLED"
	DvPStatusFailed       DvPStatus = "FAILED"
)

func ValidateDvPTransition(current, next DvPStatus) error {
	switch current {
	case DvPStatusCreated:
		if next == DvPStatusEscrowLocked || next == DvPStatusFailed {
			return nil
		}
	case DvPStatusEscrowLocked:
		if next == DvPStatusCommitted || next == DvPStatusFailed {
			return nil
		}
	case DvPStatusCommitted:
		if next == DvPStatusSettled || next == DvPStatusFailed {
			return nil
		}
	case DvPStatusSettled, DvPStatusFailed:
		return fmt.Errorf("terminal DvP state '%s' cannot transition to '%s'", current, next)
	}
	return fmt.Errorf("illegal DvP state transition from '%s' to '%s'", current, next)
}
