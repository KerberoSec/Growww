import React from 'react';

export interface MarketOrderSlippageDialogProps {
  isOpen?: boolean;
  estimatedSlippagePct?: number;
  expectedPrice?: number;
  worstCasePrice?: number;
  onConfirm?: () => void;
  onCancel?: () => void;
}

export function MarketOrderSlippageDialog({
  isOpen = true,
  estimatedSlippagePct = 1.45,
  expectedPrice = 68480,
  worstCasePrice = 69473,
  onConfirm,
  onCancel,
}: MarketOrderSlippageDialogProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
      <div className="bg-[#111620] border border-red-800 rounded-xl max-w-md w-full p-6 text-white font-sans space-y-4">
        <div className="flex items-center space-x-2 text-red-400">
          <span className="text-xl font-bold">⚠️ High Slippage Warning</span>
        </div>
        <p className="text-xs text-slate-300">
          Your order size exceeds top-of-book depth. Executing at market will push the price by{' '}
          <span className="text-red-400 font-bold font-mono">+{estimatedSlippagePct}%</span>.
        </p>

        <div className="p-3 rounded bg-slate-900 border border-slate-800 font-mono text-xs space-y-1">
          <div className="flex justify-between">
            <span className="text-slate-400">Expected Mid Price:</span>
            <span>${expectedPrice.toLocaleString()}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-slate-400">Estimated Worst Fill:</span>
            <span className="text-red-400 font-bold">${worstCasePrice.toLocaleString()}</span>
          </div>
        </div>

        <div className="flex justify-end space-x-3 pt-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 rounded bg-slate-800 text-xs text-slate-300 hover:bg-slate-700 font-mono"
          >
            Cancel Order
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 rounded bg-[#FF3B56] text-xs font-bold text-white hover:bg-red-600 font-mono"
          >
            Accept Slippage & Execute
          </button>
        </div>
      </div>
    </div>
  );
}
