import React, { useState } from 'react';

export type USDTNetwork = 'ETHEREUM_ERC20' | 'TRON_TRC20' | 'BNB_SMART_CHAIN' | 'POLYGON_POS' | 'SOLANA_SPL' | 'HYPERLEDGER_BESU';

interface NetworkConfig {
  network: USDTNetwork;
  name: string;
  depositAddress: string;
  confirmations: number;
  estimatedTime: string;
  depositFee: string;
}

export const MultichainDepositPortal: React.FC = () => {
  const [selectedNetwork, setSelectedNetwork] = useState<USDTNetwork>('HYPERLEDGER_BESU');

  const networks: Record<USDTNetwork, NetworkConfig> = {
    HYPERLEDGER_BESU: {
      network: 'HYPERLEDGER_BESU',
      name: 'Hyperledger Besu (Zero-Fee Sovereign Rail)',
      depositAddress: '0x71C8366420A01f98C339F0Fa9a888E45C108dC49',
      confirmations: 1,
      estimatedTime: '< 1 second (Instant DvP)',
      depositFee: '0.00 USDT (Zero-Fee)',
    },
    ETHEREUM_ERC20: {
      network: 'ETHEREUM_ERC20',
      name: 'Ethereum (ERC-20)',
      depositAddress: '0x8849bCd03029487F63d59e38eFa4f39E0A9B535a',
      confirmations: 12,
      estimatedTime: '3-5 minutes',
      depositFee: '0.00 USDT',
    },
    TRON_TRC20: {
      network: 'TRON_TRC20',
      name: 'Tron (TRC-20)',
      depositAddress: 'TYDzsYUE28g2e5i4oB7zLwNqV9xZ87C1fP',
      confirmations: 20,
      estimatedTime: '1-2 minutes',
      depositFee: '0.00 USDT',
    },
    SOLANA_SPL: {
      network: 'SOLANA_SPL',
      name: 'Solana (SPL Token)',
      depositAddress: '7XwN5mE8tGZ4e1qV9yB7dL3pC8sF2kH9uA6vB5nM8rP',
      confirmations: 32,
      estimatedTime: '< 10 seconds',
      depositFee: '0.00 USDT',
    },
    POLYGON_POS: {
      network: 'POLYGON_POS',
      name: 'Polygon PoS',
      depositAddress: '0x9924cBa04038488E84e68eFa5f48F29E0A9B829c',
      confirmations: 64,
      estimatedTime: '2 minutes',
      depositFee: '0.00 USDT',
    },
    BNB_SMART_CHAIN: {
      network: 'BNB_SMART_CHAIN',
      name: 'BNB Smart Chain (BEP-20)',
      depositAddress: '0x5511dBc02019488E74e68eFa4f48F29E0A9B618b',
      confirmations: 15,
      estimatedTime: '1 minute',
      depositFee: '0.00 USDT',
    },
  };

  const current = networks[selectedNetwork];

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-xl mx-auto">
      <h2 className="text-xl font-bold mb-1">Tether (USDT) Multichain Deposit Portal</h2>
      <p className="text-gray-400 text-xs mb-6">Select your preferred blockchain rail for automated on-chain verification and zero-fee minting.</p>

      {/* Network Dropdown */}
      <div className="mb-4">
        <label className="block text-xs font-semibold text-gray-300 mb-1">Select Transfer Network</label>
        <select
          value={selectedNetwork}
          onChange={e => setSelectedNetwork(e.target.value as USDTNetwork)}
          className="w-full bg-[#141824] border border-gray-700 rounded-lg p-3 text-sm text-white font-mono"
        >
          {Object.values(networks).map(n => (
            <option key={n.network} value={n.network}>
              {n.name}
            </option>
          ))}
        </select>
      </div>

      {/* Address Container */}
      <div className="bg-[#141824] p-4 rounded-lg border border-gray-800 mb-4">
        <span className="text-xs text-gray-400 block mb-1">Deposit Address ({selectedNetwork})</span>
        <div className="font-mono text-xs text-emerald-400 break-all bg-[#0B0E14] p-2.5 rounded border border-gray-850">
          {current.depositAddress}
        </div>
      </div>

      {/* Network Specifications */}
      <div className="grid grid-cols-3 gap-3 text-xs font-mono bg-[#141824] p-3 rounded-lg border border-gray-850">
        <div>
          <span className="text-gray-500 block text-[10px]">Estimated Time</span>
          <span className="text-white font-semibold">{current.estimatedTime}</span>
        </div>
        <div>
          <span className="text-gray-500 block text-[10px]">Confirmations</span>
          <span className="text-white font-semibold">{current.confirmations} Blocks</span>
        </div>
        <div>
          <span className="text-gray-500 block text-[10px]">Exchange Fee</span>
          <span className="text-emerald-400 font-bold">{current.depositFee}</span>
        </div>
      </div>
    </div>
  );
};
