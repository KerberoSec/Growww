import React from 'react';

export function TradeExecutionHistory() {
  const trades = [
    { id: 'TRD-99182', symbol: 'BTC-USDT', side: 'BUY', price: 68450.0, qty: 0.25, time: '18:22:04 IST' },
    { id: 'TRD-99181', symbol: 'RELIANCE-RWA', side: 'BUY', price: 2980.0, qty: 10.0, time: '17:45:10 IST' },
  ];

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Trade Execution History</span>
        <button className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a]">
          Download Digitally Signed Contract Note (PDF)
        </button>
      </div>
      <div className="divide-y divide-slate-800">
        {trades.map((t) => (
          <div key={t.id} className="py-2 flex justify-between">
            <span className="text-slate-400">{t.id}</span>
            <span className="font-bold">{t.symbol}</span>
            <span className={t.side === 'BUY' ? 'text-[#00F0A0]' : 'text-[#FF3B56]'}>{t.side}</span>
            <span>${t.price.toFixed(2)}</span>
            <span>{t.qty}</span>
            <span className="text-slate-500">{t.time}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
