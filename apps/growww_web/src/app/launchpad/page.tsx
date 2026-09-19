import React from 'react';
import { RwaLaunchpad } from '../../components/rwa_launchpad';

export const metadata = {
  title: 'RWA Primary Launchpad & Auctions | Growww',
  description: 'Participate in primary sovereign debt and tokenized real-world asset Dutch auctions.',
};

export default function LaunchpadPage() {
  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">RWA Primary Launchpad</h1>
        <p className="text-xs text-slate-400">Institutional sovereign bond and private credit syndication terminal.</p>
      </div>
      <RwaLaunchpad />
    </div>
  );
}
