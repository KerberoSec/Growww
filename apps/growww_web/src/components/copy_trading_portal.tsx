'use client';

import React, { useState } from 'react';

export interface MasterTrader {
  id: string;
  name: string;
  roi30d: number;
  winRate: number;
  followers: number;
  aum: string;
  profitSharePct: number;
}

export function CopyTradingPortal() {
  const [traders] = useState<MasterTrader[]>([
    {
      id: 'MT-ALPHA-QUANT',
      name: 'Alpha Quant Algorithmic Fund',
      roi30d: 38.4,
      winRate: 74.2,
      followers: 1240,
      aum: '$4,280,000',
      profitSharePct: 15,
    },
    {
      id: 'MT-DELTA-HEDGE',
      name: 'Delta Neutral Options Arbitrage',
      roi30d: 19.8,
      winRate: 88.5,
      followers: 890,
      aum: '$2,150,000',
      profitSharePct: 10,
    },
  ]);

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="grid md:grid-cols-2 gap-6">
        {traders.map((trader) => (
          <div key={trader.id} className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-4">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-bold text-lg text-white">{trader.name}</h3>
                <div className="text-xs font-mono text-slate-400 mt-0.5">{trader.id}</div>
              </div>
              <div className="px-2.5 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] text-xs font-mono font-bold">
                +{trader.roi30d}% 30D ROI
              </div>
            </div>

            <div className="grid grid-cols-3 gap-2 font-mono text-xs pt-2 border-t border-slate-800/80">
              <div>
                <div className="text-slate-500">Win Rate</div>
                <div className="text-white font-bold">{trader.winRate}%</div>
              </div>
              <div>
                <div className="text-slate-500">Followers</div>
                <div className="text-white font-bold">{trader.followers}</div>
              </div>
              <div>
                <div className="text-slate-500">Profit Share</div>
                <div className="text-white font-bold">{trader.profitSharePct}% HWM</div>
              </div>
            </div>

            <div className="pt-2">
              <button className="w-full rounded bg-[#00F0A0] py-2 text-xs font-bold text-black hover:bg-[#00d08a]">
                Copy Strategy (Max 1.0x Leverage)
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
