import React, { useState } from 'react';

export function NativeTaprootBtcDeposit() {
  const [address] = useState('bc1p5d76tf20...91m0k8q4v2');
  const [copied, setCopied] = useState(false);

  const copyToClipboard = () => {
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-amber-400">Bitcoin Real Deposit (Taproot P2TR)</span>
        <span className="text-[10px] text-slate-400">1 On-Chain Confirmation</span>
      </div>
      <div className="p-4 bg-white text-black rounded text-center font-bold text-sm w-44 mx-auto">
        [ QR Code: {address.slice(0, 10)}... ]
      </div>
      <div className="p-2 rounded bg-slate-900 flex justify-between items-center text-slate-300">
        <span className="text-[11px]">{address}</span>
        <button onClick={copyToClipboard} className="text-[#00F0A0] hover:underline">
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
    </div>
  );
}
