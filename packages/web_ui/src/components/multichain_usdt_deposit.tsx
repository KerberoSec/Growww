import React, { useState } from 'react';

export function MultichainUsdtDeposit() {
  const [network, setNetwork] = useState<'ERC-20' | 'TRC-20' | 'POLYGON' | 'SOLANA'>('POLYGON');

  const addresses = {
    'ERC-20': '0x718a...e91a (Ethereum Mainnet)',
    'TRC-20': 'TN2p9...4v1k (Tron Network)',
    'POLYGON': '0x718a...e91a (Polygon POS)',
    'SOLANA': '7Xk2...m81q (Solana SPL)',
  };

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Multichain USDT Deposit Portal</span>
        <span className="text-[#00F0A0]">0.00% Deposit Fee</span>
      </div>
      <div className="flex space-x-1">
        {(['POLYGON', 'TRC-20', 'ERC-20', 'SOLANA'] as const).map((net) => (
          <button
            key={net}
            onClick={() => setNetwork(net)}
            className={`px-2.5 py-1 rounded text-[11px] ${
              network === net ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-400'
            }`}
          >
            {net}
          </button>
        ))}
      </div>
      <div className="p-2 rounded bg-slate-900 text-slate-300">
        Address: <span className="text-[#00F0A0]">{addresses[network]}</span>
      </div>
    </div>
  );
}
