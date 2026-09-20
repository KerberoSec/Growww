'use client';

import React, { useState } from 'react';

export function DemoTradingAnalytics() {
  const [demoBalance, setDemoBalance] = useState(1000000); // 10 Lakh vINR or 10,000 vUSDT
  const [stats] = useState({
    totalTrades: 48,
    winRatePct: 68.75,
    sharpeRatio: 2.14,
    profitFactor: 2.45,
    maxDrawdownPct: 4.8,
    netPnL: 142850,
  });

  const handleReset = () => {
    setDemoBalance(1000000);
  };

  return (
    <div className="space-y-6 text-white font-sans max-w-5xl mx-auto">
      {/* Simulation Header */}
      <div className="p-4 rounded-xl border border-slate-800 bg-[#111620] flex flex-col md:flex-row justify-between md:items-center gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <span className="px-2 py-0.5 rounded bg-amber-950 text-amber-400 border border-amber-800 text-xs font-mono font-bold">
              PAPER TRADING SANDBOX
            </span>
            <span className="text-sm font-bold">Risk-Free Market Simulator</span>
          </div>
          <p className="text-xs text-slate-400 font-mono mt-1">Live NSE/BSE & Crypto order routing mirror with 0 capital risk.</p>
        </div>
        <div className="flex items-center space-x-4">
          <div className="text-right font-mono">
            <div className="text-xs text-slate-400">Virtual Portfolio NAV</div>
            <div className="text-xl font-bold text-[#00F0A0]">₹{(demoBalance + stats.netPnL).toLocaleString()}</div>
          </div>
          <button
            onClick={handleReset}
            className="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-xs font-mono text-slate-300"
          >
            Reset Balance
          </button>
        </div>
      </div>

      {/* Analytics KPI Matrix */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3 font-mono text-xs">
        <div className="p-3 rounded bg-[#111620] border border-slate-800">
          <div className="text-slate-400">TOTAL TRADES</div>
          <div className="text-lg font-bold text-white mt-1">{stats.totalTrades}</div>
        </div>
        <div className="p-3 rounded bg-[#111620] border border-slate-800">
          <div className="text-slate-400">WIN RATE</div>
          <div className="text-lg font-bold text-[#00F0A0] mt-1">{stats.winRatePct}%</div>
        </div>
        <div className="p-3 rounded bg-[#111620] border border-slate-800">
          <div className="text-slate-400">SHARPE RATIO</div>
          <div className="text-lg font-bold text-[#00F0A0] mt-1">{stats.sharpeRatio}</div>
        </div>
        <div className="p-3 rounded bg-[#111620] border border-slate-800">
          <div className="text-slate-400">PROFIT FACTOR</div>
          <div className="text-lg font-bold text-white mt-1">{stats.profitFactor}</div>
        </div>
        <div className="p-3 rounded bg-[#111620] border border-slate-800">
          <div className="text-slate-400">MAX DRAWDOWN</div>
          <div className="text-lg font-bold text-amber-400 mt-1">-{stats.maxDrawdownPct}%</div>
        </div>
        <div className="p-3 rounded bg-[#111620] border border-slate-800">
          <div className="text-slate-400">NET REALIZED PnL</div>
          <div className="text-lg font-bold text-[#00F0A0] mt-1">+₹{stats.netPnL.toLocaleString()}</div>
        </div>
      </div>

      {/* Graduation Callout */}
      <div className="p-6 rounded-xl border border-slate-800 bg-gradient-to-r from-emerald-950/40 via-[#111620] to-slate-900 flex flex-col md:flex-row justify-between md:items-center gap-4">
        <div>
          <h3 className="text-base font-bold text-white">Ready for Real Market Execution?</h3>
          <p className="text-xs text-slate-400 mt-1">
            Your paper trading consistency score is in the top 5% of simulated traders. Complete instant DigiLocker KYC to trade real equities.
          </p>
        </div>
        <a
          href="/onboarding"
          className="px-5 py-2.5 rounded bg-[#00F0A0] text-black font-bold text-xs hover:bg-[#00d08a] transition-all whitespace-nowrap"
        >
          Graduate to Real Money →
        </a>
      </div>
    </div>
  );
}
