import React, { useState } from 'react';

export interface DutchAuctionLot {
  auctionId: string;
  assetName: string;
  isin: string;
  startPriceINR: number;
  reserveFloorINR: number;
  currentDecayingPriceINR: number;
  totalTokensAvailable: number;
  clearingPriceINR?: number;
  timeRemainingSeconds: number;
  status: 'LIVE' | 'CLEARED' | 'UPCOMING';
}

export const RWADutchAuctionPortal: React.FC = () => {
  const [auctions] = useState<DutchAuctionLot[]>([
    {
      auctionId: 'AUC_SGB_01',
      assetName: 'Sovereign Gold Bond Token (SGB-2026-X)',
      isin: 'IN0020260018',
      startPriceINR: 6200,
      reserveFloorINR: 5800,
      currentDecayingPriceINR: 5980,
      totalTokensAvailable: 50000,
      timeRemainingSeconds: 3420,
      status: 'LIVE',
    },
    {
      auctionId: 'AUC_INVIT_02',
      assetName: 'National Highway Infra Trust Token (NHAI-2026)',
      isin: 'INE087201019',
      startPriceINR: 110,
      reserveFloorINR: 98,
      currentDecayingPriceINR: 103.50,
      totalTokensAvailable: 2500000,
      timeRemainingSeconds: 8400,
      status: 'LIVE',
    },
  ]);

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-5xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">RWA Primary Issuance & Dutch Auction Portal</h2>
          <p className="text-gray-400 text-xs mt-1">
            Uniform price discovery for tokenized Sovereign Gold Bonds, Real Estate, and Infrastructure Trusts under IFSCA guidelines.
          </p>
        </div>
      </div>

      <div className="space-y-4">
        {auctions.map(a => (
          <div key={a.auctionId} className="bg-[#141824] border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition">
            <div className="flex justify-between items-start mb-3">
              <div>
                <h3 className="font-bold text-base text-gray-200">{a.assetName}</h3>
                <span className="text-xs font-mono text-gray-500">ISIN: {a.isin}</span>
              </div>
              <span className="px-2.5 py-1 rounded text-xs font-mono font-bold bg-emerald-950 text-emerald-400 border border-emerald-800 animate-pulse">
                LIVE AUCTION
              </span>
            </div>

            <div className="grid grid-cols-4 gap-3 bg-[#0B0E14] p-4 rounded-lg border border-gray-850 font-mono text-xs mb-4">
              <div>
                <span className="text-[10px] text-gray-500 block">Start Price</span>
                <span className="text-gray-300 font-semibold">₹{a.startPriceINR}</span>
              </div>
              <div>
                <span className="text-[10px] text-gray-500 block">Reserve Floor</span>
                <span className="text-gray-300 font-semibold">₹{a.reserveFloorINR}</span>
              </div>
              <div>
                <span className="text-[10px] text-emerald-400 block font-bold">Current Auction Price</span>
                <span className="text-lg text-emerald-400 font-extrabold">₹{a.currentDecayingPriceINR.toFixed(2)}</span>
              </div>
              <div>
                <span className="text-[10px] text-gray-500 block">Available Supply</span>
                <span className="text-white font-semibold">{a.totalTokensAvailable.toLocaleString()} units</span>
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <button className="px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition">
                Submit Uniform Bid @ ₹{a.currentDecayingPriceINR.toFixed(2)}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
