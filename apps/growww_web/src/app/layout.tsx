import React from 'react';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Growww | Institutional & Retail RWA Trading Platform',
  description: 'Trade fractional Indian Equities, Government Securities, and Digital Assets backed 1:1 by SEBI registered custody.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="bg-[#0B0E14] text-slate-100 min-h-screen font-sans antialiased selection:bg-[#00F0A0] selection:text-black">
        <div id="root-app" className="relative flex min-h-screen flex-col">
          {children}
        </div>
      </body>
    </html>
  );
}
