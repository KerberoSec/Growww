import React from 'react';

export const HeroSection: React.FC = () => {
  return (
    <section className="relative bg-[#0A0A0A] text-white py-24 px-6 overflow-hidden border-b border-gray-900">
      <div className="max-w-6xl mx-auto text-center relative z-10">
        <div className="inline-flex items-center gap-2 bg-emerald-950/60 border border-emerald-800 text-emerald-400 px-4 py-1.5 rounded-full text-xs font-semibold uppercase tracking-wider mb-6">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          India's First 0.00% Fee Sovereign VDA & Multi-Asset Exchange
        </div>

        <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-6 bg-clip-text text-transparent bg-gradient-to-r from-white via-gray-200 to-emerald-400">
          Trade Sovereign. Trade Zero-Fee.
        </h1>

        <p className="text-lg md:text-xl text-gray-400 max-w-3xl mx-auto mb-10 leading-relaxed font-sans">
          Institutional-grade crypto VDA, Demat equities, and sovereign settlement on Hyperledger Besu QBFT. Fully compliant with Section 194S TDS, Section 115BBH, and FIU-IND AML/CFT regulations.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <button className="w-full sm:w-auto px-8 py-4 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-xl shadow-lg shadow-emerald-900/30 transition transform hover:-translate-y-0.5">
            Claim ₹10,000 Demo USDT
          </button>
          <button className="w-full sm:w-auto px-8 py-4 bg-[#161616] hover:bg-gray-800 text-gray-200 font-semibold rounded-xl border border-gray-800 transition">
            Explore Live L2 Orderbook
          </button>
        </div>

        {/* Feature Badges */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-16 pt-12 border-t border-gray-900">
          <div className="p-4 bg-[#121212] rounded-lg border border-gray-850">
            <div className="text-2xl font-bold text-emerald-400 font-mono">0.00%</div>
            <div className="text-xs text-gray-400 mt-1">Platform Maker & Taker Fee</div>
          </div>
          <div className="p-4 bg-[#121212] rounded-lg border border-gray-850">
            <div className="text-2xl font-bold text-white font-mono">&lt; 15 µs</div>
            <div className="text-xs text-gray-400 mt-1">Ultra Low-Latency Matching</div>
          </div>
          <div className="p-4 bg-[#121212] rounded-lg border border-gray-850">
            <div className="text-2xl font-bold text-emerald-400 font-mono">100%</div>
            <div className="text-xs text-gray-400 mt-1">Proof of Solvency (Merkle Tree)</div>
          </div>
          <div className="p-4 bg-[#121212] rounded-lg border border-gray-850">
            <div className="text-2xl font-bold text-white font-mono">FIU-IND</div>
            <div className="text-xs text-gray-400 mt-1">Registered & PMLA Compliant</div>
          </div>
        </div>
      </div>
    </section>
  );
};
