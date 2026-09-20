import React from 'react';

export function P2PDisputeArbitrationConsole() {
  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">Operator Arbitration Console</span>
        <span className="text-[10px] text-amber-400">Maker-Checker Dual Sign-off Required</span>
      </div>
      <div className="p-3 bg-slate-900 rounded border border-slate-800 text-slate-300">
        Dispute Case: <span className="text-white font-bold">#DISP-2026-0048</span> | Status: Reviewing Bank Slip OCR
      </div>
      <div className="flex space-x-2">
        <button className="flex-1 py-1.5 bg-red-900/60 rounded text-red-200 hover:bg-red-800">
          Reject Claim & Unlock Merchant
        </button>
        <button className="flex-1 py-1.5 bg-[#00F0A0] rounded text-black font-bold hover:bg-[#00d08a]">
          Verify UTR & Force Release
        </button>
      </div>
    </div>
  );
}
