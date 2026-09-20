import React from 'react';
import { SupervisoryDashboard } from '../../components/supervisory_dashboard';

export default function RegulatorPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Regulatory Supervisory Inspection Terminal</h1>
        <p className="text-xs text-slate-400">Direct supervisory inspection access for SEBI, IFSCA, and Exchange Surveillance Teams.</p>
      </div>
      <SupervisoryDashboard />
    </div>
  );
}
