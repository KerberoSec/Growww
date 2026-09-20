import React, { useState } from 'react';

export interface TaxLedgerRecord {
  id: string;
  timestamp: string;
  asset: string;
  tradeType: 'BUY' | 'SELL' | 'SWAP';
  grossValueInr: number;
  tdsRate: number;
  tdsDeductedInr: number;
  realizedGainInr: number;
  vdaTaxRate: number;
  vdaTaxObligationInr: number;
  txHash: string;
}

export function TaxDeductionLedger() {
  const [selectedFy, setSelectedFy] = useState<'FY 2026-27' | 'FY 2025-26'>('FY 2026-27');
  const [filterAsset, setFilterAsset] = useState<string>('ALL');

  const records: TaxLedgerRecord[] = [
    {
      id: 'TXN-9941',
      timestamp: '2026-09-20 09:14:22',
      asset: 'BTC/USDT',
      tradeType: 'SELL',
      grossValueInr: 5400000,
      tdsRate: 0.0,
      tdsDeductedInr: 0,
      realizedGainInr: 450000,
      vdaTaxRate: 30.0,
      vdaTaxObligationInr: 135000,
      txHash: '0x8f3c...4a12',
    },
    {
      id: 'TXN-9940',
      timestamp: '2026-09-19 16:45:10',
      asset: 'ETH/USDT',
      tradeType: 'SELL',
      grossValueInr: 2100000,
      tdsRate: 0.0,
      tdsDeductedInr: 0,
      realizedGainInr: 180000,
      vdaTaxRate: 30.0,
      vdaTaxObligationInr: 54000,
      txHash: '0x1b7e...92e1',
    },
    {
      id: 'TXN-9938',
      timestamp: '2026-09-18 11:20:05',
      asset: 'SOL/USDT',
      tradeType: 'SELL',
      grossValueInr: 850000,
      tdsRate: 0.0,
      tdsDeductedInr: 0,
      realizedGainInr: -45000,
      vdaTaxRate: 30.0,
      vdaTaxObligationInr: 0, // Section 115BBH: Losses cannot be set off
      txHash: '0x3d9a...7c44',
    },
    {
      id: 'TXN-9935',
      timestamp: '2026-09-17 14:02:40',
      asset: 'USDT/INR',
      tradeType: 'SELL',
      grossValueInr: 1200000,
      tdsRate: 1.0,
      tdsDeductedInr: 12000,
      realizedGainInr: 25000,
      vdaTaxRate: 30.0,
      vdaTaxObligationInr: 7500,
      txHash: '0x992b...55d2',
    },
  ];

  const filteredRecords = records.filter(
    (r) => filterAsset === 'ALL' || r.asset.includes(filterAsset)
  );

  const totalGrossVolume = records.reduce((sum, r) => sum + r.grossValueInr, 0);
  const totalGains = records.reduce((sum, r) => (r.realizedGainInr > 0 ? sum + r.realizedGainInr : sum), 0);
  const totalTdsDeducted = records.reduce((sum, r) => sum + r.tdsDeductedInr, 0);
  const totalVdaTax = records.reduce((sum, r) => sum + r.vdaTaxObligationInr, 0);
  const totalCess = totalVdaTax * 0.04;
  const effectiveTotalTax = totalVdaTax + totalCess;

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-2">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00F0A0] inline-block animate-pulse" />
            Tax Ledger: Section 194S & 115BBH VDA Accounting
          </h2>
          <p className="text-[11px] text-slate-400 mt-0.5">
            Institutional zero-TDS on-chain accounting & 30% flat capital gains tracker
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={selectedFy}
            onChange={(e) => setSelectedFy(e.target.value as any)}
            className="bg-[#121721] border border-slate-700 text-slate-200 px-3 py-1 rounded text-xs focus:outline-none focus:border-[#00F0A0]"
          >
            <option value="FY 2026-27">FY 2026-27 (Current)</option>
            <option value="FY 2025-26">FY 2025-26</option>
          </select>
          <button
            onClick={() => alert('Exporting Schedule VDA JSON & Form 26AS reconciliation report...')}
            className="px-3 py-1 bg-[#181F2C] hover:bg-slate-700 text-[#00F0A0] border border-[#00F0A0]/30 rounded font-semibold transition-colors"
          >
            Export Schedule VDA
          </button>
        </div>
      </div>

      {/* Summary KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Gross VDA Turnover</span>
          <div className="text-base font-bold text-white mt-1">₹{totalGrossVolume.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-slate-500">Across 4 settlement batches</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Taxable Gains (No Loss Offset)</span>
          <div className="text-base font-bold text-[#00F0A0] mt-1">+₹{totalGains.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-amber-400">Sec 115BBH strictly enforced</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Sec 194S TDS Deducted</span>
          <div className="text-base font-bold text-cyan-400 mt-1">₹{totalTdsDeducted.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-[#00F0A0]">0% On-Chain QBFT Verified</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">30% VDA Tax + 4% Cess</span>
          <div className="text-base font-bold text-[#FF3B56] mt-1">₹{effectiveTotalTax.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-slate-400">Net Liability (31.2% effective)</span>
        </div>
      </div>

      {/* Filter and Ledger Table */}
      <div className="space-y-2">
        <div className="flex justify-between items-center text-xs">
          <div className="flex gap-2 items-center">
            <span className="text-slate-400">Filter Asset:</span>
            {['ALL', 'BTC', 'ETH', 'SOL', 'USDT'].map((asset) => (
              <button
                key={asset}
                onClick={() => setFilterAsset(asset)}
                className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
                  filterAsset === asset
                    ? 'bg-[#00F0A0] text-black font-bold'
                    : 'bg-[#181F2C] text-slate-400 hover:text-white'
                }`}
              >
                {asset}
              </button>
            ))}
          </div>
          <span className="text-slate-500 text-[10px]">Deterministic Besu Finality</span>
        </div>

        <div className="overflow-x-auto rounded border border-slate-800">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-[#181F2C] text-slate-400 text-[11px] border-b border-slate-800">
                <th className="p-2.5">Date & Time</th>
                <th className="p-2.5">Asset</th>
                <th className="p-2.5">Type</th>
                <th className="p-2.5 text-right">Gross Value</th>
                <th className="p-2.5 text-right">TDS (194S)</th>
                <th className="p-2.5 text-right">Realized Gain</th>
                <th className="p-2.5 text-right">VDA Tax (30%)</th>
                <th className="p-2.5 text-center">Tx Hash</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 bg-[#121721]/50 text-[11px]">
              {filteredRecords.map((r) => (
                <tr key={r.id} className="hover:bg-slate-800/40 transition-colors">
                  <td className="p-2.5 text-slate-300 whitespace-nowrap">{r.timestamp}</td>
                  <td className="p-2.5 font-bold text-white">{r.asset}</td>
                  <td className="p-2.5">
                    <span
                      className={`px-1.5 py-0.5 rounded text-[10px] font-bold ${
                        r.tradeType === 'SELL' ? 'bg-[#FF3B56]/20 text-[#FF3B56]' : 'bg-[#00F0A0]/20 text-[#00F0A0]'
                      }`}
                    >
                      {r.tradeType}
                    </span>
                  </td>
                  <td className="p-2.5 text-right text-slate-200">₹{r.grossValueInr.toLocaleString('en-IN')}</td>
                  <td className="p-2.5 text-right">
                    {r.tdsDeductedInr === 0 ? (
                      <span className="text-[#00F0A0] font-bold">0% (₹0)</span>
                    ) : (
                      <span className="text-amber-400">1% (₹{r.tdsDeductedInr.toLocaleString('en-IN')})</span>
                    )}
                  </td>
                  <td className="p-2.5 text-right font-bold">
                    {r.realizedGainInr > 0 ? (
                      <span className="text-[#00F0A0]">+₹{r.realizedGainInr.toLocaleString('en-IN')}</span>
                    ) : r.realizedGainInr < 0 ? (
                      <span className="text-[#FF3B56]">-₹{Math.abs(r.realizedGainInr).toLocaleString('en-IN')}</span>
                    ) : (
                      <span className="text-slate-400">₹0</span>
                    )}
                  </td>
                  <td className="p-2.5 text-right text-[#FF3B56] font-bold">
                    ₹{r.vdaTaxObligationInr.toLocaleString('en-IN')}
                  </td>
                  <td className="p-2.5 text-center text-slate-400 hover:text-cyan-400 cursor-pointer">
                    {r.txHash}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Compliance Note */}
      <div className="p-3 bg-[#181F2C]/60 rounded border border-slate-800/80 flex items-start gap-2 text-[11px] text-slate-400">
        <span className="text-[#00F0A0] font-bold">ITD Notice Compliance:</span>
        <span>
          Under Section 115BBH of the Income Tax Act 1961, losses from one VDA cannot be set off against gains from any other VDA.
          Section 194S TDS is deducted at source for fiat offramps while zero on-chain friction applies on permissioned settlements.
        </span>
      </div>
    </div>
  );
}
