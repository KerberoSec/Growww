import React from 'react';

export const FeeComparison: React.FC = () => {
  const competitors = [
    { exchange: 'Growww NBSE', makerFee: '0.00%', takerFee: '0.00%', hiddenSpread: '0 bps', settlement: 'T+0 Instant (Besu DvP)' },
    { exchange: 'Legacy Crypto Exchange A', makerFee: '0.20%', takerFee: '0.20%', hiddenSpread: '30-50 bps', settlement: 'T+2 Off-Chain' },
    { exchange: 'Foreign Offshore Platform', makerFee: '0.10%', takerFee: '0.10%', hiddenSpread: '20 bps', settlement: 'Unverifiable' },
    { exchange: 'Traditional Broker B', makerFee: '₹20/order', takerFee: '₹20/order', hiddenSpread: 'Exchange Driven', settlement: 'T+1 Clearing Corp' },
  ];

  return (
    <section className="bg-[#0D0D0D] text-white py-20 px-6 border-b border-gray-900">
      <div className="max-w-5xl mx-auto">
        <div className="text-center mb-12">
          <h2 className="text-3xl md:text-4xl font-bold tracking-tight mb-3">Transparent Zero-Fee Trading</h2>
          <p className="text-gray-400 text-sm md:text-base">Compare Growww's sovereign exchange structure with legacy Indian and offshore platforms.</p>
        </div>

        <div className="overflow-x-auto rounded-xl border border-gray-800 bg-[#141414]">
          <table className="w-full text-left text-sm">
            <thead className="bg-[#1A1A1A] text-gray-400 text-xs font-mono uppercase">
              <tr>
                <th className="p-4">Platform</th>
                <th className="p-4">Maker Fee</th>
                <th className="p-4">Taker Fee</th>
                <th className="p-4">Hidden Spread</th>
                <th className="p-4">Settlement Security</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800 font-sans">
              {competitors.map((c, i) => (
                <tr key={c.exchange} className={i === 0 ? 'bg-emerald-950/20 font-semibold' : 'hover:bg-gray-850'}>
                  <td className="p-4 flex items-center gap-2">
                    {i === 0 && <span className="w-2 h-2 rounded-full bg-emerald-400" />}
                    <span className={i === 0 ? 'text-emerald-400 font-bold' : 'text-gray-200'}>{c.exchange}</span>
                  </td>
                  <td className={`p-4 font-mono ${i === 0 ? 'text-emerald-400 font-bold' : 'text-gray-300'}`}>{c.makerFee}</td>
                  <td className={`p-4 font-mono ${i === 0 ? 'text-emerald-400 font-bold' : 'text-gray-300'}`}>{c.takerFee}</td>
                  <td className={`p-4 font-mono ${i === 0 ? 'text-emerald-400 font-bold' : 'text-gray-400'}`}>{c.hiddenSpread}</td>
                  <td className="p-4 text-xs font-mono text-gray-300">{c.settlement}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
};
