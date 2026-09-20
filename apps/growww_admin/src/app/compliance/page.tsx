import React from 'react';
import { RegulatoryReportingPortal } from '../../components/regulatory_reporting';

export default function AdminCompliancePage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Regulatory Reporting & Audit Exports</h1>
        <p className="text-xs text-slate-400">Automated statutory filings for SEBI, RBI, FIU-IND, and IFSCA.</p>
      </div>
      <RegulatoryReportingPortal />
    </div>
  );
}
