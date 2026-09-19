import React, { useState } from 'react';

export interface eRupeeTokenBalance {
  denomination: number; // 2000, 500, 200, 100, 50
  tokenCount: number;
  totalINR: number;
}

export const RBIeRupeeSettlementDesk: React.FC = () => {
  const [balances] = useState<eRupeeTokenBalance[]>([
    { denomination: 2000, tokenCount: 50, totalINR: 100000 },
    { denomination: 500, tokenCount: 200, totalINR: 100000 },
    { denomination: 200, tokenCount: 250, totalINR: 50000 },
    { denomination: 100, tokenCount: 500, totalINR: 50000 },
  ]);

  const totalVaultBalance = balances.reduce((sum, b) => sum + b.totalINR, 0);

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-4xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">RBI Digital Rupee (e₹-W / e₹-R) Settlement Desk</h2>
          <p className="text-gray-400 text-xs mt-1">
            Central Bank Digital Currency (CBDC) nodal wallet with instant RTGS interoperability and zero counterparty risk.
          </p>
        </div>
        <div className="text-right">
          <span className="text-xs text-gray-500 block">CBDC Vault Balance</span>
          <span className="text-xl font-bold text-emerald-400 font-mono">
            ₹{totalVaultBalance.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
        {balances.map(b => (
          <div key={b.denomination} className="bg-[#141824] p-4 rounded-lg border border-gray-800 text-center font-mono">
            <span className="text-xs text-gray-500 block">₹{b.denomination} Note</span>
            <span className="text-base font-bold text-white block mt-1">{b.tokenCount} Tokens</span>
            <span className="text-[11px] text-emerald-400 font-semibold mt-0.5 block">₹{b.totalINR.toLocaleString('en-IN')}</span>
          </div>
        ))}
      </div>

      <div className="flex gap-3">
        <button className="flex-1 py-3 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition">
          Mint e₹ from Commercial Bank (RTGS)
        </button>
        <button className="flex-1 py-3 bg-[#161616] hover:bg-gray-800 text-gray-200 font-semibold rounded-lg text-xs border border-gray-800 transition">
          Redeem e₹ to Bank Account
        </button>
      </div>
    </div>
  );
};
