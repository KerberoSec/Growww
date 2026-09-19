import React, { useState } from 'react';

interface TaprootBtcDepositModalProps {
  onClose: () => void;
  userTaprootAddress: string;
}

export const TaprootBtcDepositModal: React.FC<TaprootBtcDepositModalProps> = ({
  onClose,
  userTaprootAddress,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(userTaprootAddress);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
      <div className="bg-[#0B0E14] border border-gray-800 rounded-xl max-w-md w-full p-6 text-white shadow-2xl">
        <div className="flex justify-between items-center mb-4">
          <div className="flex items-center gap-2">
            <span className="w-3 h-3 rounded-full bg-amber-500" />
            <h3 className="text-lg font-bold">Bitcoin (BTC) Deposit</h3>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white">✕</button>
        </div>

        <div className="bg-[#141824] p-3 rounded-lg border border-amber-900/30 mb-4 text-xs text-amber-300">
          <strong>Native Taproot (P2TR / Bech32m):</strong> Deposit Bitcoin directly via Native Taproot addresses for minimum mining fees and enhanced multi-sig privacy.
        </div>

        {/* QR Code Placeholder */}
        <div className="flex justify-center p-6 bg-white rounded-lg mb-4 mx-auto w-48 h-48 items-center">
          <div className="text-black font-mono text-center text-xs font-bold">
            [TAPROOT QR CODE]<br />
            <span className="text-[9px] text-gray-600 break-all">{userTaprootAddress.slice(0, 16)}...</span>
          </div>
        </div>

        {/* Address Copy Bar */}
        <div className="mb-4">
          <label className="block text-xs font-medium text-gray-400 mb-1">Your Dedicated Taproot Address</label>
          <div className="flex gap-2">
            <input
              type="text"
              readOnly
              value={userTaprootAddress}
              className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-xs text-amber-400 font-mono"
            />
            <button
              onClick={handleCopy}
              className="px-3 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-semibold"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        </div>

        <div className="space-y-1 text-[11px] text-gray-400 font-mono mb-6">
          <div className="flex justify-between">
            <span>Network:</span>
            <span className="text-white">Bitcoin Mainnet (Taproot)</span>
          </div>
          <div className="flex justify-between">
            <span>Minimum Deposit:</span>
            <span className="text-white">0.0001 BTC</span>
          </div>
          <div className="flex justify-between">
            <span>Required Confirmations:</span>
            <span className="text-white">1 Confirmation (Auto-Credit)</span>
          </div>
        </div>

        <button
          onClick={onClose}
          className="w-full py-2.5 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-lg text-xs font-semibold"
        >
          Done
        </button>
      </div>
    </div>
  );
};
