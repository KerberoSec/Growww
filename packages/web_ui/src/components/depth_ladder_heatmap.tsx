import React from 'react';

export function DepthLadderHeatmap() {
  const levels = [
    { price: 68510.0, volume: 14.5, heat: 0.95 },
    { price: 68505.0, volume: 8.2, heat: 0.65 },
    { price: 68500.0, volume: 18.9, heat: 1.0 }, // Big liquidity wall
    { price: 68495.0, volume: 4.1, heat: 0.3 },
  ];

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center text-slate-400">
        <span className="font-bold text-white">L2/L3 Cumulative Depth Heatmap</span>
        <span className="text-[#00F0A0]">Real-Time Microsecond Canvas</span>
      </div>
      <div className="space-y-1">
        {levels.map((lvl) => (
          <div key={lvl.price} className="relative flex justify-between p-1.5 rounded bg-slate-900/60 overflow-hidden">
            <div
              className="absolute inset-y-0 right-0 bg-[#00F0A0] opacity-20"
              style={{ width: `${lvl.heat * 100}%` }}
            />
            <span className="font-bold relative z-10">${lvl.price.toFixed(1)}</span>
            <span className="text-slate-300 relative z-10">{lvl.volume.toFixed(1)} BTC Wall</span>
          </div>
        ))}
      </div>
    </div>
  );
}
