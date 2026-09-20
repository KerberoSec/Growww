import React from 'react';
import Link from 'next/link';

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen grid lg:grid-cols-2 bg-[#0B0E14]">
      {/* Brand & Regulatory Assurance Left Column */}
      <div className="hidden lg:flex flex-col justify-between p-12 bg-gradient-to-br from-[#111620] via-[#0B0E14] to-[#0A192F] border-r border-slate-800">
        <div>
          <Link href="/" className="text-2xl font-black text-[#00F0A0] tracking-wider">
            GROWWW
          </Link>
          <div className="mt-16 space-y-4">
            <h1 className="text-3xl font-extrabold text-white leading-tight">
              Institutional-Grade Asset Tokenization & Trading
            </h1>
            <p className="text-slate-400 text-sm max-w-md leading-relaxed">
              100% physically backed securities held with SEBI-registered custodians (NSDL/CDSL). Real-time cryptographic proof of reserves verified on Hyperledger Besu.
            </p>
          </div>
        </div>

        <div className="space-y-4 pt-12 border-t border-slate-800/80 text-xs text-slate-500 font-mono">
          <div>STATUTORY REGULATORY DISCLOSURE</div>
          <div>Securities and Exchange Board of India (SEBI) Registered Broker INZ000000000</div>
          <div>Zero on-chain PII | FIPS 140-2 Level 3 HSM Relayed Settlement</div>
        </div>
      </div>

      {/* Auth / Onboarding Right Column */}
      <div className="flex items-center justify-center p-6 lg:p-12">
        <div className="w-full max-w-md space-y-6">
          {children}
        </div>
      </div>
    </div>
  );
}
