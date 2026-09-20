'use client';

import React, { useState } from 'react';

export interface ReserveItem {
  assetCode: string;
  assetName: string;
  custodianHolding: number; // e.g. shares in NSDL/CDSL
  onChainSupply: number;    // tokens minted on Besu
  variance: number;
  merkleRoot: string;
  status: 'RECONCILED' | 'DISCREPANCY';
  lastAttested: string;
}

export function ReserveReconciliation() {
  const [reserves, setReserves] = useState<ReserveItem[]>([
    {
      assetCode: 'RELIANCE-RWA',
      assetName: 'Reliance Industries Ltd (1:1 Backed)',
      custodianHolding: 500000,
      onChainSupply: 500000,
      variance: 0,
      merkleRoot: '0x3a4b9c1d8e7f2019a8b7c6d5e4f3a2b1c0d9e8f7',
      status: 'RECONCILED',
      lastAttested: '2026-09-19 18:00 UTC',
    },
    {
      assetCode: 'TCS-RWA',
      assetName: 'Tata Consultancy Services (1:1 Backed)',
      custodianHolding: 350000,
      onChainSupply: 350000,
      variance: 0,
      merkleRoot: '0x99a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0',
      status: 'RECONCILED',
      lastAttested: '2026-09-19 18:00 UTC',
    },
    {
      assetCode: 'GSEC-718-2033',
      assetName: '7.18% GS 2033 Sovereign Bond',
      custodianHolding: 100000000,
      onChainSupply: 100000000,
      variance: 0,
      merkleRoot: '0x11223344556677889900aabbccddeeff00112233',
      status: 'RECONCILED',
      lastAttested: '2026-09-19 18:00 UTC',
    },
  ]);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs text-slate-400 font-mono">TOTAL CUSTODIAN ASSETS</div>
          <div className="text-xl font-bold text-white font-mono mt-1">₹4,821,500,000</div>
          <div className="text-[11px] text-[#00F0A0] mt-1">100% NSDL/CDSL Custody Matched</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs text-slate-400 font-mono">ON-CHAIN TOKENS MINTED</div>
          <div className="text-xl font-bold text-white font-mono mt-1">₹4,821,500,000</div>
          <div className="text-[11px] text-[#00F0A0] mt-1">Hyperledger Besu QBFT Invariant</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs text-slate-400 font-mono">VARIANCE / UNBACKED</div>
          <div className="text-xl font-bold text-[#00F0A0] font-mono mt-1">0.00 (0.00%)</div>
          <div className="text-[11px] text-slate-400 mt-1">Automated 3-Way Reconciliation</div>
        </div>
      </div>

      <div className="rounded border border-slate-800 bg-[#111620] overflow-hidden">
        <div className="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 className="text-sm font-bold text-white">Custody vs On-Chain Audit Records</h2>
          <span className="text-xs font-mono px-2 py-0.5 rounded bg-emerald-950 text-[#00F0A0] border border-[#00F0A0]/30">
            All 3 Reservers Validated
          </span>
        </div>
        <div className="divide-y divide-slate-800">
          {reserves.map((res) => (
            <div key={res.assetCode} className="p-4 flex flex-col md:flex-row justify-between md:items-center gap-4">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="font-bold text-sm text-white">{res.assetCode}</span>
                  <span className="text-xs text-slate-400">({res.assetName})</span>
                </div>
                <div className="text-xs font-mono text-slate-500 mt-1">
                  Merkle Root: <span className="text-slate-300">{res.merkleRoot}</span>
                </div>
                <div className="text-xs font-mono text-slate-500">
                  Attested: {res.lastAttested} via Chainlink & Custody Gateway
                </div>
              </div>

              <div className="flex items-center space-x-6 text-xs font-mono">
                <div className="text-right">
                  <div className="text-slate-400">NSDL / CDSL</div>
                  <div className="text-white font-bold">{res.custodianHolding.toLocaleString()}</div>
                </div>
                <div className="text-right">
                  <div className="text-slate-400">Besu Minted</div>
                  <div className="text-white font-bold">{res.onChainSupply.toLocaleString()}</div>
                </div>
                <div className="px-2.5 py-1 rounded bg-[#00F0A0]/10 text-[#00F0A0] font-bold">
                  {res.status}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
