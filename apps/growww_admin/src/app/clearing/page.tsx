import React from 'react';
import { ClearingMemberPortal } from '../../components/clearing_member_portal';

export default function ClearingPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Clearing Member Capital Adequacy</h1>
        <p className="text-xs text-slate-400">Real-time monitoring of effective liquid net worth and Settlement Guarantee Fund (SGF).</p>
      </div>
      <ClearingMemberPortal />
    </div>
  );
}
