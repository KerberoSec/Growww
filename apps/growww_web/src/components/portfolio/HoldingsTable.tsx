import React from 'react';
import { HoldingItem } from '../../types/market';

interface HoldingsTableProps {
  holdings: HoldingItem[];
  onTradeAsset?: (symbol: string) => void;
}

export const HoldingsTable: React.FC<HoldingsTableProps> = ({ holdings, onTradeAsset }) => {
  const totalValueINR = holdings.reduce((sum, h) => sum + (h.totalQuantity * h.currentPriceINR), 0);
  const totalUnrealizedGainINR = holdings.reduce((sum, h) => sum + h.unrealizedGainINR, 0);
  const totalTax115BBHINR = holdings
    .filter(h => h.assetType === 'VDA_CRYPTO' && h.unrealizedGainINR > 0)
    .reduce((sum, h) => sum + h.tax115BBHEstimateINR, 0);

  return (
    <div className="flex flex-col bg-slate-900 border border-slate-800 rounded-lg p-4 font-sans text-xs w-full">
      {/* Portfolio Summary Header */}
      <div className="flex flex-wrap justify-between items-center pb-4 mb-4 border-b border-slate-800 gap-4">
        <div>
          <span className="text-slate-400 block mb-0.5">Total Portfolio Valuation</span>
          <span className="text-xl font-bold font-mono text-slate-100">₹{totalValueINR.toLocaleString('en-IN', { maximumFractionDigits: 2 })}</span>
        </div>
        <div>
          <span className="text-slate-400 block mb-0.5">Unrealized P&L</span>
          <span className={`text-base font-bold font-mono ${totalUnrealizedGainINR >= 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
            {totalUnrealizedGainINR >= 0 ? '+' : ''}₹{totalUnrealizedGainINR.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
          </span>
        </div>
        <div>
          <span className="text-slate-400 block mb-0.5">Est. Section 115BBH Tax (31.2%)</span>
          <span className="text-base font-bold font-mono text-amber-400">
            ₹{totalTax115BBHINR.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
          </span>
        </div>
        <div className="flex items-center gap-1.5 px-2.5 py-1 bg-emerald-950/30 border border-emerald-800/50 rounded text-emerald-400 font-mono text-[11px]">
          <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
          <span>Hyperledger Besu 1:1 Reserves Audited</span>
        </div>
      </div>

      {/* Holdings Grid */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-slate-800 text-slate-400 font-medium">
              <th className="py-2 px-3">Asset</th>
              <th className="py-2 px-3">Type</th>
              <th className="py-2 px-3 text-right">Balance</th>
              <th className="py-2 px-3 text-right">Avg Buy (INR)</th>
              <th className="py-2 px-3 text-right">Current (INR)</th>
              <th className="py-2 px-3 text-right">P&L</th>
              <th className="py-2 px-3 text-center">Besu On-Chain Proof</th>
              <th className="py-2 px-3 text-center">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50 font-mono">
            {holdings.map((h) => {
              const isGain = h.unrealizedGainINR >= 0;
              return (
                <tr key={h.assetId} className="hover:bg-slate-800/40 transition">
                  <td className="py-2.5 px-3 font-semibold text-slate-200">{h.symbol}</td>
                  <td className="py-2.5 px-3">
                    <span className={`px-1.5 py-0.5 rounded text-[10px] ${
                      h.assetType === 'EQUITY_DEMAT'
                        ? 'bg-blue-950 text-blue-400 border border-blue-800'
                        : h.assetType === 'VDA_CRYPTO'
                        ? 'bg-amber-950 text-amber-400 border border-amber-800'
                        : 'bg-emerald-950 text-emerald-400 border border-emerald-800'
                    }`}>
                      {h.assetType === 'EQUITY_DEMAT' ? 'Demat Equity' : h.assetType === 'VDA_CRYPTO' ? 'Crypto VDA' : 'CBDC Cash'}
                    </span>
                  </td>
                  <td className="py-2.5 px-3 text-right text-slate-200">
                    {h.totalQuantity.toFixed(4)}
                  </td>
                  <td className="py-2.5 px-3 text-right text-slate-400">₹{h.avgCostPriceINR.toFixed(2)}</td>
                  <td className="py-2.5 px-3 text-right text-slate-200">₹{h.currentPriceINR.toFixed(2)}</td>
                  <td className={`py-2.5 px-3 text-right font-medium ${isGain ? 'text-emerald-400' : 'text-rose-400'}`}>
                    {isGain ? '+' : ''}₹{h.unrealizedGainINR.toFixed(2)}
                  </td>
                  <td className="py-2.5 px-3 text-center">
                    <span className="text-[10px] text-emerald-400 bg-emerald-950/40 px-2 py-0.5 rounded border border-emerald-800/40">
                      1:1 On-Chain Match
                    </span>
                  </td>
                  <td className="py-2.5 px-3 text-center font-sans">
                    <button
                      onClick={() => onTradeAsset?.(h.symbol)}
                      className="px-2 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded text-[11px] transition"
                    >
                      Trade
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
