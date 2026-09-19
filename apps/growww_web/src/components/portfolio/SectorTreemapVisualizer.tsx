import React from 'react';

export interface AssetAllocation {
  symbol: string;
  name: string;
  assetClass: 'EQUITY_DEMAT' | 'CRYPTO_VDA' | 'GOV_BOND' | 'CBDC_CASH';
  sector: string;
  valueINR: number;
  weightPct: number;
  pnl24hPct: number;
}

interface SectorTreemapProps {
  allocations: AssetAllocation[];
}

export const SectorTreemapVisualizer: React.FC<SectorTreemapProps> = ({ allocations }) => {
  const totalValueINR = allocations.reduce((sum, a) => sum + a.valueINR, 0);

  const getSectorColor = (sector: string) => {
    switch (sector.toLowerCase()) {
      case 'crypto vda': return '#10B981';
      case 'banking & finance': return '#3B82F6';
      case 'it & software': return '#8B5CF6';
      case 'energy & oil': return '#F59E0B';
      case 'sovereign bonds': return '#EC4899';
      default: return '#6B7280';
    }
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 font-sans">
      <div className="flex justify-between items-center mb-4">
        <div>
          <h3 className="text-lg font-bold">Portfolio Sector Breakdown & Asset Treemap</h3>
          <p className="text-gray-400 text-xs mt-0.5">Total Portfolio Value: ₹{totalValueINR.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
        {allocations.map(asset => {
          const color = getSectorColor(asset.sector);
          return (
            <div
              key={asset.symbol}
              className="p-4 rounded-xl border border-gray-800 transition transform hover:-translate-y-0.5"
              style={{ backgroundColor: `${color}15`, borderColor: `${color}40` }}
            >
              <div className="flex justify-between items-start mb-2">
                <span className="font-mono font-bold text-sm" style={{ color }}>{asset.symbol}</span>
                <span className={`text-[10px] font-mono font-semibold px-1.5 py-0.5 rounded ${
                  asset.pnl24hPct >= 0 ? 'bg-emerald-950 text-emerald-400' : 'bg-red-950 text-red-400'
                }`}>
                  {asset.pnl24hPct >= 0 ? '+' : ''}{asset.pnl24hPct.toFixed(2)}%
                </span>
              </div>
              <div className="text-xs text-gray-300 font-medium">{asset.name}</div>
              <div className="text-[11px] text-gray-500 font-mono mt-1">{asset.sector}</div>
              <div className="mt-3 pt-2 border-t border-gray-800 flex justify-between items-center text-xs font-mono">
                <span className="font-bold text-white">₹{asset.valueINR.toLocaleString('en-IN')}</span>
                <span className="text-gray-400 font-medium">{asset.weightPct.toFixed(1)}%</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
