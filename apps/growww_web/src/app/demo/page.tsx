import React from 'react';
import { DemoTradingAnalytics } from '../../components/demo_analytics';

export const metadata = {
  title: 'Demo Paper Trading Simulator & Cockpit | Growww',
  description: 'Practice trading equities and digital assets with ₹10,00,000 risk-free virtual capital.',
};

export default function DemoPage() {
  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Paper Trading Simulation & Analytics</h1>
        <p className="text-xs text-slate-400">Mirroring live matching engine execution without financial exposure.</p>
      </div>
      <DemoTradingAnalytics />
    </div>
  );
}
