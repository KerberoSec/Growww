import React from 'react';
import { P2PArbitrationConsole } from '../../components/p2p_arbitration';

export default function P2PDisputesPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">P2P Dispute Arbitration & Compliance Desk</h1>
        <p className="text-xs text-slate-400">Escrow dispute resolution with banking UTR reconciliation and chat transcript auditing.</p>
      </div>
      <P2PArbitrationConsole />
    </div>
  );
}
