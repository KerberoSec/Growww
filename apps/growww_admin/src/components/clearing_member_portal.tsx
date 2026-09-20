'use client';

import React, { useState } from 'react';

export function ClearingMemberPortal() {
  const [metrics] = useState({
    baseMinimumCapital: '₹10,00,00,000',
    effectiveLiquidNetWorth: '₹84,50,00,000',
    peakMarginUtilizationPct: 48.2,
    sgfContribution: '₹12,40,00,000',
    status: 'ADEQUATE',
  });

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">BASE MINIMUM CAPITAL (BMC)</div>
          <div className="text-xl font-bold font-mono text-white mt-1">{metrics.baseMinimumCapital}</div>
          <div className="text-[11px] text-[#00F0A0] mt-1">SEBI Prescribed Minimum Satisfied</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">EFFECTIVE LIQUID NET WORTH</div>
          <div className="text-xl font-bold font-mono text-white mt-1">{metrics.effectiveLiquidNetWorth}</div>
          <div className="text-[11px] text-slate-400 mt-1">Cash + G-Secs + Bank Guarantees</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">PEAK MARGIN UTILIZATION</div>
          <div className="text-xl font-bold font-mono text-[#00F0A0] mt-1">{metrics.peakMarginUtilizationPct}%</div>
          <div className="text-[11px] text-slate-400 mt-1">Well below 85% Warning Threshold</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">SGF CONTRIBUTION</div>
          <div className="text-xl font-bold font-mono text-white mt-1">{metrics.sgfContribution}</div>
          <div className="text-[11px] text-[#00F0A0] mt-1">Settlement Guarantee Fund Locked</div>
        </div>
      </div>

      <div className="p-4 rounded border border-slate-800 bg-[#111620] space-y-4">
        <h2 className="text-sm font-bold">Clearing Member Capital Buffer & Intraday Stress Metric</h2>
        <div className="w-full bg-slate-800 h-3 rounded-full overflow-hidden">
          <div
            className="bg-[#00F0A0] h-full transition-all"
            style={{ width: `${metrics.peakMarginUtilizationPct}%` }}
          />
        </div>
        <div className="flex justify-between text-xs font-mono text-slate-400">
          <span>0% Utilized</span>
          <span>Buffer Remaining: {(100 - metrics.peakMarginUtilizationPct).toFixed(1)}%</span>
          <span>100% Margin Limit</span>
        </div>
      </div>
    </div>
  );
}
