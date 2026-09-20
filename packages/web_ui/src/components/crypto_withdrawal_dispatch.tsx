import React, { useState } from 'react';

export function CryptoWithdrawalDispatch() {
  const [amount, setAmount] = useState('0.10');
  const [estimatedGasUsd] = useState(1.42);

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Crypto Withdrawal Dispatch</span>
        <span className="text-slate-400">Dynamic EIP-1559 Estimator</span>
      </div>
      <div>
        <label className="text-slate-400 block mb-1">WITHDRAWAL AMOUNT</label>
        <input
          type="text"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="w-full bg-slate-900 border border-slate-700 rounded p-2 text-white"
        />
      </div>
      <div className="p-2 rounded bg-slate-900 flex justify-between text-slate-400">
        <span>Estimated Network Fee:</span>
        <span className="text-[#00F0A0] font-bold">${estimatedGasUsd} USD</span>
      </div>
    </div>
  );
}
