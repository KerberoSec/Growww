import React from 'react';
import { RiskExceptionApproval } from '../../components/risk_approval';

export default function RiskExceptionsPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Risk Exceptions & Multi-Party Approval</h1>
        <p className="text-xs text-slate-400">Cryptographic multi-party governance for institutional margin & exposure waivers.</p>
      </div>
      <RiskExceptionApproval />
    </div>
  );
}
