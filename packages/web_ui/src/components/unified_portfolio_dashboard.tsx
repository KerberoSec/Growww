import React from 'react';

export function UnifiedPortfolioDashboard() {
  const assets = [
    { category: 'NSE/BSE Demat Equities (1:1 Backed)', valueInr: 1245000, pnlPct: 14.2 },
    { category: 'Government of India Sovereign Bonds (G-Sec)', valueInr: 500000, pnlPct: 7.18 },
    { category: 'Crypto Spot Assets (BTC, ETH, USDT)', valueInr: 854000, pnlPct: 22.8 },
  ];

  const totalNav = assets.reduce((sum, a) => sum + a.valueInr, 0);

  return (
    <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] text-white font-sans space-y-6">
      <div className="flex justify-between items-center pb-4 border-b border-slate-800">
        <div>
          <div className="text-xs font-mono text-slate-400">UNIFIED NET WORTH (INR)</div>
          <div className="text-3xl font-black font-mono text-white mt-1">₹{totalNav.toLocaleString()}</div>
        </div>
        <div className="px-3 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] text-xs font-mono font-bold">
          +15.4% All-Time Gain
        </div>
      </div>

      <div className="divide-y divide-slate-800">
        {assets.map((a) => (
          <div key={a.category} className="py-3 flex justify-between items-center text-xs font-mono">
            <span className="text-slate-300 font-sans">{a.category}</span>
            <div className="text-right">
              <div className="font-bold text-white">₹{a.valueInr.toLocaleString()}</div>
              <div className="text-[#00F0A0]">+{a.pnlPct}%</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
