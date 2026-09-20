import React, { useState } from 'react';

export interface ResponsibleTradingConfig {
  dailyLossLimitInr: number;
  maxOrderValueInr: number;
  cooldownPeriodMinutes: number;
  isSelfExclusionActive: boolean;
  selfExclusionExpiresAtMs?: number;
  depositLimit24hInr: number;
}

export interface ResponsibleTradingModalProps {
  isOpen?: boolean;
  onClose?: () => void;
  initialConfig?: ResponsibleTradingConfig;
  onSave?: (config: ResponsibleTradingConfig) => void;
  onInitiateSelfExclusion?: (durationDays: number) => void;
}

const DEFAULT_CONFIG: ResponsibleTradingConfig = {
  dailyLossLimitInr: 50000,
  maxOrderValueInr: 100000,
  cooldownPeriodMinutes: 60,
  isSelfExclusionActive: false,
  depositLimit24hInr: 250000,
};

export const ResponsibleTradingModal: React.FC<ResponsibleTradingModalProps> = ({
  isOpen = true,
  onClose,
  initialConfig = DEFAULT_CONFIG,
  onSave,
  onInitiateSelfExclusion,
}) => {
  const [config, setConfig] = useState<ResponsibleTradingConfig>(initialConfig);
  const [selectedExclusionDays, setSelectedExclusionDays] = useState<number>(7);
  const [showConfirmExclusion, setShowConfirmExclusion] = useState<boolean>(false);
  const [saveSuccessMessage, setSaveSuccessMessage] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSave = () => {
    onSave?.(config);
    setSaveSuccessMessage('Responsible trading guardrails saved successfully');
    setTimeout(() => setSaveSuccessMessage(null), 3000);
  };

  const handleExclusionConfirm = () => {
    const expiry = Date.now() + selectedExclusionDays * 24 * 60 * 60 * 1000;
    const updated = {
      ...config,
      isSelfExclusionActive: true,
      selfExclusionExpiresAtMs: expiry,
    };
    setConfig(updated);
    onInitiateSelfExclusion?.(selectedExclusionDays);
    setShowConfirmExclusion(false);
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="responsible-trading-title"
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4"
    >
      <div className="w-full max-w-2xl bg-[#0B0E14] border border-[#181F2C] rounded-xl shadow-2xl p-6 text-slate-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-[#181F2C] pb-4">
          <div className="flex items-center gap-3">
            <div className="w-3 h-3 rounded-full bg-[#00F0A0] animate-pulse" />
            <h2 id="responsible-trading-title" className="text-lg font-bold text-white tracking-wide">
              Investor Protection & Responsible Trading Portal
            </h2>
          </div>
          {onClose && (
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white transition-colors p-1 rounded hover:bg-[#121721]"
              aria-label="Close modal"
            >
              ✕
            </button>
          )}
        </div>

        {/* Self-Exclusion Alert if Active */}
        {config.isSelfExclusionActive && (
          <div className="mt-4 p-4 rounded-lg bg-[#FF3B56]/10 border border-[#FF3B56]/30 text-[#FF3B56] flex flex-col gap-1">
            <span className="font-bold flex items-center gap-2">
              ⚠️ Voluntary Self-Exclusion Active
            </span>
            <span className="text-xs text-slate-300">
              Trading operations and order placement are strictly disabled until{' '}
              <strong className="font-mono text-white">
                {config.selfExclusionExpiresAtMs
                  ? new Date(config.selfExclusionExpiresAtMs).toISOString()
                  : 'Regulatory Review'}
              </strong>
              .
            </span>
          </div>
        )}

        {/* Form Controls */}
        <div className="mt-5 space-y-6">
          {/* Daily Loss Circuit Breaker Slider */}
          <div className="p-4 bg-[#121721] rounded-lg border border-[#181F2C]">
            <div className="flex justify-between items-center mb-2">
              <label htmlFor="daily-loss-slider" className="text-sm font-semibold text-slate-200">
                Daily Loss Circuit Breaker (INR)
              </label>
              <span className="font-mono text-sm font-bold text-[#00F0A0]">
                {`₹${config.dailyLossLimitInr.toLocaleString('en-IN')}`}
              </span>
            </div>
            <p className="text-xs text-slate-400 mb-3">
              Automatically locks out order submission for the remainder of the session if cumulative daily realized losses exceed this threshold.
            </p>
            <input
              id="daily-loss-slider"
              type="range"
              min={10000}
              max={500000}
              step={5000}
              value={config.dailyLossLimitInr}
              onChange={(e) =>
                setConfig({ ...config, dailyLossLimitInr: Number(e.target.value) })
              }
              className="w-full accent-[#00F0A0] bg-[#181F2C] rounded-lg cursor-pointer h-2"
            />
            <div className="flex justify-between text-[11px] font-mono text-slate-500 mt-1">
              <span>₹10,000</span>
              <span>₹250,000</span>
              <span>₹500,000</span>
            </div>
          </div>

          {/* Max Order Value Cap & 24h Deposit Limit */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 bg-[#121721] rounded-lg border border-[#181F2C]">
              <label htmlFor="max-order-val" className="block text-sm font-semibold text-slate-200 mb-1">
                Max Per-Order Value (INR)
              </label>
              <p className="text-xs text-slate-400 mb-2">Single order nominal value cap.</p>
              <input
                id="max-order-val"
                type="number"
                min={1000}
                step={1000}
                value={config.maxOrderValueInr}
                onChange={(e) =>
                  setConfig({ ...config, maxOrderValueInr: Number(e.target.value) })
                }
                className="w-full bg-[#181F2C] border border-slate-700 rounded px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-[#00F0A0]"
              />
            </div>

            <div className="p-4 bg-[#121721] rounded-lg border border-[#181F2C]">
              <label htmlFor="deposit-limit" className="block text-sm font-semibold text-slate-200 mb-1">
                24-Hour Deposit Limit (INR)
              </label>
              <p className="text-xs text-slate-400 mb-2">Maximum inbound fiat/token deposits.</p>
              <input
                id="deposit-limit"
                type="number"
                min={10000}
                step={5000}
                value={config.depositLimit24hInr}
                onChange={(e) =>
                  setConfig({ ...config, depositLimit24hInr: Number(e.target.value) })
                }
                className="w-full bg-[#181F2C] border border-slate-700 rounded px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-[#00F0A0]"
              />
            </div>
          </div>

          {/* Voluntary Self-Exclusion Section */}
          <div className="p-4 bg-[#181F2C]/60 rounded-lg border border-[#FF3B56]/30">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-bold text-white">Voluntary Self-Exclusion & Cooling-Off</h3>
                <p className="text-xs text-slate-400 mt-0.5">
                  Temporarily suspend account trading privileges to enforce discipline. Cannot be prematurely revoked.
                </p>
              </div>
              <button
                type="button"
                onClick={() => setShowConfirmExclusion(!showConfirmExclusion)}
                className="px-3 py-1.5 text-xs font-bold rounded bg-[#FF3B56] hover:bg-[#FF3B56]/80 text-white transition-colors"
              >
                {showConfirmExclusion ? 'Cancel' : 'Initiate Freeze'}
              </button>
            </div>

            {showConfirmExclusion && (
              <div className="mt-4 pt-3 border-t border-[#181F2C] space-y-3">
                <div className="flex gap-2">
                  {[1, 7, 30, 90].map((days) => (
                    <button
                      key={days}
                      type="button"
                      onClick={() => setSelectedExclusionDays(days)}
                      className={`flex-1 py-1.5 text-xs font-mono rounded border transition-colors ${
                        selectedExclusionDays === days
                          ? 'bg-[#FF3B56] border-[#FF3B56] text-white font-bold'
                          : 'bg-[#121721] border-slate-700 text-slate-300 hover:border-slate-500'
                      }`}
                    >
                      {`${days}d`}
                    </button>
                  ))}
                </div>
                <div className="flex items-center justify-between bg-[#121721] p-3 rounded border border-slate-800">
                  <span className="text-xs text-slate-300">
                    Confirm {selectedExclusionDays}-day mandatory cooling freeze?
                  </span>
                  <button
                    type="button"
                    onClick={handleExclusionConfirm}
                    className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white text-xs font-bold rounded transition-colors"
                  >
                    Confirm Lockout
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer Actions */}
        <div className="mt-6 pt-4 border-t border-[#181F2C] flex items-center justify-between">
          <span className="text-xs text-[#00F0A0] font-mono">
            {saveSuccessMessage || 'SEBI CSCRF / Investor Protection Compliant'}
          </span>
          <div className="flex gap-3">
            {onClose && (
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-sm font-semibold rounded bg-[#121721] hover:bg-[#181F2C] text-slate-300 transition-colors"
              >
                Dismiss
              </button>
            )}
            <button
              type="button"
              onClick={handleSave}
              className="px-4 py-2 text-sm font-bold rounded bg-[#00F0A0] hover:bg-[#00F0A0]/80 text-[#0B0E14] transition-colors"
            >
              Save Protection Controls
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
