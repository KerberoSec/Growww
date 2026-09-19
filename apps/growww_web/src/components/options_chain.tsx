'use client';

import React, { useState } from 'react';

export interface OptionStrike {
  strike: number;
  callBid: number;
  callAsk: number;
  callIv: number;
  callDelta: number;
  putBid: number;
  putAsk: number;
  putIv: number;
  putDelta: number;
}

export function OptionsChainMatrix() {
  const [underlyingPrice] = useState(68500);
  const [expiry, setExpiry] = useState('25-SEP-2026');

  const [strikes] = useState<OptionStrike[]>([
    { strike: 66000, callBid: 2850, callAsk: 2900, callIv: 0.42, callDelta: 0.82, putBid: 320, putAsk: 340, putIv: 0.45, putDelta: -0.18 },
    { strike: 67000, callBid: 2050, callAsk: 2100, callIv: 0.41, callDelta: 0.71, putBid: 510, putAsk: 530, putIv: 0.43, putDelta: -0.29 },
    { strike: 68000, callBid: 1350, callAsk: 1390, callIv: 0.39, callDelta: 0.58, putBid: 820, putAsk: 850, putIv: 0.40, putDelta: -0.42 },
    { strike: 69000, callBid: 810, callAsk: 840, callIv: 0.38, callDelta: 0.44, putBid: 1280, putAsk: 1320, putIv: 0.39, putDelta: -0.56 },
    { strike: 70000, callBid: 450, callAsk: 475, callIv: 0.40, callDelta: 0.31, putBid: 1910, putAsk: 1960, putIv: 0.41, putDelta: -0.69 },
  ]);

  return (
    <div className="space-y-6 text-white font-sans">
      {/* Header bar */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center bg-[#111620] p-4 rounded border border-slate-800 gap-4">
        <div className="flex items-center space-x-4">
          <span className="text-lg font-black">BTC-OPTIONS</span>
          <span className="text-xs font-mono px-2 py-0.5 rounded bg-slate-800 text-[#00F0A0]">
            Spot: ${underlyingPrice.toLocaleString()}
          </span>
        </div>
        <div className="flex items-center space-x-2 text-xs font-mono">
          <span className="text-slate-400">EXPIRY:</span>
          {['25-SEP-2026', '02-OCT-2026', '30-OCT-2026'].map((exp) => (
            <button
              key={exp}
              onClick={() => setExpiry(exp)}
              className={`px-2.5 py-1 rounded ${
                expiry === exp ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-300 hover:text-white'
              }`}
            >
              {exp}
            </button>
          ))}
        </div>
      </div>

      {/* Options Chain Grid */}
      <div className="rounded border border-slate-800 bg-[#111620] overflow-x-auto">
        <table className="w-full text-xs font-mono text-center">
          <thead>
            <tr className="bg-slate-900/80 text-slate-400 border-b border-slate-800">
              <th colSpan={4} className="py-2 text-[#00F0A0] border-r border-slate-800 uppercase">CALLS</th>
              <th className="py-2 text-white bg-slate-800/80 font-bold uppercase">STRIKE</th>
              <th colSpan={4} className="py-2 text-[#FF3B56] border-l border-slate-800 uppercase">PUTS</th>
            </tr>
            <tr className="text-slate-500 border-b border-slate-800/60 text-[11px]">
              <th className="py-1">DELTA</th>
              <th>IV</th>
              <th>BID</th>
              <th className="border-r border-slate-800">ASK</th>
              <th className="bg-slate-800/40 text-slate-300 font-bold">PRICE</th>
              <th className="border-l border-slate-800">BID</th>
              <th>ASK</th>
              <th>IV</th>
              <th>DELTA</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/40">
            {strikes.map((s) => (
              <tr key={s.strike} className="hover:bg-slate-800/20">
                <td className="py-2 text-slate-400">{s.callDelta}</td>
                <td className="text-slate-400">{(s.callIv * 100).toFixed(0)}%</td>
                <td className="text-[#00F0A0] cursor-pointer hover:underline">${s.callBid}</td>
                <td className="text-[#00F0A0] cursor-pointer hover:underline border-r border-slate-800">${s.callAsk}</td>
                <td className="font-bold bg-slate-800/40 text-amber-400">${s.strike.toLocaleString()}</td>
                <td className="text-[#FF3B56] cursor-pointer hover:underline border-l border-slate-800">${s.putBid}</td>
                <td className="text-[#FF3B56] cursor-pointer hover:underline">${s.putAsk}</td>
                <td className="text-slate-400">{(s.putIv * 100).toFixed(0)}%</td>
                <td className="text-slate-400">{s.putDelta}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
