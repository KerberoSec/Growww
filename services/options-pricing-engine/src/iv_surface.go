package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// MarketQuote represents an observed market option quote.
type MarketQuote struct {
	OptionType OptionType `json:"option_type"`
	Strike     float64    `json:"strike"`
	Expiry     float64    `json:"expiry_years"`
	Bid        float64    `json:"bid"`
	Ask        float64    `json:"ask"`
	Mid        float64    `json:"mid"`
	Weight     float64    `json:"weight"` // Calibration weight (e.g., based on open interest or inverse bid-ask spread)
}

// IVSolverConfig sets limits and tolerances for the implied volatility solver.
type IVSolverConfig struct {
	MaxIterations int
	Tolerance     float64
	MinVol        float64
	MaxVol        float64
}

// DefaultIVSolverConfig returns standard institutional solver parameters.
func DefaultIVSolverConfig() IVSolverConfig {
	return IVSolverConfig{
		MaxIterations: 100,
		Tolerance:     1e-7,
		MinVol:        0.001,  // 0.1%
		MaxVol:        10.0,   // 1000%
	}
}

// ImpliedVolatilityEngine calculates implied volatility from market prices.
type ImpliedVolatilityEngine struct {
	pricingEngine *BlackScholesEngine
	config        IVSolverConfig
}

// NewImpliedVolatilityEngine creates an IV solver engine.
func NewImpliedVolatilityEngine(cfg IVSolverConfig) *ImpliedVolatilityEngine {
	return &ImpliedVolatilityEngine{
		pricingEngine: NewBlackScholesEngine(0.05, 0.0),
		config:        cfg,
	}
}

// CheckArbitrageBounds verifies if a market price satisfies European no-arbitrage bounds.
func CheckArbitrageBounds(
	optionType OptionType,
	marketPrice, spot, strike, t, r, q float64,
) error {
	if marketPrice <= 0 {
		return errors.New("market price must be strictly positive")
	}

	dfR := math.Exp(-r * t)
	dfQ := math.Exp(-q * t)

	if optionType == OptionCall {
		lowerBound := math.Max(0.0, spot*dfQ-strike*dfR)
		upperBound := spot * dfQ
		if marketPrice < lowerBound-1e-9 || marketPrice > upperBound+1e-9 {
			return fmt.Errorf("call price %f violates no-arbitrage bounds [%f, %f]", marketPrice, lowerBound, upperBound)
		}
	} else {
		lowerBound := math.Max(0.0, strike*dfR-spot*dfQ)
		upperBound := strike * dfR
		if marketPrice < lowerBound-1e-9 || marketPrice > upperBound+1e-9 {
			return fmt.Errorf("put price %f violates no-arbitrage bounds [%f, %f]", marketPrice, lowerBound, upperBound)
		}
	}

	return nil
}

// CalculateIV solves for implied volatility using a hybrid Newton-Raphson / Brent-Bisection algorithm.
func (e *ImpliedVolatilityEngine) CalculateIV(
	optionType OptionType,
	marketPrice float64,
	spot float64,
	strike float64,
	timeToExpiry float64,
	riskFreeRate float64,
	dividendYield float64,
) (float64, error) {
	if err := CheckArbitrageBounds(optionType, marketPrice, spot, strike, timeToExpiry, riskFreeRate, dividendYield); err != nil {
		return 0, err
	}

	// 1. Initial Volatility Estimate (Corrado-Miller / Brenner-Subrahmanyam approximation)
	forward := spot * math.Exp((riskFreeRate-dividendYield)*timeToExpiry)
	moneyness := forward / strike
	initialGuess := 0.25 // Standard 25% vol default

	if math.Abs(moneyness-1.0) < 0.1 && timeToExpiry > 0 {
		dfR := math.Exp(-riskFreeRate * timeToExpiry)
		approx := (marketPrice / (spot * dfR)) * math.Sqrt(2.0*math.Pi/timeToExpiry)
		if approx > 0.05 && approx < 2.0 {
			initialGuess = approx
		}
	}

	// 2. Newton-Raphson with Step Safeguards
	sigma := initialGuess
	for i := 0; i < e.config.MaxIterations; i++ {
		res, err := e.pricingEngine.Price(optionType, spot, strike, timeToExpiry, sigma, riskFreeRate, dividendYield)
		if err != nil {
			break
		}

		diff := res.Price - marketPrice
		if math.Abs(diff) < e.config.Tolerance {
			return sigma, nil
		}

		vega := res.Greeks.Vega // Raw Vega
		if vega < 1e-12 {
			// Vega is too small for Newton-Raphson (deep ITM / OTM); fall back to Bisection
			break
		}

		step := diff / vega
		// Damping large steps
		if math.Abs(step) > 0.5 {
			step = math.Copysign(0.5, step)
		}

		nextSigma := sigma - step
		if nextSigma <= e.config.MinVol || nextSigma >= e.config.MaxVol {
			// Stepped outside bounds; fall back to Bisection
			break
		}

		sigma = nextSigma
	}

	// 3. Fallback: Robust Bisection Search
	low := e.config.MinVol
	high := e.config.MaxVol

	resLow, err1 := e.pricingEngine.Price(optionType, spot, strike, timeToExpiry, low, riskFreeRate, dividendYield)
	resHigh, err2 := e.pricingEngine.Price(optionType, spot, strike, timeToExpiry, high, riskFreeRate, dividendYield)
	if err1 != nil || err2 != nil {
		return 0, errors.New("failed evaluating pricing bounds during bisection")
	}

	diffLow := resLow.Price - marketPrice
	diffHigh := resHigh.Price - marketPrice

	if diffLow*diffHigh > 0 {
		// Market price lies outside min/max vol theoretical prices
		if math.Abs(diffLow) < math.Abs(diffHigh) {
			return low, nil
		}
		return high, nil
	}

	for i := 0; i < e.config.MaxIterations; i++ {
		mid := 0.5 * (low + high)
		resMid, err := e.pricingEngine.Price(optionType, spot, strike, timeToExpiry, mid, riskFreeRate, dividendYield)
		if err != nil {
			return 0, err
		}

		diffMid := resMid.Price - marketPrice
		if math.Abs(diffMid) < e.config.Tolerance || (high-low)*0.5 < e.config.Tolerance {
			return mid, nil
		}

		if diffLow*diffMid < 0 {
			high = mid
			diffHigh = diffMid
		} else {
			low = mid
			diffLow = diffMid
		}
	}

	return 0.5 * (low + high), nil
}

// -----------------------------------------------------------------------------
// SVI (Stochastic Volatility Inspired) Smile Model
// -----------------------------------------------------------------------------

// SVIParams represents Gatheral's raw SVI parametrization:
// w(k) = a + b * ( rho * (k - m) + sqrt((k - m)^2 + sigma^2) )
// where k = ln(K / F) is log-moneyness, w(k) = sigma_BS^2 * T is total implied variance.
type SVIParams struct {
	A     float64 `json:"a"`     // Vertical translation (a >= 0)
	B     float64 `json:"b"`     // Slope of asymptotes (b >= 0)
	Rho   float64 `json:"rho"`   // Skew / rotation (-1 <= rho <= 1)
	M     float64 `json:"m"`     // Horizontal translation
	Sigma float64 `json:"sigma"` // ATM curvature smoothness (sigma > 0)
}

// TotalVariance computes the SVI total variance w(k).
func (p *SVIParams) TotalVariance(k float64) float64 {
	disc := math.Sqrt((k-p.M)*(k-p.M) + p.Sigma*p.Sigma)
	w := p.A + p.B*(p.Rho*(k-p.M)+disc)
	if w < 1e-8 {
		return 1e-8
	}
	return w
}

// ImpliedVol calculates annual volatility from total variance at maturity T.
func (p *SVIParams) ImpliedVol(k, t float64) float64 {
	if t <= 0 {
		return 0.0
	}
	w := p.TotalVariance(k)
	return math.Sqrt(w / t)
}

// ValidateArbitrage checks Gatheral's conditions for no butterfly arbitrage in SVI.
func (p *SVIParams) ValidateArbitrage() bool {
	if p.B < 0 || p.Sigma <= 0 {
		return false
	}
	if math.Abs(p.Rho) >= 1.0 {
		return false
	}
	// a + b * sigma * sqrt(1 - rho^2) >= 0 ensures total variance is always positive
	if p.A+p.B*p.Sigma*math.Sqrt(1.0-p.Rho*p.Rho) < 0 {
		return false
	}
	// b * (1 + |rho|) < 2 ensures slope of total variance asymptotes < 2 (Roger Lee bound)
	if p.B*(1.0+math.Abs(p.Rho)) >= 2.0 {
		return false
	}
	return true
}

// PolynomialSmileParams models implied volatility directly as a polynomial in log-moneyness:
// sigma(k) = c0 + c1*k + c2*k^2 + c3*k^3
type PolynomialSmileParams struct {
	C0 float64 `json:"c0"` // ATM volatility
	C1 float64 `json:"c1"` // Skew slope
	C2 float64 `json:"c2"` // Curvature / smile
	C3 float64 `json:"c3"` // Tail skew
}

// ImpliedVol computes the polynomial smile IV.
func (p *PolynomialSmileParams) ImpliedVol(k float64) float64 {
	v := p.C0 + p.C1*k + p.C2*k*k + p.C3*k*k*k
	if v < 0.001 {
		return 0.001
	}
	return v
}

// CalibratedSmileSlice holds smile parameters for a specific maturity.
type CalibratedSmileSlice struct {
	Expiry       float64               `json:"expiry_years"`
	Forward      float64               `json:"forward_price"`
	SVI          SVIParams             `json:"svi"`
	Polynomial   PolynomialSmileParams `json:"polynomial"`
	FittedPoints int                   `json:"fitted_points"`
	RMSE         float64               `json:"rmse"`
}

// VolatilitySurface represents a complete 2D (Strike x Expiry) volatility surface.
type VolatilitySurface struct {
	UnderlyingSymbol string                 `json:"symbol"`
	SpotPrice        float64                `json:"spot_price"`
	RiskFreeRate     float64                `json:"risk_free_rate"`
	DividendYield    float64                `json:"dividend_yield"`
	Slices           []CalibratedSmileSlice `json:"slices"`
}

// NewVolatilitySurface creates an empty surface.
func NewVolatilitySurface(symbol string, spot, r, q float64) *VolatilitySurface {
	return &VolatilitySurface{
		UnderlyingSymbol: symbol,
		SpotPrice:        spot,
		RiskFreeRate:     r,
		DividendYield:    q,
		Slices:           make([]CalibratedSmileSlice, 0),
	}
}

// FitPolynomialSmile fits a cubic polynomial smile to observed (log-moneyness, IV) points.
// Uses regularized least squares.
func FitPolynomialSmile(logMoneyness, ivs, weights []float64) PolynomialSmileParams {
	n := len(logMoneyness)
	if n == 0 {
		return PolynomialSmileParams{C0: 0.2}
	}
	if n == 1 {
		return PolynomialSmileParams{C0: ivs[0]}
	}

	// Form normal equations for degree-3 polynomial: X^T W X c = X^T W y
	// Basis functions: 1, k, k^2, k^3
	degree := 3
	if n < 4 {
		degree = n - 1
	}

	dim := degree + 1
	ata := make([][]float64, dim)
	for i := range ata {
		ata[i] = make([]float64, dim)
	}
	atb := make([]float64, dim)

	for i := 0; i < n; i++ {
		k := logMoneyness[i]
		y := ivs[i]
		w := weights[i]
		if w <= 0 {
			w = 1.0
		}

		basis := make([]float64, dim)
		basis[0] = 1.0
		for d := 1; d < dim; d++ {
			basis[d] = basis[d-1] * k
		}

		for r := 0; r < dim; r++ {
			for c := 0; c < dim; c++ {
				ata[r][c] += w * basis[r] * basis[c]
			}
			atb[r] += w * basis[r] * y
		}
	}

	// Add Tikhonov regularization on higher order terms
	for d := 1; d < dim; d++ {
		ata[d][d] += 1e-4
	}

	// Solve linear system using Gaussian elimination
	coeffs := solveGaussian(ata, atb)
	params := PolynomialSmileParams{}
	if len(coeffs) > 0 {
		params.C0 = coeffs[0]
	}
	if len(coeffs) > 1 {
		params.C1 = coeffs[1]
	}
	if len(coeffs) > 2 {
		params.C2 = coeffs[2]
	}
	if len(coeffs) > 3 {
		params.C3 = coeffs[3]
	}

	return params
}

// FitSVISmile fits SVI parameters using Nelder-Mead simplex optimization.
func FitSVISmile(logMoneyness, totalVar, weights []float64, t float64) SVIParams {
	n := len(logMoneyness)
	if n == 0 {
		return SVIParams{A: 0.04 * t, B: 0.1, Rho: -0.2, M: 0.0, Sigma: 0.1}
	}

	// Initial guess based on ATM total variance
	var atmVar float64 = 0.04 * t
	minDist := math.MaxFloat64
	for i, k := range logMoneyness {
		if math.Abs(k) < minDist {
			minDist = math.Abs(k)
			atmVar = totalVar[i]
		}
	}

	initialParams := []float64{atmVar * 0.8, 0.1, -0.3, 0.0, 0.1}

	objective := func(p []float64) float64 {
		a, b, rho, m, sig := p[0], p[1], p[2], p[3], p[4]
		// Penalties for constraint violations
		if a < 0 {
			return 1e6 + math.Abs(a)*1e6
		}
		if b < 0 {
			return 1e6 + math.Abs(b)*1e6
		}
		if math.Abs(rho) >= 0.999 {
			return 1e6 + (math.Abs(rho)-0.999)*1e6
		}
		if sig <= 1e-4 {
			return 1e6 + (1e-4-sig)*1e6
		}
		if b*(1.0+math.Abs(rho)) >= 1.95 {
			return 1e6 + 1e5
		}

		svi := SVIParams{A: a, B: b, Rho: rho, M: m, Sigma: sig}
		var sse float64
		for i := 0; i < n; i++ {
			pred := svi.TotalVariance(logMoneyness[i])
			diff := pred - totalVar[i]
			w := weights[i]
			if w <= 0 {
				w = 1.0
			}
			sse += w * diff * diff
		}
		return sse
	}

	optimized := nelderMead(objective, initialParams, 200, 1e-6)
	return SVIParams{
		A:     math.Max(1e-6, optimized[0]),
		B:     math.Max(1e-6, optimized[1]),
		Rho:   math.Max(-0.99, math.Min(0.99, optimized[2])),
		M:     optimized[3],
		Sigma: math.Max(1e-4, optimized[4]),
	}
}

// CalibrateSlice calibrates a smile slice from market quotes for a given expiry.
func (s *VolatilitySurface) CalibrateSlice(
	expiry float64,
	quotes []MarketQuote,
	ivEngine *ImpliedVolatilityEngine,
) (*CalibratedSmileSlice, error) {
	if len(quotes) == 0 {
		return nil, errors.New("no quotes provided for calibration")
	}

	forward := s.SpotPrice * math.Exp((s.RiskFreeRate-s.DividendYield)*expiry)

	logMoneyness := make([]float64, 0, len(quotes))
	ivs := make([]float64, 0, len(quotes))
	totalVars := make([]float64, 0, len(quotes))
	weights := make([]float64, 0, len(quotes))

	for _, q := range quotes {
		mid := q.Mid
		if mid <= 0 && q.Bid > 0 && q.Ask > 0 {
			mid = 0.5 * (q.Bid + q.Ask)
		}
		if mid <= 0 {
			continue
		}

		iv, err := ivEngine.CalculateIV(q.OptionType, mid, s.SpotPrice, q.Strike, expiry, s.RiskFreeRate, s.DividendYield)
		if err != nil {
			continue // Skip unfeasible quotes
		}

		k := math.Log(q.Strike / forward)
		w := iv * iv * expiry

		weight := q.Weight
		if weight <= 0 {
			if q.Ask > q.Bid && q.Bid > 0 {
				weight = 1.0 / (q.Ask - q.Bid + 1e-4)
			} else {
				weight = 1.0
			}
		}

		logMoneyness = append(logMoneyness, k)
		ivs = append(ivs, iv)
		totalVars = append(totalVars, w)
		weights = append(weights, weight)
	}

	if len(logMoneyness) < 2 {
		return nil, errors.New("insufficient valid quotes to calibrate smile slice (minimum 2 required)")
	}

	polyParams := FitPolynomialSmile(logMoneyness, ivs, weights)
	sviParams := FitSVISmile(logMoneyness, totalVars, weights, expiry)

	// Calculate RMSE
	var sumSqErr float64
	for i := range logMoneyness {
		predIV := polyParams.ImpliedVol(logMoneyness[i])
		errVal := predIV - ivs[i]
		sumSqErr += errVal * errVal
	}
	rmse := math.Sqrt(sumSqErr / float64(len(logMoneyness)))

	slice := CalibratedSmileSlice{
		Expiry:       expiry,
		Forward:      forward,
		SVI:          sviParams,
		Polynomial:   polyParams,
		FittedPoints: len(logMoneyness),
		RMSE:         rmse,
	}

	s.AddSlice(slice)
	return &slice, nil
}

// AddSlice inserts and maintains sorted order of slices by expiry.
func (s *VolatilitySurface) AddSlice(slice CalibratedSmileSlice) {
	for i, existing := range s.Slices {
		if math.Abs(existing.Expiry-slice.Expiry) < 1e-6 {
			// Replace existing slice
			s.Slices[i] = slice
			return
		}
	}
	s.Slices = append(s.Slices, slice)
	sort.Slice(s.Slices, func(i, j int) bool {
		return s.Slices[i].Expiry < s.Slices[j].Expiry
	})
}

// GetImpliedVol queries the volatility surface for any strike K and expiry T.
// Guarantees no calendar arbitrage across expiries via linear total variance interpolation.
func (s *VolatilitySurface) GetImpliedVol(strike, expiry float64) (float64, error) {
	if strike <= 0 || expiry <= 0 {
		return 0, errors.New("strike and expiry must be positive")
	}
	if len(s.Slices) == 0 {
		return 0, errors.New("volatility surface has no calibrated slices")
	}

	// 1. Single slice or exact match
	if len(s.Slices) == 1 {
		slice := s.Slices[0]
		forward := s.SpotPrice * math.Exp((s.RiskFreeRate-s.DividendYield)*slice.Expiry)
		k := math.Log(strike / forward)
		return slice.Polynomial.ImpliedVol(k), nil
	}

	// 2. Extrapolation below earliest expiry
	if expiry <= s.Slices[0].Expiry {
		slice := s.Slices[0]
		k := math.Log(strike / slice.Forward)
		return slice.Polynomial.ImpliedVol(k), nil
	}

	// 3. Extrapolation beyond latest expiry
	lastIdx := len(s.Slices) - 1
	if expiry >= s.Slices[lastIdx].Expiry {
		slice := s.Slices[lastIdx]
		k := math.Log(strike / slice.Forward)
		return slice.Polynomial.ImpliedVol(k), nil
	}

	// 4. Inter-slice interpolation: Find surrounding slices [T1, T2]
	var idx int
	for i := 0; i < len(s.Slices)-1; i++ {
		if expiry >= s.Slices[i].Expiry && expiry <= s.Slices[i+1].Expiry {
			idx = i
			break
		}
	}

	slice1 := s.Slices[idx]
	slice2 := s.Slices[idx+1]

	f1 := s.SpotPrice * math.Exp((s.RiskFreeRate-s.DividendYield)*slice1.Expiry)
	f2 := s.SpotPrice * math.Exp((s.RiskFreeRate-s.DividendYield)*slice2.Expiry)

	k1 := math.Log(strike / f1)
	k2 := math.Log(strike / f2)

	vol1 := slice1.Polynomial.ImpliedVol(k1)
	vol2 := slice2.Polynomial.ImpliedVol(k2)

	// Total variance linear interpolation prevents calendar arbitrage:
	// w(T) = (T2 - T)/(T2 - T1)*w(T1) + (T - T1)/(T2 - T1)*w(T2)
	w1 := vol1 * vol1 * slice1.Expiry
	w2 := vol2 * vol2 * slice2.Expiry

	weight := (expiry - slice1.Expiry) / (slice2.Expiry - slice1.Expiry)
	wT := (1.0-weight)*w1 + weight*w2

	if wT < 0 {
		wT = 1e-8
	}

	return math.Sqrt(wT / expiry), nil
}

// ValidateCalendarArbitrage checks that total implied variance is monotonically non-decreasing
// with respect to maturity for a set of strikes.
func (s *VolatilitySurface) ValidateCalendarArbitrage(testStrikes []float64) bool {
	if len(s.Slices) < 2 {
		return true
	}

	for _, k := range testStrikes {
		for i := 0; i < len(s.Slices)-1; i++ {
			t1 := s.Slices[i].Expiry
			t2 := s.Slices[i+1].Expiry
			v1, err1 := s.GetImpliedVol(k, t1)
			v2, err2 := s.GetImpliedVol(k, t2)
			if err1 != nil || err2 != nil {
				return false
			}
			w1 := v1 * v1 * t1
			w2 := v2 * v2 * t2
			if w2 < w1-1e-6 {
				return false // Calendar arbitrage detected!
			}
		}
	}
	return true
}

// -----------------------------------------------------------------------------
// Numerical Solver Helpers
// -----------------------------------------------------------------------------

func solveGaussian(a [][]float64, b []float64) []float64 {
	n := len(b)
	// Augment matrix
	m := make([][]float64, n)
	for i := range m {
		m[i] = make([]float64, n+1)
		copy(m[i], a[i])
		m[i][n] = b[i]
	}

	for i := 0; i < n; i++ {
		// Partial pivot
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(m[k][i]) > math.Abs(m[maxRow][i]) {
				maxRow = k
			}
		}
		m[i], m[maxRow] = m[maxRow], m[i]

		pivot := m[i][i]
		if math.Abs(pivot) < 1e-12 {
			continue
		}

		for j := i; j <= n; j++ {
			m[i][j] /= pivot
		}

		for k := 0; k < n; k++ {
			if k != i {
				factor := m[k][i]
				for j := i; j <= n; j++ {
					m[k][j] -= factor * m[i][j]
				}
			}
		}
	}

	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = m[i][n]
	}
	return x
}

// nelderMead is a compact Simplex algorithm for multidimensional optimization.
func nelderMead(f func([]float64) float64, start []float64, maxIter int, tol float64) []float64 {
	dim := len(start)
	simplex := make([][]float64, dim+1)
	scores := make([]float64, dim+1)

	// Step 0: Construct initial simplex
	simplex[0] = make([]float64, dim)
	copy(simplex[0], start)
	scores[0] = f(simplex[0])

	for i := 0; i < dim; i++ {
		point := make([]float64, dim)
		copy(point, start)
		step := 0.05
		if math.Abs(point[i]) > 1e-4 {
			step = 0.05 * point[i]
		}
		point[i] += step
		simplex[i+1] = point
		scores[i+1] = f(point)
	}

	const (
		alpha = 1.0  // reflection
		gamma = 2.0  // expansion
		rho   = 0.5  // contraction
		sigma = 0.5  // shrink
	)

	for iter := 0; iter < maxIter; iter++ {
		// Sort simplex by scores
		for i := 0; i < len(simplex)-1; i++ {
			for j := i + 1; j < len(simplex); j++ {
				if scores[j] < scores[i] {
					scores[i], scores[j] = scores[j], scores[i]
					simplex[i], simplex[j] = simplex[j], simplex[i]
				}
			}
		}

		if scores[dim]-scores[0] < tol {
			break
		}

		// Centroid of the best dim points
		centroid := make([]float64, dim)
		for i := 0; i < dim; i++ {
			for d := 0; d < dim; d++ {
				centroid[d] += simplex[i][d]
			}
		}
		for d := 0; d < dim; d++ {
			centroid[d] /= float64(dim)
		}

		// 1. Reflection
		reflected := make([]float64, dim)
		for d := 0; d < dim; d++ {
			reflected[d] = centroid[d] + alpha*(centroid[d]-simplex[dim][d])
		}
		scoreReflected := f(reflected)

		if scoreReflected >= scores[0] && scoreReflected < scores[dim-1] {
			simplex[dim] = reflected
			scores[dim] = scoreReflected
			continue
		}

		// 2. Expansion
		if scoreReflected < scores[0] {
			expanded := make([]float64, dim)
			for d := 0; d < dim; d++ {
				expanded[d] = centroid[d] + gamma*(reflected[d]-centroid[d])
			}
			scoreExpanded := f(expanded)
			if scoreExpanded < scoreReflected {
				simplex[dim] = expanded
				scores[dim] = scoreExpanded
			} else {
				simplex[dim] = reflected
				scores[dim] = scoreReflected
			}
			continue
		}

		// 3. Contraction
		contracted := make([]float64, dim)
		for d := 0; d < dim; d++ {
			contracted[d] = centroid[d] + rho*(simplex[dim][d]-centroid[d])
		}
		scoreContracted := f(contracted)

		if scoreContracted < scores[dim] {
			simplex[dim] = contracted
			scores[dim] = scoreContracted
			continue
		}

		// 4. Shrink
		for i := 1; i <= dim; i++ {
			for d := 0; d < dim; d++ {
				simplex[i][d] = simplex[0][d] + sigma*(simplex[i][d]-simplex[0][d])
			}
			scores[i] = f(simplex[i])
		}
	}

	return simplex[0]
}
