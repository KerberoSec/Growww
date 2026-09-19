"""
VaR Margin Engine – NBSE Sovereign Exchange
=============================================
Production-grade Value-at-Risk engine with:
  • Parametric VaR (variance-covariance method)
  • Historical Simulation VaR
  • Monte Carlo VaR
  • Extreme Loss Margin (ELM) for tail events
  • SPAN-like portfolio margin aggregation
"""

from __future__ import annotations

import math
import random
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class VaRMethod(Enum):
    PARAMETRIC = "PARAMETRIC"
    HISTORICAL = "HISTORICAL"
    MONTE_CARLO = "MONTE_CARLO"


@dataclass(frozen=True)
class VaRResult:
    """Result of a VaR calculation."""
    method: VaRMethod
    confidence_level: float
    holding_period_days: float
    var_amount: float             # absolute VaR in portfolio currency
    var_pct: float                # VaR as percentage of portfolio value
    portfolio_value: float
    expected_shortfall: float     # CVaR / Expected Shortfall (average loss beyond VaR)


@dataclass
class PortfolioPosition:
    """A position in the margin portfolio."""
    symbol: str
    quantity: float
    current_price: float
    daily_volatility: float       # annualised σ / √252
    weight: float = 0.0           # portfolio weight (computed)


@dataclass
class MarginRequirement:
    """Margin requirement computed from VaR + ELM."""
    var_margin: float
    elm_margin: float
    total_margin: float
    var_method: VaRMethod
    confidence_level: float


# ---------------------------------------------------------------------------
# Statistical helpers
# ---------------------------------------------------------------------------

def _norm_cdf_inv(p: float) -> float:
    """
    Rational approximation for the inverse standard-normal CDF (Beasley-Springer-Moro).
    Accurate to ~1e-9 for 0.00001 < p < 0.99999.
    """
    if p <= 0 or p >= 1:
        raise ValueError("p must be in (0, 1)")

    a = [
        -3.969683028665376e+01,  2.209460984245205e+02,
        -2.759285104469687e+02,  1.383577518672690e+02,
        -3.066479806614716e+01,  2.506628277459239e+00,
    ]
    b = [
        -5.447609879822406e+01,  1.615858368580409e+02,
        -1.556989798598866e+02,  6.680131188771972e+01,
        -1.328068155288572e+01,
    ]
    c = [
        -7.784894002430293e-03, -3.223964580411365e-01,
        -2.400758277161838e+00, -2.549732539343734e+00,
         4.374664141464968e+00,  2.938163982698783e+00,
    ]
    d = [
         7.784695709041462e-03,  3.224671290700398e-01,
         2.445134137142996e+00,  3.754408661907416e+00,
    ]

    p_low = 0.02425
    p_high = 1.0 - p_low

    if p < p_low:
        q = math.sqrt(-2.0 * math.log(p))
        return (((((c[0]*q + c[1])*q + c[2])*q + c[3])*q + c[4])*q + c[5]) / \
               ((((d[0]*q + d[1])*q + d[2])*q + d[3])*q + 1.0)
    elif p <= p_high:
        q = p - 0.5
        r = q * q
        return (((((a[0]*r + a[1])*r + a[2])*r + a[3])*r + a[4])*r + a[5]) * q / \
               (((((b[0]*r + b[1])*r + b[2])*r + b[3])*r + b[4])*r + 1.0)
    else:
        q = math.sqrt(-2.0 * math.log(1.0 - p))
        return -(((((c[0]*q + c[1])*q + c[2])*q + c[3])*q + c[4])*q + c[5]) / \
                ((((d[0]*q + d[1])*q + d[2])*q + d[3])*q + 1.0)


# ---------------------------------------------------------------------------
# Engine
# ---------------------------------------------------------------------------

class VaRMarginEngine:
    """
    Value-at-Risk margin engine supporting parametric, historical simulation,
    and Monte Carlo VaR methodologies.

    Parameters
    ----------
    confidence_level : float – e.g. 0.99 for 99% VaR
    holding_period_days : float – typically 1 for daily VaR
    elm_rate : float – Extreme Loss Margin rate (e.g. 0.035 for 3.5%)
    """

    def __init__(
        self,
        confidence_level: float = 0.99,
        holding_period_days: float = 1.0,
        elm_rate: float = 0.035,
    ):
        if not 0 < confidence_level < 1:
            raise ValueError("confidence_level must be in (0, 1)")
        self.confidence_level = confidence_level
        self.z_score = _norm_cdf_inv(confidence_level)
        self.holding_period_days = holding_period_days
        self.elm_rate = elm_rate

    # ---- Parametric VaR ----------------------------------------------------

    def parametric_var(
        self,
        portfolio_value: float,
        daily_volatility: float,
    ) -> VaRResult:
        """
        Compute parametric (variance-covariance) VaR.

        VaR = Portfolio Value × Z × σ_daily × √(holding_period)
        """
        if portfolio_value <= 0:
            raise ValueError("portfolio_value must be positive")
        if daily_volatility < 0:
            raise ValueError("daily_volatility must be non-negative")

        var_amount = (
            portfolio_value
            * self.z_score
            * daily_volatility
            * math.sqrt(self.holding_period_days)
        )

        # Expected shortfall approximation (for normal distribution):
        # ES = μ + σ × φ(z) / (1 - α)  where φ is the pdf
        pdf_z = (1.0 / math.sqrt(2 * math.pi)) * math.exp(-0.5 * self.z_score ** 2)
        es_multiplier = pdf_z / (1.0 - self.confidence_level)
        es_amount = portfolio_value * es_multiplier * daily_volatility * math.sqrt(self.holding_period_days)

        return VaRResult(
            method=VaRMethod.PARAMETRIC,
            confidence_level=self.confidence_level,
            holding_period_days=self.holding_period_days,
            var_amount=var_amount,
            var_pct=var_amount / portfolio_value * 100.0,
            portfolio_value=portfolio_value,
            expected_shortfall=es_amount,
        )

    # ---- Historical Simulation VaR -----------------------------------------

    def historical_var(
        self,
        portfolio_value: float,
        historical_returns: List[float],
    ) -> VaRResult:
        """
        Compute Historical Simulation VaR from a series of daily returns.

        Sorts returns and picks the (1 - confidence) percentile loss.

        Parameters
        ----------
        portfolio_value : float
        historical_returns : List[float] – daily returns (e.g. [-0.02, 0.01, ...])
        """
        if portfolio_value <= 0:
            raise ValueError("portfolio_value must be positive")
        if len(historical_returns) < 10:
            raise ValueError("need at least 10 historical returns")

        sorted_returns = sorted(historical_returns)
        n = len(sorted_returns)
        index = int(math.floor((1.0 - self.confidence_level) * n))
        index = max(0, min(index, n - 1))

        var_return = abs(sorted_returns[index])
        var_amount = portfolio_value * var_return * math.sqrt(self.holding_period_days)

        # Expected shortfall = average of losses beyond VaR
        tail_returns = sorted_returns[:index + 1]
        es_return = abs(sum(tail_returns) / len(tail_returns)) if tail_returns else var_return
        es_amount = portfolio_value * es_return * math.sqrt(self.holding_period_days)

        return VaRResult(
            method=VaRMethod.HISTORICAL,
            confidence_level=self.confidence_level,
            holding_period_days=self.holding_period_days,
            var_amount=var_amount,
            var_pct=var_amount / portfolio_value * 100.0,
            portfolio_value=portfolio_value,
            expected_shortfall=es_amount,
        )

    # ---- Monte Carlo VaR ---------------------------------------------------

    def monte_carlo_var(
        self,
        portfolio_value: float,
        daily_volatility: float,
        drift: float = 0.0,
        num_simulations: int = 10_000,
        seed: Optional[int] = None,
    ) -> VaRResult:
        """
        Compute Monte Carlo VaR using geometric Brownian motion simulations.

        Parameters
        ----------
        portfolio_value : float
        daily_volatility : float – daily σ
        drift : float – daily expected return (usually ≈ 0 for short horizon)
        num_simulations : int – number of Monte Carlo paths
        seed : int, optional – for reproducibility
        """
        if portfolio_value <= 0:
            raise ValueError("portfolio_value must be positive")
        if daily_volatility < 0:
            raise ValueError("daily_volatility must be non-negative")
        if num_simulations < 100:
            raise ValueError("num_simulations must be >= 100")

        rng = random.Random(seed)
        steps = max(1, int(self.holding_period_days))
        terminal_values: List[float] = []

        for _ in range(num_simulations):
            value = portfolio_value
            for _ in range(steps):
                z = rng.gauss(0, 1)
                daily_return = drift + daily_volatility * z
                value *= (1.0 + daily_return)
            terminal_values.append(value)

        # PnL distribution
        pnls = sorted([v - portfolio_value for v in terminal_values])
        index = int(math.floor((1.0 - self.confidence_level) * num_simulations))
        index = max(0, min(index, num_simulations - 1))

        var_amount = abs(pnls[index])

        # Expected shortfall
        tail = pnls[:index + 1]
        es_amount = abs(sum(tail) / len(tail)) if tail else var_amount

        return VaRResult(
            method=VaRMethod.MONTE_CARLO,
            confidence_level=self.confidence_level,
            holding_period_days=self.holding_period_days,
            var_amount=var_amount,
            var_pct=var_amount / portfolio_value * 100.0,
            portfolio_value=portfolio_value,
            expected_shortfall=es_amount,
        )

    # ---- ELM (Extreme Loss Margin) -----------------------------------------

    def extreme_loss_margin(self, portfolio_value: float) -> float:
        """Compute ELM covering beyond VaR tail events."""
        return portfolio_value * self.elm_rate

    # ---- Combined Margin Requirement ---------------------------------------

    def compute_margin_requirement(
        self,
        portfolio_value: float,
        daily_volatility: float,
        method: VaRMethod = VaRMethod.PARAMETRIC,
        historical_returns: Optional[List[float]] = None,
        mc_seed: Optional[int] = None,
    ) -> MarginRequirement:
        """
        Compute total margin requirement = VaR margin + ELM.

        Parameters
        ----------
        portfolio_value : float
        daily_volatility : float
        method : VaRMethod
        historical_returns : optional list for HISTORICAL method
        mc_seed : optional seed for MONTE_CARLO method
        """
        if method == VaRMethod.PARAMETRIC:
            var_result = self.parametric_var(portfolio_value, daily_volatility)
        elif method == VaRMethod.HISTORICAL:
            if historical_returns is None:
                raise ValueError("historical_returns required for HISTORICAL method")
            var_result = self.historical_var(portfolio_value, historical_returns)
        elif method == VaRMethod.MONTE_CARLO:
            var_result = self.monte_carlo_var(
                portfolio_value, daily_volatility, seed=mc_seed,
            )
        else:
            raise ValueError(f"Unknown VaR method: {method}")

        elm = self.extreme_loss_margin(portfolio_value)

        return MarginRequirement(
            var_margin=var_result.var_amount,
            elm_margin=elm,
            total_margin=var_result.var_amount + elm,
            var_method=method,
            confidence_level=self.confidence_level,
        )
