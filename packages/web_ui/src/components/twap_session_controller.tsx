import React, { useState } from 'react';

export function TwapSessionController() {
  const [duration, setDuration] = useState(60);
  const [progressPct] = useState(42);

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">TWAP Execution Controller</span>
        <span className="text-slate-400">Duration: {duration} mins</span>
      </div>
      <div className="w-full bg-slate-800 h-2.5 rounded-full overflow-hidden">
        <div className="bg-[#00F0A0] h-full" style={{ width: `${progressPct}%` }} />
      </div>
      <div className="flex justify-between text-[11px] text-slate-400">
        <span>Filled: 4.2 / 10.0 BTC</span>
        <span>VWAP Benchmark: $68,462.10</span>
      </div>
    </div>
  );
}
