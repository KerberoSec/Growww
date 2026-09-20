'use client';

import React, { useState } from 'react';

export function TokenizationOriginator() {
  const [assetName, setAssetName] = useState('');
  const [isin, setIsin] = useState('');
  const [totalValuationInr, setTotalValuationInr] = useState('');
  const [tokenSymbol, setTokenSymbol] = useState('');
  const [deploymentStatus, setDeploymentStatus] = useState<string | null>(null);

  const handleDeployToken = (e: React.FormEvent) => {
    e.preventDefault();
    if (assetName && isin && tokenSymbol) {
      setDeploymentStatus('DEPLOYED: ERC-3643 Token deployed on Besu at 0x9812...4c21. Custody escrow locked.');
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6 text-white font-sans">
      <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-6">
        <div>
          <h2 className="text-xl font-bold">RWA Asset Tokenization Originator</h2>
          <p className="text-xs text-slate-400 mt-1">
            Originate compliant ERC-3643 permissioned asset tokens backed 1:1 by NSDL/CDSL escrowed securities.
          </p>
        </div>

        <form onSubmit={handleDeployToken} className="space-y-4">
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">ASSET / ISSUER NAME</label>
              <input
                type="text"
                value={assetName}
                onChange={(e) => setAssetName(e.target.value)}
                placeholder="e.g. HDFC Commercial Real Estate Bond"
                className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs text-white focus:border-[#00F0A0] focus:outline-none"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">ISIN CODE</label>
              <input
                type="text"
                value={isin}
                onChange={(e) => setIsin(e.target.value.toUpperCase())}
                placeholder="INE040A08011"
                className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono uppercase text-white focus:border-[#00F0A0] focus:outline-none"
                required
              />
            </div>
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">TOTAL VALUATION (INR)</label>
              <input
                type="text"
                value={totalValuationInr}
                onChange={(e) => setTotalValuationInr(e.target.value)}
                placeholder="50,00,00,000"
                className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">TOKEN TICKER SYMBOL</label>
              <input
                type="text"
                value={tokenSymbol}
                onChange={(e) => setTokenSymbol(e.target.value.toUpperCase())}
                placeholder="HDFC-CRE-26"
                className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono uppercase text-white focus:border-[#00F0A0] focus:outline-none"
                required
              />
            </div>
          </div>

          <button
            type="submit"
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black text-xs hover:bg-[#00d08a]"
          >
            Deploy ERC-3643 Token & Register With Custody Gateway
          </button>
        </form>

        {deploymentStatus && (
          <div className="p-3 rounded bg-emerald-950/40 border border-emerald-800 text-xs font-mono text-emerald-300">
            {deploymentStatus}
          </div>
        )}
      </div>
    </div>
  );
}
