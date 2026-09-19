import React, { useState } from 'react';

export interface P2PMerchantAd {
  adId: string;
  merchantName: string;
  completionRate: number; // e.g. 99.4%
  ordersCompleted: number;
  fiatCurrency: string; // INR
  cryptoAsset: string; // USDT
  side: 'BUY' | 'SELL';
  priceINR: number;
  availableCrypto: number;
  minLimitINR: number;
  maxLimitINR: number;
  paymentMethods: string[]; // UPI, IMPS, Bank Transfer
  isVerifiedMerchant: boolean;
}

interface P2PProMerchantDeskProps {
  ads: P2PMerchantAd[];
  onInitiateTrade: (adId: string, amountINR: number, side: 'BUY' | 'SELL') => void;
}

export const P2PProMerchantDesk: React.FC<P2PProMerchantDeskProps> = ({ ads, onInitiateTrade }) => {
  const [selectedSide, setSelectedSide] = useState<'BUY' | 'SELL'>('BUY');
  const [filterPaymentMethod, setFilterPaymentMethod] = useState<string>('ALL');

  const filteredAds = ads.filter(ad => {
    if (ad.side !== selectedSide) return false;
    if (filterPaymentMethod !== 'ALL' && !ad.paymentMethods.includes(filterPaymentMethod)) return false;
    return true;
  });

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-5xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">P2P Fiat & USDT Merchant Desk</h2>
          <p className="text-gray-400 text-xs mt-1">
            Zero-fee fiat on/off-ramp backed by 100% smart contract escrow on Hyperledger Besu.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setSelectedSide('BUY')}
            className={`px-5 py-2 rounded-lg font-bold text-xs transition ${
              selectedSide === 'BUY' ? 'bg-emerald-600 text-white' : 'bg-gray-850 text-gray-400'
            }`}
          >
            Buy USDT
          </button>
          <button
            onClick={() => setSelectedSide('SELL')}
            className={`px-5 py-2 rounded-lg font-bold text-xs transition ${
              selectedSide === 'SELL' ? 'bg-red-600 text-white' : 'bg-gray-850 text-gray-400'
            }`}
          >
            Sell USDT
          </button>
        </div>
      </div>

      <div className="overflow-x-auto rounded-xl border border-gray-800 bg-[#141824]">
        <table className="w-full text-left text-xs font-sans">
          <thead className="bg-[#1A2030] text-gray-400 font-mono uppercase text-[11px]">
            <tr>
              <th className="p-3.5">Merchant</th>
              <th className="p-3.5">Unit Price</th>
              <th className="p-3.5">Available / Limits</th>
              <th className="p-3.5">Payment Rails</th>
              <th className="p-3.5 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {filteredAds.map(ad => (
              <tr key={ad.adId} className="hover:bg-gray-850/50 transition">
                <td className="p-3.5">
                  <div className="font-bold text-sm text-gray-200 flex items-center gap-1.5">
                    {ad.merchantName}
                    {ad.isVerifiedMerchant && <span className="text-emerald-400 text-xs font-bold" title="SEBI/FIU Verified">✓</span>}
                  </div>
                  <div className="text-[11px] text-gray-500 font-mono">
                    {ad.ordersCompleted} orders ({ad.completionRate}% completion)
                  </div>
                </td>
                <td className="p-3.5">
                  <span className="font-mono font-bold text-base text-white">₹{ad.priceINR.toFixed(2)}</span>
                  <span className="text-gray-500 text-[10px] block font-mono">per USDT</span>
                </td>
                <td className="p-3.5 font-mono">
                  <div className="text-gray-300 font-semibold">{ad.availableCrypto.toLocaleString()} USDT</div>
                  <div className="text-gray-500 text-[11px]">
                    ₹{ad.minLimitINR.toLocaleString('en-IN')} - ₹{ad.maxLimitINR.toLocaleString('en-IN')}
                  </div>
                </td>
                <td className="p-3.5">
                  <div className="flex flex-wrap gap-1.5">
                    {ad.paymentMethods.map(m => (
                      <span key={m} className="px-2 py-0.5 rounded text-[10px] font-mono bg-gray-800 text-gray-300 border border-gray-700">
                        {m}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="p-3.5 text-right">
                  <button
                    onClick={() => onInitiateTrade(ad.adId, ad.minLimitINR, ad.side)}
                    className={`px-4 py-2 rounded-lg font-bold text-xs transition ${
                      ad.side === 'BUY'
                        ? 'bg-emerald-600 hover:bg-emerald-500 text-white'
                        : 'bg-red-600 hover:bg-red-500 text-white'
                    }`}
                  >
                    {ad.side} USDT
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
