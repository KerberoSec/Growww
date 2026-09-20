import React, { useState } from 'react';

export function DemoTradingCockpit() {
  const [balance, setBalance] = useState(10000);

  return (
    <div className="p-4 rounded border border-amber-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center">
        <span className="font-bold text-amber-400">Demo Paper Trading Cockpit</span>
        <span className="text-slate-400">Live Binance/Coinbase Mirror</span>
      </div>
      <div className="flex justify-between items-center">
        <span>Virtual Balance:</span>
        <span className="text-[#00F0A0] font-bold">${balance.toLocaleString()} vUSDT</span>
      </div>
    </div>
  );
}
