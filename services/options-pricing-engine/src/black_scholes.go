package main

import (
	"errors"
	"math"
)

// OptionType represents European option style (Call or Put).
type OptionType string

const (
	OptionCall OptionType = "CALL"
	OptionPut  OptionType = "PUT"
)

// Greeks encapsulates first-order, second-order, and cross-sensitivities.
type Greeks struct {
	Delta float64 `json:"delta"` // ∂V/∂S
	Gamma float64 `json:"gamma"` // ∂²V/∂S²
	Vega  float64 `json:"vega"`  // ∂V/∂σ (per 1.0 unit volatility)
	Vega1Pct float64 `json:"vega_1pct"` // ∂V/∂σ per 1% move in vol (Vega / 100)
	Theta float64 `json:"theta"` // ∂V/∂t (annualized)
	ThetaPerDay float64 `json:"theta_per_day"` // ∂V/∂t per calendar day (Theta / 365)
	Rho   float64 `json:"rho"`   // ∂V/∂r (annualized)
	Rho1Pct float64 `json:"rho_1pct"` // ∂V/∂r per 1% move in interest rate (Rho / 100)
	Vanna float64 `json:"vanna"` // ∂²V/∂S∂σ
	Volga float64 `json:"volga"` // ∂²V/∂σ² (Vomma)
	Charm float64 `json:"charm"` // ∂Δ/∂t (delta decay per year)
	Speed float64 `json:"speed"` // ∂³V/∂S³
}

// OptionPricingResult contains the calculated theoretical price and full Greeks.
type OptionPricingResult struct {
	OptionType    OptionType `json:"option_type"`
	Spot          float64    `json:"spot"`
	Strike        float64    `json:"strike"`
	TimeToExpiry  float64    `json:"time_to_expiry_years"`
	RiskFreeRate  float64    `json:"risk_free_rate"`
	DividendYield float64    `json:"dividend_yield"`
	Volatility    float64    `json:"volatility"`
	Price         float64    `json:"price"`
	Intrinsic     float64    `json:"intrinsic_value"`
	TimeValue     float64    `json:"time_value"`
	Greeks        Greeks     `json:"greeks"`
}

// BlackScholesEngine provides high-performance pricing and Greek calculations.
type BlackScholesEngine struct {
	DefaultRiskFreeRate  float64
	DefaultDividendYield float64
}

// NewBlackScholesEngine initializes a BlackScholesEngine with default rates.
func NewBlackScholesEngine(defaultR, defaultQ float64) *BlackScholesEngine {
	return &BlackScholesEngine{
		DefaultRiskFreeRate:  defaultR,
		DefaultDividendYield: defaultQ,
	}
}

// NormCDF computes the standard normal cumulative distribution function
// using math.Erf for machine-level precision (~1e-16).
func NormCDF(x float64) float64 {
	if math.IsNaN(x) {
		return math.NaN()
	}
	if x > 10.0 {
		return 1.0
	}
	if x < -10.0 {
		return 0.0
	}
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt2))
}

// NormPDF computes the standard normal probability density function.
func NormPDF(x float64) float64 {
	if math.IsNaN(x) {
		return math.NaN()
	}
	const invSqrt2Pi = 0.39894228040143267793994605993438 // 1 / sqrt(2 * pi)
	return invSqrt2Pi * math.Exp(-0.5*x*x)
}

// CalculateD1D2 computes the Black-Scholes d1 and d2 parameters.
func CalculateD1D2(spot, strike, t, r, q, v float64) (d1, d2 float64, err error) {
	if spot <= 0 || strike <= 0 {
		return 0, 0, errors.New("spot and strike must be strictly positive")
	}
	if t <= 0 {
		return 0, 0, errors.New("time to expiry must be strictly positive")
	}
	if v <= 0 {
		return 0, 0, errors.New("volatility must be strictly positive")
	}

	sqrtT := math.Sqrt(t)
	d1 = (math.Log(spot/strike) + (r-q+0.5*v*v)*t) / (v * sqrtT)
	d2 = d1 - v*sqrtT
	return d1, d2, nil
}

// Price calculates theoretical price and comprehensive Greeks for European options.
func (e *BlackScholesEngine) Price(
	optionType OptionType,
	spot float64,
	strike float64,
	timeToExpiry float64,
	volatility float64,
	riskFreeRate float64,
	dividendYield float64,
) (*OptionPricingResult, error) {
	if spot <= 0 {
		return nil, errors.New("spot price must be positive")
	}
	if strike <= 0 {
		return nil, errors.New("strike price must be positive")
	}
	if timeToExpiry < 0 {
		return nil, errors.New("time to expiry cannot be negative")
	}
	if volatility < 0 {
		return nil, errors.New("volatility cannot be negative")
	}

	r := riskFreeRate
	q := dividendYield

	// 1. Handling Expiry or Zero Volatility Boundary Conditions
	if timeToExpiry <= 1e-12 || volatility <= 1e-12 {
		intrinsic := 0.0
		var delta float64
		if optionType == OptionCall {
			intrinsic = math.Max(0, spot-strike)
			if spot > strike {
				delta = 1.0
			} else if spot == strike {
				delta = 0.5
			} else {
				delta = 0.0
			}
		} else {
			intrinsic = math.Max(0, strike-spot)
			if spot < strike {
				delta = -1.0
			} else if spot == strike {
				delta = -0.5
			} else {
				delta = 0.0
			}
		}

		return &OptionPricingResult{
			OptionType:    optionType,
			Spot:          spot,
			Strike:        strike,
			TimeToExpiry:  timeToExpiry,
			RiskFreeRate:  r,
			DividendYield: q,
			Volatility:    volatility,
			Price:         intrinsic,
			Intrinsic:     intrinsic,
			TimeValue:     0.0,
			Greeks: Greeks{
				Delta: delta,
			},
		}, nil
	}

	d1, d2, err := CalculateD1D2(spot, strike, timeToExpiry, r, q, volatility)
	if err != nil {
		return nil, err
	}

	sqrtT := math.Sqrt(timeToExpiry)
	dfR := math.Exp(-r * timeToExpiry) // discount factor interest rate
	dfQ := math.Exp(-q * timeToExpiry) // discount factor dividend yield

	nd1 := NormCDF(d1)
	nd2 := NormCDF(d2)
	nNegD1 := NormCDF(-d1)
	nNegD2 := NormCDF(-d2)
	pdfD1 := NormPDF(d1)

	var price float64
	var delta float64
	var theta float64
	var rho float64
	var intrinsic float64

	if optionType == OptionCall {
		price = spot*dfQ*nd1 - strike*dfR*nd2
		intrinsic = math.Max(0, spot-strike)
		delta = dfQ * nd1
		theta = -((spot*dfQ*pdfD1*volatility)/(2.0*sqrtT)) - r*strike*dfR*nd2 + q*spot*dfQ*nd1
		rho = strike * timeToExpiry * dfR * nd2
	} else {
		price = strike*dfR*nNegD2 - spot*dfQ*nNegD1
		intrinsic = math.Max(0, strike-spot)
		delta = -dfQ * nNegD1
		theta = -((spot*dfQ*pdfD1*volatility)/(2.0*sqrtT)) + r*strike*dfR*nNegD2 - q*spot*dfQ*nNegD1
		rho = -strike * timeToExpiry * dfR * nNegD2
	}

	// Protect against micro numerical precision floor
	if price < 0.0 {
		price = 0.0
	}
	timeValue := math.Max(0.0, price-intrinsic)

	// Common Greeks for Call and Put
	gamma := (dfQ * pdfD1) / (spot * volatility * sqrtT)
	vega := spot * dfQ * pdfD1 * sqrtT

	// Higher-order and Cross Greeks
	vanna := -dfQ * pdfD1 * (d2 / volatility)
	volga := vega * (d1 * d2 / volatility)
	charm := q*dfQ*nd1 - dfQ*pdfD1*(2.0*(r-q)*timeToExpiry-d2*volatility*sqrtT)/(2.0*timeToExpiry*volatility*sqrtT)
	if optionType == OptionPut {
		charm = -q*dfQ*nNegD1 - dfQ*pdfD1*(2.0*(r-q)*timeToExpiry-d2*volatility*sqrtT)/(2.0*timeToExpiry*volatility*sqrtT)
	}
	speed := -gamma / spot * (d1/(volatility*sqrtT) + 1.0)

	return &OptionPricingResult{
		OptionType:    optionType,
		Spot:          spot,
		Strike:        strike,
		TimeToExpiry:  timeToExpiry,
		RiskFreeRate:  r,
		DividendYield: q,
		Volatility:    volatility,
		Price:         price,
		Intrinsic:     intrinsic,
		TimeValue:     timeValue,
		Greeks: Greeks{
			Delta:       delta,
			Gamma:       gamma,
			Vega:        vega,
			Vega1Pct:    vega / 100.0,
			Theta:       theta,
			ThetaPerDay: theta / 365.0,
			Rho:         rho,
			Rho1Pct:     rho / 100.0,
			Vanna:       vanna,
			Volga:       volga,
			Charm:       charm,
			Speed:       speed,
		},
	}, nil
}

// ValidatePutCallParity verifies the put-call parity relationship:
// C - P = S * e^(-q*T) - K * e^(-r*T)
func ValidatePutCallParity(
	callPrice, putPrice float64,
	spot, strike, t, r, q float64,
	tolerance float64,
) (bool, float64) {
	lhs := callPrice - putPrice
	rhs := spot*math.Exp(-q*t) - strike*math.Exp(-r*t)
	diff := math.Abs(lhs - rhs)
	return diff <= tolerance, diff
}

// OptionChainLeg represents a priced strike leg in an option chain.
type OptionChainLeg struct {
	Strike       float64              `json:"strike"`
	CallResult   *OptionPricingResult `json:"call"`
	PutResult    *OptionPricingResult `json:"put"`
	Moneyness    float64              `json:"moneyness"` // Spot / Strike
	LogMoneyness float64              `json:"log_moneyness"` // ln(Strike / Spot)
}

// BuildOptionChain generates a full strike ladder around the current spot.
func (e *BlackScholesEngine) BuildOptionChain(
	spot float64,
	timeToExpiry float64,
	volatility float64,
	riskFreeRate float64,
	dividendYield float64,
	strikeStep float64,
	numStrikesEachSide int,
) ([]OptionChainLeg, error) {
	if strikeStep <= 0 {
		return nil, errors.New("strike step must be positive")
	}
	if spot <= 0 {
		return nil, errors.New("spot price must be positive")
	}

	atmStrike := math.Round(spot/strikeStep) * strikeStep
	legs := make([]OptionChainLeg, 0, 2*numStrikesEachSide+1)

	for i := -numStrikesEachSide; i <= numStrikesEachSide; i++ {
		k := atmStrike + float64(i)*strikeStep
		if k <= 0 {
			continue
		}

		callRes, err := e.Price(OptionCall, spot, k, timeToExpiry, volatility, riskFreeRate, dividendYield)
		if err != nil {
			return nil, err
		}

		putRes, err := e.Price(OptionPut, spot, k, timeToExpiry, volatility, riskFreeRate, dividendYield)
		if err != nil {
			return nil, err
		}

		legs = append(legs, OptionChainLeg{
			Strike:       k,
			CallResult:   callRes,
			PutResult:    putRes,
			Moneyness:    spot / k,
			LogMoneyness: math.Log(k / spot),
		})
	}

	return legs, nil
}
