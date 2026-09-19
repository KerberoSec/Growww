import React, { useState } from 'react';

export function TrailingStopSlider() {
  const [trailPct, setTrailPct] = useState(1.5);

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="text-slate-400">Dynamic Trailing Distance</span>
        <span className="text-[#00F0A0] font-bold">{trailPct}%</span>
      </div>
      <input
        type="range"
        min="0.5"
        max="10.0"
        step="0.1"
        value={trailPct}
        onChange={(e) => setTrailPct(parseFloat(e.target.value))}
        className="w-full accent-[#00F0A0]"
      />
    </div>
  );
}
