'use client';

import React, { useState } from 'react';

export function RwaLaunchpad() {
  const [bidAmountInr, setBidAmountInr] = useState('100000');
  const [bidYieldPct, setBidYieldPct] = useState('7.25');
  const [bidSubmitted, setBidSubmitted] = useState(false);

  const handlePlaceBid = (e: React.FormEvent) => {
    e.preventDefault();
    setBidSubmitted(true);
  };

  return (
    <div className="space-y-6 text-white font-sans max-w-4xl mx-auto">
      {/* Active Dutch Auction Card */}
      <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-6">
        <div className="flex flex-col md:flex-row justify-between md:items-center gap-4 border-b border-slate-800 pb-4">
          <div>
            <span className="text-xs font-mono px-2 py-0.5 rounded bg-emerald-950 text-[#00F0A0] border border-[#00F0A0]/30">
              PRIMARY DUTCH AUCTION LIVE
            </span>
            <h2 className="text-2xl font-bold mt-2">Government of India 91-Day T-Bills (Tokenized)</h2>
            <p className="text-xs text-slate-400 font-mono mt-0.5">ISIN: IN002026X019 | Sovereign Backing</p>
          </div>
          <div className="text-right font-mono">
            <div className="text-xs text-slate-400">Auction Closes In</div>
            <div className="text-xl font-bold text-amber-400">04h : 18m : 42s</div>
          </div>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 font-mono text-xs">
          <div className="p-3 rounded bg-slate-900 border border-slate-800">
            <div className="text-slate-400">TOTAL ISSUE SIZE</div>
            <div className="text-white font-bold text-sm mt-1">₹500 Cr</div>
          </div>
          <div className="p-3 rounded bg-slate-900 border border-slate-800">
            <div className="text-slate-400">SUBSCRIPTION</div>
            <div className="text-[#00F0A0] font-bold text-sm mt-1">3.4x (Oversubscribed)</div>
          </div>
          <div className="p-3 rounded bg-slate-900 border border-slate-800">
            <div className="text-slate-400">CUT-OFF YIELD</div>
            <div className="text-white font-bold text-sm mt-1">7.18% Indicative</div>
          </div>
          <div className="p-3 rounded bg-slate-900 border border-slate-800">
            <div className="text-slate-400">SETTLEMENT DATE</div>
            <div className="text-white font-bold text-sm mt-1">21-Sep-2026 (T+1)</div>
          </div>
        </div>

        {/* Bidding Ticket */}
        <form onSubmit={handlePlaceBid} className="space-y-4 pt-2">
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">BID INVESTMENT AMOUNT (INR)</label>
              <input
                type="number"
                value={bidAmountInr}
                onChange={(e) => setBidAmountInr(e.target.value)}
                min="10000"
                step="10000"
                className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">MINIMUM ACCEPTABLE YIELD (%)</label>
              <input
                type="number"
                value={bidYieldPct}
                onChange={(e) => setBidYieldPct(e.target.value)}
                step="0.01"
                className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
                required
              />
            </div>
          </div>

          <button
            type="submit"
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black text-xs hover:bg-[#00d08a]"
          >
            Submit Competitive Auction Bid
          </button>
        </form>

        {bidSubmitted && (
          <div className="p-4 rounded bg-emerald-950/40 border border-emerald-800 text-xs font-mono text-emerald-300">
            ✓ Bid of ₹{Number(bidAmountInr).toLocaleString()} at {bidYieldPct}% Yield submitted! UPI 2.0 Mandate created.
          </div>
        )}
      </div>
    </div>
  );
}
