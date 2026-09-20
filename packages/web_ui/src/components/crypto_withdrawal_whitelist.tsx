import React, { useState } from 'react';

export function CryptoWithdrawalWhitelist() {
  const [entries] = useState([
    { id: 'WHT-1', label: 'Ledger Hardware Cold Vault', address: '0x9481...2b8c', network: 'ETH', coolOffLeftHours: 0 },
    { id: 'WHT-2', label: 'Trezor Safe 3 Office', address: 'bc1p91...0ka7', network: 'BTC', coolOffLeftHours: 12 },
  ]);

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Withdrawal Whitelist Address Manager</span>
        <span className="text-[10px] text-amber-400">24h Security Cool-Off Enforced</span>
      </div>
      <div className="divide-y divide-slate-800">
        {entries.map((e) => (
          <div key={e.id} className="py-2 flex justify-between items-center">
            <div>
              <div className="font-bold text-white">{e.label}</div>
              <div className="text-slate-400">{e.address} ({e.network})</div>
            </div>
            <div>
              {e.coolOffLeftHours === 0 ? (
                <span className="text-[#00F0A0] font-bold">READY TO WITHDRAW</span>
              ) : (
                <span className="text-amber-400">Cool-off: {e.coolOffLeftHours}h remaining</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
