import React from 'react';
import { DmaWorkstation } from '../../components/dma_workstation';

export const metadata = {
  title: 'Institutional Direct Market Access (DMA) | Growww',
  description: 'Sub-millisecond FIX protocol trading terminal with low-latency hotkey execution.',
};

export default function DmaPage() {
  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Direct Market Access Workstation</h1>
        <p className="text-xs text-slate-400">Institutional low-latency trading terminal with hardware-accelerated order paths.</p>
      </div>
      <DmaWorkstation />
    </div>
  );
}
