import React, { useState } from 'react';

export interface UserSession {
  id: string;
  isCurrent: boolean;
  deviceType: 'DESKTOP' | 'MOBILE' | 'API_GATEWAY';
  browser: string;
  os: string;
  ipAddress: string;
  location: string;
  countryCode: string;
  createdAt: string;
  lastActive: string;
  trustScore: number;
}

export function ActiveSessionsManager() {
  const [sessions, setSessions] = useState<UserSession[]>([
    {
      id: 'SESS-901A',
      isCurrent: true,
      deviceType: 'DESKTOP',
      browser: 'Chrome 128.0 (macOS)',
      os: 'macOS Sonoma 14.6',
      ipAddress: '49.207.214.18',
      location: 'Mumbai, Maharashtra, India',
      countryCode: 'IN',
      createdAt: '2026-09-20 07:15:00',
      lastActive: 'Active now',
      trustScore: 99,
    },
    {
      id: 'SESS-842C',
      isCurrent: false,
      deviceType: 'MOBILE',
      browser: 'Growww iOS App 4.12',
      os: 'iOS 18.0 (iPhone 16 Pro)',
      ipAddress: '157.48.112.54',
      location: 'Bengaluru, Karnataka, India',
      countryCode: 'IN',
      createdAt: '2026-09-19 14:22:11',
      lastActive: '45 mins ago',
      trustScore: 95,
    },
    {
      id: 'SESS-719E',
      isCurrent: false,
      deviceType: 'API_GATEWAY',
      browser: 'FIX 4.4 Engine / Golang client',
      os: 'Linux x86_64',
      ipAddress: '103.21.144.12',
      location: 'GIFT City, Gujarat, India',
      countryCode: 'IN',
      createdAt: '2026-09-18 00:01:00',
      lastActive: '2 mins ago',
      trustScore: 98,
    },
    {
      id: 'SESS-632B',
      isCurrent: false,
      deviceType: 'DESKTOP',
      browser: 'Firefox 129.0 (Windows)',
      os: 'Windows 11 Pro',
      ipAddress: '182.74.88.90',
      location: 'New Delhi, India',
      countryCode: 'IN',
      createdAt: '2026-09-15 10:10:00',
      lastActive: '3 days ago',
      trustScore: 82,
    },
  ]);

  const [revokingId, setRevokingId] = useState<string | null>(null);
  const [showRevokeAllConfirm, setShowRevokeAllConfirm] = useState<boolean>(false);

  const handleRevokeSession = (sessionId: string) => {
    setRevokingId(sessionId);
    setTimeout(() => {
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
      setRevokingId(null);
    }, 400);
  };

  const handleRevokeAllOthers = () => {
    setSessions((prev) => prev.filter((s) => s.isCurrent));
    setShowRevokeAllConfirm(false);
  };

  const otherSessionsCount = sessions.filter((s) => !s.isCurrent).length;

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-2">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00F0A0] inline-block animate-pulse" />
            Active Sessions & Remote Revocation Portal
          </h2>
          <p className="text-[11px] text-slate-400 mt-0.5">
            Real-time IP geolocation telemetry, device fingerprinting, and cryptographic session kill switch
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            disabled={otherSessionsCount === 0}
            onClick={() => setShowRevokeAllConfirm(true)}
            className={`px-3 py-1.5 rounded font-bold text-xs transition-colors ${
              otherSessionsCount > 0
                ? 'bg-[#FF3B56]/20 hover:bg-[#FF3B56]/30 text-[#FF3B56] border border-[#FF3B56]/40'
                : 'bg-slate-800 text-slate-600 cursor-not-allowed border border-transparent'
            }`}
          >
            Revoke All Other Sessions ({otherSessionsCount})
          </button>
        </div>
      </div>

      {/* Revoke All Confirmation Modal */}
      {showRevokeAllConfirm && (
        <div className="p-4 rounded-lg bg-[#181F2C] border border-[#FF3B56]/40 space-y-3">
          <div className="flex items-center justify-between">
            <span className="font-bold text-[#FF3B56] text-sm flex items-center gap-2">
              <span>⚠️</span> Revoke All {otherSessionsCount} Remote Sessions?
            </span>
            <button
              onClick={() => setShowRevokeAllConfirm(false)}
              className="text-slate-400 hover:text-white"
            >
              ✕
            </button>
          </div>
          <p className="text-[11px] text-slate-300">
            This will immediately invalidate all session tokens, disconnect active WebSocket streams, and require full 2FA re-authentication on all other devices except this browser.
          </p>
          <div className="flex justify-end gap-2 pt-1">
            <button
              onClick={() => setShowRevokeAllConfirm(false)}
              className="px-3 py-1.5 rounded bg-slate-800 text-slate-300 text-xs"
            >
              Cancel
            </button>
            <button
              onClick={handleRevokeAllOthers}
              className="px-4 py-1.5 rounded bg-[#FF3B56] hover:bg-[#E02E46] text-white font-bold text-xs"
            >
              Confirm Revocation
            </button>
          </div>
        </div>
      )}

      {/* Session List */}
      <div className="space-y-3">
        {sessions.map((sess) => (
          <div
            key={sess.id}
            className={`p-4 rounded-lg border transition-colors flex flex-col md:flex-row justify-between items-start md:items-center gap-3 ${
              sess.isCurrent
                ? 'bg-[#121721] border-[#00F0A0]/40 ring-1 ring-[#00F0A0]/20'
                : 'bg-[#121721] border-slate-800 hover:border-slate-700'
            }`}
          >
            <div className="space-y-1.5">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="font-bold text-white text-xs">{sess.browser}</span>
                {sess.isCurrent ? (
                  <span className="px-2 py-0.5 rounded bg-[#00F0A0]/20 text-[#00F0A0] text-[10px] font-bold border border-[#00F0A0]/30">
                    CURRENT DEVICE
                  </span>
                ) : (
                  <span className="px-2 py-0.5 rounded bg-slate-800 text-slate-400 text-[10px] font-medium">
                    {sess.deviceType}
                  </span>
                )}
                <span className="text-[10px] text-slate-500">ID: {sess.id}</span>
              </div>

              <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-slate-400">
                <span className="flex items-center gap-1">
                  <span>📍</span> {sess.location}
                </span>
                <span className="flex items-center gap-1">
                  <span>🌐</span> IP: <strong className="text-slate-300">{sess.ipAddress}</strong>
                </span>
                <span>OS: {sess.os}</span>
              </div>

              <div className="text-[10px] text-slate-500 flex gap-3">
                <span>First Login: {sess.createdAt}</span>
                <span>•</span>
                <span className={sess.isCurrent ? 'text-[#00F0A0]' : 'text-slate-400'}>
                  Last Activity: {sess.lastActive}
                </span>
              </div>
            </div>

            <div className="flex items-center gap-3 w-full md:w-auto justify-between md:justify-end border-t md:border-t-0 pt-2 md:pt-0 border-slate-800">
              <div className="text-right">
                <div className="text-[10px] text-slate-500 uppercase">Trust Score</div>
                <div className="text-xs font-bold text-[#00F0A0]">{sess.trustScore}%</div>
              </div>

              {!sess.isCurrent ? (
                <button
                  disabled={revokingId === sess.id}
                  onClick={() => handleRevokeSession(sess.id)}
                  className="px-3 py-1.5 rounded bg-slate-800 hover:bg-[#FF3B56]/20 hover:text-[#FF3B56] hover:border-[#FF3B56]/40 border border-slate-700 text-slate-300 text-xs font-semibold transition-colors"
                >
                  {revokingId === sess.id ? 'Terminating...' : 'Log Out Device'}
                </button>
              ) : (
                <span className="text-[10px] text-slate-500 italic px-2">This Browser</span>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Security Info Footer */}
      <div className="p-3 bg-[#181F2C]/60 rounded border border-slate-800 text-[11px] text-slate-400 flex items-center justify-between">
        <span>Session tokens expire automatically after 30 days of inactivity or upon password reset.</span>
        <span className="text-[#00F0A0] font-bold">TLS 1.3 Strict-Transport</span>
      </div>
    </div>
  );
}

export const WebActiveSessionsIpGeolocationRemoteLogout = ActiveSessionsManager;
