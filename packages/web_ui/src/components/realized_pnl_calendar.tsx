import React from 'react';

export function RealizedPnLCalendar() {
  const days = [
    { day: 1, pnl: 4500 }, { day: 2, pnl: -1200 }, { day: 3, pnl: 8200 },
    { day: 4, pnl: 3100 }, { day: 5, pnl: -400 },  { day: 6, pnl: 0 },
    { day: 7, pnl: 12400 }, { day: 8, pnl: 5600 }, { day: 9, pnl: -2300 },
  ];

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Realized PnL Calendar & Tax Reports</span>
        <button className="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300">
          Export Schedule VDA (CSV)
        </button>
      </div>
      <div className="grid grid-cols-7 gap-1">
        {days.map((d) => (
          <div
            key={d.day}
            className={`p-2 rounded text-center ${
              d.pnl > 0 ? 'bg-[#00F0A0]/20 text-[#00F0A0] border border-[#00F0A0]/30' :
              d.pnl < 0 ? 'bg-[#FF3B56]/20 text-[#FF3B56] border border-[#FF3B56]/30' :
              'bg-slate-900 text-slate-500'
            }`}
          >
            <div className="text-[10px] text-slate-400">Day {d.day}</div>
            <div className="font-bold text-[11px] mt-0.5">
              {d.pnl > 0 ? `+₹${d.pnl}` : d.pnl < 0 ? `-₹${Math.abs(d.pnl)}` : '₹0'}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
