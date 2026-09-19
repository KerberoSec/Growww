"""
Options Pricing Engine – NBSE Sovereign Exchange
=================================================
Production-grade Black-Scholes pricing with:
  • European Call / Put pricing
  • Full Greeks: delta, gamma, theta, vega, rho
  • Implied-volatility Newton-Raphson solver
  • Options chain builder for strike ladders
"""

from __future__ import annotations

import math
from dataclasses import dataclass, field
from enum import Enum
from typing import List, Optional


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class OptionType(Enum):
    CALL = "CALL"
    PUT = "PUT"


@dataclass(frozen=True)
class Greeks:
    """Full set of option sensitivities."""
    delta: float
    gamma: float
    theta: float  # per calendar day
    vega: float   # per 1% move in vol
    rho: float    # per 1% move in rate


@dataclass(frozen=True)
class OptionPriceResult:
    """Result of pricing a single European option."""
    option_type: OptionType
    spot: float
    strike: float
    time_to_expiry: float  # years
    risk_free_rate: float
    volatility: float
    price: float
    greeks: Greeks


@dataclass
class OptionsChainLeg:
    """Single leg in an options chain."""
    strike: float
    option_type: OptionType
    price: float
    greeks: Greeks
    implied_vol: Optional[float] = None


@dataclass
class OptionsChain:
    """Full options chain for a symbol at a given expiry."""
    symbol: str
    spot: float
    expiry_years: float
    risk_free_rate: float
    legs: List[OptionsChainLeg] = field(default_factory=list)


# ---------------------------------------------------------------------------
# Normal CDF / PDF (Abramowitz & Stegun 7.1.26, ε < 1.5 × 10⁻⁷)
# ---------------------------------------------------------------------------

def _norm_cdf(x: float) -> float:
    """Standard-normal cumulative distribution function."""
    if x >= 0.0:
        t = 1.0 / (1.0 + 0.2316419 * x)
        poly = t * (0.319381530 + t * (-0.356563782 + t * (1.781477937 + t * (-1.821255978 + t * 1.330274429))))
        return 1.0 - (1.0 / math.sqrt(2.0 * math.pi)) * math.exp(-x * x / 2.0) * poly
    else:
        return 1.0 - _norm_cdf(-x)


def _norm_pdf(x: float) -> float:
    """Standard-normal probability density function."""
    return (1.0 / math.sqrt(2.0 * math.pi)) * math.exp(-x * x / 2.0)


# ---------------------------------------------------------------------------
# Core Black-Scholes Calculator
# ---------------------------------------------------------------------------

class BlackScholesEngine:
    """
    Black-Scholes option pricing engine for European-style options.

    Parameters
    ----------
    risk_free_rate : float
        Annualised continuous-compounding risk-free rate (e.g. 0.05 for 5 %).
    """

    def __init__(self, risk_free_rate: float = 0.05):
        if risk_free_rate < 0:
            raise ValueError("risk_free_rate must be non-negative")
        self.risk_free_rate = risk_free_rate

    # ---- internal helpers --------------------------------------------------

    @staticmethod
    def _d1(s: float, k: float, t: float, r: float, sigma: float) -> float:
        return (math.log(s / k) + (r + 0.5 * sigma * sigma) * t) / (sigma * math.sqrt(t))

    @staticmethod
    def _d2(d1: float, sigma: float, t: float) -> float:
        return d1 - sigma * math.sqrt(t)

    # ---- pricing -----------------------------------------------------------

    def price(
        self,
        option_type: OptionType,
        spot: float,
        strike: float,
        time_to_expiry: float,
        volatility: float,
        *,
        risk_free_rate: Optional[float] = None,
    ) -> OptionPriceResult:
        """
        Price a European option and compute full Greeks.

        Parameters
        ----------
        option_type : OptionType
        spot : float – current underlying price
        strike : float – option strike price
        time_to_expiry : float – time to expiry in years (must be > 0)
        volatility : float – annualised volatility (e.g. 0.30 for 30 %)
        risk_free_rate : float, optional – override the engine default

        Returns
        -------
        OptionPriceResult

        Raises
        ------
        ValueError – on invalid inputs
        """
        r = risk_free_rate if risk_free_rate is not None else self.risk_free_rate
        self._validate(spot, strike, time_to_expiry, volatility)

        # Edge-case: expiry reached
        if time_to_expiry <= 1e-12:
            intrinsic = max(spot - strike, 0.0) if option_type == OptionType.CALL else max(strike - spot, 0.0)
            return OptionPriceResult(
                option_type=option_type, spot=spot, strike=strike,
                time_to_expiry=0.0, risk_free_rate=r, volatility=volatility,
                price=intrinsic,
                greeks=Greeks(
                    delta=1.0 if option_type == OptionType.CALL and spot > strike else (
                        -1.0 if option_type == OptionType.PUT and spot < strike else 0.0),
                    gamma=0.0, theta=0.0, vega=0.0, rho=0.0,
                ),
            )

        t = time_to_expiry
        s, k, v = spot, strike, volatility

        d1 = self._d1(s, k, t, r, v)
        d2 = self._d2(d1, v, t)

        discount = math.exp(-r * t)
        sqrt_t = math.sqrt(t)
        pdf_d1 = _norm_pdf(d1)

        if option_type == OptionType.CALL:
            price_val = s * _norm_cdf(d1) - k * discount * _norm_cdf(d2)
            delta = _norm_cdf(d1)
            theta = (-(s * pdf_d1 * v) / (2.0 * sqrt_t)
                     - r * k * discount * _norm_cdf(d2))
            rho = k * t * discount * _norm_cdf(d2)
        else:
            price_val = k * discount * _norm_cdf(-d2) - s * _norm_cdf(-d1)
            delta = _norm_cdf(d1) - 1.0
            theta = (-(s * pdf_d1 * v) / (2.0 * sqrt_t)
                     + r * k * discount * _norm_cdf(-d2))
            rho = -k * t * discount * _norm_cdf(-d2)

        gamma = pdf_d1 / (s * v * sqrt_t)
        vega = s * pdf_d1 * sqrt_t  # per unit vol; divide by 100 for per-%

        # Theta per calendar day
        theta_day = theta / 365.0

        greeks = Greeks(
            delta=delta,
            gamma=gamma,
            theta=theta_day,
            vega=vega / 100.0,   # per 1% vol move
            rho=rho / 100.0,     # per 1% rate move
        )

        return OptionPriceResult(
            option_type=option_type, spot=spot, strike=strike,
            time_to_expiry=t, risk_free_rate=r, volatility=v,
            price=price_val, greeks=greeks,
        )

    # ---- implied volatility ------------------------------------------------

    def implied_volatility(
        self,
        option_type: OptionType,
        market_price: float,
        spot: float,
        strike: float,
        time_to_expiry: float,
        *,
        risk_free_rate: Optional[float] = None,
        initial_guess: float = 0.30,
        tol: float = 1e-8,
        max_iter: int = 100,
    ) -> float:
        """
        Solve for implied volatility via Newton-Raphson on vega.

        Returns
        -------
        float – annualised implied volatility

        Raises
        ------
        ValueError – if convergence fails
        """
        r = risk_free_rate if risk_free_rate is not None else self.risk_free_rate
        sigma = initial_guess

        for _ in range(max_iter):
            result = self.price(option_type, spot, strike, time_to_expiry, sigma, risk_free_rate=r)
            diff = result.price - market_price
            # vega stored per 1% move; need per unit vol for Newton step
            vega_unit = result.greeks.vega * 100.0
            if abs(vega_unit) < 1e-15:
                break
            sigma -= diff / vega_unit
            sigma = max(sigma, 1e-6)  # floor at near-zero
            if abs(diff) < tol:
                return sigma

        raise ValueError(
            f"Implied volatility solver failed to converge after {max_iter} iterations "
            f"(last sigma={sigma:.6f}, residual={diff:.2e})"
        )

    # ---- options chain builder ---------------------------------------------

    def build_chain(
        self,
        symbol: str,
        spot: float,
        time_to_expiry: float,
        volatility: float,
        *,
        strike_step: float = 100.0,
        num_strikes: int = 10,
        risk_free_rate: Optional[float] = None,
    ) -> OptionsChain:
        """
        Build a symmetric options chain around ATM.

        Parameters
        ----------
        symbol : str – underlying symbol
        spot : float – current price
        time_to_expiry : float – years
        volatility : float – annualised vol
        strike_step : float – gap between strikes
        num_strikes : int – number of strikes on *each* side of ATM
        risk_free_rate : float, optional

        Returns
        -------
        OptionsChain
        """
        r = risk_free_rate if risk_free_rate is not None else self.risk_free_rate
        atm_strike = round(spot / strike_step) * strike_step

        chain = OptionsChain(
            symbol=symbol, spot=spot, expiry_years=time_to_expiry, risk_free_rate=r,
        )

        for i in range(-num_strikes, num_strikes + 1):
            k = atm_strike + i * strike_step
            if k <= 0:
                continue
            for otype in (OptionType.CALL, OptionType.PUT):
                res = self.price(otype, spot, k, time_to_expiry, volatility, risk_free_rate=r)
                chain.legs.append(OptionsChainLeg(
                    strike=k, option_type=otype,
                    price=res.price, greeks=res.greeks,
                ))

        return chain

    # ---- validation --------------------------------------------------------

    @staticmethod
    def _validate(spot: float, strike: float, t: float, v: float) -> None:
        if spot <= 0:
            raise ValueError(f"spot must be positive, got {spot}")
        if strike <= 0:
            raise ValueError(f"strike must be positive, got {strike}")
        if t < 0:
            raise ValueError(f"time_to_expiry must be non-negative, got {t}")
        if v < 0:
            raise ValueError(f"volatility must be non-negative, got {v}")
