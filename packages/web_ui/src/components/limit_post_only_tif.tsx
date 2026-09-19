import React, { useState } from 'react';

export function LimitPostOnlyTifSelector() {
  const [postOnly, setPostOnly] = useState(true);
  const [tif, setTif] = useState<'GTC' | 'IOC' | 'FOK'>('GTC');

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <label className="flex items-center space-x-2 cursor-pointer">
          <input
            type="checkbox"
            checked={postOnly}
            onChange={(e) => setPostOnly(e.target.checked)}
            className="accent-[#00F0A0] rounded"
          />
          <span className="text-white font-bold">Post-Only (Maker Guarantee)</span>
        </label>
        <span className="text-[10px] text-slate-400">0.00% Maker Fee</span>
      </div>

      <div className="flex items-center space-x-2 pt-1">
        <span className="text-slate-400">Time-in-Force:</span>
        {(['GTC', 'IOC', 'FOK'] as const).map((mode) => (
          <button
            key={mode}
            onClick={() => setTif(mode)}
            className={`px-2.5 py-1 rounded text-[11px] ${
              tif === mode ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-400'
            }`}
          >
            {mode}
          </button>
        ))}
      </div>
    </div>
  );
}
