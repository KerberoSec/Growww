import React from 'react';

export interface DailyPnL {
  date: string; // YYYY-MM-DD
  realizedPnLINR: number;
  tradesCount: number;
}

interface RealizedPnLHeatmapProps {
  data: DailyPnL[];
}

export const RealizedPnLCalendarHeatmap: React.FC<RealizedPnLHeatmapProps> = ({ data }) => {
  const totalPnL = data.reduce((sum, d) => sum + d.realizedPnLINR, 0);

  const getHeatmapColor = (pnl: number) => {
    if (pnl > 50000) return 'bg-emerald-500 text-black';
    if (pnl > 10000) return 'bg-emerald-600 text-white';
    if (pnl > 0) return 'bg-emerald-800 text-gray-200';
    if (pnl === 0) return 'bg-gray-850 text-gray-500';
    if (pnl > -10000) return 'bg-red-800 text-gray-200';
    if (pnl > -50000) return 'bg-red-600 text-white';
    return 'bg-red-500 text-black';
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h3 className="text-lg font-bold">Realized PnL Calendar Heatmap</h3>
          <p className="text-gray-400 text-xs mt-0.5">30-Day Trading Session Performance Matrix</p>
        </div>
        <div className="text-right">
          <span className="text-xs text-gray-500 block">Total Realized PnL</span>
          <span className={`text-lg font-bold font-mono ${totalPnL >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
            {totalPnL >= 0 ? '+' : ''}₹{totalPnL.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-7 gap-2">
        {data.map(day => (
          <div
            key={day.date}
            className={`p-2.5 rounded-lg border border-gray-800/80 text-center font-mono ${getHeatmapColor(day.realizedPnLINR)}`}
          >
            <div className="text-[10px] opacity-75">{day.date.slice(5)}</div>
            <div className="text-xs font-bold mt-1">
              {day.realizedPnLINR !== 0 ? `₹${(day.realizedPnLINR / 1000).toFixed(1)}k` : '-'}
            </div>
            <div className="text-[9px] opacity-60 mt-0.5">{day.tradesCount} trades</div>
          </div>
        ))}
      </div>
    </div>
  );
};
