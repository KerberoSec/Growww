import React, { useState } from 'react';

export interface ProDockLayoutProps {
  layoutName?: string;
  children?: React.ReactNode;
}

export function ProDockLayout({ layoutName = 'DEFAULT_PRO', children }: ProDockLayoutProps) {
  const [activePanels, setActivePanels] = useState<string[]>([
    'CHART', 'ORDERBOOK', 'ORDER_ENTRY', 'TRADES', 'POSITIONS'
  ]);

  const togglePanel = (panel: string) => {
    setActivePanels((prev) =>
      prev.includes(panel) ? prev.filter((p) => p !== panel) : [...prev, panel]
    );
  };

  return (
    <div className="flex flex-col h-full w-full bg-[#0B0E14] text-white font-sans">
      <div className="h-9 bg-[#111620] border-b border-slate-800 px-3 flex items-center justify-between text-xs font-mono">
        <div className="flex items-center space-x-2">
          <span className="font-bold text-[#00F0A0]">DOCKVIEW:</span>
          <span>{layoutName}</span>
        </div>
        <div className="flex space-x-1">
          {['CHART', 'ORDERBOOK', 'ORDER_ENTRY', 'TRADES', 'POSITIONS'].map((p) => (
            <button
              key={p}
              onClick={() => togglePanel(p)}
              className={`px-2 py-0.5 rounded text-[10px] ${
                activePanels.includes(p) ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-400'
              }`}
            >
              {p}
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 p-2 grid grid-cols-12 gap-2">
        {children || (
          <div className="col-span-12 rounded border border-dashed border-slate-800 flex items-center justify-center text-slate-500 font-mono text-xs">
            Multi-Dock Tiling Workspace Active ({activePanels.length} Panels Mounted)
          </div>
        )}
      </div>
    </div>
  );
}
