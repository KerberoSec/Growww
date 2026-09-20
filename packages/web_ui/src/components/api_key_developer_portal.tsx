import React, { useState } from 'react';

export interface ApiKeyItem {
  id: string;
  label: string;
  maskedKey: string;
  scopes: ('read' | 'trade' | 'withdraw')[];
  ipWhitelist: string[];
  rateLimitPerSec: number;
  createdAt: string;
  lastUsedAt?: string;
  isActive: boolean;
}

export interface ApiKeyDeveloperPortalProps {
  initialKeys?: ApiKeyItem[];
  onCreateKey?: (label: string, scopes: string[], ips: string[]) => void;
  onRevokeKey?: (id: string) => void;
}

export const ApiKeyDeveloperPortal: React.FC<ApiKeyDeveloperPortalProps> = ({
  initialKeys = [
    {
      id: 'key_1',
      label: 'Algo Trading Bot Alpha',
      maskedKey: 'gw_live_99f2****************881a',
      scopes: ['read', 'trade'],
      ipWhitelist: ['203.0.113.15/32', '198.51.100.0/24'],
      rateLimitPerSec: 50,
      createdAt: '2026-08-15',
      lastUsedAt: '2026-09-20 11:45:00',
      isActive: true,
    },
    {
      id: 'key_2',
      label: 'Portfolio Reporting Daemon',
      maskedKey: 'gw_live_12ab****************44fe',
      scopes: ['read'],
      ipWhitelist: ['10.0.0.0/16'],
      rateLimitPerSec: 10,
      createdAt: '2026-09-01',
      lastUsedAt: '2026-09-20 09:12:30',
      isActive: true,
    },
  ],
  onCreateKey,
  onRevokeKey,
}) => {
  const [keys, setKeys] = useState<ApiKeyItem[]>(initialKeys);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [newLabel, setNewLabel] = useState('');
  const [readScope, setReadScope] = useState(true);
  const [tradeScope, setTradeScope] = useState(true);
  const [withdrawScope, setWithdrawScope] = useState(false);
  const [cidrList, setCidrList] = useState('');

  const handleCreate = () => {
    if (!newLabel.trim()) return;

    const scopes: ('read' | 'trade' | 'withdraw')[] = [];
    if (readScope) scopes.push('read');
    if (tradeScope) scopes.push('trade');
    if (withdrawScope) scopes.push('withdraw');

    const ips = cidrList
      .split('\n')
      .map((ip) => ip.trim())
      .filter((ip) => ip.length > 0);

    const newKey: ApiKeyItem = {
      id: `key_${Date.now()}`,
      label: newLabel,
      maskedKey: `gw_live_${Math.random().toString(36).substring(2, 6)}****************${Math.random().toString(36).substring(2, 6)}`,
      scopes,
      ipWhitelist: ips.length > 0 ? ips : ['Any IP'],
      rateLimitPerSec: tradeScope ? 50 : 10,
      createdAt: new Date().toISOString().split('T')[0],
      isActive: true,
    };

    setKeys([...keys, newKey]);
    onCreateKey?.(newLabel, scopes, ips);
    setIsCreateModalOpen(false);
    setNewLabel('');
    setCidrList('');
  };

  const handleRevoke = (id: string) => {
    setKeys(keys.filter((k) => k.id !== id));
    onRevokeKey?.(id);
  };

  return (
    <div className="p-6 bg-[#0B0E14] text-white rounded-xl border border-gray-800 max-w-5xl mx-auto shadow-2xl">
      <div className="flex justify-between items-center mb-6 pb-4 border-b border-gray-800">
        <div>
          <h2 className="text-xl font-bold text-[#00F0A0] tracking-wide">Developer API Key Portal</h2>
          <p className="text-xs text-gray-400 mt-1">Manage programmatic trading keys, CIDR IP restrictions, and HMAC secrets</p>
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="px-4 py-2 bg-[#00F0A0] text-black font-semibold text-xs rounded hover:bg-[#00d68f] transition shadow-lg"
        >
          + Generate New API Key
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="border-b border-gray-800 text-gray-400 uppercase tracking-wider">
              <th className="py-3 px-4">Label</th>
              <th className="py-3 px-4">API Key</th>
              <th className="py-3 px-4">Scopes</th>
              <th className="py-3 px-4">IP Whitelist (CIDR)</th>
              <th className="py-3 px-4">Rate Limit</th>
              <th className="py-3 px-4">Created</th>
              <th className="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800 font-mono">
            {keys.map((key) => (
              <tr key={key.id} className="hover:bg-gray-900/50 transition">
                <td className="py-3 px-4 font-sans font-medium text-white">{key.label}</td>
                <td className="py-3 px-4 text-gray-400">{key.maskedKey}</td>
                <td className="py-3 px-4">
                  <div className="flex gap-1 flex-wrap">
                    {key.scopes.map((s) => (
                      <span
                        key={s}
                        className={`px-2 py-0.5 rounded text-[10px] font-sans uppercase font-bold ${
                          s === 'withdraw'
                            ? 'bg-red-950/60 text-red-400 border border-red-800'
                            : s === 'trade'
                            ? 'bg-blue-950/60 text-blue-400 border border-blue-800'
                            : 'bg-gray-800 text-gray-300'
                        }`}
                      >
                        {s}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="py-3 px-4 text-gray-300">
                  {key.ipWhitelist.map((ip) => (
                    <div key={ip} className="text-[11px]">{ip}</div>
                  ))}
                </td>
                <td className="py-3 px-4 text-gray-300">{`${key.rateLimitPerSec} req/s`}</td>
                <td className="py-3 px-4 text-gray-400">{key.createdAt}</td>
                <td className="py-3 px-4 text-right">
                  <button
                    onClick={() => handleRevoke(key.id)}
                    className="text-red-400 hover:text-red-300 font-sans text-xs underline"
                  >
                    Revoke
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {isCreateModalOpen && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
          <div className="bg-[#121721] p-6 rounded-xl border border-gray-700 max-w-md w-full shadow-2xl">
            <h3 className="text-lg font-bold text-white mb-4">Create New Programmatic API Key</h3>

            <div className="space-y-4 text-xs font-sans">
              <div>
                <label className="block text-gray-400 mb-1">Key Label / Purpose</label>
                <input
                  type="text"
                  placeholder="e.g. Market Making Bot #2"
                  value={newLabel}
                  onChange={(e) => setNewLabel(e.target.value)}
                  className="w-full bg-[#181F2C] border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-[#00F0A0]"
                />
              </div>

              <div>
                <label className="block text-gray-400 mb-1">Access Scopes</label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-1.5 text-gray-300">
                    <input type="checkbox" checked={readScope} onChange={(e) => setReadScope(e.target.checked)} />
                    Read
                  </label>
                  <label className="flex items-center gap-1.5 text-gray-300">
                    <input type="checkbox" checked={tradeScope} onChange={(e) => setTradeScope(e.target.checked)} />
                    Trade
                  </label>
                  <label className="flex items-center gap-1.5 text-red-400">
                    <input type="checkbox" checked={withdrawScope} onChange={(e) => setWithdrawScope(e.target.checked)} />
                    Withdraw
                  </label>
                </div>
              </div>

              <div>
                <label className="block text-gray-400 mb-1">CIDR IP Whitelist (One per line)</label>
                <textarea
                  rows={3}
                  placeholder="203.0.113.15/32&#10;198.51.100.0/24"
                  value={cidrList}
                  onChange={(e) => setCidrList(e.target.value)}
                  className="w-full bg-[#181F2C] border border-gray-700 rounded px-3 py-2 text-white font-mono focus:outline-none focus:border-[#00F0A0]"
                />
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                <button
                  onClick={() => setIsCreateModalOpen(false)}
                  className="px-4 py-2 bg-gray-800 text-gray-300 rounded hover:bg-gray-700"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreate}
                  className="px-4 py-2 bg-[#00F0A0] text-black font-semibold rounded hover:bg-[#00d68f]"
                >
                  Confirm & Generate
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ApiKeyDeveloperPortal;
