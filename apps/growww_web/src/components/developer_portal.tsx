'use client';

import React, { useState } from 'react';

export function DeveloperPortal() {
  const [apiKey, setApiKey] = useState('gw_test_91a8c4f2e0...8b');
  const [faucetAddress, setFaucetAddress] = useState('');
  const [faucetClaimStatus, setFaucetClaimStatus] = useState<string | null>(null);

  const handleGenerateKey = () => {
    setApiKey(`gw_live_${Math.random().toString(36).substring(2, 15)}_${Date.now().toString(36)}`);
  };

  const handleClaimFaucet = (e: React.FormEvent) => {
    e.preventDefault();
    if (faucetAddress.startsWith('0x') && faucetAddress.length === 42) {
      setFaucetClaimStatus('SUCCESS: Credited 10,000 vUSDT & 1.00 vETH on Besu Testnet');
    } else {
      setFaucetClaimStatus('ERROR: Invalid Ethereum address format');
    }
  };

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-8 text-white">
      <div>
        <h1 className="text-3xl font-black">Developer Portal & Web3 Faucet</h1>
        <p className="text-sm text-slate-400 mt-1">
          High-performance REST, gRPC-Web, and FIX 4.4 API integrations for algorithmic trading.
        </p>
      </div>

      {/* API Key Management */}
      <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-4">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-bold">Sandbox API Credentials</h2>
            <p className="text-xs text-slate-400">Use this token for paper trading and testnet execution.</p>
          </div>
          <button
            onClick={handleGenerateKey}
            className="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-xs font-mono text-[#00F0A0]"
          >
            Regenerate Secret
          </button>
        </div>
        <div className="p-3 rounded bg-black/40 border border-slate-800 font-mono text-xs flex justify-between items-center">
          <span className="text-slate-300">{apiKey}</span>
          <span className="text-slate-500 text-[10px]">Tier-2 Sandboxed</span>
        </div>
      </div>

      {/* Testnet Faucet */}
      <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-4">
        <h2 className="text-lg font-bold">Besu Testnet & Virtual Demat Faucet</h2>
        <p className="text-xs text-slate-400">
          Request test collateral to test order routing, options chains, and DvP settlement.
        </p>
        <form onSubmit={handleClaimFaucet} className="space-y-4">
          <div>
            <label className="block text-xs font-mono text-slate-400 mb-1">EVM RECIPIENT ADDRESS</label>
            <input
              type="text"
              value={faucetAddress}
              onChange={(e) => setFaucetAddress(e.target.value)}
              placeholder="0x1234567890123456789012345678901234567890"
              className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
              required
            />
          </div>
          <button
            type="submit"
            className="rounded bg-[#00F0A0] px-5 py-2.5 font-bold text-black text-xs hover:bg-[#00d08a]"
          >
            Request 10,000 vUSDT Faucet
          </button>
        </form>

        {faucetClaimStatus && (
          <div
            className={`p-3 rounded text-xs font-mono ${
              faucetClaimStatus.startsWith('SUCCESS')
                ? 'bg-emerald-950/50 border border-emerald-800 text-emerald-300'
                : 'bg-red-950/50 border border-red-800 text-red-300'
            }`}
          >
            {faucetClaimStatus}
          </div>
        )}
      </div>

      {/* Interactive API Explorer snippet */}
      <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-3 font-mono text-xs">
        <div className="text-slate-400 font-bold uppercase">Sample Market Data Request (cURL)</div>
        <pre className="p-4 rounded bg-black/60 border border-slate-800 text-[#00F0A0] overflow-x-auto">
{`curl -X GET "https://api.growww.in/v1/market/orderbook?pair=BTC-USDT&depth=20" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Accept: application/json"`}
        </pre>
      </div>
    </div>
  );
}
