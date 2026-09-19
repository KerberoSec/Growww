import React from 'react';

export function DemoAnalyticsSharpe() {
  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Risk-Adjusted Performance (Sharpe & Calmar)</span>
        <span className="text-[#00F0A0] font-bold">Sharpe: 2.14</span>
      </div>
      <div className="h-28 bg-black/40 rounded border border-slate-800 flex items-center justify-center text-slate-500">
        [ Cumulative Equity Curve Canvas - Upward Drift ]
      </div>
    </div>
  );
}
