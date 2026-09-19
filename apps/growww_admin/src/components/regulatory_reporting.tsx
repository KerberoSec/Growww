'use client';

import React, { useState } from 'react';

export interface RegulatoryReport {
  id: string;
  regulator: 'SEBI' | 'RBI' | 'FIU_IND' | 'IFSCA';
  title: string;
  frequency: string;
  wormHash: string;
  status: 'SUBMITTED' | 'GENERATING' | 'READY_FOR_FILING';
  filingDeadline: string;
}

export function RegulatoryReportingPortal() {
  const [reports] = useState<RegulatoryReport[]>([
    {
      id: 'REP-SEBI-2026-M09',
      regulator: 'SEBI',
      title: 'Monthly Broker Clearing & Settlement Reconciliation (Annexure A)',
      frequency: 'MONTHLY',
      wormHash: 'sha256:4d8a1c927f08b3e...99c',
      status: 'READY_FOR_FILING',
      filingDeadline: '2026-09-22 18:30 IST',
    },
    {
      id: 'REP-FIU-STR-492',
      regulator: 'FIU_IND',
      title: 'Suspicious Transaction Report (STR Batch 492 under PMLA)',
      frequency: 'AD-HOC',
      wormHash: 'sha256:a1b2c3d4e5f6...789',
      status: 'SUBMITTED',
      filingDeadline: '2026-09-18 23:59 IST',
    },
    {
      id: 'REP-IFSCA-FPI-81',
      regulator: 'IFSCA',
      title: 'GIFT City International Investor Outbound LRS Report',
      frequency: 'WEEKLY',
      wormHash: 'sha256:9f8e7d6c5b4a...123',
      status: 'READY_FOR_FILING',
      filingDeadline: '2026-09-24 17:00 IST',
    },
  ]);

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 font-mono text-xs">
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-slate-400">SEBI (DOMESTIC)</div>
          <div className="text-lg font-bold text-[#00F0A0] mt-1">100% Compliant</div>
          <div className="text-[10px] text-slate-500 mt-1">Circular SEBI/HO/MIRSD/2026</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-slate-400">FIU-IND (PMLA)</div>
          <div className="text-lg font-bold text-white mt-1">WORM Validated</div>
          <div className="text-[10px] text-slate-500 mt-1">7-Year Tamper-Proof Vault</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-slate-400">RBI (FEMA / LRS)</div>
          <div className="text-lg font-bold text-white mt-1">Real-time Stream</div>
          <div className="text-[10px] text-slate-500 mt-1">Automated Nostro/Vostro</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-slate-400">IFSCA (GIFT CITY)</div>
          <div className="text-lg font-bold text-[#00F0A0] mt-1">Active Sandbox</div>
          <div className="text-[10px] text-slate-500 mt-1">Dual-Entity Compliance</div>
        </div>
      </div>

      <div className="rounded border border-slate-800 bg-[#111620] overflow-hidden">
        <div className="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 className="text-sm font-bold">Statutory Compliance Export Hub</h2>
          <span className="text-xs font-mono text-slate-400">Export format: Signed XML / CSV / PDF</span>
        </div>
        <div className="divide-y divide-slate-800">
          {reports.map((rep) => (
            <div key={rep.id} className="p-4 flex flex-col md:flex-row justify-between md:items-center gap-4">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="px-2 py-0.5 rounded bg-slate-800 font-mono text-xs font-bold text-[#00F0A0]">
                    {rep.regulator}
                  </span>
                  <span className="font-bold text-sm text-white">{rep.title}</span>
                </div>
                <div className="text-xs font-mono text-slate-500 mt-1">
                  Report ID: {rep.id} | Deadline: {rep.filingDeadline}
                </div>
                <div className="text-[11px] font-mono text-slate-400">
                  WORM Storage Checksum: {rep.wormHash}
                </div>
              </div>

              <div className="flex items-center space-x-4">
                <span
                  className={`text-xs font-mono px-2 py-1 rounded ${
                    rep.status === 'SUBMITTED' ? 'bg-[#00F0A0]/20 text-[#00F0A0]' : 'bg-amber-900/30 text-amber-300'
                  }`}
                >
                  {rep.status}
                </span>
                <button className="px-3 py-1.5 rounded bg-slate-800 text-xs font-mono hover:bg-slate-700 text-slate-200">
                  Export Packet
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
