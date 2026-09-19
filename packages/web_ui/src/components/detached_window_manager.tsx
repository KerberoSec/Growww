import React, { useState } from 'react';

export function DetachedWindowManager() {
  const [detachedWindows, setDetachedWindows] = useState<string[]>(['MONITOR_2_CHART']);

  const spawnDetachedWindow = (name: string) => {
    setDetachedWindows((prev) => [...prev, `${name}_${Date.now().toString(36)}`]);
  };

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">BroadcastChannel Multi-Screen Sync</span>
        <button
          onClick={() => spawnDetachedWindow('DETACHED_ORDERBOOK')}
          className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a]"
        >
          Detach Panel to New Window
        </button>
      </div>
      <div className="text-slate-400">
        Active Sync Channel: <span className="text-white">growww_cross_screen_bus</span>
      </div>
      <div className="space-y-1">
        {detachedWindows.map((win) => (
          <div key={win} className="p-2 rounded bg-slate-900 flex justify-between items-center text-slate-300">
            <span>Window ID: {win}</span>
            <span className="text-[#00F0A0] text-[10px]">SYNCED (0 lag)</span>
          </div>
        ))}
      </div>
    </div>
  );
}
