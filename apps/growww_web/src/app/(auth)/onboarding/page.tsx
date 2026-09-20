import React from 'react';
import { KycFlow } from '../../../components/onboarding/kyc_flow';

export const metadata = {
  title: 'Investor Onboarding & DigiLocker KYC | Growww',
  description: 'Complete paperless digital KYC verification for real-world asset trading.',
};

export default function OnboardingPage() {
  return (
    <div className="w-full">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-black tracking-tight text-white">Investor Verification</h1>
        <p className="text-xs text-slate-400 mt-1">SEBI Compliant Tier-2 Electronic KYC</p>
      </div>
      <KycFlow />
    </div>
  );
}
