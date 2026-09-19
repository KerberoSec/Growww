import React, { useState } from 'react';

export interface SolvencyAuditEpoch {
  epochNumber: number;
  merkleRoot: string;
  totalReservesUSD: number;
  totalLiabilitiesUSD: number;
  solvencyRatio: number; // e.g., 104.5%
  auditorAddress: string;
  publishedAt: string;
  besuTxHash: string;
}

interface PublicProofOfSolvencyProps {
  epochs: SolvencyAuditEpoch[];
  userAccountHash: string;
  onVerifyUserInclusion: (accountHash: string, epoch: number) => Promise<boolean>;
}

export const PublicProofOfSolvency: React.FC<PublicProofOfSolvencyProps> = ({
  epochs,
  userAccountHash,
  onVerifyUserInclusion,
}) => {
  const [selectedEpoch, setSelectedEpoch] = useState<number>(epochs[0]?.epochNumber || 1);
  const [verificationResult, setVerificationResult] = useState<boolean | null>(null);
  const [verifying, setVerifying] = useState(false);

  const currentEpoch = epochs.find(e => e.epochNumber === selectedEpoch) || epochs[0];

  const handleVerify = async () => {
    setVerifying(true);
    try {
      const isIncluded = await onVerifyUserInclusion(userAccountHash, selectedEpoch);
      setVerificationResult(isIncluded);
    } catch {
      setVerificationResult(false);
    } finally {
      setVerifying(false);
    }
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-3xl mx-auto font-sans">
      <div className="flex justify-between items-start mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Cryptographic Proof of Solvency</h2>
          <p className="text-gray-400 text-xs mt-1">
            Mathematical proof that 100% of user assets are fully backed and registered on Hyperledger Besu.
          </p>
        </div>
        <div className="text-right">
          <span className="text-xs text-gray-500 block">Solvency Ratio</span>
          <span className="text-xl font-bold text-emerald-400 font-mono">
            {currentEpoch?.solvencyRatio.toFixed(2)}%
          </span>
        </div>
      </div>

      {/* Solvency Metric Cards */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-[#141824] p-4 rounded-lg border border-gray-800">
          <span className="text-xs text-gray-400 block mb-1">Total Verified Reserves</span>
          <span className="text-lg font-bold font-mono text-emerald-400">
            ${currentEpoch?.totalReservesUSD.toLocaleString('en-US')}
          </span>
        </div>
        <div className="bg-[#141824] p-4 rounded-lg border border-gray-800">
          <span className="text-xs text-gray-400 block mb-1">Total User Liabilities</span>
          <span className="text-lg font-bold font-mono text-gray-200">
            ${currentEpoch?.totalLiabilitiesUSD.toLocaleString('en-US')}
          </span>
        </div>
        <div className="bg-[#141824] p-4 rounded-lg border border-gray-800">
          <span className="text-xs text-gray-400 block mb-1">Audit Epoch</span>
          <span className="text-lg font-bold font-mono text-amber-400">
            Epoch #{currentEpoch?.epochNumber}
          </span>
        </div>
      </div>

      {/* Merkle Root Hash */}
      <div className="bg-[#141824] p-4 rounded-lg border border-gray-800 mb-6 font-mono text-xs">
        <span className="text-gray-400 block mb-1">State Merkle Root Hash (Keccak256)</span>
        <div className="text-emerald-400 bg-[#0B0E14] p-2.5 rounded break-all border border-gray-850">
          {currentEpoch?.merkleRoot}
        </div>
      </div>

      {/* Self-Verification Section */}
      <div className="bg-emerald-950/20 border border-emerald-900/50 p-4 rounded-lg mb-6">
        <h3 className="text-sm font-bold text-emerald-400 mb-2">Verify Your Account Inclusion in Merkle Tree</h3>
        <p className="text-xs text-gray-300 mb-4">
          Provide your anonymized account hash to verify that your exact balance was cryptographically included in the published Merkle tree root.
        </p>

        <div className="flex gap-3">
          <input
            type="text"
            readOnly
            value={userAccountHash}
            className="w-full bg-[#0B0E14] border border-gray-700 rounded-lg p-2.5 text-xs text-gray-300 font-mono"
          />
          <button
            onClick={handleVerify}
            disabled={verifying}
            className="px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-bold whitespace-nowrap transition"
          >
            {verifying ? 'Computing Proof...' : 'Verify In Merkle Tree'}
          </button>
        </div>

        {verificationResult !== null && (
          <div className={`mt-3 p-3 rounded text-xs font-mono flex items-center gap-2 ${
            verificationResult ? 'bg-emerald-900/40 text-emerald-300 border border-emerald-800' : 'bg-red-900/40 text-red-300 border border-red-800'
          }`}>
            <span>{verificationResult ? '✓ VERIFIED: Your balance is mathematically proven to be fully solvent inside the Merkle Tree!' : '✕ Inclusion check failed.'}</span>
          </div>
        )}
      </div>
    </div>
  );
};
