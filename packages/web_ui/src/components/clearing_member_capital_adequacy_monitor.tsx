import React, { useState } from 'react';

export interface SectoralHeadroom {
  sector: string;
  regulatoryCapPct: number;
  currentUtilizationPct: number;
  totalLimitInrCr: number;
  utilizedInrCr: number;
  fpiCategory: 'CAT_I' | 'CAT_II';
}

export function ClearingMemberCapitalAdequacyMonitor() {
  const [selectedMember] = useState<string>('GROWWW-CM-0092');

  const [capitalMetrics] = useState({
    netWorthInrCr: 250.0,
    baseMinimumCapitalInrCr: 50.0,
    totalCollateralDepositedCr: 180.0,
    initialMarginRequirementCr: 95.4,
    extremeLossMarginCr: 12.6,
    mtmMarginCr: 4.8,
    peakIntradayExposureCr: 420.0,
    maxExposureLimitCr: 600.0,
    capitalAdequacyRatioPct: 218.5, // Min SEBI requirement: 100%
  });

  const [sectors] = useState<SectoralHeadroom[]>([
    {
      sector: 'Banking & Financial Services',
      regulatoryCapPct: 74.0,
      currentUtilizationPct: 58.2,
      totalLimitInrCr: 1480.0,
      utilizedInrCr: 861.36,
      fpiCategory: 'CAT_I',
    },
    {
      sector: 'Information Technology',
      regulatoryCapPct: 100.0,
      currentUtilizationPct: 42.1,
      totalLimitInrCr: 2000.0,
      utilizedInrCr: 842.0,
      fpiCategory: 'CAT_I',
    },
    {
      sector: 'Defense & Aerospace',
      regulatoryCapPct: 49.0,
      currentUtilizationPct: 44.8, // Close to cap!
      totalLimitInrCr: 490.0,
      utilizedInrCr: 219.52,
      fpiCategory: 'CAT_II',
    },
    {
      sector: 'Crypto-Asset Derivatives Index',
      regulatoryCapPct: 20.0,
      currentUtilizationPct: 16.5,
      totalLimitInrCr: 500.0,
      utilizedInrCr: 82.5,
      fpiCategory: 'CAT_I',
    },
    {
      sector: 'Telecommunications',
      regulatoryCapPct: 100.0,
      currentUtilizationPct: 88.4,
      totalLimitInrCr: 1200.0,
      utilizedInrCr: 1060.8,
      fpiCategory: 'CAT_I',
    },
  ]);

  const totalMarginRequired =
    capitalMetrics.initialMarginRequirementCr +
    capitalMetrics.extremeLossMarginCr +
    capitalMetrics.mtmMarginCr;

  const marginHeadroomCr = capitalMetrics.totalCollateralDepositedCr - totalMarginRequired;
  const exposureUtilizationPct = (capitalMetrics.peakIntradayExposureCr / capitalMetrics.maxExposureLimitCr) * 100;

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-2">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00F0A0] inline-block" />
            Clearing Member Capital Adequacy & FPI Sectoral Headroom Monitor
          </h2>
          <p className="text-[11px] text-slate-400 mt-0.5">
            SEBI Institutional Risk Surveillance • Base Minimum Capital (BMC) & Cross-Border Sector Limits
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 text-slate-300 font-bold">
            CM ID: {selectedMember}
          </span>
          <span className="px-2 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] font-bold">
            COMPLIANT (CAR: {capitalMetrics.capitalAdequacyRatioPct}%)
          </span>
        </div>
      </div>

      {/* Capital Adequacy & Margin Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Net Worth & BMC</span>
          <div className="text-base font-bold text-white mt-1">₹{capitalMetrics.netWorthInrCr} Cr</div>
          <span className="text-[10px] text-slate-400">BMC Min: ₹{capitalMetrics.baseMinimumCapitalInrCr} Cr</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Collateral Deposited</span>
          <div className="text-base font-bold text-cyan-400 mt-1">₹{capitalMetrics.totalCollateralDepositedCr} Cr</div>
          <span className="text-[10px] text-slate-400">Cash + Sovereign G-Secs</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Total Margin Obligation</span>
          <div className="text-base font-bold text-amber-400 mt-1">₹{totalMarginRequired.toFixed(1)} Cr</div>
          <span className="text-[10px] text-slate-500">
            SPAN ₹{capitalMetrics.initialMarginRequirementCr}Cr + ELM ₹{capitalMetrics.extremeLossMarginCr}Cr
          </span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Available Margin Buffer</span>
          <div className="text-base font-bold text-[#00F0A0] mt-1">₹{marginHeadroomCr.toFixed(1)} Cr</div>
          <span className="text-[10px] text-[#00F0A0]">Free Collateral Surplus</span>
        </div>
      </div>

      {/* Intraday Exposure Limit Gauge */}
      <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 space-y-2">
        <div className="flex justify-between items-center text-xs">
          <span className="font-bold text-slate-200">Intraday Peak Exposure vs Clearing Limit</span>
          <span className="text-slate-300">
            ₹{capitalMetrics.peakIntradayExposureCr} Cr / ₹{capitalMetrics.maxExposureLimitCr} Cr ({exposureUtilizationPct.toFixed(1)}%)
          </span>
        </div>
        <div className="w-full bg-slate-800 rounded-full h-2.5 overflow-hidden">
          <div
            className={`h-full transition-all ${
              exposureUtilizationPct > 85 ? 'bg-[#FF3B56]' : exposureUtilizationPct > 70 ? 'bg-amber-400' : 'bg-[#00F0A0]'
            }`}
            style={{ width: `${Math.min(100, exposureUtilizationPct)}%` }}
          />
        </div>
        <div className="flex justify-between text-[10px] text-slate-500">
          <span>0 Cr</span>
          <span>Warning Threshold (75%)</span>
          <span>Risk Halt Threshold (90%)</span>
          <span>Max Limit ₹{capitalMetrics.maxExposureLimitCr} Cr</span>
        </div>
      </div>

      {/* FPI Limit Monitor Sectoral Headroom Bar Table */}
      <div className="space-y-2">
        <div className="flex justify-between items-center text-xs">
          <span className="font-bold text-white">FPI Sectoral Investment Cap & Headroom Bar</span>
          <span className="text-[11px] text-slate-400">NSDL / CDSL Consolidated Foreign Cap Feed</span>
        </div>

        <div className="overflow-x-auto rounded border border-slate-800">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-[#181F2C] text-slate-400 text-[11px] border-b border-slate-800">
                <th className="p-2.5">Sector</th>
                <th className="p-2.5 text-center">FPI Category</th>
                <th className="p-2.5 text-right">Regulatory Cap</th>
                <th className="p-2.5 text-right">Utilized / Limit (Cr)</th>
                <th className="p-2.5 w-48">Headroom Bar</th>
                <th className="p-2.5 text-center">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 bg-[#121721]/50 text-[11px]">
              {sectors.map((s) => {
                const headroomPct = Math.max(0, s.regulatoryCapPct - s.currentUtilizationPct);
                const isCritical = headroomPct < 5.0;
                const isWarning = headroomPct < 15.0 && !isCritical;

                return (
                  <tr key={s.sector} className="hover:bg-slate-800/40 transition-colors">
                    <td className="p-2.5 font-bold text-white">{s.sector}</td>
                    <td className="p-2.5 text-center">
                      <span className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px] text-slate-300">
                        {s.fpiCategory}
                      </span>
                    </td>
                    <td className="p-2.5 text-right font-bold text-slate-200">{s.regulatoryCapPct}%</td>
                    <td className="p-2.5 text-right text-slate-300">
                      ₹{s.utilizedInrCr.toFixed(0)} / ₹{s.totalLimitInrCr} Cr
                    </td>
                    <td className="p-2.5">
                      <div className="space-y-1">
                        <div className="flex justify-between text-[10px]">
                          <span className="text-slate-400">{s.currentUtilizationPct}% used</span>
                          <span className={isCritical ? 'text-[#FF3B56] font-bold' : isWarning ? 'text-amber-400' : 'text-[#00F0A0]'}>
                            {headroomPct.toFixed(1)}% buffer
                          </span>
                        </div>
                        <div className="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
                          <div
                            className={`h-full ${
                              isCritical ? 'bg-[#FF3B56]' : isWarning ? 'bg-amber-400' : 'bg-[#00F0A0]'
                            }`}
                            style={{ width: `${Math.min(100, (s.currentUtilizationPct / s.regulatoryCapPct) * 100)}%` }}
                          />
                        </div>
                      </div>
                    </td>
                    <td className="p-2.5 text-center">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                          isCritical
                            ? 'bg-[#FF3B56]/20 text-[#FF3B56] border border-[#FF3B56]/30'
                            : isWarning
                            ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                            : 'bg-[#00F0A0]/20 text-[#00F0A0] border border-[#00F0A0]/30'
                        }`}
                      >
                        {isCritical ? 'RED BREACH RISK' : isWarning ? 'AMBER ALERT' : 'OPEN HEADROOM'}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export const FpiLimitMonitorSectoralHeadroomBar = ClearingMemberCapitalAdequacyMonitor;
export const WebClearingMemberCapitalAdequacyMonitor = ClearingMemberCapitalAdequacyMonitor;
