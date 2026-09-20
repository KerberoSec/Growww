'use client';

import React, { useState } from 'react';

export interface CommodityVaultItem {
  commodity: string;
  vaultLocation: string;
  purity: string;
  allocatedKg: number;
  vaultReceiptId: string;
  assayStatus: 'CERTIFIED' | 'AUDIT_PENDING';
  redeemable: boolean;
}

export function CommodityVaultPortal() {
  const [items] = useState<CommodityVaultItem[]>([
    {
      commodity: 'Physical Gold (LBMA Good Delivery)',
      vaultLocation: 'Brinks Vault, GIFT City IFSC (SEZ)',
      purity: '999.9 Fine Gold',
      allocatedKg: 50.0,
      vaultReceiptId: 'EWR-GLD-2026-00412',
      assayStatus: 'CERTIFIED',
      redeemable: true,
    },
    {
      commodity: 'Physical Silver 30kg Bars',
      vaultLocation: 'Malca-Amit Vault, Ahmedabad, Gujarat',
      purity: '999.0 Fine Silver',
      allocatedKg: 1200.0,
      vaultReceiptId: 'EWR-SLV-2026-00891',
      assayStatus: 'CERTIFIED',
      redeemable: true,
    },
  ]);

  return (
    <div className="space-y-6 text-white font-sans">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">TOTAL ALLOCATED GOLD</div>
          <div className="text-xl font-bold font-mono text-amber-400 mt-1">50.000 KG</div>
          <div className="text-[11px] text-slate-400 mt-1">Insured by Lloyd’s Syndicate</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">TOTAL ALLOCATED SILVER</div>
          <div className="text-xl font-bold font-mono text-slate-200 mt-1">1,200.000 KG</div>
          <div className="text-[11px] text-slate-400 mt-1">Assayed by NABL Laboratory</div>
        </div>
        <div className="p-4 rounded border border-slate-800 bg-[#111620]">
          <div className="text-xs font-mono text-slate-400">ELECTRONIC VAULT RECEIPTS</div>
          <div className="text-xl font-bold font-mono text-[#00F0A0] mt-1">2 Active EWRs</div>
          <div className="text-[11px] text-slate-400 mt-1">WDRA & SEBI Depository Registered</div>
        </div>
      </div>

      <div className="rounded border border-slate-800 bg-[#111620] overflow-hidden">
        <div className="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 className="text-sm font-bold">Allocated Physical Vault Inventory</h2>
          <span className="text-xs font-mono text-slate-400">100% Title Segregated</span>
        </div>
        <div className="divide-y divide-slate-800">
          {items.map((item) => (
            <div key={item.vaultReceiptId} className="p-4 flex flex-col md:flex-row justify-between md:items-center gap-4">
              <div>
                <div className="font-bold text-sm text-white">{item.commodity}</div>
                <div className="text-xs font-mono text-slate-400 mt-0.5">
                  Vault: {item.vaultLocation} | Purity: {item.purity}
                </div>
                <div className="text-xs font-mono text-slate-500 mt-1">
                  EWR: {item.vaultReceiptId} | Status: <span className="text-[#00F0A0]">{item.assayStatus}</span>
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <div className="text-right font-mono text-xs">
                  <div className="text-slate-400">Holding Size</div>
                  <div className="font-bold text-white text-sm">{item.allocatedKg} KG</div>
                </div>
                <button className="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-xs font-mono text-slate-200">
                  Request Physical Delivery
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
