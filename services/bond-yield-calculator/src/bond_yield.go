package main

import (
	"fmt"
	"math"
)

type Bond struct {
	FaceValue      float64
	CouponRate     float64 // annual as decimal
	MarketPrice    float64
	YearsToMaturity float64
	Frequency      int // coupons per year
}

func (b *Bond) CurrentYield() float64 {
	if b.MarketPrice <= 0 {
		return 0
	}
	return (b.FaceValue * b.CouponRate) / b.MarketPrice
}

func (b *Bond) YieldToMaturity() (float64, error) {
	if b.MarketPrice <= 0 || b.FaceValue <= 0 || b.YearsToMaturity <= 0 {
		return 0, fmt.Errorf("invalid bond parameters")
	}
	coupon := b.FaceValue * b.CouponRate / float64(b.Frequency)
	n := b.YearsToMaturity * float64(b.Frequency)
	guess := b.CouponRate
	for i := 0; i < 200; i++ {
		price := 0.0
		for t := 1.0; t <= n; t++ {
			price += coupon / math.Pow(1+guess/float64(b.Frequency), t)
		}
		price += b.FaceValue / math.Pow(1+guess/float64(b.Frequency), n)
		diff := price - b.MarketPrice
		if math.Abs(diff) < 0.0001 {
			return guess, nil
		}
		dprice := 0.0
		for t := 1.0; t <= n; t++ {
			dprice -= t * coupon / (float64(b.Frequency) * math.Pow(1+guess/float64(b.Frequency), t+1))
		}
		dprice -= n * b.FaceValue / (float64(b.Frequency) * math.Pow(1+guess/float64(b.Frequency), n+1))
		if dprice == 0 {
			break
		}
		guess -= diff / dprice
	}
	return guess, nil
}

func (b *Bond) ModifiedDuration() float64 {
	ytm, _ := b.YieldToMaturity()
	coupon := b.FaceValue * b.CouponRate / float64(b.Frequency)
	n := b.YearsToMaturity * float64(b.Frequency)
	y := ytm / float64(b.Frequency)
	macD := 0.0
	pvTotal := 0.0
	for t := 1.0; t <= n; t++ {
		pv := coupon / math.Pow(1+y, t)
		macD += t * pv / float64(b.Frequency)
		pvTotal += pv
	}
	pvFace := b.FaceValue / math.Pow(1+y, n)
	macD += b.YearsToMaturity * pvFace
	pvTotal += pvFace
	if pvTotal == 0 {
		return 0
	}
	macD /= pvTotal
	return macD / (1 + y)
}
