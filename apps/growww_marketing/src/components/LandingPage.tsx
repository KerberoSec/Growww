import React from 'react';
import { HeroSection } from './HeroSection';
import { FeeComparison } from './FeeComparison';
import { RegulatoryTrust } from './RegulatoryTrust';

export const LandingPage: React.FC = () => {
  return (
    <div className="min-h-screen bg-[#0A0A0A] text-white flex flex-col justify-between font-sans">
      <header className="border-b border-gray-900 bg-[#0A0A0A]/90 backdrop-blur sticky top-0 z-40 px-6 py-4">
        <div className="max-w-6xl mx-auto flex justify-between items-center">
          <div className="flex items-center gap-3">
            <span className="text-xl font-extrabold text-emerald-400 tracking-tight">GROWWW</span>
            <span className="text-xs bg-emerald-950 text-emerald-400 border border-emerald-800 px-2 py-0.5 rounded font-mono font-semibold">NBSE</span>
          </div>
          <nav className="flex items-center gap-6 text-sm text-gray-300">
            <a href="#features" className="hover:text-white transition">Features</a>
            <a href="#fees" className="hover:text-white transition">0.00% Fees</a>
            <a href="#compliance" className="hover:text-white transition">Compliance</a>
            <button className="bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg font-semibold text-xs transition">
              Launch App
            </button>
          </nav>
        </div>
      </header>

      <main className="flex-1">
        <HeroSection />
        <FeeComparison />
        <RegulatoryTrust />
      </main>

      <footer className="border-t border-gray-900 bg-[#070707] py-8 px-6 text-center text-xs text-gray-500">
        <div className="max-w-6xl mx-auto flex flex-col sm:flex-row justify-between items-center gap-4">
          <p>© 2026 Growww Sovereign Exchange (NBSE). All rights reserved.</p>
          <div className="flex gap-4">
            <a href="/terms" className="hover:underline">Terms of Service</a>
            <a href="/privacy" className="hover:underline">DPDP Privacy Policy</a>
            <a href="/scores" className="hover:underline">SEBI SCORES Redressal</a>
          </div>
        </div>
      </footer>
    </div>
  );
};
