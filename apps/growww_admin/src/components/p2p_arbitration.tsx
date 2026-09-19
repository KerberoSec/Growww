'use client';

import React, { useState } from 'react';

export function P2PArbitrationConsole() {
  const [disputes, setDisputes] = useState([
    {
      id: 'DISP-P2P-9021',
      orderId: 'P2P-ORD-11029',
      buyer: 'trader_rajesh',
      seller: 'merchant_vikram',
      amountInr: 50000,
      usdtAmount: 561.8,
      claimedUtr: 'UTR491820194820',
      status: 'UNDER_INVESTIGATION',
    },
  ]);

  const handleResolve = (id: string, action: 'RELEASE_TO_BUYER' | 'REFUND_SELLER') => {
    setDisputes((prev) =>
      prev.map((d) => (d.id === id ? { ...d, status: `RESOLVED_${action}` } : d))
    );
  };

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="bg-[#111620] p-4 rounded border border-slate-800 flex justify-between items-center">
        <div>
          <div className="text-sm font-bold">P2P Fiat Dispute Arbitration Desk</div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">Maker-checker escrow intervention & bank UTR verification</div>
        </div>
        <div className="text-xs font-mono text-amber-400">1 Active Case</div>
      </div>

      <div className="divide-y divide-slate-800 rounded border border-slate-800 bg-[#111620]">
        {disputes.map((d) => (
          <div key={d.id} className="p-4 space-y-4">
            <div className="flex justify-between items-start">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="font-bold text-sm">{d.id}</span>
                  <span className="text-xs px-2 py-0.5 rounded bg-slate-800 font-mono text-[#00F0A0]">{d.orderId}</span>
                </div>
                <div className="text-xs text-slate-400 font-mono mt-1">
                  Buyer: {d.buyer} | Seller: {d.seller}
                </div>
              </div>
              <div className="text-right font-mono text-xs">
                <div className="text-[#00F0A0] font-bold">₹{d.amountInr.toLocaleString()}</div>
                <div className="text-slate-400">{d.usdtAmount} USDT in Escrow</div>
              </div>
            </div>

            <div className="p-3 rounded bg-slate-900 border border-slate-800 font-mono text-xs space-y-1">
              <div>Claimed Bank UTR: <span className="text-white font-bold">{d.claimedUtr}</span></div>
              <div className="text-slate-400">Bank Statement API Verification: UTR match confirmed at receiving beneficiary bank.</div>
            </div>

            <div className="flex justify-between items-center pt-2">
              <span className="text-xs font-mono text-amber-400">{d.status}</span>
              <div className="space-x-3">
                <button
                  onClick={() => handleResolve(d.id, 'REFUND_SELLER')}
                  className="px-3 py-1.5 rounded bg-slate-800 text-xs font-mono text-red-300 hover:bg-slate-700"
                >
                  Cancel & Return to Seller
                </button>
                <button
                  onClick={() => handleResolve(d.id, 'RELEASE_TO_BUYER')}
                  className="px-4 py-1.5 rounded bg-[#00F0A0] text-xs font-bold text-black hover:bg-[#00d08a]"
                >
                  Confirm Bank UTR & Release USDT
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
