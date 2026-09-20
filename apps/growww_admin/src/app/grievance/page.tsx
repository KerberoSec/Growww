import React from 'react';
import { ScoresGrievancePortal } from '../../components/scores_grievance';

export default function GrievancePage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">SEBI SCORES 2.0 Grievance Management</h1>
        <p className="text-xs text-slate-400">Statutory investor complaint escalation, conciliation, and Online Dispute Resolution (ODR).</p>
      </div>
      <ScoresGrievancePortal />
    </div>
  );
}
