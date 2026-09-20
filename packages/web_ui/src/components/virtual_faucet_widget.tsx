import React, { useState } from 'react';

export function VirtualFaucetWidget() {
  const [claimed, setClaimed] = useState(false);

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">Virtual Sandbox Faucet</span>
        <button
          onClick={() => setClaimed(true)}
          className="px-3 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a]"
        >
          {claimed ? 'Claimed (+10,000 vUSDT)' : 'Claim 10,000 vUSDT & 1 vBTC'}
        </button>
      </div>
    </div>
  );
}
