import React from 'react';
import { CopyTradingPortal } from '../../components/copy_trading_portal';

export const metadata = {
  title: 'Pro-Trader Copy Trading & Strategies | Growww',
  description: 'Replicate vetted master trader strategies with high-water mark profit sharing.',
};

export default function CopyTradingPage() {
  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Master Trader & Copy Trading Portal</h1>
        <p className="text-xs text-slate-400">Institutional copy-trading engine with real-time risk parity and follower scaling.</p>
      </div>
      <CopyTradingPortal />
    </div>
  );
}
