import React, { useState } from 'react';

export interface EarnVault {
  vaultId: string;
  asset: string;
  vaultType: 'FLEXIBLE_SAVINGS' | 'FIXED_30D' | 'FIXED_90D' | 'INSTITUTIONAL_COLLATERAL';
  apyPct: number; // e.g. 6.5%
  minDeposit: number;
  totalDepositedUSD: number;
  lockupDays: number;
  interestPayoutFrequency: string; // Daily
}

export const CryptoEarnVaults: React.FC = () => {
  const [vaults] = useState<EarnVault[]>([
    { vaultId: 'V1', asset: 'USDT', vaultType: 'FLEXIBLE_SAVINGS', apyPct: 8.5, minDeposit: 10, totalDepositedUSD: 4500000, lockupDays: 0, interestPayoutFrequency: 'Daily Compound' },
    { vaultId: 'V2', asset: 'USDT', vaultType: 'FIXED_30D', apyPct: 11.2, minDeposit: 100, totalDepositedUSD: 8200000, lockupDays: 30, interestPayoutFrequency: 'Maturity Payout' },
    { vaultId: 'V3', asset: 'BTC', vaultType: 'FLEXIBLE_SAVINGS', apyPct: 4.2, minDeposit: 0.001, totalDepositedUSD: 15400000, lockupDays: 0, interestPayoutFrequency: 'Daily Compound' },
    { vaultId: 'V4', asset: 'ETH', vaultType: 'FIXED_90D', apyPct: 5.8, minDeposit: 0.1, totalDepositedUSD: 6800000, lockupDays: 90, interestPayoutFrequency: 'Maturity Payout' },
  ]);

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-5xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Crypto Earn & Liquidity Staking Vaults</h2>
          <p className="text-gray-400 text-xs mt-1">
            Put your idle VDA assets to work with automated yield generation backed by 100% on-chain collateralization.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {vaults.map(v => (
          <div key={v.vaultId} className="bg-[#141824] border border-gray-800 rounded-xl p-5 flex flex-col justify-between hover:border-gray-700 transition">
            <div>
              <div className="flex justify-between items-center mb-3">
                <span className="font-bold text-base text-white">{v.asset}</span>
                <span className="text-xs font-mono font-bold text-emerald-400 bg-emerald-950/60 border border-emerald-800 px-2 py-0.5 rounded">
                  {v.apyPct}% APY
                </span>
              </div>

              <div className="text-xs text-gray-400 mb-4">{v.vaultType.replace('_', ' ')}</div>

              <div className="space-y-1.5 text-[11px] font-mono text-gray-400 mb-4">
                <div className="flex justify-between">
                  <span>Lockup:</span>
                  <span className="text-gray-200">{v.lockupDays === 0 ? 'Flexible (Anytime)' : `${v.lockupDays} Days`}</span>
                </div>
                <div className="flex justify-between">
                  <span>Min Deposit:</span>
                  <span className="text-gray-200">{v.minDeposit} {v.asset}</span>
                </div>
                <div className="flex justify-between">
                  <span>Payouts:</span>
                  <span className="text-gray-200">{v.interestPayoutFrequency}</span>
                </div>
              </div>
            </div>

            <button className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition">
              Subscribe to Vault
            </button>
          </div>
        ))}
      </div>
    </div>
  );
};
