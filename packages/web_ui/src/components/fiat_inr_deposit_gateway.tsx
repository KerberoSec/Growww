import React, { useState } from 'react';

export function FiatInrDepositGateway() {
  const [amount, setAmount] = useState('25000');
  const [rail, setRail] = useState<'UPI_QR' | 'IMPS'>('UPI_QR');

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold">Instant Fiat INR Funding</span>
        <span className="text-[#00F0A0]">0.00% Zero-Fee Deposit</span>
      </div>
      <div>
        <label className="text-slate-400 block mb-1">DEPOSIT AMOUNT (INR)</label>
        <input
          type="text"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="w-full bg-slate-900 border border-slate-700 rounded p-2 text-white"
        />
      </div>
      <div className="flex space-x-2">
        <button
          onClick={() => setRail('UPI_QR')}
          className={`flex-1 py-1.5 rounded ${rail === 'UPI_QR' ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-300'}`}
        >
          NPCI UPI 2.0 Dynamic QR
        </button>
        <button
          onClick={() => setRail('IMPS')}
          className={`flex-1 py-1.5 rounded ${rail === 'IMPS' ? 'bg-[#00F0A0] text-black font-bold' : 'bg-slate-800 text-slate-300'}`}
        >
          Virtual Account IMPS/NEFT
        </button>
      </div>
    </div>
  );
}
