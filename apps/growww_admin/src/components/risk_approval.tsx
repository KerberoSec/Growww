'use client';

import React, { useState } from 'react';

export interface RiskException {
  id: string;
  type: 'MARGIN_HAIRCUT_OVERRIDE' | 'EXPOSURE_CAP_INCREASE' | 'CIRCUIT_BREAKER_SUSPENSION';
  account: string;
  requestedBy: string;
  requiredSignatures: number;
  currentSignatures: string[];
  status: 'PENDING' | 'EXECUTED' | 'EXPIRED';
  expiresAt: string;
}

export function RiskExceptionApproval() {
  const [exceptions, setExceptions] = useState<RiskException[]>([
    {
      id: 'EXC-8821',
      type: 'EXPOSURE_CAP_INCREASE',
      account: 'INST-ALPHA-CAPITAL',
      requestedBy: 'risk-officer-01@growww.in',
      requiredSignatures: 3,
      currentSignatures: ['0x4a...81', '0x9c...3f'],
      status: 'PENDING',
      expiresAt: '2026-09-19 23:59 UTC',
    },
    {
      id: 'EXC-8822',
      type: 'MARGIN_HAIRCUT_OVERRIDE',
      account: 'MM-LIQUIDITY-PARTNERS',
      requestedBy: 'risk-officer-02@growww.in',
      requiredSignatures: 3,
      currentSignatures: ['0x4a...81'],
      status: 'PENDING',
      expiresAt: '2026-09-20 04:00 UTC',
    },
  ]);

  const [connectedSigner] = useState('0x71...b4');

  const handleSign = (id: string) => {
    setExceptions((prev) =>
      prev.map((item) => {
        if (item.id === id && !item.currentSignatures.includes(connectedSigner)) {
          const updatedSignatures = [...item.currentSignatures, connectedSigner];
          const isExecuted = updatedSignatures.length >= item.requiredSignatures;
          return {
            ...item,
            currentSignatures: updatedSignatures,
            status: isExecuted ? 'EXECUTED' : 'PENDING',
          };
        }
        return item;
      })
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center bg-[#111620] p-4 rounded border border-slate-800">
        <div className="text-xs font-mono text-slate-300">
          <span>Connected HSM Signer: </span>
          <span className="text-[#00F0A0] font-bold">{connectedSigner}</span>
        </div>
        <div className="text-xs font-mono px-2 py-1 rounded bg-slate-800 text-amber-400">
          FIPS 140-2 Level 3 M-of-N Multisig
        </div>
      </div>

      <div className="divide-y divide-slate-800 rounded border border-slate-800 bg-[#111620] overflow-hidden">
        {exceptions.map((item) => (
          <div key={item.id} className="p-4 flex flex-col md:flex-row justify-between md:items-center gap-4">
            <div className="space-y-1">
              <div className="flex items-center space-x-3">
                <span className="font-mono text-sm font-bold text-white">{item.id}</span>
                <span className="text-xs px-2 py-0.5 rounded bg-slate-800 text-[#00F0A0] font-mono">
                  {item.type}
                </span>
                <span className="text-xs font-mono text-slate-400">Account: {item.account}</span>
              </div>
              <div className="text-xs text-slate-400 font-mono">
                Initiated by {item.requestedBy} | Expires: {item.expiresAt}
              </div>
              <div className="flex items-center space-x-2 text-xs font-mono pt-1">
                <span className="text-slate-400">Approvals:</span>
                <span className="text-[#00F0A0] font-bold">
                  {item.currentSignatures.length} / {item.requiredSignatures}
                </span>
                <div className="flex space-x-1">
                  {Array.from({ length: item.requiredSignatures }).map((_, i) => (
                    <span
                      key={i}
                      className={`inline-block h-2 w-5 rounded ${
                        i < item.currentSignatures.length ? 'bg-[#00F0A0]' : 'bg-slate-700'
                      }`}
                    />
                  ))}
                </div>
              </div>
            </div>

            <div>
              {item.status === 'EXECUTED' ? (
                <span className="px-3 py-1.5 rounded bg-[#00F0A0]/20 text-[#00F0A0] text-xs font-bold font-mono">
                  EXECUTED ON-CHAIN
                </span>
              ) : (
                <button
                  onClick={() => handleSign(item.id)}
                  disabled={item.currentSignatures.includes(connectedSigner)}
                  className="px-4 py-2 rounded bg-[#00F0A0] text-black font-bold text-xs disabled:opacity-40 hover:bg-[#00d08a]"
                >
                  {item.currentSignatures.includes(connectedSigner) ? 'Signed' : 'Cosign Exception'}
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
