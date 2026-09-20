import React, { useState } from 'react';

export function AntiPhishingShield() {
  const [currentPhrase, setCurrentPhrase] = useState<string>('GROWWW-ALPHA-9942');
  const [newPhraseInput, setNewPhraseInput] = useState<string>('');
  const [isEditing, setIsEditing] = useState<boolean>(false);
  const [authCode, setAuthCode] = useState<string>('');
  const [previewTab, setPreviewTab] = useState<'EMAIL' | 'SMS' | 'PUSH'>('EMAIL');
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);

  const handleUpdatePhrase = () => {
    if (!newPhraseInput.trim() || authCode.length < 6) return;
    setCurrentPhrase(newPhraseInput.trim());
    setIsEditing(false);
    setNewPhraseInput('');
    setAuthCode('');
    setSaveSuccess(true);
    setTimeout(() => setSaveSuccess(false), 4000);
  };

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-2">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00F0A0] inline-block" />
            Anti-Phishing Phrase Configuration & Verification Shield
          </h2>
          <p className="text-[11px] text-slate-400 mt-0.5">
            Cryptographic assurance banner embedded in all authentic Growww emails, SMS, and trade alerts
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="px-2.5 py-1 rounded bg-[#00F0A0]/10 border border-[#00F0A0]/30 text-[#00F0A0] font-bold text-xs flex items-center gap-1.5">
            <span>🛡️</span> Shield Active
          </span>
        </div>
      </div>

      {saveSuccess && (
        <div className="p-3 bg-[#00F0A0]/10 border border-[#00F0A0]/40 rounded text-[#00F0A0] text-xs font-bold">
          ✓ Anti-phishing phrase updated successfully! Your new phrase will appear on all future notifications.
        </div>
      )}

      {/* Configuration & Preview Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Phrase Configuration Card */}
        <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 space-y-3">
          <div className="flex justify-between items-center">
            <span className="font-bold text-white text-xs">Configured Anti-Phishing Code</span>
            <button
              onClick={() => {
                setIsEditing(!isEditing);
                setNewPhraseInput(currentPhrase);
              }}
              className="px-2.5 py-1 rounded bg-[#181F2C] hover:bg-slate-700 text-slate-300 border border-slate-700 text-[11px]"
            >
              {isEditing ? 'Cancel' : 'Change Phrase'}
            </button>
          </div>

          {!isEditing ? (
            <div className="p-3 rounded bg-[#0B0E14] border border-slate-800 flex items-center justify-between">
              <div>
                <span className="text-[10px] text-slate-500 uppercase block font-bold">Your Secret Phrase</span>
                <span className="text-sm font-bold text-[#00F0A0] tracking-wider">{currentPhrase}</span>
              </div>
              <span className="text-[10px] text-slate-400 bg-[#181F2C] px-2 py-1 rounded">2FA Protected</span>
            </div>
          ) : (
            <div className="space-y-3 p-3 rounded bg-[#0B0E14] border border-slate-800">
              <div>
                <label className="text-slate-400 text-[10px] block mb-1">New Secret Anti-Phishing Phrase (4-20 characters)</label>
                <input
                  type="text"
                  placeholder="e.g. SECURE-GROWWW-2026"
                  maxLength={20}
                  value={newPhraseInput}
                  onChange={(e) => setNewPhraseInput(e.target.value)}
                  className="w-full bg-[#121721] border border-slate-700 px-3 py-1.5 rounded text-white text-xs focus:border-[#00F0A0] focus:outline-none font-bold"
                />
              </div>
              <div>
                <label className="text-slate-400 text-[10px] block mb-1">2FA / TOTP 6-Digit Verification Code</label>
                <input
                  type="text"
                  placeholder="Enter 6-digit code"
                  maxLength={6}
                  value={authCode}
                  onChange={(e) => setAuthCode(e.target.value)}
                  className="w-40 bg-[#121721] border border-slate-700 px-3 py-1.5 rounded text-white text-xs focus:border-[#00F0A0] focus:outline-none"
                />
              </div>
              <button
                onClick={handleUpdatePhrase}
                disabled={!newPhraseInput.trim() || authCode.length < 6}
                className={`px-4 py-1.5 rounded font-bold text-xs ${
                  newPhraseInput.trim() && authCode.length >= 6
                    ? 'bg-[#00F0A0] text-black hover:bg-[#00D090]'
                    : 'bg-slate-800 text-slate-600 cursor-not-allowed'
                }`}
              >
                Confirm & Update Phrase
              </button>
            </div>
          )}

          <div className="text-[11px] text-slate-400 space-y-1.5">
            <div className="font-bold text-slate-300">How this protects you:</div>
            <ul className="list-disc list-inside space-y-1 text-[10px] text-slate-400">
              <li>Every legitimate email and withdrawal SMS from Growww contains your secret phrase.</li>
              <li>Phishing attackers and cloned websites cannot reproduce your private phrase.</li>
              <li>Never disclose your password, private seed phrases, or OTP to anyone.</li>
            </ul>
          </div>
        </div>

        {/* Live Notification Preview Card */}
        <div className="p-4 rounded-lg bg-[#121721] border border-slate-800 space-y-3">
          <div className="flex justify-between items-center">
            <span className="font-bold text-white text-xs">Live Verification Shield Preview</span>
            <div className="flex gap-1">
              {(['EMAIL', 'SMS', 'PUSH'] as const).map((tab) => (
                <button
                  key={tab}
                  onClick={() => setPreviewTab(tab)}
                  className={`px-2 py-0.5 rounded text-[10px] font-semibold ${
                    previewTab === tab ? 'bg-[#00F0A0] text-black' : 'bg-[#181F2C] text-slate-400 hover:text-white'
                  }`}
                >
                  {tab}
                </button>
              ))}
            </div>
          </div>

          {/* Email Preview */}
          {previewTab === 'EMAIL' && (
            <div className="p-3 rounded bg-[#181F2C] border border-slate-700 text-[11px] space-y-2">
              <div className="border-b border-slate-700 pb-1.5 flex justify-between text-slate-400 text-[10px]">
                <span>From: security@growww.institutional</span>
                <span>To: institutional-desk@growww.in</span>
              </div>
              {/* Anti-Phishing Banner inside Email */}
              <div className="p-2 rounded bg-[#0B0E14] border border-[#00F0A0]/40 flex items-center justify-between">
                <div className="flex items-center gap-1.5 text-[#00F0A0] font-bold text-[10px]">
                  <span>🛡️ Official Growww Verification Shield</span>
                </div>
                <div className="text-white font-bold text-[11px] bg-[#121721] px-2 py-0.5 rounded border border-slate-700">
                  {currentPhrase}
                </div>
              </div>
              <div className="text-slate-300 space-y-1">
                <div className="font-bold text-white">Withdrawal Authorization Request</div>
                <p className="text-[10px] text-slate-400">
                  A withdrawal request for 2.50 BTC to 0x9481...2b8c has been initiated. If you did not make this request, lock your account immediately.
                </p>
              </div>
            </div>
          )}

          {/* SMS Preview */}
          {previewTab === 'SMS' && (
            <div className="p-3 rounded bg-[#181F2C] border border-slate-700 text-[11px] space-y-2">
              <div className="text-slate-400 text-[10px]">Growww SMS Alert - +91 98****0012</div>
              <div className="p-2.5 rounded bg-[#0B0E14] border border-slate-800 text-slate-200 text-[10px] leading-relaxed">
                [Growww] Security Code: 849201. Your anti-phishing code is <strong className="text-[#00F0A0]">{currentPhrase}</strong>. If this phrase does not match, report phishing immediately to 1800-GROWWW.
              </div>
            </div>
          )}

          {/* Push Notification Preview */}
          {previewTab === 'PUSH' && (
            <div className="p-3 rounded bg-[#181F2C] border border-slate-700 text-[11px] space-y-2">
              <div className="p-2.5 rounded bg-[#0B0E14] border border-slate-800 flex items-start gap-2">
                <span className="text-base">🔔</span>
                <div className="space-y-0.5">
                  <div className="font-bold text-white flex items-center gap-1 text-[10px]">
                    Growww Pro Terminal • <span className="text-[#00F0A0] font-mono">{currentPhrase}</span>
                  </div>
                  <div className="text-[10px] text-slate-400">
                    Order Filled: BUY 1.45 BTC @ ₹5,420,000 / BTC. Executed on NBSE QBFT Engine.
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export const AntiPhishingPhraseVerificationShield = AntiPhishingShield;
export const WebAntiPhishingPhraseVerificationShield = AntiPhishingShield;
