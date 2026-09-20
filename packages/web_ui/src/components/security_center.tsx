import React, { useState } from 'react';

export interface PasskeyItem {
  id: string;
  name: string;
  type: 'WEBAUTHN_BIOMETRIC' | 'FIDO2_HARDWARE';
  addedAt: string;
  lastUsed: string;
}

export function SecurityCenter() {
  const [totpEnabled, setTotpEnabled] = useState<boolean>(true);
  const [require2faForWithdrawals, setRequire2faForWithdrawals] = useState<boolean>(true);
  const [require2faForApiKeys, setRequire2faForApiKeys] = useState<boolean>(true);
  const [require2faForLargeTrades, setRequire2faForLargeTrades] = useState<boolean>(true);

  const [passkeys, setPasskeys] = useState<PasskeyItem[]>([
    {
      id: 'PK-01',
      name: "MacBook Pro Touch ID",
      type: 'WEBAUTHN_BIOMETRIC',
      addedAt: '2026-08-12',
      lastUsed: '2026-09-20 08:30',
    },
    {
      id: 'PK-02',
      name: 'YubiKey 5C NFC Primary',
      type: 'FIDO2_HARDWARE',
      addedAt: '2026-07-04',
      lastUsed: '2026-09-19 19:15',
    },
  ]);

  const [showTotpModal, setShowTotpModal] = useState<boolean>(false);
  const [showAddKeyModal, setShowAddKeyModal] = useState<boolean>(false);
  const [newKeyName, setNewKeyName] = useState<string>('');
  const [newKeyType, setNewKeyType] = useState<'WEBAUTHN_BIOMETRIC' | 'FIDO2_HARDWARE'>('FIDO2_HARDWARE');

  // Compute security score
  let score = 30; // base score
  if (totpEnabled) score += 30;
  if (passkeys.some((k) => k.type === 'WEBAUTHN_BIOMETRIC')) score += 20;
  if (passkeys.some((k) => k.type === 'FIDO2_HARDWARE')) score += 20;

  const handleRegisterPasskey = () => {
    if (!newKeyName.trim()) return;
    const newEntry: PasskeyItem = {
      id: `PK-${Date.now().toString().slice(-4)}`,
      name: newKeyName,
      type: newKeyType,
      addedAt: new Date().toISOString().split('T')[0],
      lastUsed: 'Just now',
    };
    setPasskeys((prev) => [...prev, newEntry]);
    setNewKeyName('');
    setShowAddKeyModal(false);
  };

  const handleDeletePasskey = (id: string) => {
    setPasskeys((prev) => prev.filter((k) => k.id !== id));
  };

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-2">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00F0A0] inline-block" />
            Security Center: Multi-Factor Authentication & Cryptographic Keys
          </h2>
          <p className="text-[11px] text-slate-400 mt-0.5">
            FIDO2 WebAuthn Passkeys, YubiKey Hardware Enclaves & Time-based One-Time Passwords (TOTP)
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="px-2.5 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] font-bold text-xs">
            Institutional Tier 3 Active
          </span>
        </div>
      </div>

      {/* Security Health Score Banner */}
      <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="text-xs font-bold text-slate-200">Account Security Strength</div>
          <div className="text-[11px] text-slate-400">
            Hardware keys and biometric passkeys protect against SIM-swapping, phishing, and man-in-the-middle attacks.
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="text-right">
            <div className="text-xl font-bold text-[#00F0A0]">{score} / 100</div>
            <div className="text-[10px] text-slate-400">
              {score >= 90 ? 'Maximum Protection' : score >= 60 ? 'Strong Protection' : 'Needs Improvement'}
            </div>
          </div>
          <div className="w-12 h-12 rounded-full border-4 border-[#00F0A0] flex items-center justify-center font-bold text-xs text-[#00F0A0]">
            {score}%
          </div>
        </div>
      </div>

      {/* Primary 2FA Methods */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* TOTP Authenticator Card */}
        <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 space-y-3">
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-white">Authenticator App (TOTP)</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400 font-bold">
                RFC 6238
              </span>
            </div>
            <span
              className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                totpEnabled ? 'bg-[#00F0A0]/20 text-[#00F0A0]' : 'bg-slate-800 text-slate-500'
              }`}
            >
              {totpEnabled ? 'ACTIVE' : 'DISABLED'}
            </span>
          </div>
          <p className="text-[11px] text-slate-400">
            Use Google Authenticator, Authy, or 1Password to generate 6-digit dynamic verification codes.
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setShowTotpModal(true)}
              className="px-3 py-1.5 rounded bg-[#181F2C] hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold"
            >
              {totpEnabled ? 'Reconfigure TOTP' : 'Setup Authenticator'}
            </button>
            <button
              onClick={() => setTotpEnabled(!totpEnabled)}
              className={`px-3 py-1.5 rounded text-xs font-semibold ${
                totpEnabled ? 'bg-[#FF3B56]/20 hover:bg-[#FF3B56]/30 text-[#FF3B56]' : 'bg-[#00F0A0] text-black'
              }`}
            >
              {totpEnabled ? 'Disable' : 'Enable'}
            </button>
          </div>
        </div>

        {/* Passkeys & Hardware Keys Card */}
        <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 space-y-3">
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-white">Passkeys & Security Keys</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-[#00F0A0]/20 text-[#00F0A0] font-bold">
                FIDO2 / WebAuthn
              </span>
            </div>
            <button
              onClick={() => setShowAddKeyModal(true)}
              className="px-2.5 py-1 rounded bg-[#00F0A0] hover:bg-[#00D090] text-black font-bold text-[10px]"
            >
              + Add Key
            </button>
          </div>
          <p className="text-[11px] text-slate-400">
            Hardware enclaves (TouchID, FaceID) and FIDO2 physical keys (YubiKey 5 Series).
          </p>
          <div className="space-y-1.5 max-h-36 overflow-y-auto pr-1">
            {passkeys.map((pk) => (
              <div
                key={pk.id}
                className="flex justify-between items-center p-2 rounded bg-[#181F2C] border border-slate-800 text-[11px]"
              >
                <div>
                  <div className="font-bold text-white flex items-center gap-1.5">
                    {pk.name}
                    <span className="text-[9px] px-1 rounded bg-slate-700 text-slate-300">
                      {pk.type === 'WEBAUTHN_BIOMETRIC' ? 'Biometric' : 'YubiKey'}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500">Last used: {pk.lastUsed}</div>
                </div>
                <button
                  onClick={() => handleDeletePasskey(pk.id)}
                  className="text-slate-500 hover:text-[#FF3B56] text-[10px]"
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* High-Security Action Policy Controls */}
      <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 space-y-3">
        <span className="text-xs font-bold text-white block">Institutional Security Enforcement Policies</span>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-[11px]">
          <label className="flex items-center gap-2 p-2.5 rounded bg-[#181F2C] border border-slate-800 cursor-pointer">
            <input
              type="checkbox"
              checked={require2faForWithdrawals}
              onChange={(e) => setRequire2faForWithdrawals(e.target.checked)}
              className="accent-[#00F0A0]"
            />
            <span className="text-slate-300">Enforce on All Withdrawals</span>
          </label>

          <label className="flex items-center gap-2 p-2.5 rounded bg-[#181F2C] border border-slate-800 cursor-pointer">
            <input
              type="checkbox"
              checked={require2faForApiKeys}
              onChange={(e) => setRequire2faForApiKeys(e.target.checked)}
              className="accent-[#00F0A0]"
            />
            <span className="text-slate-300">Enforce on API Key Changes</span>
          </label>

          <label className="flex items-center gap-2 p-2.5 rounded bg-[#181F2C] border border-slate-800 cursor-pointer">
            <input
              type="checkbox"
              checked={require2faForLargeTrades}
              onChange={(e) => setRequire2faForLargeTrades(e.target.checked)}
              className="accent-[#00F0A0]"
            />
            <span className="text-slate-300">Enforce on Orders &gt; ₹10L</span>
          </label>
        </div>
      </div>

      {/* TOTP Setup Modal */}
      {showTotpModal && (
        <div className="p-4 rounded bg-[#181F2C] border border-slate-700 space-y-3">
          <div className="flex justify-between items-center border-b border-slate-700 pb-2">
            <span className="font-bold text-white text-sm">Enroll Google Authenticator / TOTP</span>
            <button onClick={() => setShowTotpModal(false)} className="text-slate-400 hover:text-white">
              ✕
            </button>
          </div>
          <div className="flex flex-col sm:flex-row items-center gap-4">
            <div className="w-24 h-24 bg-white p-1 rounded flex items-center justify-center text-black font-bold text-[9px] text-center border">
              [QR CODE ENCODED SECRET]
            </div>
            <div className="space-y-2 text-slate-300 text-[11px] flex-1">
              <div>1. Scan QR code in your Authenticator app.</div>
              <div>
                2. Secret Key: <span className="font-mono text-[#00F0A0] bg-[#121721] px-2 py-0.5 rounded">JBSWY3DPEHPK3PXP</span>
              </div>
              <div className="flex gap-2 items-center pt-1">
                <input
                  type="text"
                  placeholder="Enter 6-digit code"
                  maxLength={6}
                  className="bg-[#121721] border border-slate-600 px-3 py-1 rounded text-white text-xs w-36 focus:border-[#00F0A0] focus:outline-none"
                />
                <button
                  onClick={() => {
                    setTotpEnabled(true);
                    setShowTotpModal(false);
                  }}
                  className="px-3 py-1 rounded bg-[#00F0A0] text-black font-bold text-xs"
                >
                  Verify & Save
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Passkey / Hardware Key Enrollment Modal */}
      {showAddKeyModal && (
        <div className="p-4 rounded bg-[#181F2C] border border-slate-700 space-y-3">
          <div className="flex justify-between items-center border-b border-slate-700 pb-2">
            <span className="font-bold text-white text-sm">Register FIDO2 WebAuthn Credential</span>
            <button onClick={() => setShowAddKeyModal(false)} className="text-slate-400 hover:text-white">
              ✕
            </button>
          </div>
          <div className="space-y-3">
            <div>
              <label className="text-slate-400 text-[11px] block mb-1">Key Nickname</label>
              <input
                type="text"
                placeholder="e.g. YubiKey 5C NFC Backup or Office Laptop"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="w-full bg-[#121721] border border-slate-600 px-3 py-1.5 rounded text-white text-xs focus:border-[#00F0A0] focus:outline-none"
              />
            </div>
            <div>
              <label className="text-slate-400 text-[11px] block mb-1">Credential Type</label>
              <select
                value={newKeyType}
                onChange={(e) => setNewKeyType(e.target.value as any)}
                className="bg-[#121721] border border-slate-600 px-3 py-1.5 rounded text-white text-xs focus:border-[#00F0A0] focus:outline-none w-full"
              >
                <option value="FIDO2_HARDWARE">FIDO2 Hardware Key (YubiKey / Nitrokey)</option>
                <option value="WEBAUTHN_BIOMETRIC">Platform Biometrics (Touch ID / Face ID / Windows Hello)</option>
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                onClick={() => setShowAddKeyModal(false)}
                className="px-3 py-1.5 rounded bg-slate-800 text-slate-300 text-xs"
              >
                Cancel
              </button>
              <button
                onClick={handleRegisterPasskey}
                className="px-4 py-1.5 rounded bg-[#00F0A0] text-black font-bold text-xs"
              >
                Enroll Key via WebAuthn
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export const WebSecurityCenterTotpPasskeysYubikeyFido2 = SecurityCenter;
