import React from 'react';
import { ReserveReconciliation } from '../../components/reserve_reconciliation';

export default function AdminReservesPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Proof of Reserve Reconciliation</h1>
        <p className="text-xs text-slate-400">Continuous 1:1 physical asset backing verification against SEBI registered custodians.</p>
      </div>
      <ReserveReconciliation />
    </div>
  );
}
