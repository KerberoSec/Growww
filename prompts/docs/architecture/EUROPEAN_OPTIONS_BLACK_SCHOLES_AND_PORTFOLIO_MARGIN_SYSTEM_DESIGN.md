# European Options Black-Scholes & Portfolio Margin System Design

**Specification ID:** SPEC-ARCH-020-OPT  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Quantitative Options Pricing & Risk Capital Optimization  
**Owner:** Quantitative Financial Engineering Group  

---

## 1. Executive Summary & Zero-Fee Structure
The Options and Volatility trading engine prices European vanilla and exotic options contracts on BTC and major crypto assets:
- **Universal Zero-Fee Trading**: Strictly **0.00% commission** on options trades (No fee at all).
- **Sub-5 Microsecond Pricing**: Analytical Black-Scholes-Merton (BSM) evaluation accelerated with vectorized SIMD / AVX-512 CPU instructions.
- **SPAN-Style Portfolio Margin**: Offsets margin across correlated options and underlying spot/futures positions, granting up to 85% margin relief for hedged strategies.

---

## 2. Analytical Black-Scholes-Merton & Greeks Formulations

### 2.1 Pricing Equations:
$$C(S, K, T) = S N(d_1) - K e^{-r T} N(d_2)$$
$$P(S, K, T) = K e^{-r T} N(-d_2) - S N(-d_1)$$
$$d_1 = \frac{\ln(S/K) + \left(r + \frac{\sigma^2}{2}\right)T}{\sigma \sqrt{T}}, \quad d_2 = d_1 - \sigma \sqrt{T}$$

### 2.2 First and Second Order Analytical Greeks:
- **Delta ($\Delta$)**: $\frac{\partial V}{\partial S} = N(d_1)$ (Calls), $N(d_1) - 1$ (Puts)
- **Gamma ($\Gamma$)**: $\frac{\partial^2 V}{\partial S^2} = \frac{N'(d_1)}{S \sigma \sqrt{T}}$
- **Theta ($\Theta$)**: $\frac{\partial V}{\partial T} = -\frac{S N'(d_1) \sigma}{2 \sqrt{T}} - r K e^{-r T} N(d_2)$
- **Vega ($\mathcal{V}$)**: $\frac{\partial V}{\partial \sigma} = S \sqrt{T} N'(d_1)$
- **Rho ($\rho$)**: $\frac{\partial V}{\partial r} = K T e^{-r T} N(d_2)$

---

## 3. Portfolio Margin Simulation Matrix (16 Scenarios)
Portfolio margin simulates total portfolio value across 16 extreme scenarios:
1. Underlying Spot Price Shocks: $\pm 3\%, \pm 6\%, \pm 10\%, \pm 15\%$.
2. Implied Volatility Shocks: $\pm 5\%, \pm 10\%, \pm 20\%$.
3. Extreme Tail Stress Test: $\pm 30\%$ with 35% loss weighting.

$$\text{Portfolio Margin Requirement} = \max_{k \in [1, 16]} (\text{ScenarioLoss}_k) + \text{ShortOptionMinimumBuffer}$$
