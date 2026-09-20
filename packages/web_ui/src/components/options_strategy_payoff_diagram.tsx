import React, { useState, useMemo } from 'react';

export type OptionType = 'CALL' | 'PUT';
export type PositionAction = 'BUY' | 'SELL';

export interface OptionLeg {
  id: string;
  type: OptionType;
  action: PositionAction;
  strike: number;
  premium: number;
  quantity: number;
  ivPct: number;
}

export interface StrategyGreeks {
  delta: number;
  gamma: number;
  theta: number;
  vega: number;
}

export interface StrategySummary {
  netPremium: number; // positive = net debit, negative = net credit
  maxProfit: number | 'Unlimited';
  maxLoss: number | 'Unlimited';
  breakevens: number[];
  riskRewardRatio: string;
  greeks: StrategyGreeks;
}

// Standard Normal CDF approximation (Abramowitz & Stegun)
function normalCdf(x: number): number {
  const a1 = 0.254829592;
  const a2 = -0.284496736;
  const a3 = 1.421413741;
  const a4 = -1.453152027;
  const a5 = 1.061405429;
  const p = 0.3275911;

  const sign = x < 0 ? -1 : 1;
  const absX = Math.abs(x) / Math.SQRT2;
  const t = 1.0 / (1.0 + p * absX);
  const erf = 1.0 - ((((a5 * t + a4) * t + a3) * t + a2) * t + a1) * t * Math.exp(-absX * absX);
  return 0.5 * (1.0 + sign * erf);
}

function normalPdf(x: number): number {
  return (1.0 / Math.sqrt(2 * Math.PI)) * Math.exp(-0.5 * x * x);
}

// Black-Scholes European Option Pricing & Greeks
function blackScholes(
  spot: number,
  strike: number,
  dteDays: number,
  volPct: number,
  ratePct: number,
  type: OptionType
) {
  if (dteDays <= 0) {
    const intrinsic = type === 'CALL' ? Math.max(0, spot - strike) : Math.max(0, strike - spot);
    return {
      price: intrinsic,
      delta: type === 'CALL' ? (spot > strike ? 1 : 0) : (spot < strike ? -1 : 0),
      gamma: 0,
      theta: 0,
      vega: 0,
    };
  }

  const T = Math.max(dteDays, 0.001) / 365.0;
  const sigma = Math.max(volPct, 1) / 100.0;
  const r = ratePct / 100.0;
  const sqrtT = Math.sqrt(T);

  const d1 = (Math.log(spot / strike) + (r + 0.5 * sigma * sigma) * T) / (sigma * sqrtT);
  const d2 = d1 - sigma * sqrtT;

  const nd1 = normalCdf(d1);
  const nd2 = normalCdf(d2);
  const nNegD1 = normalCdf(-d1);
  const nNegD2 = normalCdf(-d2);
  const npdfD1 = normalPdf(d1);

  let price = 0;
  let delta = 0;
  let theta = 0;

  if (type === 'CALL') {
    price = spot * nd1 - strike * Math.exp(-r * T) * nd2;
    delta = nd1;
    theta = (-(spot * npdfD1 * sigma) / (2 * sqrtT) - r * strike * Math.exp(-r * T) * nd2) / 365.0;
  } else {
    price = strike * Math.exp(-r * T) * nNegD2 - spot * nNegD1;
    delta = nd1 - 1.0;
    theta = (-(spot * npdfD1 * sigma) / (2 * sqrtT) + r * strike * Math.exp(-r * T) * nNegD2) / 365.0;
  }

  const gamma = npdfD1 / (spot * sigma * sqrtT);
  const vega = (spot * sqrtT * npdfD1) / 100.0; // change per 1% IV

  return { price, delta, gamma, theta, vega };
}

export function OptionsStrategyPayoffVisualizer() {
  const [selectedAsset, setSelectedAsset] = useState<'BTC' | 'ETH' | 'NIFTY'>('BTC');
  const [spotPrice, setSpotPrice] = useState<number>(65000);
  const [simulatedPrice, setSimulatedPrice] = useState<number>(65000);
  const [dte, setDte] = useState<number>(14); // 14 days to expiry
  const [riskFreeRate] = useState<number>(5.0); // 5%
  const [globalIv, setGlobalIv] = useState<number>(50); // 50%

  // Preset Strategy template generator
  const createPreset = (preset: string, spot: number): OptionLeg[] => {
    const roundTo = (val: number, step: number) => Math.round(val / step) * step;
    const step = spot > 10000 ? 1000 : spot > 1000 ? 50 : 100;
    const atm = roundTo(spot, step);

    switch (preset) {
      case 'BULL_CALL_SPREAD':
        return [
          { id: '1', type: 'CALL', action: 'BUY', strike: atm, premium: roundTo(spot * 0.035, 10), quantity: 1, ivPct: globalIv },
          { id: '2', type: 'CALL', action: 'SELL', strike: atm + step * 2, premium: roundTo(spot * 0.015, 10), quantity: 1, ivPct: globalIv },
        ];
      case 'BEAR_PUT_SPREAD':
        return [
          { id: '1', type: 'PUT', action: 'BUY', strike: atm, premium: roundTo(spot * 0.035, 10), quantity: 1, ivPct: globalIv },
          { id: '2', type: 'PUT', action: 'SELL', strike: atm - step * 2, premium: roundTo(spot * 0.015, 10), quantity: 1, ivPct: globalIv },
        ];
      case 'IRON_CONDOR':
        return [
          { id: '1', type: 'PUT', action: 'BUY', strike: atm - step * 3, premium: roundTo(spot * 0.008, 10), quantity: 1, ivPct: globalIv },
          { id: '2', type: 'PUT', action: 'SELL', strike: atm - step, premium: roundTo(spot * 0.022, 10), quantity: 1, ivPct: globalIv },
          { id: '3', type: 'CALL', action: 'SELL', strike: atm + step, premium: roundTo(spot * 0.022, 10), quantity: 1, ivPct: globalIv },
          { id: '4', type: 'CALL', action: 'BUY', strike: atm + step * 3, premium: roundTo(spot * 0.008, 10), quantity: 1, ivPct: globalIv },
        ];
      case 'STRADDLE':
        return [
          { id: '1', type: 'CALL', action: 'BUY', strike: atm, premium: roundTo(spot * 0.038, 10), quantity: 1, ivPct: globalIv },
          { id: '2', type: 'PUT', action: 'BUY', strike: atm, premium: roundTo(spot * 0.038, 10), quantity: 1, ivPct: globalIv },
        ];
      case 'STRANGLE':
        return [
          { id: '1', type: 'PUT', action: 'BUY', strike: atm - step * 2, premium: roundTo(spot * 0.018, 10), quantity: 1, ivPct: globalIv },
          { id: '2', type: 'CALL', action: 'BUY', strike: atm + step * 2, premium: roundTo(spot * 0.018, 10), quantity: 1, ivPct: globalIv },
        ];
      case 'LONG_CALL':
      default:
        return [
          { id: '1', type: 'CALL', action: 'BUY', strike: atm, premium: roundTo(spot * 0.035, 10), quantity: 1, ivPct: globalIv },
        ];
    }
  };

  const [legs, setLegs] = useState<OptionLeg[]>(() => createPreset('BULL_CALL_SPREAD', 65000));

  const handleAssetChange = (asset: 'BTC' | 'ETH' | 'NIFTY') => {
    setSelectedAsset(asset);
    const newSpot = asset === 'BTC' ? 65000 : asset === 'ETH' ? 3500 : 24500;
    setSpotPrice(newSpot);
    setSimulatedPrice(newSpot);
    setLegs(createPreset('BULL_CALL_SPREAD', newSpot));
  };

  const applyPreset = (preset: string) => {
    setLegs(createPreset(preset, spotPrice));
  };

  const addLeg = () => {
    const newId = (legs.length + 1).toString();
    const newLeg: OptionLeg = {
      id: newId,
      type: 'CALL',
      action: 'BUY',
      strike: spotPrice,
      premium: Math.round(spotPrice * 0.02),
      quantity: 1,
      ivPct: globalIv,
    };
    setLegs([...legs, newLeg]);
  };

  const removeLeg = (id: string) => {
    setLegs(legs.filter((l) => l.id !== id));
  };

  const updateLeg = (id: string, updates: Partial<OptionLeg>) => {
    setLegs(legs.map((l) => (l.id === id ? { ...l, ...updates } : l)));
  };

  // Payoff calculations
  const calculateExpiryPnL = (price: number, legList: OptionLeg[]): number => {
    return legList.reduce((acc, leg) => {
      let intrinsic = 0;
      if (leg.type === 'CALL') {
        intrinsic = Math.max(0, price - leg.strike);
      } else {
        intrinsic = Math.max(0, leg.strike - price);
      }

      const legPnL = leg.action === 'BUY'
        ? (intrinsic - leg.premium) * leg.quantity
        : (leg.premium - intrinsic) * leg.quantity;

      return acc + legPnL;
    }, 0);
  };

  const calculateT0PnL = (price: number, legList: OptionLeg[], dteDays: number): number => {
    return legList.reduce((acc, leg) => {
      const bs = blackScholes(price, leg.strike, dteDays, leg.ivPct, riskFreeRate, leg.type);
      const legPnL = leg.action === 'BUY'
        ? (bs.price - leg.premium) * leg.quantity
        : (leg.premium - bs.price) * leg.quantity;
      return acc + legPnL;
    }, 0);
  };

  // Calculate Net Greeks at spot
  const strategySummary = useMemo<StrategySummary>(() => {
    const netPremium = legs.reduce((acc, leg) => {
      const signedPrem = leg.action === 'BUY' ? leg.premium * leg.quantity : -leg.premium * leg.quantity;
      return acc + signedPrem;
    }, 0);

    let netDelta = 0;
    let netGamma = 0;
    let netTheta = 0;
    let netVega = 0;

    legs.forEach((leg) => {
      const bs = blackScholes(spotPrice, leg.strike, dte, leg.ivPct, riskFreeRate, leg.type);
      const sign = leg.action === 'BUY' ? 1 : -1;
      netDelta += sign * bs.delta * leg.quantity;
      netGamma += sign * bs.gamma * leg.quantity;
      netTheta += sign * bs.theta * leg.quantity;
      netVega += sign * bs.vega * leg.quantity;
    });

    // Min / Max over a wide range to determine Max Profit / Loss
    const testMin = spotPrice * 0.2;
    const testMax = spotPrice * 2.5;
    const steps = 150;
    const stepSize = (testMax - testMin) / steps;

    let minPnL = Infinity;
    let maxPnL = -Infinity;
    const breakevens: number[] = [];
    let prevPnL = calculateExpiryPnL(testMin, legs);

    for (let i = 0; i <= steps; i++) {
      const p = testMin + i * stepSize;
      const pnl = calculateExpiryPnL(p, legs);

      if (pnl < minPnL) minPnL = pnl;
      if (pnl > maxPnL) maxPnL = pnl;

      // Breakeven crossing check
      if ((prevPnL < 0 && pnl >= 0) || (prevPnL > 0 && pnl <= 0)) {
        breakevens.push(Math.round(p));
      }
      prevPnL = pnl;
    }

    // Check boundary behavior
    const highEndPnL = calculateExpiryPnL(spotPrice * 5, legs);
    const lowEndPnL = calculateExpiryPnL(0.01, legs);

    const isProfitUnlimited = highEndPnL > maxPnL * 1.5 || (legs.some(l => l.action === 'BUY' && l.type === 'CALL') && !legs.some(l => l.action === 'SELL' && l.type === 'CALL' && l.quantity >= 1));
    const isLossUnlimited = highEndPnL < minPnL * 1.5 || (legs.some(l => l.action === 'SELL' && l.type === 'CALL') && !legs.some(l => l.action === 'BUY' && l.type === 'CALL' && l.quantity >= 1));

    const finalMaxProfit = isProfitUnlimited ? 'Unlimited' : Math.round(maxPnL);
    const finalMaxLoss = isLossUnlimited ? 'Unlimited' : Math.round(minPnL);

    const rrRatio =
      finalMaxProfit === 'Unlimited' || finalMaxLoss === 'Unlimited'
        ? 'N/A'
        : finalMaxLoss !== 0
        ? (Math.abs(Number(finalMaxProfit)) / Math.abs(Number(finalMaxLoss))).toFixed(2)
        : '1:0';

    return {
      netPremium: Math.round(netPremium),
      maxProfit: finalMaxProfit,
      maxLoss: finalMaxLoss,
      breakevens,
      riskRewardRatio: rrRatio,
      greeks: {
        delta: Number(netDelta.toFixed(4)),
        gamma: Number(netGamma.toFixed(6)),
        theta: Number(netTheta.toFixed(2)),
        vega: Number(netVega.toFixed(2)),
      },
    };
  }, [legs, spotPrice, dte, riskFreeRate, globalIv]);

  // Diagram plot points (Width: 600, Height: 260)
  const rangeMin = spotPrice * 0.7;
  const rangeMax = spotPrice * 1.3;
  const plotWidth = 600;
  const plotHeight = 240;
  const numSamples = 60;

  const points = useMemo(() => {
    const result: { price: number; expiryPnL: number; t0PnL: number }[] = [];
    const step = (rangeMax - rangeMin) / numSamples;
    for (let i = 0; i <= numSamples; i++) {
      const p = rangeMin + i * step;
      result.push({
        price: p,
        expiryPnL: calculateExpiryPnL(p, legs),
        t0PnL: calculateT0PnL(p, legs, dte),
      });
    }
    return result;
  }, [rangeMin, rangeMax, legs, dte]);

  // Y-axis scaling
  const allPnLs = points.flatMap((pt) => [pt.expiryPnL, pt.t0PnL]);
  const maxAbsPnL = Math.max(...allPnLs.map((val) => Math.abs(val)), 500);
  const yRange = maxAbsPnL * 1.25;

  const toSvgX = (price: number) => {
    return ((price - rangeMin) / (rangeMax - rangeMin)) * plotWidth;
  };

  const toSvgY = (pnl: number) => {
    // 0 is at middle (plotHeight / 2)
    const mid = plotHeight / 2;
    return mid - (pnl / yRange) * mid;
  };

  // SVG Paths
  const expiryPathD = points.reduce((acc, pt, idx) => {
    const x = toSvgX(pt.price);
    const y = toSvgY(pt.expiryPnL);
    return idx === 0 ? `M ${x} ${y}` : `${acc} L ${x} ${y}`;
  }, '');

  const t0PathD = points.reduce((acc, pt, idx) => {
    const x = toSvgX(pt.price);
    const y = toSvgY(pt.t0PnL);
    return idx === 0 ? `M ${x} ${y}` : `${acc} L ${x} ${y}`;
  }, '');

  const zeroY = toSvgY(0);
  const spotX = toSvgX(spotPrice);
  const simX = toSvgX(simulatedPrice);
  const simExpiryPnL = Math.round(calculateExpiryPnL(simulatedPrice, legs));
  const simT0PnL = Math.round(calculateT0PnL(simulatedPrice, legs, dte));

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#00F0A0]" />
            <h2 className="text-sm font-bold text-white tracking-wide">
              Institutional Options Strategy Payoff Diagram & PnL Visualizer
            </h2>
          </div>
          <p className="text-[11px] text-slate-400 mt-0.5">
            Real-Time Analytical Black-Scholes Greeks • Multi-Leg Payoff Simulation • Expiry & T+0 Curve
          </p>
        </div>

        {/* Asset Switcher */}
        <div className="flex items-center gap-2 bg-[#121722] p-1 rounded border border-slate-800">
          {(['BTC', 'ETH', 'NIFTY'] as const).map((asset) => (
            <button
              key={asset}
              onClick={() => handleAssetChange(asset)}
              className={`px-3 py-1 rounded text-xs font-bold transition-colors ${
                selectedAsset === asset
                  ? 'bg-[#00F0A0] text-black shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              {asset} ({asset === 'BTC' ? '$65K' : asset === 'ETH' ? '$3.5K' : '₹24.5K'})
            </button>
          ))}
        </div>
      </div>

      {/* Preset Strategy Buttons */}
      <div className="flex flex-wrap items-center gap-2 text-[11px]">
        <span className="text-slate-400 font-bold">Presets:</span>
        {[
          { label: 'Bull Call Spread', key: 'BULL_CALL_SPREAD' },
          { label: 'Bear Put Spread', key: 'BEAR_PUT_SPREAD' },
          { label: 'Iron Condor', key: 'IRON_CONDOR' },
          { label: 'Straddle', key: 'STRADDLE' },
          { label: 'Strangle', key: 'STRANGLE' },
          { label: 'Long Call', key: 'LONG_CALL' },
        ].map((p) => (
          <button
            key={p.key}
            onClick={() => applyPreset(p.key)}
            className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 hover:border-[#00F0A0] text-slate-300 hover:text-white transition-colors"
          >
            {p.label}
          </button>
        ))}
      </div>

      {/* Key Metric Badges */}
      <div className="grid grid-cols-2 md:grid-cols-6 gap-2">
        <div className="p-2.5 rounded bg-[#121722] border border-slate-800">
          <div className="text-slate-400 text-[10px]">NET PREMIUM</div>
          <div className={`text-sm font-bold mt-0.5 ${strategySummary.netPremium <= 0 ? 'text-[#00F0A0]' : 'text-amber-400'}`}>
            {strategySummary.netPremium <= 0 ? `Credit: +$${Math.abs(strategySummary.netPremium)}` : `Debit: -$${strategySummary.netPremium}`}
          </div>
        </div>

        <div className="p-2.5 rounded bg-[#121722] border border-slate-800">
          <div className="text-slate-400 text-[10px]">MAX PROFIT</div>
          <div className="text-sm font-bold text-[#00F0A0] mt-0.5">
            {typeof strategySummary.maxProfit === 'number' ? `+$${strategySummary.maxProfit}` : strategySummary.maxProfit}
          </div>
        </div>

        <div className="p-2.5 rounded bg-[#121722] border border-slate-800">
          <div className="text-slate-400 text-[10px]">MAX LOSS</div>
          <div className="text-sm font-bold text-rose-400 mt-0.5">
            {typeof strategySummary.maxLoss === 'number' ? `-$${Math.abs(strategySummary.maxLoss)}` : strategySummary.maxLoss}
          </div>
        </div>

        <div className="p-2.5 rounded bg-[#121722] border border-slate-800">
          <div className="text-slate-400 text-[10px]">BREAKEVENS</div>
          <div className="text-xs font-bold text-slate-200 mt-1 truncate">
            {strategySummary.breakevens.length > 0
              ? strategySummary.breakevens.map((b) => `$${b}`).join(', ')
              : 'None'}
          </div>
        </div>

        <div className="p-2.5 rounded bg-[#121722] border border-slate-800">
          <div className="text-slate-400 text-[10px]">RISK / REWARD</div>
          <div className="text-sm font-bold text-indigo-300 mt-0.5">
            {strategySummary.riskRewardRatio}
          </div>
        </div>

        <div className="p-2.5 rounded bg-[#121722] border border-slate-800">
          <div className="text-slate-400 text-[10px]">NET DELTA / THETA</div>
          <div className="text-xs font-bold text-slate-200 mt-1">
            Δ {strategySummary.greeks.delta} | Θ {strategySummary.greeks.theta}/d
          </div>
        </div>
      </div>

      {/* Interactive Payoff Canvas */}
      <div className="p-4 rounded-lg bg-[#0E121B] border border-slate-800 relative">
        <div className="flex justify-between items-center mb-2">
          <div className="flex items-center gap-4 text-[11px]">
            <span className="flex items-center gap-1.5 text-cyan-400">
              <span className="w-3 h-0.5 bg-cyan-400 inline-block" />
              At Expiry Payoff
            </span>
            <span className="flex items-center gap-1.5 text-amber-400">
              <span className="w-3 h-0.5 border-t-2 border-dashed border-amber-400 inline-block" />
              T+0 Curve ({dte} DTE)
            </span>
            <span className="flex items-center gap-1.5 text-slate-400">
              <span className="w-2 h-2 bg-slate-500 rounded-full inline-block" />
              Spot: ${spotPrice.toLocaleString()}
            </span>
          </div>
          <div className="text-[11px] text-slate-300 font-bold">
            Simulated Spot: <span className="text-[#00F0A0]">${simulatedPrice.toLocaleString()}</span>
            {' '}| Expiry PnL: <span className={simExpiryPnL >= 0 ? 'text-[#00F0A0]' : 'text-rose-400'}>
              {simExpiryPnL >= 0 ? `+$${simExpiryPnL}` : `-$${Math.abs(simExpiryPnL)}`}
            </span>
            {' '}| T+0 PnL: <span className={simT0PnL >= 0 ? 'text-amber-400' : 'text-rose-400'}>
              {simT0PnL >= 0 ? `+$${simT0PnL}` : `-$${Math.abs(simT0PnL)}`}
            </span>
          </div>
        </div>

        {/* SVG Curve */}
        <div className="w-full overflow-x-auto">
          <svg
            viewBox={`0 0 ${plotWidth} ${plotHeight}`}
            className="w-full h-56 select-none"
          >
            <defs>
              <linearGradient id="profitFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#00F0A0" stopOpacity="0.15" />
                <stop offset="100%" stopColor="#00F0A0" stopOpacity="0.0" />
              </linearGradient>
              <linearGradient id="lossFill" x1="0" y1="1" x2="0" y2="0">
                <stop offset="0%" stopColor="#FF4D4D" stopOpacity="0.15" />
                <stop offset="100%" stopColor="#FF4D4D" stopOpacity="0.0" />
              </linearGradient>
            </defs>

            {/* Grid & Zero Axis */}
            <line
              x1="0"
              y1={zeroY}
              x2={plotWidth}
              y2={zeroY}
              stroke="#334155"
              strokeWidth="1.5"
            />
            <text x="5" y={zeroY - 4} fill="#64748b" fontSize="10">
              $0 PnL
            </text>

            {/* Current Spot Line */}
            <line
              x1={spotX}
              y1="0"
              x2={spotX}
              y2={plotHeight}
              stroke="#64748b"
              strokeDasharray="3,3"
              strokeWidth="1"
            />
            <text x={spotX + 4} y="15" fill="#94a3b8" fontSize="9">
              Spot ${spotPrice}
            </text>

            {/* Simulated Price Line */}
            <line
              x1={simX}
              y1="0"
              x2={simX}
              y2={plotHeight}
              stroke="#00F0A0"
              strokeWidth="1.5"
            />
            <circle cx={simX} cy={toSvgY(simExpiryPnL)} r="4" fill="#00F0A0" />
            <circle cx={simX} cy={toSvgY(simT0PnL)} r="3" fill="#f59e0b" />

            {/* Payoff Curves */}
            <path
              d={expiryPathD}
              fill="none"
              stroke="#38bdf8"
              strokeWidth="2.5"
            />
            <path
              d={t0PathD}
              fill="none"
              stroke="#f59e0b"
              strokeDasharray="4,4"
              strokeWidth="1.8"
            />
          </svg>
        </div>

        {/* Spot Scrubber Slider */}
        <div className="mt-3 flex items-center gap-4 text-slate-400">
          <span className="text-[10px] w-28">Underlying Scrubber:</span>
          <input
            type="range"
            min={rangeMin}
            max={rangeMax}
            step={(rangeMax - rangeMin) / 100}
            value={simulatedPrice}
            onChange={(e) => setSimulatedPrice(Number(e.target.value))}
            className="flex-1 accent-[#00F0A0] cursor-pointer"
          />
          <button
            onClick={() => setSimulatedPrice(spotPrice)}
            className="px-2 py-0.5 bg-[#1e293b] text-slate-300 rounded text-[10px] hover:text-white"
          >
            Reset Spot
          </button>
        </div>
      </div>

      {/* Model Controls: DTE & IV */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 bg-[#121722] p-3 rounded border border-slate-800 text-[11px]">
        <div className="space-y-1">
          <div className="flex justify-between">
            <span className="text-slate-400">Days to Expiration (DTE):</span>
            <span className="text-white font-bold">{dte} Days</span>
          </div>
          <input
            type="range"
            min="0"
            max="90"
            value={dte}
            onChange={(e) => setDte(Number(e.target.value))}
            className="w-full accent-[#00F0A0] cursor-pointer"
          />
        </div>

        <div className="space-y-1">
          <div className="flex justify-between">
            <span className="text-slate-400">Implied Volatility (IV):</span>
            <span className="text-white font-bold">{globalIv}%</span>
          </div>
          <input
            type="range"
            min="10"
            max="150"
            value={globalIv}
            onChange={(e) => {
              const iv = Number(e.target.value);
              setGlobalIv(iv);
              setLegs(legs.map((l) => ({ ...l, ivPct: iv })));
            }}
            className="w-full accent-[#00F0A0] cursor-pointer"
          />
        </div>
      </div>

      {/* Legs Table */}
      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <span className="font-bold text-slate-300">Strategy Legs ({legs.length})</span>
          <button
            onClick={addLeg}
            className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a] transition-colors text-[11px]"
          >
            + Add Option Leg
          </button>
        </div>

        <div className="border border-slate-800 rounded overflow-hidden">
          <table className="w-full text-left text-[11px]">
            <thead className="bg-[#121722] text-slate-400 border-b border-slate-800">
              <tr>
                <th className="p-2">Action</th>
                <th className="p-2">Type</th>
                <th className="p-2">Strike ($)</th>
                <th className="p-2">Premium ($)</th>
                <th className="p-2">Contracts</th>
                <th className="p-2">IV %</th>
                <th className="p-2 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800 bg-[#0E121B]">
              {legs.map((leg) => (
                <tr key={leg.id} className="hover:bg-slate-800/40">
                  <td className="p-2">
                    <select
                      value={leg.action}
                      onChange={(e) => updateLeg(leg.id, { action: e.target.value as PositionAction })}
                      className="bg-[#181F2C] border border-slate-700 rounded px-1.5 py-0.5 text-slate-200"
                    >
                      <option value="BUY">BUY</option>
                      <option value="SELL">SELL</option>
                    </select>
                  </td>
                  <td className="p-2">
                    <select
                      value={leg.type}
                      onChange={(e) => updateLeg(leg.id, { type: e.target.value as OptionType })}
                      className="bg-[#181F2C] border border-slate-700 rounded px-1.5 py-0.5 text-slate-200"
                    >
                      <option value="CALL">CALL</option>
                      <option value="PUT">PUT</option>
                    </select>
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      value={leg.strike}
                      onChange={(e) => updateLeg(leg.id, { strike: Number(e.target.value) })}
                      className="w-24 bg-[#181F2C] border border-slate-700 rounded px-1.5 py-0.5 text-slate-200"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      value={leg.premium}
                      onChange={(e) => updateLeg(leg.id, { premium: Number(e.target.value) })}
                      className="w-20 bg-[#181F2C] border border-slate-700 rounded px-1.5 py-0.5 text-slate-200"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min="1"
                      value={leg.quantity}
                      onChange={(e) => updateLeg(leg.id, { quantity: Math.max(1, Number(e.target.value)) })}
                      className="w-16 bg-[#181F2C] border border-slate-700 rounded px-1.5 py-0.5 text-slate-200"
                    />
                  </td>
                  <td className="p-2">
                    <span className="text-slate-300">{leg.ivPct}%</span>
                  </td>
                  <td className="p-2 text-right">
                    <button
                      onClick={() => removeLeg(leg.id)}
                      disabled={legs.length <= 1}
                      className="text-rose-400 hover:text-rose-300 disabled:opacity-30 text-[10px]"
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export const WebOptionsStrategyPayoffDiagramPnLVisualizer = OptionsStrategyPayoffVisualizer;
