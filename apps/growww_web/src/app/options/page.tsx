import React from 'react';
import { OptionsChainMatrix } from '../../components/options_chain';

export const metadata = {
  title: 'Options Chain & Strategy Builder | Growww',
  description: 'Real-time European options chain matrix with live Greeks and volatility surfaces.',
};

export default function OptionsPage() {
  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Options Matrix & Greeks Terminal</h1>
        <p className="text-xs text-slate-400">Institutional Black-Scholes pricing with sub-10ms volatility surface calibration.</p>
      </div>
      <OptionsChainMatrix />
    </div>
  );
}
