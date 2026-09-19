package main

import (
	"math"
	"testing"
)

func TestCurrentYield(t *testing.T) {
	b := Bond{FaceValue: 1000, CouponRate: 0.07, MarketPrice: 950, YearsToMaturity: 10, Frequency: 2}
	cy := b.CurrentYield()
	expected := 70.0 / 950.0
	if math.Abs(cy-expected) > 0.0001 {
		t.Errorf("current yield: got %f, want %f", cy, expected)
	}
}

func TestYTMConverges(t *testing.T) {
	b := Bond{FaceValue: 1000, CouponRate: 0.08, MarketPrice: 1000, YearsToMaturity: 5, Frequency: 2}
	ytm, err := b.YieldToMaturity()
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(ytm-0.08) > 0.001 {
		t.Errorf("par bond YTM should equal coupon: got %f", ytm)
	}
}

func TestDiscountBondYTM(t *testing.T) {
	b := Bond{FaceValue: 1000, CouponRate: 0.06, MarketPrice: 950, YearsToMaturity: 5, Frequency: 2}
	ytm, _ := b.YieldToMaturity()
	if ytm <= 0.06 {
		t.Errorf("discount bond YTM should exceed coupon: got %f", ytm)
	}
}

func TestModifiedDuration(t *testing.T) {
	b := Bond{FaceValue: 1000, CouponRate: 0.07, MarketPrice: 1000, YearsToMaturity: 10, Frequency: 2}
	md := b.ModifiedDuration()
	if md <= 0 || md > 15 {
		t.Errorf("modified duration out of range: %f", md)
	}
}
