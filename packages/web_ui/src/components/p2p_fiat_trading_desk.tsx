import React from 'react';

export function P2PFiatTradingDesk() {
  const ads = [
    { id: 'AD-991', merchant: 'SwiftP2P_Merchant', priceInr: 89.25, availableUsdt: 4200, limit: '₹10,000 - ₹2,00,000' },
    { id: 'AD-992', merchant: 'Apex_Liquidity_Desk', priceInr: 89.30, availableUsdt: 12500, limit: '₹50,000 - ₹5,00,000' },
  ];

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">P2P Fiat-to-USDT Pro Merchant Desk</span>
        <button className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a]">
          Post Ad
        </button>
      </div>
      <div className="divide-y divide-slate-800">
        {ads.map((ad) => (
          <div key={ad.id} className="py-2 flex justify-between items-center">
            <div>
              <div className="font-bold text-white">{ad.merchant}</div>
              <div className="text-slate-400">Limit: {ad.limit}</div>
            </div>
            <div className="text-right">
              <div className="font-bold text-[#00F0A0]">₹{ad.priceInr} / USDT</div>
              <div className="text-slate-400">{ad.availableUsdt} USDT Available</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
