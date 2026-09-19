import React, { useEffect, useState } from 'react';

export function HotkeyExecutionManager() {
  const [lastExecutedHotkey, setLastExecutedHotkey] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.shiftKey && e.key.toUpperCase() === 'B') {
        setLastExecutedHotkey('BUY_BEST_BID (Shift+B)');
      } else if (e.shiftKey && e.key.toUpperCase() === 'S') {
        setLastExecutedHotkey('SELL_BEST_ASK (Shift+S)');
      } else if (e.key === 'Escape') {
        setLastExecutedHotkey('PANIC_CANCEL_ALL (Esc)');
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-2">
      <div className="flex justify-between items-center">
        <span className="font-bold text-[#00F0A0]">Hotkey Execution Manager</span>
        <span className="text-[10px] text-slate-400">Listener Active</span>
      </div>
      <div className="text-slate-400">
        Last Hotkey Triggered: <span className="text-white font-bold">{lastExecutedHotkey || 'None'}</span>
      </div>
    </div>
  );
}
