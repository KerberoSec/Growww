package main

import (
	"math"
	"testing"
)

// Hull textbook benchmark: S=100, K=100, T=1.0, r=0.05, q=0.0, v=0.20
func TestBlackScholesKnownBenchmark(t *testing.T) {
	engine := NewBlackScholesEngine(0.05, 0.0)

	spot := 100.0
	strike := 100.0
	expiry := 1.0
	rate := 0.05
	div := 0.0
	vol := 0.20

	callRes, err := engine.Price(OptionCall, spot, strike, expiry, vol, rate, div)
	if err != nil {
		t.Fatalf("Call pricing failed: %v", err)
	}

	putRes, err := engine.Price(OptionPut, spot, strike, expiry, vol, rate, div)
	if err != nil {
		t.Fatalf("Put pricing failed: %v", err)
	}

	expectedCall := 10.45058
	expectedPut := 5.57352

	if math.Abs(callRes.Price-expectedCall) > 0.001 {
		t.Errorf("Call price mismatch: got %.5f, expected %.5f", callRes.Price, expectedCall)
	}
	if math.Abs(putRes.Price-expectedPut) > 0.001 {
		t.Errorf("Put price mismatch: got %.5f, expected %.5f", putRes.Price, expectedPut)
	}

	// Verify Put-Call Parity: C - P = S - K * e^(-r*T)
	valid, diff := ValidatePutCallParity(callRes.Price, putRes.Price, spot, strike, expiry, rate, div, 1e-7)
	if !valid {
		t.Errorf("Put-call parity failed with difference: %e", diff)
	}
}

// Numerical differentiation test for Greeks
func TestGreeksFiniteDifference(t *testing.T) {
	engine := NewBlackScholesEngine(0.06, 0.02)
	s := 150.0
	k := 140.0
	expiry := 0.75
	v := 0.28
	r := 0.06
	q := 0.02

	baseRes, err := engine.Price(OptionCall, s, k, expiry, v, r, q)
	if err != nil {
		t.Fatalf("Base price error: %v", err)
	}

	epsS := 0.01
	resUpS, _ := engine.Price(OptionCall, s+epsS, k, expiry, v, r, q)
	resDownS, _ := engine.Price(OptionCall, s-epsS, k, expiry, v, r, q)

	// Numerical Delta
	numDelta := (resUpS.Price - resDownS.Price) / (2.0 * epsS)
	if math.Abs(baseRes.Greeks.Delta-numDelta) > 1e-3 {
		t.Errorf("Delta mismatch: analytical=%.6f, numerical=%.6f", baseRes.Greeks.Delta, numDelta)
	}

	// Numerical Gamma
	numGamma := (resUpS.Price - 2.0*baseRes.Price + resDownS.Price) / (epsS * epsS)
	if math.Abs(baseRes.Greeks.Gamma-numGamma) > 1e-3 {
		t.Errorf("Gamma mismatch: analytical=%.6f, numerical=%.6f", baseRes.Greeks.Gamma, numGamma)
	}

	// Numerical Vega
	epsV := 0.001
	resUpV, _ := engine.Price(OptionCall, s, k, expiry, v+epsV, r, q)
	resDownV, _ := engine.Price(OptionCall, s, k, expiry, v-epsV, r, q)
	numVega := (resUpV.Price - resDownV.Price) / (2.0 * epsV)
	if math.Abs(baseRes.Greeks.Vega-numVega) > 1e-2 {
		t.Errorf("Vega mismatch: analytical=%.6f, numerical=%.6f", baseRes.Greeks.Vega, numVega)
	}

	// Numerical Rho
	epsR := 0.0001
	resUpR, _ := engine.Price(OptionCall, s, k, expiry, v, r+epsR, q)
	resDownR, _ := engine.Price(OptionCall, s, k, expiry, v, r-epsR, q)
	numRho := (resUpR.Price - resDownR.Price) / (2.0 * epsR)
	if math.Abs(baseRes.Greeks.Rho-numRho) > 1e-2 {
		t.Errorf("Rho mismatch: analytical=%.6f, numerical=%.6f", baseRes.Greeks.Rho, numRho)
	}
}

// Test boundary conditions (expiry, zero vol, deep ITM/OTM)
func TestBoundaryConditions(t *testing.T) {
	engine := NewBlackScholesEngine(0.05, 0.0)

	// 1. Time to expiry = 0 (Immediate Expiry)
	resCall0, err := engine.Price(OptionCall, 120.0, 100.0, 0.0, 0.20, 0.05, 0.0)
	if err != nil {
		t.Fatalf("Unexpected error for t=0: %v", err)
	}
	if resCall0.Price != 20.0 || resCall0.Greeks.Delta != 1.0 {
		t.Errorf("Expected expired ITM call price=20.0, delta=1.0, got price=%.2f, delta=%.2f", resCall0.Price, resCall0.Greeks.Delta)
	}

	resPut0, err := engine.Price(OptionPut, 120.0, 100.0, 0.0, 0.20, 0.05, 0.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resPut0.Price != 0.0 || resPut0.Greeks.Delta != 0.0 {
		t.Errorf("Expected expired OTM put price=0.0, delta=0.0, got price=%.2f, delta=%.2f", resPut0.Price, resPut0.Greeks.Delta)
	}

	// 2. Zero Volatility
	resZeroVol, err := engine.Price(OptionCall, 110.0, 100.0, 0.5, 0.0, 0.05, 0.0)
	if err != nil {
		t.Fatalf("Unexpected error for vol=0: %v", err)
	}
	if resZeroVol.Price != 10.0 {
		t.Errorf("Expected zero vol call intrinsic=10.0, got %.2f", resZeroVol.Price)
	}

	// 3. Deep OTM option (Price should approach 0, Delta ~ 0)
	resDeepOTM, err := engine.Price(OptionCall, 50.0, 200.0, 0.1, 0.2, 0.05, 0.0)
	if err != nil {
		t.Fatalf("Deep OTM pricing error: %v", err)
	}
	if resDeepOTM.Price > 1e-6 || resDeepOTM.Greeks.Delta > 1e-4 {
		t.Errorf("Deep OTM option should have zero price and delta, got price=%e, delta=%e", resDeepOTM.Price, resDeepOTM.Greeks.Delta)
	}

	// 4. Invalid arguments
	if _, err := engine.Price(OptionCall, -100, 100, 1, 0.2, 0.05, 0); err == nil {
		t.Error("Expected error on negative spot")
	}
	if _, err := engine.Price(OptionCall, 100, -100, 1, 0.2, 0.05, 0); err == nil {
		t.Error("Expected error on negative strike")
	}
}

// Test Option Chain generation
func TestBuildOptionChain(t *testing.T) {
	engine := NewBlackScholesEngine(0.05, 0.01)
	chain, err := engine.BuildOptionChain(25000.0, 0.25, 0.18, 0.06, 0.01, 200.0, 5)
	if err != nil {
		t.Fatalf("Failed to build option chain: %v", err)
	}

	if len(chain) != 11 {
		t.Fatalf("Expected 11 strikes in chain, got %d", len(chain))
	}

	for _, leg := range chain {
		if leg.CallResult == nil || leg.PutResult == nil {
			t.Fatalf("Leg strike %.1f missing call or put result", leg.Strike)
		}
		// Put-call parity must hold on every leg
		valid, diff := ValidatePutCallParity(
			leg.CallResult.Price, leg.PutResult.Price,
			25000.0, leg.Strike, 0.25, 0.06, 0.01, 1e-4,
		)
		if !valid {
			t.Errorf("Put-call parity failed at strike %.1f with diff %e", leg.Strike, diff)
		}
	}
}

// Test Implied Volatility Solver (Call and Put Roundtrips)
func TestImpliedVolatilitySolver(t *testing.T) {
	bsEngine := NewBlackScholesEngine(0.05, 0.0)
	ivEngine := NewImpliedVolatilityEngine(DefaultIVSolverConfig())

	spot := 500.0
	strikes := []float64{450.0, 480.0, 500.0, 520.0, 550.0}
	expiry := 0.5
	rate := 0.04
	div := 0.01
	targetVols := []float64{0.15, 0.25, 0.35, 0.45, 0.60}

	for _, k := range strikes {
		for _, targetVol := range targetVols {
			// Call Test
			callRes, err := bsEngine.Price(OptionCall, spot, k, expiry, targetVol, rate, div)
			if err != nil {
				t.Fatalf("Call price error: %v", err)
			}

			solvedCallIV, err := ivEngine.CalculateIV(OptionCall, callRes.Price, spot, k, expiry, rate, div)
			if err != nil {
				t.Fatalf("Failed to solve Call IV for strike=%.1f, vol=%.2f: %v", k, targetVol, err)
			}
			if math.Abs(solvedCallIV-targetVol) > 1e-4 {
				t.Errorf("Call IV mismatch at K=%.1f: expected=%.4f, solved=%.4f", k, targetVol, solvedCallIV)
			}

			// Put Test
			putRes, err := bsEngine.Price(OptionPut, spot, k, expiry, targetVol, rate, div)
			if err != nil {
				t.Fatalf("Put price error: %v", err)
			}

			solvedPutIV, err := ivEngine.CalculateIV(OptionPut, putRes.Price, spot, k, expiry, rate, div)
			if err != nil {
				t.Fatalf("Failed to solve Put IV for strike=%.1f, vol=%.2f: %v", k, targetVol, err)
			}
			if math.Abs(solvedPutIV-targetVol) > 1e-4 {
				t.Errorf("Put IV mismatch at K=%.1f: expected=%.4f, solved=%.4f", k, targetVol, solvedPutIV)
			}
		}
	}
}

// Test Arbitrage Bounds Enforcement in IV Solver
func TestArbitrageBoundEnforcement(t *testing.T) {
	ivEngine := NewImpliedVolatilityEngine(DefaultIVSolverConfig())
	spot := 100.0
	strike := 80.0
	expiry := 1.0
	rate := 0.05
	div := 0.0

	// Intrinsic lower bound for call is ~ S - K*e^(-r*T) = 100 - 80*0.9512 = 23.90
	// Price below lower bound (e.g., $10.00) should return an arbitrage error
	_, err := ivEngine.CalculateIV(OptionCall, 10.0, spot, strike, expiry, rate, div)
	if err == nil {
		t.Error("Expected arbitrage bound error for call price below intrinsic, got nil")
	}

	// Call price above spot (e.g. $110.00) violates upper bound
	_, err = ivEngine.CalculateIV(OptionCall, 110.0, spot, strike, expiry, rate, div)
	if err == nil {
		t.Error("Expected arbitrage bound error for call price exceeding spot, got nil")
	}
}

// Test Volatility Smile & Surface Calibration
func TestVolatilitySurfaceCalibration(t *testing.T) {
	bsEngine := NewBlackScholesEngine(0.05, 0.0)
	ivEngine := NewImpliedVolatilityEngine(DefaultIVSolverConfig())

	spot := 1000.0
	surface := NewVolatilitySurface("NIFTY50", spot, 0.06, 0.01)

	expiries := []float64{0.25, 0.50, 1.00}
	strikes := []float64{850.0, 900.0, 950.0, 1000.0, 1050.0, 1100.0, 1150.0}

	for _, exp := range expiries {
		quotes := make([]MarketQuote, 0, len(strikes))
		for _, k := range strikes {
			// Create a realistic synthetic volatility smile: ATM vol = 20%, skew = -0.15, curvature = 0.3
			logM := math.Log(k / spot)
			trueVol := 0.20 - 0.15*logM + 0.30*logM*logM
			callRes, err := bsEngine.Price(OptionCall, spot, k, exp, trueVol, 0.06, 0.01)
			if err != nil {
				continue
			}

			quotes = append(quotes, MarketQuote{
				OptionType: OptionCall,
				Strike:     k,
				Expiry:     exp,
				Mid:        callRes.Price,
				Bid:        callRes.Price - 0.5,
				Ask:        callRes.Price + 0.5,
			})
		}

		slice, err := surface.CalibrateSlice(exp, quotes, ivEngine)
		if err != nil {
			t.Fatalf("Failed to calibrate smile slice at T=%.2f: %v", exp, err)
		}
		if slice.RMSE > 0.05 {
			t.Errorf("Calibration RMSE too high: %f", slice.RMSE)
		}
	}

	// Query Surface at intermediate maturity and strikes
	queryVol, err := surface.GetImpliedVol(1000.0, 0.375) // Halfway between 0.25 and 0.50
	if err != nil {
		t.Fatalf("Failed to query surface: %v", err)
	}
	if queryVol < 0.15 || queryVol > 0.30 {
		t.Errorf("Interpolated volatility out of expected range: %.4f", queryVol)
	}

	// Verify Calendar Arbitrage: Total variance must be non-decreasing across expiries
	if !surface.ValidateCalendarArbitrage(strikes) {
		t.Error("Calendar arbitrage detected on calibrated volatility surface")
	}
}
