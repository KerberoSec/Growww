'use client';

import React, { useState, useEffect } from 'react';

export interface OrderbookLevel {
  price: number;
  size: number;
  total: number;
}

export function ProOrderbook() {
  const [bids, setBids] = useState<OrderbookLevel[]>([
    { price: 68480.0, size: 1.25, total: 1.25 },
    { price: 68475.5, size: 2.40, total: 3.65 },
    { price: 68470.0, size: 0.85, total: 4.50 },
    { price: 68465.0, size: 3.10, total: 7.60 },
    { price: 68460.0, size: 1.50, total: 9.10 },
  ]);

  const [asks, setAsks] = useState<OrderbookLevel[]>([
    { price: 68485.0, size: 0.95, total: 0.95 },
    { price: 68490.5, size: 1.80, total: 2.75 },
    { price: 68495.0, size: 2.15, total: 4.90 },
    { price: 68500.0, size: 4.20, total: 9.10 },
    { price: 68505.0, size: 1.10, total: 10.20 },
  ]);

  const [lastPrice, setLastPrice] = useState(68482.5);

  useEffect(() => {
    const interval = setInterval(() => {
      const delta = (Math.random() - 0.5) * 5;
      setLastPrice((prev) => Math.round((prev + delta) * 10) / 10);
    }, 800);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="rounded-lg border border-slate-800 bg-[#111620] p-4 text-white font-mono text-xs w-full max-w-md">
      <div className="flex justify-between items-center pb-2 border-b border-slate-800 text-slate-400">
        <span>Order Book (BTC/USDT)</span>
        <span className="text-[10px] text-[#00F0A0]">0.1 Tick Size</span>
      </div>

      <div className="grid grid-cols-3 text-slate-500 py-1.5 text-[11px]">
        <span>PRICE (USDT)</span>
        <span className="text-right">SIZE (BTC)</span>
        <span className="text-right">TOTAL</span>
      </div>

      {/* Asks (Red) */}
      <div className="space-y-0.5">
        {asks.slice().reverse().map((level) => {
          const depthPct = Math.min((level.total / 12) * 100, 100);
          return (
            <div key={level.price} className="relative flex justify-between py-0.5 px-1 hover:bg-slate-800/40">
              <div
                className="absolute inset-y-0 right-0 bg-[#FF3B56]/15 pointer-events-none"
                style={{ width: `${depthPct}%` }}
              />
              <span className="text-[#FF3B56] relative z-10">{level.price.toFixed(1)}</span>
              <span className="text-right text-slate-300 relative z-10">{level.size.toFixed(2)}</span>
              <span className="text-right text-slate-400 relative z-10">{level.total.toFixed(2)}</span>
            </div>
          );
        })}
      </div>

      {/* Spread / Mid-market */}
      <div className="py-2 my-1 border-y border-slate-800 flex justify-between items-center px-1">
        <span className="text-base font-bold text-[#00F0A0]">${lastPrice.toFixed(1)}</span>
        <span className="text-[11px] text-slate-400">Spread: $5.00 (0.007%)</span>
      </div>

      {/* Bids (Green) */}
      <div className="space-y-0.5">
        {bids.map((level) => {
          const depthPct = Math.min((level.total / 12) * 100, 100);
          return (
            <div key={level.price} className="relative flex justify-between py-0.5 px-1 hover:bg-slate-800/40">
              <div
                className="absolute inset-y-0 right-0 bg-[#00F0A0]/15 pointer-events-none"
                style={{ width: `${depthPct}%` }}
              />
              <span className="text-[#00F0A0] relative z-10">{level.price.toFixed(1)}</span>
              <span className="text-right text-slate-300 relative z-10">{level.size.toFixed(2)}</span>
              <span className="text-right text-slate-400 relative z-10">{level.total.toFixed(2)}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
