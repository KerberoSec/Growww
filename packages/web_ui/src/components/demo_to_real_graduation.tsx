import React from 'react';

export function DemoToRealGraduation() {
  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">Demo-to-Real Graduation Funnel</span>
        <span className="text-amber-400">Milestone 3 of 3 Unlocked</span>
      </div>
      <p className="text-slate-400">
        You have maintained positive expected value for 30 consecutive simulated days. Transition your verified strategy to live order books.
      </p>
      <button className="px-3 py-1.5 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a]">
        Proceed to DigiLocker Aadhaar KYC
      </button>
    </div>
  );
}
