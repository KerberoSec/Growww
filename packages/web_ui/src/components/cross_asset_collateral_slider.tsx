import React, { useState, useMemo } from 'react';

export interface CollateralAsset {
  id: string;
  symbol: string;
  name: string;
  category: 'FIAT' | 'CRYPTO' | 'GOV_BOND' | 'EQUITY';
  nominalValueUsd: number;
  haircutPct: number;
  currency: string;
  rawAmount: string;
}

export interface MarginBucketAllocation {
  id: string;
  name: string;
  color: string;
  weightPct: number;
  imRequiredUsd: number;
  mmRequiredUsd: number;
}

export function CrossAssetCollateralSlider() {
  // Collateral Assets in multi-asset treasury
  const [assets, setAssets] = useState<CollateralAsset[]>([
    { id: '1', symbol: 'INR_CASH', name: 'INR Cash Balance (SEBI Segregated)', category: 'FIAT', nominalValueUsd: 30000, haircutPct: 0, currency: 'INR', rawAmount: '₹2,500,000' },
    { id: '2', symbol: 'USDC', name: 'USD Coin (ERC-20 Institutional)', category: 'CRYPTO', nominalValueUsd: 50000, haircutPct: 2, currency: 'USDC', rawAmount: '50,000 USDC' },
    { id: '3', symbol: 'G_SEC', name: '7.18% GS 2033 Sovereign Bond', category: 'GOV_BOND', nominalValueUsd: 20000, haircutPct: 10, currency: 'INR', rawAmount: '₹1,660,000' },
    { id: '4', symbol: 'DEMAT_EQUITY', name: 'Demat Bluechip Shares (Pledged)', category: 'EQUITY', nominalValueUsd: 40000, haircutPct: 20, currency: 'INR', rawAmount: '₹3,320,000' },
    { id: '5', symbol: 'WBTC', name: 'Native Bitcoin (Pledged Vault)', category: 'CRYPTO', nominalValueUsd: 65000, haircutPct: 25, currency: 'BTC', rawAmount: '1.00 BTC' },
    { id: '6', symbol: 'ST_ETH', name: 'Lido Staked ETH (stETH)', category: 'CRYPTO', nominalValueUsd: 35000, haircutPct: 30, currency: 'ETH', rawAmount: '10.0 stETH' },
  ]);

  // Trading / Hedging Segments
  const [buckets, setBuckets] = useState<MarginBucketAllocation[]>([
    { id: 'fno', name: 'Equity & Index F&O', color: '#38bdf8', weightPct: 40, imRequiredUsd: 45000, mmRequiredUsd: 32000 },
    { id: 'crypto_perp', name: 'Crypto Perpetuals & Options', color: '#00F0A0', weightPct: 30, imRequiredUsd: 38000, mmRequiredUsd: 28000 },
    { id: 'rwa_lending', name: 'RWA Yield & Repo Credit', color: '#a855f7', weightPct: 15, imRequiredUsd: 12000, mmRequiredUsd: 9000 },
    { id: 'free_buffer', name: 'Unencumbered Buffer', color: '#94a3b8', weightPct: 15, imRequiredUsd: 0, mmRequiredUsd: 0 },
  ]);

  // Calculations
  const totals = useMemo(() => {
    const totalNominal = assets.reduce((sum, a) => sum + a.nominalValueUsd, 0);
    const totalEffective = assets.reduce(
      (sum, a) => sum + a.nominalValueUsd * (1 - a.haircutPct / 100),
      0
    );
    const totalHaircutDeduction = totalNominal - totalEffective;
    const avgHaircutPct = (totalHaircutDeduction / totalNominal) * 100;

    const totalImRequired = buckets.reduce((sum, b) => sum + b.imRequiredUsd, 0);
    const totalMmRequired = buckets.reduce((sum, b) => sum + b.mmRequiredUsd, 0);

    const marginUtilizationPct = totalEffective > 0 ? (totalImRequired / totalEffective) * 100 : 0;
    const capitalEfficiency = totalImRequired > 0 ? totalEffective / totalImRequired : 1.0;

    return {
      totalNominal,
      totalEffective,
      totalHaircutDeduction,
      avgHaircutPct,
      totalImRequired,
      totalMmRequired,
      marginUtilizationPct,
      capitalEfficiency,
    };
  }, [assets, buckets]);

  // Slider change for bucket weight
  const handleSliderChange = (changedId: string, newWeight: number) => {
    const clampedNew = Math.max(0, Math.min(100, newWeight));
    const otherBuckets = buckets.filter((b) => b.id !== changedId);
    const currentOthersSum = otherBuckets.reduce((sum, b) => sum + b.weightPct, 0);

    const targetOthersSum = 100 - clampedNew;

    let updatedBuckets: MarginBucketAllocation[];
    if (currentOthersSum === 0) {
      const evenShare = targetOthersSum / otherBuckets.length;
      updatedBuckets = buckets.map((b) =>
        b.id === changedId ? { ...b, weightPct: clampedNew } : { ...b, weightPct: Math.round(evenShare) }
      );
    } else {
      const scale = targetOthersSum / currentOthersSum;
      updatedBuckets = buckets.map((b) => {
        if (b.id === changedId) return { ...b, weightPct: clampedNew };
        return { ...b, weightPct: Math.round(b.weightPct * scale) };
      });
    }

    // Fix rounding discrepancies
    const sumTotal = updatedBuckets.reduce((sum, b) => sum + b.weightPct, 0);
    if (sumTotal !== 100 && updatedBuckets.length > 0) {
      const diff = 100 - sumTotal;
      const lastIndex = updatedBuckets.findIndex((b) => b.id !== changedId);
      if (lastIndex !== -1) {
        updatedBuckets[lastIndex].weightPct += diff;
      }
    }

    setBuckets(updatedBuckets);
  };

  // Auto-optimize allocation
  const handleAutoOptimize = () => {
    // Optimizes weighting to match IM requirement proportions + 15% safety buffer
    const totalIm = buckets.reduce((sum, b) => sum + b.imRequiredUsd, 0);
    if (totalIm === 0) return;

    const availableEffective = totals.totalEffective;
    const bufferPct = 20; // 20% free buffer
    const allocatablePct = 80;

    const optimized = buckets.map((b) => {
      if (b.id === 'free_buffer') {
        return { ...b, weightPct: bufferPct };
      }
      const proportional = Math.round((b.imRequiredUsd / totalIm) * allocatablePct);
      return { ...b, weightPct: proportional };
    });

    const sumTotal = optimized.reduce((sum, b) => sum + b.weightPct, 0);
    if (sumTotal !== 100) {
      const buffer = optimized.find((b) => b.id === 'free_buffer');
      if (buffer) buffer.weightPct += 100 - sumTotal;
    }

    setBuckets(optimized);
  };

  const getStatusColor = (pct: number) => {
    if (pct < 70) return 'text-[#00F0A0]';
    if (pct < 85) return 'text-amber-400';
    return 'text-rose-400';
  };

  const getStatusBadge = (pct: number) => {
    if (pct < 70) return { label: 'HEALTHY BUFFER', bg: 'bg-[#00F0A0]/10 text-[#00F0A0] border-[#00F0A0]/30' };
    if (pct < 85) return { label: 'MARGIN SQUEEZE WARNING', bg: 'bg-amber-400/10 text-amber-400 border-amber-400/30' };
    return { label: 'CRITICAL LIQUIDATION RISK', bg: 'bg-rose-400/10 text-rose-400 border-rose-400/30' };
  };

  const statusBadge = getStatusBadge(totals.marginUtilizationPct);

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#00F0A0]" />
            <h2 className="text-sm font-bold text-white tracking-wide">
              Cross-Asset Collateral Allocation Slider & Margin Optimizer
            </h2>
          </div>
          <p className="text-[11px] text-slate-400 mt-0.5">
            Cross-Margining Pool • Haircut-Adjusted Purchasing Power • Multi-Tranche Risk Segregation
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className={`px-2.5 py-1 rounded border text-[11px] font-bold ${statusBadge.bg}`}>
            {statusBadge.label}
          </span>
          <button
            onClick={handleAutoOptimize}
            className="px-3 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a] transition-colors text-[11px]"
          >
            ⚡ Auto-Balance Allocation
          </button>
        </div>
      </div>

      {/* Primary KPI Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div className="p-3 rounded bg-[#121722] border border-slate-800">
          <span className="text-slate-400 text-[10px]">TOTAL NOMINAL ASSETS</span>
          <div className="text-base font-bold text-white mt-1">
            ${totals.totalNominal.toLocaleString()}
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">
            Haircut Drag: -${Math.round(totals.totalHaircutDeduction).toLocaleString()} ({totals.avgHaircutPct.toFixed(1)}%)
          </div>
        </div>

        <div className="p-3 rounded bg-[#121722] border border-slate-800">
          <span className="text-slate-400 text-[10px]">EFFECTIVE COLLATERAL</span>
          <div className="text-base font-bold text-[#00F0A0] mt-1">
            ${Math.round(totals.totalEffective).toLocaleString()}
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">
            Haircut-adjusted purchasing power
          </div>
        </div>

        <div className="p-3 rounded bg-[#121722] border border-slate-800">
          <span className="text-slate-400 text-[10px]">MARGIN UTILIZATION</span>
          <div className={`text-base font-bold mt-1 ${getStatusColor(totals.marginUtilizationPct)}`}>
            {totals.marginUtilizationPct.toFixed(1)}%
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">
            Total IM: ${totals.totalImRequired.toLocaleString()}
          </div>
        </div>

        <div className="p-3 rounded bg-[#121722] border border-slate-800">
          <span className="text-slate-400 text-[10px]">CAPITAL EFFICIENCY</span>
          <div className="text-base font-bold text-indigo-300 mt-1">
            {totals.capitalEfficiency.toFixed(2)}x
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">
            Collateral to Margin Requirement
          </div>
        </div>
      </div>

      {/* Allocation Stack Bar */}
      <div className="p-3 rounded bg-[#0E121B] border border-slate-800 space-y-2">
        <div className="flex justify-between items-center text-[11px]">
          <span className="text-slate-300 font-bold">Collateral Allocation Distribution</span>
          <span className="text-slate-400">Total Allocated: 100%</span>
        </div>
        <div className="w-full h-4 bg-slate-800 rounded-full overflow-hidden flex">
          {buckets.map((b) => (
            <div
              key={b.id}
              style={{ width: `${b.weightPct}%`, backgroundColor: b.color }}
              className="h-full transition-all duration-200"
              title={`${b.name}: ${b.weightPct}%`}
            />
          ))}
        </div>
        <div className="flex flex-wrap gap-4 pt-1">
          {buckets.map((b) => (
            <div key={b.id} className="flex items-center gap-1.5 text-[11px]">
              <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: b.color }} />
              <span className="text-slate-300">{b.name}:</span>
              <span className="font-bold text-white">{b.weightPct}%</span>
              <span className="text-slate-400 text-[10px]">
                (${Math.round((totals.totalEffective * b.weightPct) / 100).toLocaleString()})
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Dynamic Sliders Section */}
      <div className="space-y-3">
        <h3 className="font-bold text-slate-200">Rebalance Segment Margin Sliders</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {buckets.map((bucket) => {
            const allocatedUsd = Math.round((totals.totalEffective * bucket.weightPct) / 100);
            const isDeficit = allocatedUsd < bucket.imRequiredUsd && bucket.imRequiredUsd > 0;
            return (
              <div
                key={bucket.id}
                className="p-3 rounded bg-[#121722] border border-slate-800 space-y-2"
              >
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: bucket.color }} />
                    <span className="font-bold text-white">{bucket.name}</span>
                  </div>
                  <span className="font-bold text-sm" style={{ color: bucket.color }}>
                    {bucket.weightPct}%
                  </span>
                </div>

                <input
                  type="range"
                  min="0"
                  max="100"
                  value={bucket.weightPct}
                  onChange={(e) => handleSliderChange(bucket.id, Number(e.target.value))}
                  className="w-full cursor-pointer accent-[#00F0A0]"
                />

                <div className="flex justify-between text-[10px] text-slate-400">
                  <span>Effective Collateral: ${allocatedUsd.toLocaleString()}</span>
                  <span>
                    IM Req: ${bucket.imRequiredUsd.toLocaleString()}
                    {isDeficit && (
                      <span className="ml-1 text-rose-400 font-bold">(DEFICIT!)</span>
                    )}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Multi-Asset Collateral Inventory Table */}
      <div className="space-y-2">
        <span className="font-bold text-slate-300">Pledged Collateral Asset Pool & Haircut Matrix</span>
        <div className="border border-slate-800 rounded overflow-hidden">
          <table className="w-full text-left text-[11px]">
            <thead className="bg-[#121722] text-slate-400 border-b border-slate-800">
              <tr>
                <th className="p-2.5">Asset</th>
                <th className="p-2.5">Category</th>
                <th className="p-2.5">Holding Amount</th>
                <th className="p-2.5">Nominal Value (USD)</th>
                <th className="p-2.5">Haircut %</th>
                <th className="p-2.5">Effective Margin (USD)</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800 bg-[#0E121B]">
              {assets.map((asset) => {
                const effValue = Math.round(asset.nominalValueUsd * (1 - asset.haircutPct / 100));
                return (
                  <tr key={asset.id} className="hover:bg-slate-800/40">
                    <td className="p-2.5 font-bold text-white">
                      {asset.symbol}
                      <span className="block font-normal text-slate-400 text-[10px]">{asset.name}</span>
                    </td>
                    <td className="p-2.5">
                      <span className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px]">
                        {asset.category}
                      </span>
                    </td>
                    <td className="p-2.5 text-slate-300">{asset.rawAmount}</td>
                    <td className="p-2.5 font-bold text-white">${asset.nominalValueUsd.toLocaleString()}</td>
                    <td className="p-2.5">
                      <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold ${
                        asset.haircutPct === 0
                          ? 'bg-[#00F0A0]/10 text-[#00F0A0]'
                          : asset.haircutPct <= 15
                          ? 'bg-blue-400/10 text-blue-400'
                          : 'bg-amber-400/10 text-amber-400'
                      }`}>
                        {asset.haircutPct}%
                      </span>
                    </td>
                    <td className="p-2.5 font-bold text-[#00F0A0]">
                      ${effValue.toLocaleString()}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export const WebCrossAssetCollateralAllocationSlider = CrossAssetCollateralSlider;
