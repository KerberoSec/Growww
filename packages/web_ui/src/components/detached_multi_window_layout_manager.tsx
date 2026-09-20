import React, { useState, useEffect } from 'react';

export interface DetachedWindowInfo {
  id: string;
  panelType: 'ORDERBOOK_LADDER' | 'TRADINGVIEW_CHART' | 'RISK_PORTFOLIO' | 'INSTITUTIONAL_RFQ' | 'DEPTH_HEATMAP';
  title: string;
  monitorAssignment: string;
  state: 'ACTIVE' | 'BACKGROUND' | 'MINIMIZED';
  lastPingMs: number;
  syncPacketsCount: number;
  windowCoordinates: { x: number; y: number; width: number; height: number };
}

export interface SyncMessagePacket {
  sequenceId: number;
  type: 'SYMBOL_SYNC' | 'ORDER_EVENT' | 'CROSSHAIR_SYNC' | 'HEARTBEAT_PING';
  payload: Record<string, unknown>;
  timestamp: string;
}

export function DetachedMultiWindowLayoutManager() {
  const [broadcastChannelName] = useState<string>('growww_multiwindow_bus');
  const [syncStatus, setSyncStatus] = useState<'CONNECTED' | 'DISCONNECTED'>('CONNECTED');
  const [symbolSyncEnabled, setSymbolSyncEnabled] = useState<boolean>(true);
  const [crosshairSyncEnabled, setCrosshairSyncEnabled] = useState<boolean>(true);
  const [sharedSymbol, setSharedSymbol] = useState<string>('BTC/USDT');
  const [layoutPreset, setLayoutPreset] = useState<string>('3_SCREEN_DESK');
  const [sequenceCounter, setSequenceCounter] = useState<number>(1042);

  const [detachedWindows, setDetachedWindows] = useState<DetachedWindowInfo[]>([
    {
      id: 'win-chart-aux-01',
      panelType: 'TRADINGVIEW_CHART',
      title: 'TradingView 4x Multi-Chart Grid',
      monitorAssignment: 'Monitor 2 (Right - 4K)',
      state: 'ACTIVE',
      lastPingMs: 0.6,
      syncPacketsCount: 4521,
      windowCoordinates: { x: 3840, y: 0, width: 3840, height: 2160 },
    },
    {
      id: 'win-depth-aux-02',
      panelType: 'ORDERBOOK_LADDER',
      title: 'L2/L3 Ultra-DOM Orderbook Ladder',
      monitorAssignment: 'Monitor 1 (Left - 1440p)',
      state: 'ACTIVE',
      lastPingMs: 0.4,
      syncPacketsCount: 8940,
      windowCoordinates: { x: -2560, y: 0, width: 2560, height: 1440 },
    },
    {
      id: 'win-risk-aux-03',
      panelType: 'RISK_PORTFOLIO',
      title: 'Institutional Risk & Collateral Monitor',
      monitorAssignment: 'Monitor 3 (Vertical - 1080p)',
      state: 'BACKGROUND',
      lastPingMs: 1.2,
      syncPacketsCount: 1240,
      windowCoordinates: { x: 0, y: 1440, width: 1080, height: 1920 },
    },
  ]);

  const [recentSyncPackets, setRecentSyncPackets] = useState<SyncMessagePacket[]>([
    {
      sequenceId: 1040,
      type: 'SYMBOL_SYNC',
      payload: { symbol: 'BTC/USDT', spot: 65240 },
      timestamp: '11:23:45.102',
    },
    {
      sequenceId: 1041,
      type: 'CROSSHAIR_SYNC',
      payload: { timestamp: 1726811625, price: 65240 },
      timestamp: '11:23:46.400',
    },
  ]);

  // Broadcast channel setup if supported in environment
  useEffect(() => {
    let channel: BroadcastChannel | null = null;
    try {
      if (typeof window !== 'undefined' && 'BroadcastChannel' in window) {
        channel = new BroadcastChannel(broadcastChannelName);
        channel.onmessage = (event) => {
          if (event.data?.type === 'HEARTBEAT_PING') {
            channel?.postMessage({ type: 'HEARTBEAT_PONG', timestamp: Date.now() });
          }
        };
      }
    } catch {
      // Fallback or Node test environment
    }

    return () => {
      channel?.close();
    };
  }, [broadcastChannelName]);

  const handleDetachPanel = (
    panelType: DetachedWindowInfo['panelType'],
    title: string
  ) => {
    const newId = `win-${panelType.toLowerCase().slice(0, 5)}-${Date.now().toString(36)}`;
    const newWindow: DetachedWindowInfo = {
      id: newId,
      panelType,
      title,
      monitorAssignment: 'Monitor 2 (Detected External)',
      state: 'ACTIVE',
      lastPingMs: 0.5,
      syncPacketsCount: 1,
      windowCoordinates: { x: 1920, y: 100, width: 1200, height: 800 },
    };

    setDetachedWindows((prev) => [...prev, newWindow]);

    // Dispatch broadcast event
    const newSeq = sequenceCounter + 1;
    setSequenceCounter(newSeq);
    setRecentSyncPackets((prev) => [
      {
        sequenceId: newSeq,
        type: 'ORDER_EVENT',
        payload: { action: 'SPAWN_WINDOW', id: newId, title },
        timestamp: new Date().toLocaleTimeString(),
      },
      ...prev.slice(0, 4),
    ]);
  };

  const handleReDockWindow = (id: string) => {
    setDetachedWindows((prev) => prev.filter((w) => w !== undefined && w.id !== id));
    const newSeq = sequenceCounter + 1;
    setSequenceCounter(newSeq);
    setRecentSyncPackets((prev) => [
      {
        sequenceId: newSeq,
        type: 'ORDER_EVENT',
        payload: { action: 'RE_DOCK_WINDOW', id },
        timestamp: new Date().toLocaleTimeString(),
      },
      ...prev.slice(0, 4),
    ]);
  };

  const handleSendSyncPulse = () => {
    const newSeq = sequenceCounter + 1;
    setSequenceCounter(newSeq);
    const newPacket: SyncMessagePacket = {
      sequenceId: newSeq,
      type: 'HEARTBEAT_PING',
      payload: { origin: 'MAIN_WORKSTATION', activeSymbol: sharedSymbol },
      timestamp: new Date().toLocaleTimeString(),
    };

    setRecentSyncPackets((prev) => [newPacket, ...prev.slice(0, 4)]);

    // Update window ping latency and packet counter
    setDetachedWindows((prev) =>
      prev.map((w) => ({
        ...w,
        lastPingMs: Number((0.2 + Math.random() * 0.6).toFixed(2)),
        syncPacketsCount: w.syncPacketsCount + 1,
      }))
    );
  };

  const handleSymbolChange = (sym: string) => {
    setSharedSymbol(sym);
    if (symbolSyncEnabled) {
      const newSeq = sequenceCounter + 1;
      setSequenceCounter(newSeq);
      setRecentSyncPackets((prev) => [
        {
          sequenceId: newSeq,
          type: 'SYMBOL_SYNC',
          payload: { symbol: sym },
          timestamp: new Date().toLocaleTimeString(),
        },
        ...prev.slice(0, 4),
      ]);
    }
  };

  const handleApplyPreset = (preset: string) => {
    setLayoutPreset(preset);
    if (preset === '1_SCREEN_DOCKED') {
      setDetachedWindows([]);
    } else if (preset === '2_SCREEN_PRO') {
      setDetachedWindows(detachedWindows.slice(0, 1));
    }
  };

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#00F0A0]" />
            <h2 className="text-sm font-bold text-white tracking-wide">
              Detached Multi-Window Layout Manager & Cross-Screen State Sync
            </h2>
          </div>
          <p className="text-[11px] text-slate-400 mt-0.5">
            BroadcastChannel Inter-Process Bus • Multi-Monitor Topology • Zero-Lag Symbol & Crosshair Mirroring
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-slate-400 text-[11px]">Sync Bus:</span>
          <span className="px-2 py-0.5 bg-[#181F2C] border border-slate-700 text-[#00F0A0] font-bold rounded">
            {broadcastChannelName}
          </span>
          <span className="px-2 py-0.5 rounded bg-[#00F0A0]/10 text-[#00F0A0] border border-[#00F0A0]/30 font-bold text-[10px]">
            {syncStatus}
          </span>
        </div>
      </div>

      {/* Multi-Window Topology & Preset Controls */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 bg-[#121722] p-3 rounded border border-slate-800">
        <div>
          <span className="text-[10px] text-slate-400 block font-bold">DESK DISPLAY TOPOLOGY</span>
          <div className="flex gap-1.5 mt-1.5">
            {[
              { id: '1_SCREEN_DOCKED', label: '1-Screen' },
              { id: '2_SCREEN_PRO', label: '2-Screen Dual' },
              { id: '3_SCREEN_DESK', label: '3-Screen Desk' },
              { id: '4_SCREEN_COMMAND', label: '4-Screen Pro' },
            ].map((p) => (
              <button
                key={p.id}
                onClick={() => handleApplyPreset(p.id)}
                className={`flex-1 py-1 rounded text-[10px] font-bold ${
                  layoutPreset === p.id
                    ? 'bg-[#00F0A0] text-black'
                    : 'bg-[#181F2C] text-slate-300 hover:text-white'
                }`}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>

        <div>
          <span className="text-[10px] text-slate-400 block font-bold">SHARED INSTRUMENT SYMBOL</span>
          <div className="flex items-center gap-2 mt-1.5">
            <select
              value={sharedSymbol}
              onChange={(e) => handleSymbolChange(e.target.value)}
              className="flex-1 bg-[#181F2C] border border-slate-700 rounded p-1 text-white font-bold"
            >
              <option value="BTC/USDT">BTC/USDT ($65,240)</option>
              <option value="ETH/USDT">ETH/USDT ($3,480)</option>
              <option value="SOL/USDT">SOL/USDT ($154)</option>
              <option value="NIFTY-50">NIFTY 50 (₹24,500)</option>
            </select>
          </div>
        </div>

        <div>
          <span className="text-[10px] text-slate-400 block font-bold">SYNC POLICIES</span>
          <div className="flex items-center gap-4 mt-2">
            <label className="flex items-center gap-1.5 cursor-pointer text-slate-300 text-[11px]">
              <input
                type="checkbox"
                checked={symbolSyncEnabled}
                onChange={(e) => setSymbolSyncEnabled(e.target.checked)}
                className="accent-[#00F0A0]"
              />
              Symbol Sync
            </label>
            <label className="flex items-center gap-1.5 cursor-pointer text-slate-300 text-[11px]">
              <input
                type="checkbox"
                checked={crosshairSyncEnabled}
                onChange={(e) => setCrosshairSyncEnabled(e.target.checked)}
                className="accent-[#00F0A0]"
              />
              Crosshair Sync
            </label>
          </div>
        </div>
      </div>

      {/* Detach Quick Action Bar */}
      <div className="flex flex-wrap items-center justify-between gap-2 p-3 bg-[#0E121B] rounded border border-slate-800">
        <div className="flex items-center gap-2">
          <span className="text-slate-300 font-bold">Pop-Out New Screen:</span>
          <button
            onClick={() => handleDetachPanel('TRADINGVIEW_CHART', 'TradingView Detached Chart')}
            className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 hover:border-[#00F0A0] text-slate-200 hover:text-white text-[11px]"
          >
            + Chart Canvas
          </button>
          <button
            onClick={() => handleDetachPanel('ORDERBOOK_LADDER', 'L2/L3 Detached Depth Ladder')}
            className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 hover:border-[#00F0A0] text-slate-200 hover:text-white text-[11px]"
          >
            + DOM Ladder
          </button>
          <button
            onClick={() => handleDetachPanel('INSTITUTIONAL_RFQ', 'Institutional RFQ Blotter')}
            className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 hover:border-[#00F0A0] text-slate-200 hover:text-white text-[11px]"
          >
            + RFQ Blotter
          </button>
        </div>

        <button
          onClick={handleSendSyncPulse}
          className="px-3 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a] transition-colors text-[11px]"
        >
          ⚡ Send Heartbeat Pulse
        </button>
      </div>

      {/* Detached Windows Grid / Table */}
      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <span className="font-bold text-slate-300">
            Active Detached Secondary Windows ({detachedWindows.length})
          </span>
          <span className="text-slate-400 text-[10px]">
            Average Sync Roundtrip: ~0.5ms (BroadcastChannel IPC)
          </span>
        </div>

        <div className="border border-slate-800 rounded overflow-hidden">
          <table className="w-full text-left text-[11px]">
            <thead className="bg-[#121722] text-slate-400 border-b border-slate-800">
              <tr>
                <th className="p-2.5">Window ID</th>
                <th className="p-2.5">Panel Component</th>
                <th className="p-2.5">Monitor Topology</th>
                <th className="p-2.5">Status</th>
                <th className="p-2.5">Latency</th>
                <th className="p-2.5">Sync Packets</th>
                <th className="p-2.5 text-right">Window Control</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800 bg-[#0E121B]">
              {detachedWindows.length === 0 ? (
                <tr>
                  <td colSpan={7} className="p-4 text-center text-slate-500">
                    No secondary windows detached. Workstation is in single-screen docked mode.
                  </td>
                </tr>
              ) : (
                detachedWindows.map((win) => (
                  <tr key={win.id} className="hover:bg-slate-800/40">
                    <td className="p-2.5 font-mono text-slate-300 font-bold">{win.id}</td>
                    <td className="p-2.5 text-white">
                      {win.title}
                      <span className="block text-[10px] text-slate-400 font-mono">
                        {win.panelType}
                      </span>
                    </td>
                    <td className="p-2.5 text-slate-300">{win.monitorAssignment}</td>
                    <td className="p-2.5">
                      <span
                        className={`px-1.5 py-0.5 rounded text-[10px] font-bold ${
                          win.state === 'ACTIVE'
                            ? 'bg-[#00F0A0]/10 text-[#00F0A0]'
                            : 'bg-amber-400/10 text-amber-400'
                        }`}
                      >
                        {win.state}
                      </span>
                    </td>
                    <td className="p-2.5 text-[#00F0A0] font-mono">{win.lastPingMs}ms</td>
                    <td className="p-2.5 text-slate-300 font-mono">{win.syncPacketsCount}</td>
                    <td className="p-2.5 text-right">
                      <button
                        onClick={() => handleReDockWindow(win.id)}
                        className="px-2.5 py-1 rounded bg-[#181F2C] border border-slate-700 hover:border-rose-500 text-rose-300 hover:text-rose-200 text-[10px]"
                      >
                        Re-Dock to Main
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Sync Bus Packet Monitor */}
      <div className="p-3 rounded bg-[#121722] border border-slate-800 space-y-1.5">
        <div className="flex justify-between items-center text-[10px] text-slate-400">
          <span className="font-bold uppercase tracking-wider">
            Live Inter-Screen Bus Packet Telemetry (Seq #{sequenceCounter})
          </span>
          <span>Protocol: JSON RPC over BroadcastChannel</span>
        </div>
        <div className="space-y-1">
          {recentSyncPackets.map((pkt) => (
            <div
              key={pkt.sequenceId}
              className="flex items-center justify-between text-[10px] p-1.5 bg-[#0E121B] rounded border border-slate-800 font-mono"
            >
              <div className="flex items-center gap-2">
                <span className="text-[#00F0A0] font-bold">#{pkt.sequenceId}</span>
                <span className="text-indigo-300 font-bold">{pkt.type}</span>
                <span className="text-slate-400">{JSON.stringify(pkt.payload)}</span>
              </div>
              <span className="text-slate-500">{pkt.timestamp}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export const WebDetachedMultiWindowLayoutManagerAndStateSync = DetachedMultiWindowLayoutManager;
