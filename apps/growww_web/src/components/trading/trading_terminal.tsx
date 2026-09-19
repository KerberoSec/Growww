'use client';

import React, { useState, useEffect } from 'react';

export interface TradingTerminalProps {
  ticker?: string;
}

export function TradingTerminal({ ticker = 'BTC-USDT' }: TradingTerminalProps) {
  const [currentPrice, setCurrentPrice] = useState(68450.25);
  const [priceChange24h, setPriceChange24h] = useState(2.84);
  const [orderSide, setOrderSide] = useState<'BUY' | 'SELL'>('BUY');
  const [orderType, setOrderType] = useState<'LIMIT' | 'MARKET'>('LIMIT');
  const [priceInput, setPriceInput] = useState('68450.00');
  const [quantityInput, setQuantityInput] = useState('0.15');

  // Simulated live ticker feed
  useEffect(() => {
    const timer = setInterval(() => {
      const delta = (Math.random() - 0.48) * 12;
      setCurrentPrice((prev) => Math.round((prev + delta) * 100) / 100);
    }, 1200);
    return () => clearInterval(timer);
  }, []);

  return (
    <div className="flex flex-1 flex-col lg:flex-row h-full overflow-hidden bg-[#0B0E14] text-white">
      {/* Chart & Market Data Central Column */}
      <div className="flex-1 flex flex-col border-r border-slate-800">
        {/* Market Stats Bar */}
        <div className="h-12 border-b border-slate-800 bg-[#111620] px-4 flex items-center justify-between text-xs font-mono">
          <div className="flex items-center space-x-4">
            <span className="text-sm font-bold text-white">{ticker}</span>
            <span className="text-lg font-bold text-[#00F0A0]">
              ${currentPrice.toLocaleString('en-US', { minimumFractionDigits: 2 })}
            </span>
            <span className={priceChange24h >= 0 ? 'text-[#00F0A0]' : 'text-[#FF3B56]'}>
              +{priceChange24h}%
            </span>
          </div>
          <div className="hidden sm:flex space-x-6 text-slate-400">
            <div>24h High: <span className="text-white">$69,210.00</span></div>
            <div>24h Low: <span className="text-white">$67,110.00</span></div>
            <div>24h Vol: <span className="text-white">1,482.4 BTC</span></div>
          </div>
        </div>

        {/* Lightweight Charts Canvas Placeholder */}
        <div className="flex-1 min-h-[360px] bg-[#0A0D13] flex flex-col items-center justify-center border-b border-slate-800 p-6 relative">
          <div className="absolute top-3 left-4 flex space-x-2 text-xs font-mono text-slate-400">
            <button className="px-2 py-0.5 rounded bg-slate-800 text-[#00F0A0]">1m</button>
            <button className="px-2 py-0.5 rounded hover:bg-slate-800">5m</button>
            <button className="px-2 py-0.5 rounded hover:bg-slate-800">15m</button>
            <button className="px-2 py-0.5 rounded hover:bg-slate-800">1H</button>
            <button className="px-2 py-0.5 rounded hover:bg-slate-800">1D</button>
          </div>
          <div className="text-center font-mono text-xs text-slate-500">
            <div className="text-[#00F0A0] text-sm mb-1 font-sans">TradingView Lightweight Charts v4.x Bridge</div>
            <div>[ Interactive WebGL Candlestick & Volume Canvas Connected ]</div>
            <div className="mt-2 text-[11px] text-slate-600">Sub-10ms tick update frequency | Hyperledger Besu Block Anchor: #12,984,102</div>
          </div>
        </div>

        {/* Bottom Positions & Orders Tray */}
        <div className="h-44 bg-[#111620] flex flex-col">
          <div className="h-8 border-b border-slate-800 px-3 flex items-center space-x-4 text-xs font-medium text-slate-400">
            <button className="text-[#00F0A0] border-b-2 border-[#00F0A0] h-full flex items-center">Open Orders (0)</button>
            <button className="hover:text-white h-full flex items-center">Positions (1)</button>
            <button className="hover:text-white h-full flex items-center">Trade History</button>
          </div>
          <div className="p-3 text-xs font-mono text-slate-400">
            <div className="grid grid-cols-6 text-slate-500 pb-1 border-b border-slate-800/60">
              <span>CONTRACT</span>
              <span>SIZE</span>
              <span>ENTRY</span>
              <span>MARK</span>
              <span>UNREALIZED PnL</span>
              <span>ACTION</span>
            </div>
            <div className="grid grid-cols-6 py-2 items-center text-slate-200">
              <span className="text-[#00F0A0]">BTC/USDT SPOT</span>
              <span>+0.250 BTC</span>
              <span>$67,820.00</span>
              <span>${currentPrice.toFixed(2)}</span>
              <span className="text-[#00F0A0]">+$157.56 (+0.93%)</span>
              <button className="text-xs text-red-400 hover:text-red-300">Close</button>
            </div>
          </div>
        </div>
      </div>

      {/* Right Rail: Order Book & Execution Ticket */}
      <div className="w-full lg:w-80 flex flex-col bg-[#111620]">
        {/* Order Book Depth (Compact) */}
        <div className="flex-1 p-3 border-b border-slate-800 overflow-y-auto">
          <div className="text-xs font-mono text-slate-400 mb-2 flex justify-between">
            <span>Price (USDT)</span>
            <span>Size (BTC)</span>
          </div>
          {/* Asks */}
          <div className="space-y-1 font-mono text-xs">
            <div className="flex justify-between text-[#FF3B56] hover:bg-slate-800/40 px-1 py-0.5 rounded cursor-pointer">
              <span>68,460.00</span><span>0.412</span>
            </div>
            <div className="flex justify-between text-[#FF3B56] hover:bg-slate-800/40 px-1 py-0.5 rounded cursor-pointer">
              <span>68,455.50</span><span>1.205</span>
            </div>
            <div className="flex justify-between text-[#FF3B56] hover:bg-slate-800/40 px-1 py-0.5 rounded cursor-pointer">
              <span>68,452.00</span><span>0.850</span>
            </div>
          </div>
          {/* Spread */}
          <div className="py-2 my-1 border-y border-slate-800/80 text-center font-mono text-sm font-bold text-white">
            ${currentPrice.toFixed(2)}
          </div>
          {/* Bids */}
          <div className="space-y-1 font-mono text-xs">
            <div className="flex justify-between text-[#00F0A0] hover:bg-slate-800/40 px-1 py-0.5 rounded cursor-pointer">
              <span>68,448.00</span><span>0.720</span>
            </div>
            <div className="flex justify-between text-[#00F0A0] hover:bg-slate-800/40 px-1 py-0.5 rounded cursor-pointer">
              <span>68,445.00</span><span>2.110</span>
            </div>
            <div className="flex justify-between text-[#00F0A0] hover:bg-slate-800/40 px-1 py-0.5 rounded cursor-pointer">
              <span>68,440.00</span><span>1.540</span>
            </div>
          </div>
        </div>

        {/* Order Entry Form */}
        <div className="p-4 space-y-3">
          {/* Buy/Sell Tabs */}
          <div className="grid grid-cols-2 gap-1 rounded bg-slate-900 p-1">
            <button
              onClick={() => setOrderSide('BUY')}
              className={`rounded py-1.5 text-xs font-bold transition-all ${
                orderSide === 'BUY' ? 'bg-[#00F0A0] text-black shadow' : 'text-slate-400 hover:text-white'
              }`}
            >
              BUY
            </button>
            <button
              onClick={() => setOrderSide('SELL')}
              className={`rounded py-1.5 text-xs font-bold transition-all ${
                orderSide === 'SELL' ? 'bg-[#FF3B56] text-white shadow' : 'text-slate-400 hover:text-white'
              }`}
            >
              SELL
            </button>
          </div>

          {/* Limit / Market Switch */}
          <div className="flex space-x-2 text-xs font-mono text-slate-400">
            <button
              onClick={() => setOrderType('LIMIT')}
              className={orderType === 'LIMIT' ? 'text-white font-bold' : ''}
            >
              Limit
            </button>
            <span>|</span>
            <button
              onClick={() => setOrderType('MARKET')}
              className={orderType === 'MARKET' ? 'text-white font-bold' : ''}
            >
              Market
            </button>
          </div>

          {/* Inputs */}
          <div>
            <label className="block text-[11px] font-mono text-slate-400 mb-0.5">PRICE</label>
            <input
              type="text"
              value={priceInput}
              onChange={(e) => setPriceInput(e.target.value)}
              className="w-full rounded bg-slate-900 border border-slate-700 px-2.5 py-1.5 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-mono text-slate-400 mb-0.5">QUANTITY</label>
            <input
              type="text"
              value={quantityInput}
              onChange={(e) => setQuantityInput(e.target.value)}
              className="w-full rounded bg-slate-900 border border-slate-700 px-2.5 py-1.5 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
            />
          </div>

          {/* Submit Action */}
          <button
            className={`w-full rounded py-2.5 text-xs font-bold uppercase transition-colors ${
              orderSide === 'BUY' ? 'bg-[#00F0A0] text-black hover:bg-[#00d08a]' : 'bg-[#FF3B56] text-white hover:bg-red-600'
            }`}
          >
            {orderSide} {ticker.split('-')[0]}
          </button>
        </div>
      </div>
    </div>
  );
}
