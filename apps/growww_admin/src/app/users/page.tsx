import React from 'react';
import { KycReviewDashboard } from '../../components/kyc_review';

export default function AdminUsersPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-black text-white">User Management & KYC Review</h1>
          <p className="text-xs text-slate-400">Maker-checker approval queue under SEBI PMLA guidelines.</p>
        </div>
      </div>
      <KycReviewDashboard />
    </div>
  );
}
