import React from 'react';
import { ProOrderbook } from '../../../components/orderbook/pro_orderbook';

export const metadata = {
  title: 'BTC/USDT Spot Pro Terminal | Growww',
  description: 'Ultra low latency Level-2 order book with real-time cumulative depth and tick streams.',
};

export default function BtcUsdtTradePage() {
  return (
    <div className="max-w-7xl mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-black text-white">BTC/USDT Pro-Trading Terminal</h1>
          <p className="text-xs text-slate-400">Microsecond in-memory matching with on-chain settlement proof.</p>
        </div>
      </div>
      <div className="flex justify-center">
        <ProOrderbook />
      </div>
    </div>
  );
}
