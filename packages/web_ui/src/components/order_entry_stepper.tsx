import React, { useState } from 'react';

export function OrderEntryStepper() {
  const [qty, setQty] = useState(0.25);
  const [percentage, setPercentage] = useState(25);

  const adjustQty = (delta: number) => {
    setQty((prev) => Math.max(0.01, Math.round((prev + delta) * 100) / 100));
  };

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="text-slate-400">Order Size Stepper</span>
        <span className="text-[#00F0A0] font-bold">{qty.toFixed(2)} BTC</span>
      </div>
      <div className="flex space-x-2">
        <button onClick={() => adjustQty(-0.05)} className="px-3 py-1 bg-slate-800 rounded hover:bg-slate-700">-</button>
        <input
          type="number"
          value={qty}
          onChange={(e) => setQty(parseFloat(e.target.value) || 0)}
          className="flex-1 bg-slate-900 border border-slate-700 rounded px-2 text-center text-white"
        />
        <button onClick={() => adjustQty(0.05)} className="px-3 py-1 bg-slate-800 rounded hover:bg-slate-700">+</button>
      </div>

      <div className="flex justify-between space-x-1 pt-1">
        {[25, 50, 75, 100].map((pct) => (
          <button
            key={pct}
            onClick={() => setPercentage(pct)}
            className={`flex-1 py-1 rounded text-[11px] ${
              percentage === pct ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-400'
            }`}
          >
            {pct}%
          </button>
        ))}
      </div>
    </div>
  );
}
