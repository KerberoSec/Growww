# Options & Volatility Trading Engine Specification

This document defines the mathematical models, pricing engines, Greeks calculations, implied volatility (IV) surface calibration, and portfolio margin algorithms for European and American options on the Growww RWA Exchange.

---

## 1. Options Pricing Engine & Black-Scholes-Merton Model

### 1.1 Black-Scholes-Merton (BSM) Analytical Pricing
Options contracts on BTC, ETH, and Tokenized Blue-Chip equities (Reliance, TCS, Apple) are priced in real time ($< 5\mu\text{s}$) using the BSM formulation:

$$C = S e^{-q T} N(d_1) - K e^{-r T} N(d_2)$$
$$P = K e^{-r T} N(-d_2) - S e^{-q T} N(-d_1)$$

Where:
$$d_1 = \frac{\ln(S / K) + \left(r - q + \frac{\sigma^2}{2}\right)T}{\sigma \sqrt{T}}, \quad d_2 = d_1 - \sigma \sqrt{T}$$

- $S$: Underlying Spot / Mark Price.
- $K$: Strike Price.
- $T$: Time to Expiration in Years.
- $r$: Risk-free Interest Rate (RBI 91-day T-Bill rate / US SOFR).
- $q$: Continuous Dividend Yield (for equities).
- $\sigma$: Implied Volatility (IV).
- $N(x)$: Cumulative Standard Normal Distribution.

### 1.2 Real-Time Analytical Greeks Engine
The engine computes 1st and 2nd order Greeks for every strike on every underlying price update:

$$\text{Delta } (\Delta) = \frac{\partial V}{\partial S} = e^{-q T} N(d_1) \quad (\text{Calls}), \quad -e^{-q T} N(-d_1) \quad (\text{Puts})$$
$$\text{Gamma } (\Gamma) = \frac{\partial^2 V}{\partial S^2} = \frac{e^{-q T} N'(d_1)}{S \sigma \sqrt{T}}$$
$$\text{Theta } (\Theta) = \frac{\partial V}{\partial T} = -\frac{S \sigma e^{-q T} N'(d_1)}{2 \sqrt{T}} - r K e^{-r T} N(d_2) + q S e^{-q T} N(d_1)$$
$$\text{Vega } (\mathcal{V}) = \frac{\partial V}{\partial \sigma} = S e^{-q T} \sqrt{T} N'(d_1)$$
$$\text{Rho } (\rho) = \frac{\partial V}{\partial r} = K T e^{-r T} N(d_2) \quad (\text{Calls}), \quad -K T e^{-r T} N(-d_2) \quad (\text{Puts})$$

---

## 2. Implied Volatility (IV) Surface & SABR Model

### 2.1 Dynamic Volatility Smile Fitting (SABR Formulation)
To prevent arbitrage across strike prices and maturities, the exchange fits the Hagan et al. SABR model to calculate local volatility $\sigma(K, T)$:

$$\sigma_{SABR}(K, F) = \frac{\alpha}{(F K)^{(1-\beta)/2} \left[1 + \frac{(1-\beta)^2}{24}\ln^2(F/K) + \dots\right]} \cdot \left(\frac{z}{\chi(z)}\right) \cdot \left[1 + \left(\frac{(1-\beta)^2}{24}\frac{\alpha^2}{(FK)^{1-\beta}} + \frac{\rho \beta \nu \alpha}{4(FK)^{(1-\beta)/2}} + \frac{2-3\rho^2}{24}\nu^2\right)T\right]$$

---

## 3. Portfolio Margin (SPAN-Style Risk Offsetting)

### 3.1 Portfolio Risk Simulation Matrix
Rather than requiring separate margin for every individual option leg, **Portfolio Margin** simulates the entire account portfolio across **16 Risk Scenarios**:
1. Underlying Price Shifts: $\pm 3\%, \pm 6\%, \pm 10\%, \pm 15\%$.
2. Volatility Shifts: $\pm 5\%, \pm 10\%, \pm 20\%$.
3. Extreme Move Stress Test: $\pm 30\%$ move with a 35% loss weighting.

### 3.2 Maximum Theoretical Loss & Maintenance Margin
$$\text{Portfolio Margin Requirement} = \max_{\text{Scenarios } k}\left(\sum_{i=1}^M \Delta \text{Value}_{i, k}\right) + \text{ShortOptionMinimumBuffer}$$

- Long call and short call spreads (e.g. Bull Call Spreads) receive **up to 85% margin reduction**.
- Delta-hedged portfolios (Long Perp + Short Call) require margin only for gamma/vega tail risk.
