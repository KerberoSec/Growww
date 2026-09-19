import React, { useState } from 'react';

export interface SecurityMethod {
  id: string;
  name: string;
  type: 'PASSKEY' | 'YUBIKEY_FIDO2' | 'TOTP_AUTHENTICATOR' | 'SMS_OTP';
  enabled: boolean;
  addedAt?: string;
}

export const SecurityCenterPasskeys: React.FC = () => {
  const [methods, setMethods] = useState<SecurityMethod[]>([
    { id: '1', name: 'MacBook Touch ID / Windows Hello', type: 'PASSKEY', enabled: true, addedAt: '2026-08-15' },
    { id: '2', name: 'YubiKey 5C NFC (FIDO2 Hardware Key)', type: 'YUBIKEY_FIDO2', enabled: true, addedAt: '2026-09-01' },
    { id: '3', name: 'Google Authenticator (TOTP)', type: 'TOTP_AUTHENTICATOR', enabled: true, addedAt: '2026-07-20' },
    { id: '4', name: 'SMS OTP Backup (Fallback)', type: 'SMS_OTP', enabled: false },
  ]);

  const toggleMethod = (id: string) => {
    setMethods(methods.map(m => m.id === id ? { ...m, enabled: !m.enabled } : m));
  };

  return (
    <div className="bg-[#0B0E14] text-white p-6 rounded-xl border border-gray-800 max-w-2xl mx-auto font-sans">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-lg bg-emerald-950 flex items-center justify-center text-emerald-400 font-bold border border-emerald-800">
          🛡
        </div>
        <div>
          <h2 className="text-xl font-bold">Security Center & Hardware Authentication</h2>
          <p className="text-gray-400 text-xs">FIDO2 / WebAuthn Biometrics & Hardware Security Key Protection</p>
        </div>
      </div>

      <div className="space-y-3 mb-6">
        {methods.map(method => (
          <div
            key={method.id}
            className="flex justify-between items-center p-4 bg-[#141824] rounded-lg border border-gray-800 hover:border-gray-700 transition"
          >
            <div>
              <div className="font-semibold text-sm text-gray-200">{method.name}</div>
              <div className="text-[11px] text-gray-500 font-mono mt-0.5">
                Type: {method.type} {method.addedAt && `• Enrolled: ${method.addedAt}`}
              </div>
            </div>
            <button
              onClick={() => toggleMethod(method.id)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
                method.enabled
                  ? 'bg-emerald-900/40 text-emerald-400 border border-emerald-800'
                  : 'bg-gray-800 text-gray-400 border border-gray-700'
              }`}
            >
              {method.enabled ? 'Enabled' : 'Disabled'}
            </button>
          </div>
        ))}
      </div>

      <div className="flex gap-3">
        <button className="flex-1 py-3 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-xs transition">
          + Register New Hardware Key (FIDO2)
        </button>
        <button className="flex-1 py-3 bg-[#161616] hover:bg-gray-800 text-gray-200 font-semibold rounded-lg text-xs border border-gray-800 transition">
          Generate Emergency Recovery Codes
        </button>
      </div>
    </div>
  );
};
