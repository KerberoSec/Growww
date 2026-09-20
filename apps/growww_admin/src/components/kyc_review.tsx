'use client';

import React, { useState } from 'react';

export interface KycApplicant {
  id: string;
  pan: string;
  name: string;
  matchScore: number;
  status: 'PENDING_MAKER' | 'PENDING_CHECKER' | 'APPROVED' | 'REJECTED';
  sanctionsHit: boolean;
  appliedAt: string;
}

export function KycReviewDashboard() {
  const [applicants, setApplicants] = useState<KycApplicant[]>([
    {
      id: 'USR-7019',
      pan: 'ABCDE1234F',
      name: 'Rohan Sharma',
      matchScore: 0.98,
      status: 'PENDING_CHECKER',
      sanctionsHit: false,
      appliedAt: '2026-09-19 14:22 UTC',
    },
    {
      id: 'USR-7020',
      pan: 'FGHIJ5678K',
      name: 'Vikram Mehta',
      matchScore: 0.81,
      status: 'PENDING_MAKER',
      sanctionsHit: false,
      appliedAt: '2026-09-19 15:05 UTC',
    },
    {
      id: 'USR-7021',
      pan: 'KLMNO9012P',
      name: 'Ananya Verma',
      matchScore: 0.99,
      status: 'PENDING_MAKER',
      sanctionsHit: false,
      appliedAt: '2026-09-19 15:10 UTC',
    },
  ]);

  const [selectedId, setSelectedId] = useState<string>('USR-7019');

  const handleApprove = (id: string) => {
    setApplicants((prev) =>
      prev.map((app) => (app.id === id ? { ...app, status: 'APPROVED' } : app))
    );
  };

  const handleReject = (id: string) => {
    setApplicants((prev) =>
      prev.map((app) => (app.id === id ? { ...app, status: 'REJECTED' } : app))
    );
  };

  const currentApp = applicants.find((a) => a.id === selectedId) || applicants[0];

  return (
    <div className="flex flex-col lg:flex-row h-full min-h-[500px] border border-slate-800 rounded-lg bg-[#111620] overflow-hidden text-white font-sans">
      {/* Left List of Applicants */}
      <div className="w-full lg:w-96 border-r border-slate-800 flex flex-col">
        <div className="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 className="text-sm font-bold tracking-wide">KYC Verification Queue</h2>
          <span className="text-xs font-mono px-2 py-0.5 rounded bg-slate-800 text-[#00F0A0]">
            {applicants.filter((a) => a.status.startsWith('PENDING')).length} Pending
          </span>
        </div>
        <div className="divide-y divide-slate-800/60 overflow-y-auto flex-1">
          {applicants.map((app) => (
            <div
              key={app.id}
              onClick={() => setSelectedId(app.id)}
              className={`p-3 cursor-pointer transition-colors ${
                app.id === selectedId ? 'bg-slate-800/80 border-l-4 border-[#00F0A0]' : 'hover:bg-slate-800/30'
              }`}
            >
              <div className="flex justify-between items-center text-xs">
                <span className="font-bold">{app.name}</span>
                <span className="font-mono text-slate-400">{app.id}</span>
              </div>
              <div className="flex justify-between items-center mt-1 text-[11px] font-mono">
                <span className="text-slate-400">PAN: {app.pan}</span>
                <span
                  className={
                    app.status === 'APPROVED' ? 'text-[#00F0A0]' :
                    app.status === 'REJECTED' ? 'text-[#FF3B56]' : 'text-amber-400'
                  }
                >
                  {app.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Right Details Panel */}
      <div className="flex-1 p-6 flex flex-col justify-between">
        <div>
          <div className="flex justify-between items-start pb-4 border-b border-slate-800">
            <div>
              <h1 className="text-xl font-bold">{currentApp.name}</h1>
              <p className="text-xs font-mono text-slate-400 mt-0.5">Applied: {currentApp.appliedAt}</p>
            </div>
            <div className="text-right font-mono">
              <div className="text-xs text-slate-400">Facial Match Confidence</div>
              <div className="text-lg font-bold text-[#00F0A0]">{(currentApp.matchScore * 100).toFixed(0)}%</div>
            </div>
          </div>

          {/* Verification Cards */}
          <div className="grid grid-cols-2 gap-4 mt-6">
            <div className="p-4 rounded border border-slate-800 bg-slate-900/60 font-mono text-xs space-y-2">
              <div className="text-slate-400">UIDAI / DigiLocker Records</div>
              <div>Aadhaar Masked: XXXX-XXXX-4819</div>
              <div>Name Match: 100% Exact</div>
              <div>DOB: 14-Aug-1992</div>
              <div>Address: Mumbai, MH, IN</div>
            </div>
            <div className="p-4 rounded border border-slate-800 bg-slate-900/60 font-mono text-xs space-y-2">
              <div className="text-slate-400">AML & Sanctions Screening</div>
              <div>UN Sanctions: Clear</div>
              <div>OFAC / FATF Red List: Clear</div>
              <div>PEP Classification: Non-PEP</div>
              <div>Risk Level: Low (Tier-2 Authorized)</div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="pt-6 border-t border-slate-800 flex justify-end space-x-3">
          <button
            onClick={() => handleReject(currentApp.id)}
            disabled={currentApp.status === 'REJECTED'}
            className="px-4 py-2 rounded bg-red-900/40 border border-red-700 text-red-300 font-medium text-xs hover:bg-red-900/70"
          >
            Reject Application
          </button>
          <button
            onClick={() => handleApprove(currentApp.id)}
            disabled={currentApp.status === 'APPROVED'}
            className="px-5 py-2 rounded bg-[#00F0A0] text-black font-bold text-xs hover:bg-[#00d08a]"
          >
            Authorize Tier-2 (Checker Sign)
          </button>
        </div>
      </div>
    </div>
  );
}
