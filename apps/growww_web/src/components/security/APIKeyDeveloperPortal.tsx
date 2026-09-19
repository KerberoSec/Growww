import React, { useState } from 'react';

export interface APIKeyRecord {
  keyId: string;
  label: string;
  apiKey: string;
  secretMasked: string;
  scopes: ('READ' | 'TRADE' | 'WITHDRAW')[];
  ipWhitelistCIDR: string[];
  rateLimitTier: 'RETAIL_100_REQ_S' | 'PRO_500_REQ_S' | 'HFT_UNLIMITED';
  createdAt: string;
  lastUsedAt?: string;
  status: 'ACTIVE' | 'REVOKED';
}

export const APIKeyDeveloperPortal: React.FC = () => {
  const [keys, setKeys] = useState<APIKeyRecord[]>([
    {
      keyId: 'KEY_01',
      label: 'Hummingbot Market Maker Bot',
      apiKey: 'grw_live_9f8a7c6e5d4c3b2a',
      secretMasked: '••••••••••••••••••••••••••••••••',
      scopes: ['READ', 'TRADE'],
      ipWhitelistCIDR: ['103.21.244.0/24', '14.139.120.5'],
      rateLimitTier: 'PRO_500_REQ_S',
      createdAt: '2026-08-10',
      lastUsedAt: 'Just now',
      status: 'ACTIVE',
    },
    {
      keyId: 'KEY_02',
      label: 'Read-Only Tax Accounting Sync',
      apiKey: 'grw_live_1a2b3c4d5e6f7a8b',
      secretMasked: '••••••••••••••••••••••••••••••••',
      scopes: ['READ'],
      ipWhitelistCIDR: [],
      rateLimitTier: 'RETAIL_100_REQ_S',
      createdAt: '2026-09-01',
      lastUsedAt: '2 hours ago',
      status: 'ACTIVE',
    },
  ]);

  const [newLabel, setNewLabel] = useState('');
  const [newCIDR, setNewCIDR] = useState('');
  const [selectedScopes, setSelectedScopes] = useState<('READ' | 'TRADE' | 'WITHDRAW')[]>(['READ', 'TRADE']);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    const newKey: APIKeyRecord = {
      keyId: `KEY_${Date.now()}`,
      label: newLabel || 'Trading API Key',
      apiKey: `grw_live_${Math.random().toString(36).substring(2, 18)}`,
      secretMasked: '••••••••••••••••••••••••••••••••',
      scopes: selectedScopes,
      ipWhitelistCIDR: newCIDR ? newCIDR.split(',').map(s => s.trim()) : [],
      rateLimitTier: 'PRO_500_REQ_S',
      createdAt: new Date().toISOString().split('T')[0],
      status: 'ACTIVE',
    };
    setKeys([newKey, ...keys]);
    setShowCreateModal(false);
    setNewLabel('');
    setNewCIDR('');
  };

  const revokeKey = (keyId: string) => {
    setKeys(keys.map(k => k.keyId === keyId ? { ...k, status: 'REVOKED' } : k));
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-5xl mx-auto font-sans">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">API Key Developer Portal</h2>
          <p className="text-gray-400 text-xs mt-1">
            Programmatic trading keys with strict IP CIDR whitelisting, permission scoping, and sub-millisecond REST/WebSocket access.
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition"
        >
          + Create New API Key
        </button>
      </div>

      <div className="space-y-4">
        {keys.map(k => (
          <div
            key={k.keyId}
            className={`p-5 rounded-xl border transition ${
              k.status === 'ACTIVE' ? 'bg-[#141824] border-gray-800' : 'bg-[#0E1017] border-gray-850 opacity-60'
            }`}
          >
            <div className="flex justify-between items-start mb-3">
              <div>
                <h3 className="font-bold text-sm text-gray-200">{k.label}</h3>
                <div className="font-mono text-xs text-emerald-400 mt-0.5">{k.apiKey}</div>
              </div>
              <div className="flex items-center gap-2">
                <span className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold ${
                  k.status === 'ACTIVE' ? 'bg-emerald-950 text-emerald-400 border border-emerald-800' : 'bg-red-950 text-red-400'
                }`}>
                  {k.status}
                </span>
                {k.status === 'ACTIVE' && (
                  <button
                    onClick={() => revokeKey(k.keyId)}
                    className="px-3 py-1 bg-red-950/60 hover:bg-red-900 text-red-400 border border-red-800 rounded text-xs"
                  >
                    Revoke
                  </button>
                )}
              </div>
            </div>

            <div className="grid grid-cols-3 gap-3 text-xs font-mono bg-[#0B0E14] p-3 rounded-lg border border-gray-850 mb-3">
              <div>
                <span className="text-[10px] text-gray-500 block">Permission Scopes</span>
                <div className="flex gap-1.5 mt-1">
                  {k.scopes.map(s => (
                    <span key={s} className="px-1.5 py-0.5 bg-gray-800 text-gray-300 rounded text-[10px]">
                      {s}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <span className="text-[10px] text-gray-500 block">IP CIDR Whitelist</span>
                <span className="text-gray-300 text-[11px]">
                  {k.ipWhitelistCIDR.length > 0 ? k.ipWhitelistCIDR.join(', ') : 'Unrestricted (0.0.0.0/0)'}
                </span>
              </div>
              <div>
                <span className="text-[10px] text-gray-500 block">Rate Limit Tier</span>
                <span className="text-white text-[11px] font-semibold">{k.rateLimitTier}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
          <div className="bg-[#0B0E14] border border-gray-700 rounded-xl max-w-md w-full p-6 text-white shadow-2xl">
            <h3 className="text-lg font-bold mb-4">Generate API Key</h3>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">Key Label</label>
                <input
                  type="text"
                  value={newLabel}
                  onChange={e => setNewLabel(e.target.value)}
                  placeholder="e.g. Arbitrage Trading Bot"
                  className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2.5 text-xs text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">IP Whitelist CIDRs (Comma-separated)</label>
                <input
                  type="text"
                  value={newCIDR}
                  onChange={e => setNewCIDR(e.target.value)}
                  placeholder="e.g. 103.21.244.0/24, 192.168.1.1"
                  className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2.5 text-xs text-white font-mono"
                />
              </div>
              <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 bg-gray-800 text-gray-300 rounded-lg text-xs"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-bold"
                >
                  Generate Key & Secret
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
