import React, { useState } from 'react';

export interface EquityHolding {
  isin: string;
  symbol: string;
  name: string;
  depository: 'CDSL' | 'NSDL';
  totalQty: number;
  pledgedQty: number;
  cmpInr: number;
  haircutPct: number;
}

export function MarginPledgePortal() {
  const [holdings, setHoldings] = useState<EquityHolding[]>([
    {
      isin: 'INE002A01018',
      symbol: 'RELIANCE',
      name: 'Reliance Industries Ltd',
      depository: 'CDSL',
      totalQty: 100,
      pledgedQty: 40,
      cmpInr: 2950.0,
      haircutPct: 15,
    },
    {
      isin: 'INE009A01021',
      symbol: 'INFY',
      name: 'Infosys Ltd',
      depository: 'CDSL',
      totalQty: 250,
      pledgedQty: 100,
      cmpInr: 1620.0,
      haircutPct: 15,
    },
    {
      isin: 'INE040A01034',
      symbol: 'HDFCBANK',
      name: 'HDFC Bank Ltd',
      depository: 'NSDL',
      totalQty: 200,
      pledgedQty: 0,
      cmpInr: 1680.0,
      haircutPct: 15,
    },
    {
      isin: 'INE155A01022',
      symbol: 'TATAMOTORS',
      name: 'Tata Motors Ltd',
      depository: 'CDSL',
      totalQty: 500,
      pledgedQty: 200,
      cmpInr: 980.0,
      haircutPct: 20,
    },
  ]);

  const [selectedIsin, setSelectedIsin] = useState<string | null>(null);
  const [pledgeAction, setPledgeAction] = useState<'PLEDGE' | 'UNPLEDGE'>('PLEDGE');
  const [qtyInput, setQtyInput] = useState<number>(10);
  const [otpSent, setOtpSent] = useState<boolean>(false);
  const [otpValue, setOtpValue] = useState<string>('');

  const selectedHolding = holdings.find((h) => h.isin === selectedIsin);

  // Calculations
  const totalPortfolioValue = holdings.reduce((sum, h) => sum + h.totalQty * h.cmpInr, 0);
  const totalPledgedValue = holdings.reduce((sum, h) => sum + h.pledgedQty * h.cmpInr, 0);
  const totalMarginUnlocked = holdings.reduce(
    (sum, h) => sum + h.pledgedQty * h.cmpInr * (1 - h.haircutPct / 100),
    0
  );
  const marginUtilized = 285000; // Simulated active margin used in open crypto positions
  const marginUtilizationPct = totalMarginUnlocked > 0 ? (marginUtilized / totalMarginUnlocked) * 100 : 0;

  const handleAction = (isin: string, action: 'PLEDGE' | 'UNPLEDGE') => {
    setSelectedIsin(isin);
    setPledgeAction(action);
    setOtpSent(false);
    setOtpValue('');
    const h = holdings.find((x) => x.isin === isin);
    if (h) {
      const max = action === 'PLEDGE' ? h.totalQty - h.pledgedQty : h.pledgedQty;
      setQtyInput(Math.min(10, max));
    }
  };

  const handleConfirmAction = () => {
    if (!selectedHolding) return;
    if (pledgeAction === 'PLEDGE') {
      setHoldings((prev) =>
        prev.map((h) =>
          h.isin === selectedHolding.isin
            ? { ...h, pledgedQty: Math.min(h.totalQty, h.pledgedQty + qtyInput) }
            : h
        )
      );
    } else {
      setHoldings((prev) =>
        prev.map((h) =>
          h.isin === selectedHolding.isin
            ? { ...h, pledgedQty: Math.max(0, h.pledgedQty - qtyInput) }
            : h
        )
      );
    }
    setSelectedIsin(null);
  };

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-2">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00F0A0] inline-block" />
            Margin Pledge Portal: Demat Equity to Instant Crypto Margin
          </h2>
          <p className="text-[11px] text-slate-400 mt-0.5">
            SEBI & Depository (NSDL/CDSL) compliant re-pledge mechanism with automated risk haircuts
          </p>
        </div>
        <div className="flex items-center gap-2 text-[11px]">
          <span className="px-2 py-1 rounded bg-[#181F2C] border border-slate-700 text-slate-300">
            BO ID: 1208160099482110 (CDSL)
          </span>
        </div>
      </div>

      {/* KPI Overview */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Total Demat Holdings</span>
          <div className="text-base font-bold text-white mt-1">₹{totalPortfolioValue.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-slate-500">4 Scrips Eligible</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Collateral Pledged</span>
          <div className="text-base font-bold text-cyan-400 mt-1">₹{totalPledgedValue.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-slate-400">NSDL / CDSL Lien marked</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Margin Unlocked</span>
          <div className="text-base font-bold text-[#00F0A0] mt-1">₹{totalMarginUnlocked.toLocaleString('en-IN')}</div>
          <span className="text-[10px] text-[#00F0A0]">Instant Crypto Trading Power</span>
        </div>

        <div className="bg-[#121721] p-3 rounded border border-slate-800">
          <span className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Margin Utilization</span>
          <div className="text-base font-bold text-amber-400 mt-1">
            {marginUtilizationPct.toFixed(1)}%
          </div>
          <div className="w-full bg-slate-800 rounded-full h-1.5 mt-1.5 overflow-hidden">
            <div
              className={`h-full ${marginUtilizationPct > 80 ? 'bg-[#FF3B56]' : 'bg-[#00F0A0]'}`}
              style={{ width: `${Math.min(100, marginUtilizationPct)}%` }}
            />
          </div>
        </div>
      </div>

      {/* Holdings Table */}
      <div className="overflow-x-auto rounded border border-slate-800">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-[#181F2C] text-slate-400 text-[11px] border-b border-slate-800">
              <th className="p-2.5">Scrip / ISIN</th>
              <th className="p-2.5 text-center">Depository</th>
              <th className="p-2.5 text-right">Available / Total Qty</th>
              <th className="p-2.5 text-right">CMP (INR)</th>
              <th className="p-2.5 text-right">Haircut</th>
              <th className="p-2.5 text-right">Pledged Qty</th>
              <th className="p-2.5 text-right">Margin Unlocked</th>
              <th className="p-2.5 text-center">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60 bg-[#121721]/50 text-[11px]">
            {holdings.map((h) => {
              const availableQty = h.totalQty - h.pledgedQty;
              const unlockedMargin = h.pledgedQty * h.cmpInr * (1 - h.haircutPct / 100);

              return (
                <tr key={h.isin} className="hover:bg-slate-800/40 transition-colors">
                  <td className="p-2.5">
                    <div className="font-bold text-white">{h.symbol}</div>
                    <div className="text-[10px] text-slate-500">{h.isin} - {h.name}</div>
                  </td>
                  <td className="p-2.5 text-center">
                    <span className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px] font-bold text-slate-300">
                      {h.depository}
                    </span>
                  </td>
                  <td className="p-2.5 text-right text-slate-200">
                    <span className="font-bold text-[#00F0A0]">{availableQty}</span> / {h.totalQty}
                  </td>
                  <td className="p-2.5 text-right text-slate-200">₹{h.cmpInr.toLocaleString('en-IN')}</td>
                  <td className="p-2.5 text-right text-amber-400 font-bold">{h.haircutPct}%</td>
                  <td className="p-2.5 text-right font-bold text-cyan-400">{h.pledgedQty}</td>
                  <td className="p-2.5 text-right font-bold text-[#00F0A0]">
                    ₹{unlockedMargin.toLocaleString('en-IN')}
                  </td>
                  <td className="p-2.5 text-center">
                    <div className="flex justify-center gap-1.5">
                      <button
                        disabled={availableQty === 0}
                        onClick={() => handleAction(h.isin, 'PLEDGE')}
                        className={`px-2 py-1 rounded font-bold text-[10px] transition-colors ${
                          availableQty > 0
                            ? 'bg-[#00F0A0]/20 hover:bg-[#00F0A0]/30 text-[#00F0A0] border border-[#00F0A0]/40'
                            : 'bg-slate-800 text-slate-600 cursor-not-allowed'
                        }`}
                      >
                        Pledge
                      </button>
                      <button
                        disabled={h.pledgedQty === 0}
                        onClick={() => handleAction(h.isin, 'UNPLEDGE')}
                        className={`px-2 py-1 rounded font-bold text-[10px] transition-colors ${
                          h.pledgedQty > 0
                            ? 'bg-[#FF3B56]/20 hover:bg-[#FF3B56]/30 text-[#FF3B56] border border-[#FF3B56]/40'
                            : 'bg-slate-800 text-slate-600 cursor-not-allowed'
                        }`}
                      >
                        Unpledge
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Action Modal / Drawer */}
      {selectedHolding && (
        <div className="p-4 rounded-lg bg-[#181F2C] border border-slate-700 space-y-3">
          <div className="flex justify-between items-center border-b border-slate-700 pb-2">
            <span className="font-bold text-white text-sm">
              {pledgeAction === 'PLEDGE' ? 'Pledge Shares for Margin' : 'Unpledge Collateral'} - {selectedHolding.symbol}
            </span>
            <button
              onClick={() => setSelectedIsin(null)}
              className="text-slate-400 hover:text-white text-sm font-bold"
            >
              ✕
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-slate-400 text-[11px] block">
                Quantity to {pledgeAction === 'PLEDGE' ? 'Pledge' : 'Unpledge'} (Max:{' '}
                {pledgeAction === 'PLEDGE'
                  ? selectedHolding.totalQty - selectedHolding.pledgedQty
                  : selectedHolding.pledgedQty}
                )
              </label>
              <div className="flex items-center gap-2">
                <input
                  type="number"
                  min="1"
                  max={
                    pledgeAction === 'PLEDGE'
                      ? selectedHolding.totalQty - selectedHolding.pledgedQty
                      : selectedHolding.pledgedQty
                  }
                  value={qtyInput}
                  onChange={(e) => setQtyInput(Math.max(1, parseInt(e.target.value) || 1))}
                  className="bg-[#121721] border border-slate-600 px-3 py-1.5 rounded text-white w-32 focus:border-[#00F0A0] focus:outline-none font-bold"
                />
                <button
                  onClick={() =>
                    setQtyInput(
                      pledgeAction === 'PLEDGE'
                        ? selectedHolding.totalQty - selectedHolding.pledgedQty
                        : selectedHolding.pledgedQty
                    )
                  }
                  className="px-2 py-1 bg-slate-800 text-slate-300 rounded hover:bg-slate-700 text-[10px]"
                >
                  MAX
                </button>
              </div>

              <div className="text-[11px] text-slate-400 space-y-1 pt-1">
                <div>Haircut Applicable: <span className="text-amber-400 font-bold">{selectedHolding.haircutPct}%</span></div>
                <div>
                  Margin Impact:{' '}
                  <span className="text-[#00F0A0] font-bold">
                    ₹{(qtyInput * selectedHolding.cmpInr * (1 - selectedHolding.haircutPct / 100)).toLocaleString('en-IN')}
                  </span>
                </div>
              </div>
            </div>

            <div className="space-y-2 bg-[#121721] p-3 rounded border border-slate-800">
              <div className="text-[11px] text-slate-300 font-bold">CDSL / NSDL e-DIS Authentication</div>
              {!otpSent ? (
                <div className="space-y-2">
                  <p className="text-[10px] text-slate-400">
                    Authorization required via Depository OTP sent to registered mobile/email.
                  </p>
                  <button
                    onClick={() => setOtpSent(true)}
                    className="px-3 py-1.5 rounded bg-blue-600 hover:bg-blue-500 text-white font-bold text-xs"
                  >
                    Send Depository OTP
                  </button>
                </div>
              ) : (
                <div className="space-y-2">
                  <span className="text-[10px] text-[#00F0A0]">OTP Sent to Depository Mobile</span>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      placeholder="Enter 6-digit OTP"
                      value={otpValue}
                      onChange={(e) => setOtpValue(e.target.value)}
                      maxLength={6}
                      className="bg-[#0B0E14] border border-slate-700 px-3 py-1 rounded text-white text-xs w-36 focus:border-[#00F0A0] focus:outline-none"
                    />
                    <button
                      onClick={handleConfirmAction}
                      disabled={otpValue.length < 4}
                      className={`px-3 py-1 rounded font-bold text-xs ${
                        otpValue.length >= 4
                          ? 'bg-[#00F0A0] text-black hover:bg-[#00D090]'
                          : 'bg-slate-800 text-slate-600 cursor-not-allowed'
                      }`}
                    >
                      Confirm {pledgeAction}
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export const MarginPledgeDepositorySharesPortal = MarginPledgePortal;
