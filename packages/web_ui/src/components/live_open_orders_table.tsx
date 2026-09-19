import React, { useState } from 'react';

export function LiveOpenOrdersTable() {
  const [orders, setOrders] = useState([
    { id: 'ORD-101', side: 'BUY', price: 68420.0, qty: 0.5 },
    { id: 'ORD-102', side: 'BUY', price: 68410.0, qty: 1.2 },
    { id: 'ORD-103', side: 'SELL', price: 68510.0, qty: 0.8 },
  ]);

  const cancelAll = (side?: string) => {
    if (side) {
      setOrders((prev) => prev.filter((o) => o.side !== side));
    } else {
      setOrders([]);
    }
  };

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Live Open Orders ({orders.length})</span>
        <div className="space-x-2">
          <button onClick={() => cancelAll('BUY')} className="px-2 py-1 bg-slate-800 rounded hover:bg-slate-700 text-slate-300">
            Cancel All Bids
          </button>
          <button onClick={() => cancelAll('SELL')} className="px-2 py-1 bg-slate-800 rounded hover:bg-slate-700 text-slate-300">
            Cancel All Asks
          </button>
          <button onClick={() => cancelAll()} className="px-2 py-1 bg-red-900/50 rounded hover:bg-red-800 text-red-300">
            Cancel All
          </button>
        </div>
      </div>
      <div className="divide-y divide-slate-800/60">
        {orders.map((o) => (
          <div key={o.id} className="py-1.5 flex justify-between">
            <span className={o.side === 'BUY' ? 'text-[#00F0A0]' : 'text-[#FF3B56]'}>{o.side}</span>
            <span>${o.price.toFixed(1)}</span>
            <span>{o.qty} BTC</span>
          </div>
        ))}
      </div>
    </div>
  );
}
