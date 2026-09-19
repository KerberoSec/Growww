import React from 'react';
import Link from 'next/link';

export default function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col bg-[#0B0E14]">
      {/* Top Ticker Ribbon */}
      <header className="h-14 border-b border-slate-800/80 bg-[#111620] px-4 flex items-center justify-between">
        <div className="flex items-center space-x-6">
          <Link href="/" className="flex items-center space-x-2">
            <span className="text-xl font-extrabold tracking-tight text-[#00F0A0]">GROWWW</span>
            <span className="text-xs px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 font-mono">PRO</span>
          </Link>
          <nav className="hidden md:flex items-center space-x-4 text-sm text-slate-300 font-medium">
            <Link href="/trade/btc-usdt" className="hover:text-[#00F0A0] transition-colors">Trade</Link>
            <Link href="/portfolio" className="hover:text-[#00F0A0] transition-colors">Portfolio</Link>
            <Link href="/options" className="hover:text-[#00F0A0] transition-colors">Options</Link>
            <Link href="/launchpad" className="hover:text-[#00F0A0] transition-colors">Launchpad</Link>
            <Link href="/demo" className="hover:text-[#00F0A0] transition-colors text-amber-400">Demo Simulator</Link>
          </nav>
        </div>

        <div className="flex items-center space-x-4 text-xs font-mono">
          <div className="flex items-center space-x-1.5 px-2.5 py-1 rounded bg-slate-800/60 border border-slate-700/50">
            <span className="inline-block h-2 w-2 rounded-full bg-[#00F0A0] animate-pulse"></span>
            <span className="text-slate-200">BESU QBFT: 13371</span>
          </div>
          <div className="px-2.5 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] font-semibold">
            ₹1,25,480.00
          </div>
        </div>
      </header>

      {/* Main App Content Body */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {children}
      </main>
    </div>
  );
}
