import React, { useState, useEffect } from 'react';

export type HotkeyCategory = 'ORDER_ENTRY' | 'RISK_MANAGEMENT' | 'NAVIGATION' | 'MACRO_STRATEGY';

export interface HotkeyMacro {
  id: string;
  name: string;
  category: HotkeyCategory;
  description: string;
  keyCombination: string; // e.g. "Shift+B", "Ctrl+Alt+C"
  enabled: boolean;
  isDangerous?: boolean;
}

const DEFAULT_HOTKEYS: HotkeyMacro[] = [
  {
    id: 'buy_best_bid',
    name: 'Buy at Best Bid',
    category: 'ORDER_ENTRY',
    description: 'Places passive limit buy order at the highest bid on the book',
    keyCombination: 'Shift+B',
    enabled: true,
  },
  {
    id: 'sell_best_ask',
    name: 'Sell at Best Ask',
    category: 'ORDER_ENTRY',
    description: 'Places passive limit sell order at lowest ask on the book',
    keyCombination: 'Shift+S',
    enabled: true,
  },
  {
    id: 'cross_spread_buy',
    name: 'Aggressive Market Buy 10%',
    category: 'ORDER_ENTRY',
    description: 'Immediately takes liquidity for 10% of portfolio buying power',
    keyCombination: 'Alt+B',
    enabled: true,
  },
  {
    id: 'cross_spread_sell',
    name: 'Aggressive Market Sell 10%',
    category: 'ORDER_ENTRY',
    description: 'Immediately sells 10% of position value at market',
    keyCombination: 'Alt+S',
    enabled: true,
  },
  {
    id: 'panic_cancel_all',
    name: 'Panic: Cancel All Orders',
    category: 'RISK_MANAGEMENT',
    description: 'Cancels all open limit, stop, iceberg and TWAP orders',
    keyCombination: 'Escape',
    enabled: true,
    isDangerous: true,
  },
  {
    id: 'flatten_position',
    name: 'Flatten 100% Position Market',
    category: 'RISK_MANAGEMENT',
    description: 'Immediately liquidates active instrument inventory at best market price',
    keyCombination: 'Shift+X',
    enabled: true,
    isDangerous: true,
  },
  {
    id: 'flip_position',
    name: 'Flip Position Long / Short',
    category: 'RISK_MANAGEMENT',
    description: 'Cancels reverse orders and inverts current net position',
    keyCombination: 'Alt+F',
    enabled: true,
  },
  {
    id: 'focus_orderbook',
    name: 'Focus L2 Orderbook Ladder',
    category: 'NAVIGATION',
    description: 'Transfers keyboard focus to DOM depth ladder',
    keyCombination: 'Ctrl+L',
    enabled: true,
  },
  {
    id: 'toggle_heatmap',
    name: 'Toggle L3 Depth Heatmap',
    category: 'NAVIGATION',
    description: 'Toggles between orderbook ladder and historical liquidity heatmap',
    keyCombination: 'Ctrl+H',
    enabled: true,
  },
  {
    id: 'macro_iceberg_clip',
    name: 'Macro: Iceberg 10x Slices',
    category: 'MACRO_STRATEGY',
    description: 'Splits current order size into 10 randomized tranches',
    keyCombination: 'Alt+I',
    enabled: true,
  },
  {
    id: 'macro_twap_5m',
    name: 'Macro: TWAP 5-Minute Execution',
    category: 'MACRO_STRATEGY',
    description: 'Schedules linear TWAP schedule over next 300 seconds',
    keyCombination: 'Alt+T',
    enabled: true,
  },
];

const PRESETS: Record<string, HotkeyMacro[]> = {
  INSTITUTIONAL_DEFAULT: DEFAULT_HOTKEYS,
  BLOOMBERG_EMSX: DEFAULT_HOTKEYS.map((hk) => {
    if (hk.id === 'buy_best_bid') return { ...hk, keyCombination: 'F1' };
    if (hk.id === 'sell_best_ask') return { ...hk, keyCombination: 'F2' };
    if (hk.id === 'panic_cancel_all') return { ...hk, keyCombination: 'F12' };
    return hk;
  }),
  TRADINGVIEW_PRO: DEFAULT_HOTKEYS.map((hk) => {
    if (hk.id === 'buy_best_bid') return { ...hk, keyCombination: 'Shift+B' };
    if (hk.id === 'sell_best_ask') return { ...hk, keyCombination: 'Shift+S' };
    if (hk.id === 'panic_cancel_all') return { ...hk, keyCombination: 'Ctrl+Alt+C' };
    return hk;
  }),
};

export function HotkeyMacroConfigModal({ isOpen = true, onClose }: { isOpen?: boolean; onClose?: () => void }) {
  const [macros, setMacros] = useState<HotkeyMacro[]>(DEFAULT_HOTKEYS);
  const [selectedCategory, setSelectedCategory] = useState<HotkeyCategory | 'ALL'>('ALL');
  const [recordingId, setRecordingId] = useState<string | null>(null);
  const [lastTriggeredMacro, setLastTriggeredMacro] = useState<string | null>(null);
  const [activePreset, setActivePreset] = useState<string>('INSTITUTIONAL_DEFAULT');
  const [conflictWarning, setConflictWarning] = useState<string | null>(null);

  // Key recording listener
  useEffect(() => {
    if (!recordingId) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();

      // Don't record standalone modifier keys
      if (['Control', 'Shift', 'Alt', 'Meta'].includes(e.key)) {
        return;
      }

      const parts: string[] = [];
      if (e.ctrlKey) parts.push('Ctrl');
      if (e.altKey) parts.push('Alt');
      if (e.shiftKey) parts.push('Shift');
      if (e.metaKey) parts.push('Meta');

      let keyName = e.key;
      if (keyName === ' ') keyName = 'Space';
      else if (keyName.length === 1) keyName = keyName.toUpperCase();

      parts.push(keyName);
      const combo = parts.join('+');

      // Check conflict with dangerous browser reserved shortcuts
      const browserReserved = ['Ctrl+W', 'Ctrl+R', 'Ctrl+T', 'Ctrl+N'];
      if (browserReserved.includes(combo)) {
        setConflictWarning(`Warning: "${combo}" is reserved by the web browser and cannot be used.`);
        setRecordingId(null);
        return;
      }

      // Check conflict with other macros
      const existing = macros.find((m) => m.id !== recordingId && m.keyCombination.toLowerCase() === combo.toLowerCase());
      if (existing) {
        setConflictWarning(`Shortcut "${combo}" conflicts with "${existing.name}".`);
      } else {
        setConflictWarning(null);
      }

      setMacros((prev) =>
        prev.map((m) => (m.id === recordingId ? { ...m, keyCombination: combo } : m))
      );
      setRecordingId(null);
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [recordingId, macros]);

  // Global key listener for simulator
  useEffect(() => {
    if (recordingId) return;

    const handleSimulation = (e: KeyboardEvent) => {
      const parts: string[] = [];
      if (e.ctrlKey) parts.push('Ctrl');
      if (e.altKey) parts.push('Alt');
      if (e.shiftKey) parts.push('Shift');
      if (e.metaKey) parts.push('Meta');

      let keyName = e.key;
      if (keyName === ' ') keyName = 'Space';
      else if (keyName.length === 1) keyName = keyName.toUpperCase();

      parts.push(keyName);
      const pressedCombo = parts.join('+');

      const matched = macros.find(
        (m) => m.enabled && m.keyCombination.toLowerCase() === pressedCombo.toLowerCase()
      );

      if (matched) {
        setLastTriggeredMacro(`${matched.name} (${matched.keyCombination})`);
      }
    };

    window.addEventListener('keydown', handleSimulation);
    return () => window.removeEventListener('keydown', handleSimulation);
  }, [recordingId, macros]);

  const toggleMacro = (id: string) => {
    setMacros((prev) => prev.map((m) => (m.id === id ? { ...m, enabled: !m.enabled } : m)));
  };

  const loadPreset = (presetKey: string) => {
    if (PRESETS[presetKey]) {
      setMacros(PRESETS[presetKey]);
      setActivePreset(presetKey);
      setConflictWarning(null);
    }
  };

  const handleExportJson = () => {
    const jsonStr = JSON.stringify(macros, null, 2);
    navigator?.clipboard?.writeText?.(jsonStr);
  };

  const filteredMacros =
    selectedCategory === 'ALL'
      ? macros
      : macros.filter((m) => m.category === selectedCategory);

  if (!isOpen) return null;

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4 max-w-4xl mx-auto shadow-2xl">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#00F0A0]" />
            <h2 className="text-sm font-bold text-white tracking-wide">
              High-Frequency Hotkey Macro Configuration Modal
            </h2>
          </div>
          <p className="text-[11px] text-slate-400 mt-0.5">
            Sub-millisecond Keyboard Execution Maps • Modifiers Key Capture • Safety Interlocks
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleExportJson}
            className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 hover:border-slate-500 text-slate-300 hover:text-white text-[11px]"
          >
            📋 Copy JSON Profile
          </button>
          {onClose && (
            <button
              onClick={onClose}
              className="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-[11px]"
            >
              ✕ Close
            </button>
          )}
        </div>
      </div>

      {/* Preset & Conflict Warning Row */}
      <div className="flex flex-wrap items-center justify-between gap-3 bg-[#121722] p-2.5 rounded border border-slate-800 text-[11px]">
        <div className="flex items-center gap-2">
          <span className="text-slate-400 font-bold">Desk Preset:</span>
          {Object.keys(PRESETS).map((p) => (
            <button
              key={p}
              onClick={() => loadPreset(p)}
              className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                activePreset === p
                  ? 'bg-[#00F0A0] text-black'
                  : 'bg-[#181F2C] text-slate-400 hover:text-white'
              }`}
            >
              {p.replace('_', ' ')}
            </button>
          ))}
        </div>

        <div className="text-slate-400 text-[10px]">
          Live Test Simulator Active: Press configured key on keyboard
        </div>
      </div>

      {conflictWarning && (
        <div className="p-2.5 rounded bg-rose-950/40 border border-rose-600/50 text-rose-300 text-[11px] flex items-center gap-2">
          <span>⚠️</span>
          <span>{conflictWarning}</span>
        </div>
      )}

      {/* Category Tabs */}
      <div className="flex gap-2 border-b border-slate-800 pb-2 text-[11px]">
        {(['ALL', 'ORDER_ENTRY', 'RISK_MANAGEMENT', 'NAVIGATION', 'MACRO_STRATEGY'] as const).map(
          (cat) => (
            <button
              key={cat}
              onClick={() => setSelectedCategory(cat)}
              className={`px-3 py-1 rounded font-bold transition-colors ${
                selectedCategory === cat
                  ? 'bg-[#00F0A0] text-black'
                  : 'bg-[#121722] text-slate-400 hover:text-slate-200'
              }`}
            >
              {cat.replace('_', ' ')}
            </button>
          )
        )}
      </div>

      {/* Hotkeys Table */}
      <div className="border border-slate-800 rounded overflow-hidden">
        <table className="w-full text-left text-[11px]">
          <thead className="bg-[#121722] text-slate-400 border-b border-slate-800">
            <tr>
              <th className="p-2.5">Enabled</th>
              <th className="p-2.5">Action Name</th>
              <th className="p-2.5">Category</th>
              <th className="p-2.5">Assigned Key Combination</th>
              <th className="p-2.5">Description</th>
              <th className="p-2.5 text-right">Rebind Key</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800 bg-[#0E121B]">
            {filteredMacros.map((macro) => {
              const isRecording = recordingId === macro.id;
              return (
                <tr key={macro.id} className="hover:bg-slate-800/40">
                  <td className="p-2.5">
                    <input
                      type="checkbox"
                      checked={macro.enabled}
                      onChange={() => toggleMacro(macro.id)}
                      className="accent-[#00F0A0] cursor-pointer"
                    />
                  </td>
                  <td className="p-2.5 font-bold text-white">
                    <span className="flex items-center gap-1.5">
                      {macro.isDangerous && <span className="text-rose-400">⚡</span>}
                      {macro.name}
                    </span>
                  </td>
                  <td className="p-2.5">
                    <span className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 text-[10px]">
                      {macro.category}
                    </span>
                  </td>
                  <td className="p-2.5">
                    <span
                      className={`px-2 py-1 rounded font-mono font-bold text-xs ${
                        isRecording
                          ? 'bg-amber-400 text-black animate-pulse'
                          : 'bg-[#181F2C] border border-slate-700 text-[#00F0A0]'
                      }`}
                    >
                      {isRecording ? 'Press Keys Now...' : macro.keyCombination}
                    </span>
                  </td>
                  <td className="p-2.5 text-slate-400 text-[10px] max-w-xs">{macro.description}</td>
                  <td className="p-2.5 text-right">
                    <button
                      onClick={() => setRecordingId(isRecording ? null : macro.id)}
                      className={`px-2.5 py-1 rounded font-bold text-[10px] ${
                        isRecording
                          ? 'bg-rose-500 text-white'
                          : 'bg-[#181F2C] border border-slate-700 text-slate-300 hover:text-white hover:border-[#00F0A0]'
                      }`}
                    >
                      {isRecording ? 'Cancel' : 'Record'}
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Simulator Event Indicator */}
      <div className="p-3 rounded bg-[#121722] border border-slate-800 flex justify-between items-center">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          <span className="text-slate-400 text-[11px]">Last Hotkey Event Triggered:</span>
          <span className="text-white font-bold text-xs">
            {lastTriggeredMacro || 'None (Press any configured shortcut)'}
          </span>
        </div>
        <button
          onClick={() => setLastTriggeredMacro(null)}
          className="text-slate-400 hover:text-slate-200 text-[10px]"
        >
          Clear
        </button>
      </div>
    </div>
  );
}

export const WebHighFrequencyHotkeyMacroConfigurationModal = HotkeyMacroConfigModal;
