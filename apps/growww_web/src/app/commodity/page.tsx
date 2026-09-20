import React from 'react';
import { CommodityVaultPortal } from '../../components/commodity_vault';

export const metadata = {
  title: 'Commodity Physical Delivery & Vaults | Growww',
  description: 'Physical vaulted bullion redemption and institutional vault receipt management.',
};

export default function CommodityPage() {
  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Commodity Physical Delivery Portal</h1>
        <p className="text-xs text-slate-400">Institutional vaulted bullion with electronic warehouse receipts (EWR).</p>
      </div>
      <CommodityVaultPortal />
    </div>
  );
}
