import React from 'react';

export const RegulatoryTrust: React.FC = () => {
  return (
    <section className="bg-[#0A0A0A] text-white py-20 px-6">
      <div className="max-w-5xl mx-auto text-center">
        <h2 className="text-3xl md:text-4xl font-bold tracking-tight mb-4">Sovereign Compliance by Design</h2>
        <p className="text-gray-400 max-w-2xl mx-auto mb-12 text-sm md:text-base">
          Built strictly in accordance with Indian regulatory statutes, financial intelligence mandates, and real-time tax withholding frameworks.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="p-6 bg-[#141414] rounded-xl border border-gray-800 text-left">
            <div className="w-10 h-10 rounded-lg bg-emerald-950 flex items-center justify-center text-emerald-400 font-bold mb-4 border border-emerald-800">
              §
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Section 194S TDS</h3>
            <p className="text-gray-400 text-xs leading-relaxed">
              Automated 1% TDS deduction at transaction execution. Instant generation of Form 16A and Challan ITNS 281 audit trails directly into your tax dashboard.
            </p>
          </div>

          <div className="p-6 bg-[#141414] rounded-xl border border-gray-800 text-left">
            <div className="w-10 h-10 rounded-lg bg-emerald-950 flex items-center justify-center text-emerald-400 font-bold mb-4 border border-emerald-800">
              ✓
            </div>
            <h3 className="text-lg font-bold text-white mb-2">FIU-IND Reporting</h3>
            <p className="text-gray-400 text-xs leading-relaxed">
              Integrated PMLA Compliance with automated CTR (Cash Transaction Report) and STR (Suspicious Transaction Report) XML dispatch to Financial Intelligence Unit India.
            </p>
          </div>

          <div className="p-6 bg-[#141414] rounded-xl border border-gray-800 text-left">
            <div className="w-10 h-10 rounded-lg bg-emerald-950 flex items-center justify-center text-emerald-400 font-bold mb-4 border border-emerald-800">
              ⚖
            </div>
            <h3 className="text-lg font-bold text-white mb-2">SEBI SCORES 2.0 SLA</h3>
            <p className="text-gray-400 text-xs leading-relaxed">
              Investor grievance redressal with strict 21-day statutory SLA tracking, automated Action Taken Reports (ATR), and Smart ODR conciliation gateway.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
};
