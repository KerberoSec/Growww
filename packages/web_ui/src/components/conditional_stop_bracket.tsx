import React, { useState } from 'react';

export function ConditionalStopBracket() {
  const [takeProfit, setTakeProfit] = useState('72000.00');
  const [stopLoss, setStopLoss] = useState('66500.00');

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">OCO Bracket Entry Panel</span>
        <span className="text-[10px] text-slate-400">One-Cancels-the-Other</span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="text-slate-400 block mb-1">TAKE PROFIT LIMIT</label>
          <input
            type="text"
            value={takeProfit}
            onChange={(e) => setTakeProfit(e.target.value)}
            className="w-full bg-slate-900 border border-slate-700 rounded p-2 text-white"
          />
        </div>
        <div>
          <label className="text-slate-400 block mb-1">STOP LOSS TRIGGER</label>
          <input
            type="text"
            value={stopLoss}
            onChange={(e) => setStopLoss(e.target.value)}
            className="w-full bg-slate-900 border border-slate-700 rounded p-2 text-white"
          />
        </div>
      </div>
    </div>
  );
}
