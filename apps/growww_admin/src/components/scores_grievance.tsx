'use client';

import React, { useState } from 'react';

export interface GrievanceItem {
  id: string;
  scoresRef: string;
  investorName: string;
  category: string;
  daysRemainingSla: number;
  status: 'OPEN' | 'INVESTIGATING' | 'RESOLVED_ATR_FILED';
}

export function ScoresGrievancePortal() {
  const [complaints] = useState<GrievanceItem[]>([
    {
      id: 'GRV-2026-0012',
      scoresRef: 'SEBIE/MH26/0001928/1',
      investorName: 'Kavita Sundaram',
      category: 'Delayed Demat Fractional Dividend Credit',
      daysRemainingSla: 14,
      status: 'INVESTIGATING',
    },
    {
      id: 'GRV-2026-0013',
      scoresRef: 'SEBIE/DL26/0002041/1',
      investorName: 'Amit Patel',
      category: 'UPI Refund Timeout Settlement Inquiry',
      daysRemainingSla: 18,
      status: 'OPEN',
    },
  ]);

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="bg-[#111620] p-4 rounded border border-slate-800 flex justify-between items-center">
        <div>
          <div className="text-sm font-bold">SEBI SCORES 2.0 & ODR Grievance Gateway</div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">Strict 21-Calendar-Day Resolution SLA Tracking</div>
        </div>
        <div className="px-3 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] text-xs font-mono font-bold">
          0 Breached SLAs
        </div>
      </div>

      <div className="rounded border border-slate-800 bg-[#111620] overflow-hidden">
        <div className="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 className="text-sm font-bold">Active Investor Complaints Queue</h2>
          <span className="text-xs font-mono text-slate-400">Automated ATR Submissions</span>
        </div>
        <div className="divide-y divide-slate-800">
          {complaints.map((c) => (
            <div key={c.id} className="p-4 flex flex-col md:flex-row justify-between md:items-center gap-4">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="font-bold text-sm text-white">{c.scoresRef}</span>
                  <span className="text-xs px-2 py-0.5 rounded bg-slate-800 font-mono text-slate-300">{c.id}</span>
                </div>
                <div className="text-xs text-slate-300 mt-1">Investor: {c.investorName} | Subject: {c.category}</div>
                <div className="text-[11px] font-mono text-amber-400 mt-1">
                  SLA Timer: {c.daysRemainingSla} days remaining to file Action Taken Report (ATR)
                </div>
              </div>
              <div>
                <button className="px-4 py-1.5 rounded bg-[#00F0A0] text-black font-bold text-xs hover:bg-[#00d08a]">
                  Compose ATR Response
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
