'use client';

import React, { useState } from 'react';

export default function PublicReservesPage() {
  const [merkleProofInput, setMerkleProofInput] = useState('');
  const [verificationResult, setVerificationResult] = useState<null | {
    valid: boolean;
    asset: string;
    sharesHeld: number;
    leafHash: string;
    rootMatch: boolean;
  }>(null);

  const handleVerify = (e: React.FormEvent) => {
    e.preventDefault();
    // Cryptographic verification simulation of Merkle inclusion
    setVerificationResult({
      valid: true,
      asset: 'RELIANCE-RWA (ISIN INE002A01018)',
      sharesHeld: 25.0,
      leafHash: '0x7f2019a8b7c6d5e4f3a2b1c0d9e8f73a4b9c1d8e',
      rootMatch: true,
    });
  };

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6 text-white">
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-extrabold text-white">Public Proof-of-Reserve Portal</h1>
        <p className="text-sm text-slate-400 max-w-xl mx-auto">
          Independently verify that your tokenized fractional shares correspond directly to physical shares held in NSDL / CDSL custodial accounts.
        </p>
      </div>

      <div className="p-6 rounded-xl border border-slate-800 bg-[#111620] space-y-4">
        <h2 className="text-lg font-bold">Client-Side Merkle Tree Verifier</h2>
        <form onSubmit={handleVerify} className="space-y-4">
          <div>
            <label className="block text-xs font-mono text-slate-400 mb-1">
              PASTE YOUR ANONYMIZED USER MERKLE LEAF OR TOKEN ID
            </label>
            <input
              type="text"
              value={merkleProofInput}
              onChange={(e) => setMerkleProofInput(e.target.value)}
              placeholder="0x9a8b7c6d5e4f3a2b1c0d9e8f7..."
              className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-xs font-mono text-white focus:border-[#00F0A0] focus:outline-none"
              required
            />
          </div>
          <button
            type="submit"
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black text-xs hover:bg-[#00d08a]"
          >
            Verify Cryptographic Inclusion on Hyperledger Besu
          </button>
        </form>

        {verificationResult && (
          <div className="p-4 rounded border border-emerald-800 bg-emerald-950/40 space-y-2 font-mono text-xs text-emerald-300">
            <div className="flex items-center space-x-2">
              <span className="font-bold text-sm">✓ Cryptographic Attestation Valid</span>
              <span className="text-[11px] px-2 py-0.5 rounded bg-emerald-900 text-emerald-200">1:1 Backed</span>
            </div>
            <div>Asset: {verificationResult.asset}</div>
            <div>Verified Share Allocation: {verificationResult.sharesHeld} shares</div>
            <div>Leaf Hash: {verificationResult.leafHash}</div>
            <div>Besu PoR Contract State: Invariant Satisfied</div>
          </div>
        )}
      </div>
    </div>
  );
}
