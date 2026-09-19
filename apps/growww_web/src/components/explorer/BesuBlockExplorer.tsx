import React, { useState } from 'react';

export interface BesuBlock {
  blockNumber: number;
  blockHash: string;
  validatorAddress: string;
  txCount: number;
  gasUsed: number;
  timestamp: string;
}

export interface BesuTransaction {
  txHash: string;
  blockNumber: number;
  from: string;
  to: string;
  settlementType: 'ATOMIC_DVP' | 'P2P_ESCROW' | 'POR_SNAPSHOT' | 'MULTI_SIG_WITHDRAWAL';
  status: 'SUCCESS' | 'FAILED';
  timestamp: string;
}

interface BesuExplorerProps {
  latestBlocks: BesuBlock[];
  recentTransactions: BesuTransaction[];
}

export const BesuBlockExplorer: React.FC<BesuExplorerProps> = ({
  latestBlocks,
  recentTransactions,
}) => {
  const [searchQuery, setSearchQuery] = useState('');

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-5xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Hyperledger Besu QBFT Settlement Explorer</h2>
          <p className="text-gray-400 text-xs mt-1">
            Real-time public ledger of on-chain sovereign DvP transactions, validator signatures, and Merkle solvency epochs.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
          <span className="text-xs font-mono text-emerald-400 font-semibold">QBFT Consensus Live</span>
        </div>
      </div>

      {/* Search Bar */}
      <div className="mb-6">
        <input
          type="text"
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          placeholder="Search by Tx Hash (0x...), Block Number, or Account Address..."
          className="w-full bg-[#141824] border border-gray-700 rounded-lg p-3 text-xs text-white font-mono placeholder-gray-500"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Latest Blocks */}
        <div className="bg-[#141824] p-5 rounded-xl border border-gray-800">
          <h3 className="font-bold text-sm mb-4 text-gray-200">Latest Validated Blocks</h3>
          <div className="space-y-2.5">
            {latestBlocks.slice(0, 6).map(b => (
              <div key={b.blockNumber} className="p-3 bg-[#0B0E14] rounded-lg border border-gray-850 flex justify-between items-center font-mono text-xs">
                <div>
                  <span className="font-bold text-emerald-400 block">#{b.blockNumber}</span>
                  <span className="text-[10px] text-gray-500">Validator: {b.validatorAddress.slice(0, 10)}...</span>
                </div>
                <div className="text-right">
                  <span className="text-gray-300 block">{b.txCount} txns</span>
                  <span className="text-[10px] text-gray-500">{b.timestamp}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Settlement Transactions */}
        <div className="bg-[#141824] p-5 rounded-xl border border-gray-800">
          <h3 className="font-bold text-sm mb-4 text-gray-200">Recent Settlement Transactions</h3>
          <div className="space-y-2.5">
            {recentTransactions.slice(0, 6).map(tx => (
              <div key={tx.txHash} className="p-3 bg-[#0B0E14] rounded-lg border border-gray-850 font-mono text-xs flex justify-between items-center">
                <div>
                  <span className="text-blue-400 font-bold block">{tx.txHash.slice(0, 14)}...</span>
                  <span className="text-[10px] text-gray-400">{tx.settlementType}</span>
                </div>
                <div className="text-right">
                  <span className={`text-[10px] font-bold px-2 py-0.5 rounded ${
                    tx.status === 'SUCCESS' ? 'bg-emerald-950 text-emerald-400 border border-emerald-800' : 'bg-red-950 text-red-400'
                  }`}>
                    {tx.status}
                  </span>
                  <span className="text-[10px] text-gray-500 block mt-0.5">{tx.timestamp}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
