import React from 'react';

export function HoldingsSectorTreemap() {
  const sectors = [
    { name: 'Financial Services & Banking', weight: 32.5, color: 'bg-emerald-800' },
    { name: 'Information Technology', weight: 24.0, color: 'bg-teal-800' },
    { name: 'Energy & Oil/Gas', weight: 18.5, color: 'bg-cyan-800' },
    { name: 'Sovereign Debt (G-Sec)', weight: 15.0, color: 'bg-blue-800' },
    { name: 'Crypto Assets', weight: 10.0, color: 'bg-indigo-800' },
  ];

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Holdings Sector Treemap Visualizer</span>
        <span className="text-[#00F0A0]">100% Portfolio Coverage</span>
      </div>
      <div className="grid grid-cols-12 gap-1 h-36">
        {sectors.map((s) => (
          <div
            key={s.name}
            className={`${s.color} rounded p-2 flex flex-col justify-between overflow-hidden`}
            style={{ gridColumn: `span ${Math.max(2, Math.round((s.weight / 100) * 12))}` }}
          >
            <span className="truncate text-[11px] font-bold">{s.name}</span>
            <span className="text-right text-[10px] text-white/80">{s.weight}%</span>
          </div>
        ))}
      </div>
    </div>
  );
}
