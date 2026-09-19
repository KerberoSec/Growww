import React, { useState } from 'react';

export function IcebergOrderPanel() {
  const [totalQty, setTotalQty] = useState('10.0');
  const [displayQty, setDisplayQty] = useState('1.0');

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">Iceberg Order Slicing</span>
        <span className="text-[10px] text-slate-400">Algorithmic Hidden Depth</span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="text-slate-400 block mb-1">TOTAL QUANTITY</label>
          <input
            type="text"
            value={totalQty}
            onChange={(e) => setTotalQty(e.target.value)}
            className="w-full bg-slate-900 border border-slate-700 rounded p-2 text-white"
          />
        </div>
        <div>
          <label className="text-slate-400 block mb-1">VISIBLE TRANCHE SIZE</label>
          <input
            type="text"
            value={displayQty}
            onChange={(e) => setDisplayQty(e.target.value)}
            className="w-full bg-slate-900 border border-slate-700 rounded p-2 text-white"
          />
        </div>
      </div>
    </div>
  );
}
