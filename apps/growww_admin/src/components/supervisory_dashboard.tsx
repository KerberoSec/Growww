'use client';

import React, { useState } from 'react';

export function SupervisoryDashboard() {
  const [surveillanceAlerts] = useState([
    {
      id: 'ALR-CIRC-019',
      category: 'SPOOFING_DETECTION',
      instrument: 'RELIANCE-RWA',
      severity: 'HIGH',
      description: 'Rapid cancel-to-fill ratio (>94%) detected from sub-account SA-8821.',
      timestamp: '2026-09-19 18:14:02 IST',
    },
    {
      id: 'ALR-CIRC-020',
      category: 'FRONT_RUNNING_PREVENTION',
      instrument: 'BTC/USDT',
      severity: 'LOW',
      description: 'Zero pattern deviation detected. Pre-trade margin gate enforced.',
      timestamp: '2026-09-19 18:10:45 IST',
    },
  ]);

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="bg-[#111620] p-4 rounded border border-slate-800 flex justify-between items-center">
        <div>
          <div className="text-sm font-bold">SEBI & IFSCA Read-Only Supervisory Portal</div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">Air-gapped regulatory inspection node</div>
        </div>
        <div className="px-3 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] text-xs font-mono font-bold">
          Active Surveillance Engine
        </div>
      </div>

      <div className="rounded border border-slate-800 bg-[#111620] overflow-hidden">
        <div className="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 className="text-sm font-bold">Market Surveillance & Pattern Alerts</h2>
          <span className="text-xs font-mono text-slate-400">SEBI Surveillance Master Circular</span>
        </div>
        <div className="divide-y divide-slate-800">
          {surveillanceAlerts.map((alr) => (
            <div key={alr.id} className="p-4 flex flex-col md:flex-row justify-between md:items-center gap-4">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="font-mono text-xs text-slate-400 font-bold">{alr.id}</span>
                  <span className="px-2 py-0.5 rounded bg-slate-800 text-xs font-mono text-amber-400">
                    {alr.category}
                  </span>
                  <span className="font-bold text-sm text-white">{alr.instrument}</span>
                </div>
                <p className="text-xs text-slate-300 mt-1">{alr.description}</p>
                <div className="text-[11px] font-mono text-slate-500 mt-1">{alr.timestamp}</div>
              </div>
              <div>
                <button className="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-xs font-mono text-slate-200">
                  Inspect Order Audit Log
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
