import React, { useState } from 'react';

export const ResponsibleTradingControls: React.FC = () => {
  const [dailyLossLimitINR, setDailyLossLimitINR] = useState<number>(50000);
  const [maxLeverageCap, setMaxLeverageCap] = useState<number>(5);
  const [selfExclusionDays, setSelfExclusionDays] = useState<number>(0);
  const [saved, setSaved] = useState(false);

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-2xl mx-auto font-sans">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-lg bg-amber-950 flex items-center justify-center text-amber-400 font-bold border border-amber-800">
          ⚖
        </div>
        <div>
          <h2 className="text-xl font-bold">Responsible Trading & Circuit Breaker Limits</h2>
          <p className="text-gray-400 text-xs">Self-imposed risk boundaries, position caps, and voluntary cooling-off periods.</p>
        </div>
      </div>

      <form onSubmit={handleSave} className="space-y-4 mb-6">
        <div className="bg-[#141824] p-4 rounded-lg border border-gray-800">
          <label className="block text-xs font-semibold text-gray-300 mb-1">Max Daily Realized Loss Limit (INR)</label>
          <input
            type="number"
            value={dailyLossLimitINR}
            onChange={e => setDailyLossLimitINR(Number(e.target.value))}
            className="w-full bg-[#0B0E14] border border-gray-700 rounded-lg p-2.5 text-xs text-white font-mono"
          />
          <span className="text-[10px] text-gray-500 font-mono mt-1 block">Trading locks automatically if realized losses reach this limit.</span>
        </div>

        <div className="bg-[#141824] p-4 rounded-lg border border-gray-800">
          <label className="block text-xs font-semibold text-gray-300 mb-1">Max Leverage Ceiling ({maxLeverageCap}x)</label>
          <input
            type="range"
            min={1}
            max={20}
            value={maxLeverageCap}
            onChange={e => setMaxLeverageCap(Number(e.target.value))}
            className="w-full accent-emerald-500"
          />
          <div className="flex justify-between text-[10px] font-mono text-gray-400 mt-1">
            <span>1x (No Leverage)</span>
            <span>5x</span>
            <span>10x</span>
            <span>20x (Max SEBI Regulated)</span>
          </div>
        </div>

        <div className="bg-red-950/20 p-4 rounded-lg border border-red-900/40">
          <label className="block text-xs font-semibold text-red-400 mb-1">Voluntary Self-Exclusion Cooling-Off (Days)</label>
          <select
            value={selfExclusionDays}
            onChange={e => setSelfExclusionDays(Number(e.target.value))}
            className="w-full bg-[#0B0E14] border border-gray-700 rounded-lg p-2.5 text-xs text-white font-mono"
          >
            <option value={0}>No Active Exclusion</option>
            <option value={1}>24 Hours Cooling-Off</option>
            <option value={7}>7 Days Break</option>
            <option value={30}>30 Days Self-Exclusion</option>
            <option value={180}>6 Months Strict Lock</option>
          </select>
          <span className="text-[10px] text-gray-400 mt-1 block">During exclusion, order placement and deposits are disabled; withdrawals remain accessible.</span>
        </div>

        <button
          type="submit"
          className="w-full py-3 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition"
        >
          {saved ? '✓ Protection Limits Committed' : 'Save Risk Safeguards'}
        </button>
      </form>
    </div>
  );
};
