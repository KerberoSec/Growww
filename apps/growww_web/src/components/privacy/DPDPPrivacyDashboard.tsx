import React, { useState } from 'react';

export interface ConsentItem {
  id: string;
  category: string;
  description: string;
  statutoryRequired: boolean;
  granted: boolean;
  grantedAt?: string;
}

export const DPDPPrivacyDashboard: React.FC = () => {
  const [consents, setConsents] = useState<ConsentItem[]>([
    { id: '1', category: 'SEBI / PMLA KYC Records', description: 'Mandatory PAN, Aadhaar, and CKYC data retained for 5 years per statutory obligation.', statutoryRequired: true, granted: true, grantedAt: '2026-01-10' },
    { id: '2', category: 'FIU-IND Reporting Data', description: 'Trade transaction and wallet address telemetry for statutory anti-money laundering monitoring.', statutoryRequired: true, granted: true, grantedAt: '2026-01-10' },
    { id: '3', category: 'Personalized Market Intelligence', description: 'AI recommendation and conversational assistant portfolio analysis.', statutoryRequired: false, granted: true, grantedAt: '2026-03-15' },
    { id: '4', category: 'Marketing & Promotional Notifications', description: 'Email, SMS, and WhatsApp trade alerts, product launches, and fee reports.', statutoryRequired: false, granted: false },
  ]);

  const [erasureRequested, setErasureRequested] = useState(false);

  const toggleConsent = (id: string) => {
    setConsents(consents.map(c => {
      if (c.id === id && !c.statutoryRequired) {
        return { ...c, granted: !c.granted };
      }
      return c;
    }));
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-3xl mx-auto font-sans">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-lg bg-blue-950 flex items-center justify-center text-blue-400 font-bold border border-blue-800">
          📜
        </div>
        <div>
          <h2 className="text-xl font-bold">Digital Personal Data Protection (DPDP) Act 2023</h2>
          <p className="text-gray-400 text-xs">Manage your personal data consents, view processing logs, or exercise your Right to Erasure.</p>
        </div>
      </div>

      <div className="space-y-3 mb-6">
        {consents.map(c => (
          <div
            key={c.id}
            className="p-4 bg-[#141824] rounded-lg border border-gray-800 flex justify-between items-start"
          >
            <div className="max-w-xl">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-sm text-gray-200">{c.category}</span>
                {c.statutoryRequired && (
                  <span className="text-[10px] font-mono font-bold bg-amber-950/80 text-amber-400 border border-amber-800 px-1.5 py-0.5 rounded">
                    Statutory Mandatory (PMLA/SEBI)
                  </span>
                )}
              </div>
              <p className="text-xs text-gray-400 mt-1 leading-relaxed">{c.description}</p>
              {c.grantedAt && <span className="text-[10px] text-gray-500 font-mono block mt-1">Consent Logged: {c.grantedAt}</span>}
            </div>
            <button
              onClick={() => toggleConsent(c.id)}
              disabled={c.statutoryRequired}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
                c.granted
                  ? 'bg-emerald-900/40 text-emerald-400 border border-emerald-800'
                  : 'bg-gray-800 text-gray-400 border border-gray-700'
              } ${c.statutoryRequired ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-80'}`}
            >
              {c.granted ? 'Active' : 'Revoked'}
            </button>
          </div>
        ))}
      </div>

      {/* Right to Erasure Card */}
      <div className="bg-red-950/20 border border-red-900/40 p-4 rounded-lg mb-6">
        <h3 className="text-sm font-bold text-red-400 mb-1">Right to Erasure / Account Crypto-Shredding</h3>
        <p className="text-xs text-gray-300 mb-3 leading-relaxed">
          Request crypto-shredding and irreversible erasure of non-statutory personal identifiers. Statutory transaction ledgers will be pseudonymized and archived in compliance with PMLA Section 12.
        </p>
        <button
          onClick={() => setErasureRequested(true)}
          className="px-4 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg text-xs font-bold transition"
        >
          {erasureRequested ? '✓ Erasure Docket Registered' : 'Initiate Data Erasure Workflow'}
        </button>
      </div>
    </div>
  );
};
