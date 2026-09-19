import React, { useState } from 'react';

export interface MasterTrader {
  traderId: string;
  name: string;
  strategyName: string;
  winRatePct: number; // e.g. 78.5%
  roi30dPct: number; // e.g. +34.2%
  maxDrawdownPct: number; // e.g. 8.4%
  followersCount: number;
  totalAUM_INR: number;
  profitSharePct: number; // e.g. 10%
  minFollowINR: number;
}

interface CopyTradingMarketplaceProps {
  traders: MasterTrader[];
  onFollowTrader: (traderId: string, investmentINR: number, maxSlippageBps: number) => void;
}

export const CopyTradingMarketplace: React.FC<CopyTradingMarketplaceProps> = ({
  traders,
  onFollowTrader,
}) => {
  const [selectedTrader, setSelectedTrader] = useState<MasterTrader | null>(null);
  const [investmentAmount, setInvestmentAmount] = useState<number>(10000);
  const [maxSlippageBps, setMaxSlippageBps] = useState<number>(50); // 0.50%

  const handleFollow = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTrader) return;
    onFollowTrader(selectedTrader.traderId, investmentAmount, maxSlippageBps);
    setSelectedTrader(null);
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-5xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Copy Trading Marketplace</h2>
          <p className="text-gray-400 text-xs mt-1">
            Mirror verified master traders in real-time with automatic equity-proportional sizing and slippage protection.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {traders.map(t => (
          <div
            key={t.traderId}
            className="bg-[#141824] border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition flex flex-col justify-between"
          >
            <div>
              <div className="flex justify-between items-start mb-3">
                <div>
                  <h3 className="font-bold text-sm text-gray-200">{t.name}</h3>
                  <span className="text-[11px] text-emerald-400 font-mono">{t.strategyName}</span>
                </div>
                <span className="text-xs bg-emerald-950 text-emerald-400 border border-emerald-800 px-2 py-0.5 rounded font-mono font-bold">
                  +{t.roi30dPct.toFixed(1)}% 30D
                </span>
              </div>

              <div className="grid grid-cols-3 gap-2 bg-[#0B0E14] p-3 rounded-lg border border-gray-850 text-center font-mono text-xs mb-4">
                <div>
                  <span className="text-[10px] text-gray-500 block">Win Rate</span>
                  <span className="font-bold text-white">{t.winRatePct.toFixed(1)}%</span>
                </div>
                <div>
                  <span className="text-[10px] text-gray-500 block">Max DD</span>
                  <span className="font-bold text-red-400">{t.maxDrawdownPct.toFixed(1)}%</span>
                </div>
                <div>
                  <span className="text-[10px] text-gray-500 block">Followers</span>
                  <span className="font-bold text-white">{t.followersCount}</span>
                </div>
              </div>

              <div className="text-[11px] text-gray-400 font-mono space-y-1 mb-4">
                <div className="flex justify-between">
                  <span>Strategy AUM:</span>
                  <span className="text-gray-200 font-semibold">₹{t.totalAUM_INR.toLocaleString('en-IN')}</span>
                </div>
                <div className="flex justify-between">
                  <span>Profit Sharing:</span>
                  <span className="text-gray-200 font-semibold">{t.profitSharePct}% of gains</span>
                </div>
                <div className="flex justify-between">
                  <span>Min Capital:</span>
                  <span className="text-gray-200 font-semibold">₹{t.minFollowINR.toLocaleString('en-IN')}</span>
                </div>
              </div>
            </div>

            <button
              onClick={() => setSelectedTrader(t)}
              className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition shadow-lg shadow-emerald-950/40"
            >
              Copy Strategy
            </button>
          </div>
        ))}
      </div>

      {/* Copy Configuration Modal */}
      {selectedTrader && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
          <div className="bg-[#0B0E14] border border-gray-700 rounded-xl max-w-md w-full p-6 text-white shadow-2xl">
            <h3 className="text-lg font-bold mb-1">Mirror {selectedTrader.name}</h3>
            <p className="text-gray-400 text-xs mb-4">Strategy: {selectedTrader.strategyName}</p>

            <form onSubmit={handleFollow} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">Allocated Capital (INR)</label>
                <input
                  type="number"
                  value={investmentAmount}
                  onChange={e => setInvestmentAmount(Number(e.target.value))}
                  min={selectedTrader.minFollowINR}
                  className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2.5 text-xs text-white font-mono"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">Max Slippage Protector (Basis Points)</label>
                <input
                  type="number"
                  value={maxSlippageBps}
                  onChange={e => setMaxSlippageBps(Number(e.target.value))}
                  min={10}
                  max={200}
                  className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2.5 text-xs text-white font-mono"
                  required
                />
                <span className="text-[10px] text-gray-500 font-mono mt-1 block">50 bps = 0.50% max execution variance</span>
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setSelectedTrader(null)}
                  className="px-4 py-2 bg-gray-800 text-gray-300 rounded-lg text-xs"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-bold"
                >
                  Confirm Mirror Allocation
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
