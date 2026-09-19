import React, { useState } from 'react';
import { PriceBandSnapshot, ManualHaltRequest, MarketLULDState } from '../types/admin';

interface CircuitBreakerControlProps {
  priceBands: PriceBandSnapshot[];
  onTriggerManualHalt: (req: ManualHaltRequest) => void;
  onResumeTrading: (isin: string, symbol: string) => void;
}

export const CircuitBreakerControl: React.FC<CircuitBreakerControlProps> = ({
  priceBands,
  onTriggerManualHalt,
  onResumeTrading,
}) => {
  const [selectedHaltSecurity, setSelectedHaltSecurity] = useState<PriceBandSnapshot | null>(null);
  const [haltReason, setHaltReason] = useState('');
  const [durationMinutes, setDurationMinutes] = useState(15);

  const getLULDStateBadge = (state: MarketLULDState) => {
    switch (state) {
      case 'REGULAR':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-emerald-900/40 text-emerald-400 border border-emerald-800">REGULAR CONTINUOUS</span>;
      case 'STRADDLE':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-yellow-900/40 text-yellow-400 border border-yellow-800">STRADDLE STATE</span>;
      case 'LIMIT_STATE':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-orange-900/40 text-orange-400 border border-orange-800">LIMIT STATE (15s GRACE)</span>;
      case 'PAUSED_CALL_AUCTION':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-purple-900/40 text-purple-400 border border-purple-800">CALL AUCTION</span>;
      case 'HALTED':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-red-900/40 text-red-400 border border-red-800">MARKET HALTED</span>;
      case 'RESUMING':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-blue-900/40 text-blue-400 border border-blue-800">RESUMING (PRICE DISCOVERY)</span>;
    }
  };

  const handleHaltSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedHaltSecurity) return;

    onTriggerManualHalt({
      isin: selectedHaltSecurity.isin,
      symbol: selectedHaltSecurity.symbol,
      reason: haltReason,
      durationMinutes,
      makerAdminId: 'ADMIN_CHIEF_SURVEILLANCE_01',
    });

    setSelectedHaltSecurity(null);
    setHaltReason('');
  };

  return (
    <div className="bg-[#121212] text-white p-6 rounded-xl border border-gray-800">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">LULD Price Bands & Circuit Breaker Dashboard</h2>
          <p className="text-gray-400 text-sm mt-1">Real-time Dynamic Price Collars, Automated Limit States & Emergency Market Halts</p>
        </div>
      </div>

      <div className="overflow-x-auto rounded-lg border border-gray-800 mb-6">
        <table className="w-full text-left text-sm">
          <thead className="bg-[#1A1A1A] text-gray-400 font-mono text-xs uppercase">
            <tr>
              <th className="p-3">Symbol</th>
              <th className="p-3">Tier</th>
              <th className="p-3">Reference Price</th>
              <th className="p-3">Lower Band (-5%)</th>
              <th className="p-3">Current Price</th>
              <th className="p-3">Upper Band (+5%)</th>
              <th className="p-3">LULD State</th>
              <th className="p-3">Controls</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {priceBands.map(pb => (
              <tr key={pb.isin} className="hover:bg-gray-850 transition">
                <td className="p-3 font-mono font-bold text-white">{pb.symbol}</td>
                <td className="p-3 font-mono text-xs text-gray-400">{pb.tier}</td>
                <td className="p-3 font-mono">₹{(pb.referencePricePaise / 100).toFixed(2)}</td>
                <td className="p-3 font-mono text-red-400">₹{(pb.lowerPriceBandPaise / 100).toFixed(2)}</td>
                <td className="p-3 font-mono font-bold text-emerald-400">₹{(pb.currentPricePaise / 100).toFixed(2)}</td>
                <td className="p-3 font-mono text-emerald-400">₹{(pb.upperPriceBandPaise / 100).toFixed(2)}</td>
                <td className="p-3">{getLULDStateBadge(pb.currentState)}</td>
                <td className="p-3 flex gap-2">
                  {pb.currentState === 'HALTED' ? (
                    <button
                      onClick={() => onResumeTrading(pb.isin, pb.symbol)}
                      className="px-3 py-1 bg-emerald-700 hover:bg-emerald-600 text-white rounded text-xs font-semibold"
                    >
                      Resume Auction
                    </button>
                  ) : (
                    <button
                      onClick={() => setSelectedHaltSecurity(pb)}
                      className="px-3 py-1 bg-red-700 hover:bg-red-600 text-white rounded text-xs font-semibold"
                    >
                      Emergency Halt
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Manual Halt Modal */}
      {selectedHaltSecurity && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
          <div className="bg-[#1A1A1A] border border-red-800 rounded-xl max-w-lg w-full p-6 text-white shadow-2xl">
            <h3 className="text-xl font-bold text-red-400 mb-2">Trigger Emergency Trading Halt</h3>
            <p className="text-gray-400 text-sm mb-4">You are triggering a mandatory market halt for <span className="text-white font-mono font-bold">{selectedHaltSecurity.symbol}</span> ({selectedHaltSecurity.isin}).</p>

            <form onSubmit={handleHaltSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-300 mb-1">Regulatory & Risk Justification</label>
                <textarea
                  rows={3}
                  value={haltReason}
                  onChange={e => setHaltReason(e.target.value)}
                  placeholder="Specify abnormal order flow, volatility spike, or regulatory directive..."
                  className="w-full bg-[#121212] border border-gray-700 rounded p-2 text-sm text-white"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-300 mb-1">Halt Duration (Minutes)</label>
                <input
                  type="number"
                  value={durationMinutes}
                  onChange={e => setDurationMinutes(Number(e.target.value))}
                  min={5}
                  max={120}
                  className="w-full bg-[#121212] border border-gray-700 rounded p-2 text-sm text-white font-mono"
                />
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setSelectedHaltSecurity(null)}
                  className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded text-sm"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-red-600 hover:bg-red-500 text-white rounded text-sm font-bold"
                >
                  Broadcast Halt Order
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
