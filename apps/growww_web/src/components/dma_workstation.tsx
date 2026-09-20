'use client';

import React, { useState } from 'react';

export function DmaWorkstation() {
  const [fixStatus] = useState({
    sessionId: 'FIX.4.4:GROWWW_PROD_DMA_01',
    seqIn: 104820,
    seqOut: 104821,
    latencyUs: 42,
    colocation: 'Equinix MB1 (Mumbai BKC) Cross-Connect',
  });

  return (
    <div className="space-y-6 text-white font-sans">
      {/* Session Header */}
      <div className="flex flex-col md:flex-row justify-between bg-[#111620] p-4 rounded border border-slate-800 gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <span className="inline-block h-2.5 w-2.5 rounded-full bg-[#00F0A0] animate-pulse" />
            <span className="font-bold text-sm">Direct Market Access (DMA) Workstation</span>
          </div>
          <p className="text-xs font-mono text-slate-400 mt-1">Colocation: {fixStatus.colocation}</p>
        </div>
        <div className="flex items-center space-x-6 text-xs font-mono">
          <div>
            <span className="text-slate-400">SESSION: </span>
            <span className="text-white">{fixStatus.sessionId}</span>
          </div>
          <div className="px-2.5 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] font-bold">
            {fixStatus.latencyUs} μs ROUNDTRIP
          </div>
        </div>
      </div>

      {/* Hotkey matrix & Micro Order Ladder */}
      <div className="grid md:grid-cols-2 gap-6">
        <div className="p-4 rounded border border-slate-800 bg-[#111620] space-y-4">
          <h2 className="text-sm font-bold text-slate-200">High-Frequency Hotkey Bindings</h2>
          <div className="space-y-2 font-mono text-xs">
            <div className="flex justify-between p-2 rounded bg-slate-900 border border-slate-800">
              <span className="text-[#00F0A0] font-bold">Shift + B</span>
              <span className="text-slate-300">Instant Best-Bid Buy (0.50 BTC)</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900 border border-slate-800">
              <span className="text-[#FF3B56] font-bold">Shift + S</span>
              <span className="text-slate-300">Instant Best-Ask Sell (0.50 BTC)</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900 border border-slate-800">
              <span className="text-amber-400 font-bold">Esc</span>
              <span className="text-slate-300">Panic Cancel All Open DMA Orders</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900 border border-slate-800">
              <span className="text-blue-400 font-bold">Ctrl + Space</span>
              <span className="text-slate-300">Flatten All Active Spot & Perp Positions</span>
            </div>
          </div>
        </div>

        <div className="p-4 rounded border border-slate-800 bg-[#111620] space-y-4">
          <h2 className="text-sm font-bold text-slate-200">FPGA Drop-Copy Stream</h2>
          <div className="p-3 rounded bg-black/50 border border-slate-800 font-mono text-[11px] text-slate-400 space-y-1">
            <div>8=FIX.4.4|35=8|49=GROWWW|56=INST_ALPHA|37=ORD_99182|150=2|39=2|31=68450.00|32=0.250|</div>
            <div>8=FIX.4.4|35=8|49=GROWWW|56=INST_ALPHA|37=ORD_99183|150=0|39=0|38=0.500|44=68448.50|</div>
            <div className="text-[#00F0A0]">✓ Stream synchronized: 0 dropped packets on Solarflare NIC</div>
          </div>
        </div>
      </div>
    </div>
  );
}
