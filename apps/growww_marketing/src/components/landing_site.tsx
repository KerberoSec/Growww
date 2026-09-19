import React from 'react';

export function MarketingLandingSite() {
  return (
    <div className="min-h-screen bg-[#0B0E14] text-white selection:bg-[#00F0A0] selection:text-black">
      {/* Navigation Header */}
      <nav className="border-b border-slate-800 bg-[#0B0E14]/80 backdrop-blur sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <span className="text-2xl font-black text-[#00F0A0] tracking-wider">GROWWW</span>
            <span className="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-300 font-mono">NBSE</span>
          </div>
          <div className="hidden md:flex items-center space-x-8 text-sm font-medium text-slate-300">
            <a href="#features" className="hover:text-[#00F0A0] transition-colors">Tokenized Equities</a>
            <a href="#reserves" className="hover:text-[#00F0A0] transition-colors">Proof of Reserves</a>
            <a href="#institutions" className="hover:text-[#00F0A0] transition-colors">GIFT City IFSC</a>
            <a href="#disclosures" className="hover:text-[#00F0A0] transition-colors">Statutory Disclosures</a>
          </div>
          <div>
            <a
              href="http://localhost:3000/trade/btc-usdt"
              className="px-4 py-2 rounded bg-[#00F0A0] text-black text-xs font-bold hover:bg-[#00d08a] transition-all"
            >
              Open Trading WebApp →
            </a>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="py-24 px-6 max-w-7xl mx-auto text-center relative overflow-hidden">
        <div className="inline-block px-3 py-1 rounded-full border border-[#00F0A0]/30 bg-[#00F0A0]/10 text-xs font-mono text-[#00F0A0] mb-6">
          SEBI-Registered Custody (NSDL/CDSL) • Hyperledger Besu QBFT Settlement
        </div>
        <h1 className="text-5xl md:text-7xl font-black tracking-tight max-w-4xl mx-auto leading-tight">
          Trade Fractional Equities & Digital Assets With{' '}
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#00F0A0] to-[#00B4D8]">
            Zero Friction
          </span>
        </h1>
        <p className="mt-6 text-lg text-slate-400 max-w-2xl mx-auto">
          Every fractional share is backed 1:1 by real securities held in depository accounts. Continuous automated cryptographic attestation on-chain.
        </p>
        <div className="mt-8 flex justify-center space-x-4">
          <a
            href="http://localhost:3000/onboarding"
            className="px-6 py-3 rounded bg-[#00F0A0] text-black font-bold text-sm hover:bg-[#00d08a] shadow-lg shadow-[#00F0A0]/20"
          >
            Start Instant DigiLocker KYC
          </a>
          <a
            href="http://localhost:3000/demo"
            className="px-6 py-3 rounded border border-slate-700 bg-slate-900 text-white font-bold text-sm hover:bg-slate-800"
          >
            Explore Demo Cockpit
          </a>
        </div>
      </section>

      {/* Trust & Transparency Grid */}
      <section id="features" className="py-16 px-6 max-w-7xl mx-auto border-t border-slate-800">
        <div className="grid md:grid-cols-3 gap-8">
          <div className="p-6 rounded-xl border border-slate-800 bg-[#111620]">
            <div className="h-10 w-10 rounded bg-[#00F0A0]/10 text-[#00F0A0] flex items-center justify-center font-mono font-bold mb-4">
              01
            </div>
            <h3 className="text-lg font-bold">1:1 Physical Depository Backing</h3>
            <p className="mt-2 text-sm text-slate-400">
              Fractional equity tokens directly reflect actual NSDL/CDSL custodian holdings audited 24/7.
            </p>
          </div>

          <div className="p-6 rounded-xl border border-slate-800 bg-[#111620]">
            <div className="h-10 w-10 rounded bg-[#00F0A0]/10 text-[#00F0A0] flex items-center justify-center font-mono font-bold mb-4">
              02
            </div>
            <h3 className="text-lg font-bold">Deterministic 2s Settlement</h3>
            <p className="mt-2 text-sm text-slate-400">
              Hyperledger Besu consortium network guarantees instant DvP finality without counterparty settlement risk.
            </p>
          </div>

          <div className="p-6 rounded-xl border border-slate-800 bg-[#111620]">
            <div className="h-10 w-10 rounded bg-[#00F0A0]/10 text-[#00F0A0] flex items-center justify-center font-mono font-bold mb-4">
              03
            </div>
            <h3 className="text-lg font-bold">Statutory Privacy (Zero PII)</h3>
            <p className="mt-2 text-sm text-slate-400">
              Compliant with DPDP Act 2023. Zero personally identifiable information is stored on-chain.
            </p>
          </div>
        </div>
      </section>

      {/* Statutory Footer Disclosures */}
      <footer id="disclosures" className="border-t border-slate-800 py-12 px-6 bg-[#080B10] text-xs text-slate-500 font-mono">
        <div className="max-w-7xl mx-auto space-y-4">
          <div className="font-bold text-slate-400 uppercase">Statutory Regulatory & Risk Disclosures</div>
          <p>
            Securities and Exchange Board of India (SEBI) Registration No: INZ000000000. Member of NSE and BSE.
            Investments in securities markets are subject to market risks; read all related documents carefully before investing.
          </p>
          <p>
            Cryptographic tokens and digital assets are orchestrated under sandbox guidelines. Proof-of-reserve attestation is performed via automated Merkle root comparison against depository balances.
          </p>
          <div className="pt-4 border-t border-slate-800 flex justify-between">
            <span>© 2026 Growww NBSE Technologies Pvt Ltd. All rights reserved.</span>
            <span>Version 1.0.0-PRO (Ritik Branch)</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
