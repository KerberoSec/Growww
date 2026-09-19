import React, { useState } from 'react';
import { SurveillanceAlert, AlertSeverity, SurveillanceCategory } from '../types/admin';

interface SurveillanceAlertsProps {
  alerts: SurveillanceAlert[];
  onReviewAlert: (alertId: string, action: 'ESCALATE_FIU' | 'DISMISS' | 'INVESTIGATE') => void;
}

export const SurveillanceAlerts: React.FC<SurveillanceAlertsProps> = ({
  alerts,
  onReviewAlert,
}) => {
  const [selectedCategory, setSelectedCategory] = useState<SurveillanceCategory | 'ALL'>('ALL');
  const [selectedSeverity, setSelectedSeverity] = useState<AlertSeverity | 'ALL'>('ALL');

  const filteredAlerts = alerts.filter(a => {
    if (selectedCategory !== 'ALL' && a.category !== selectedCategory) return false;
    if (selectedSeverity !== 'ALL' && a.severity !== selectedSeverity) return false;
    return true;
  });

  const getSeverityBadge = (severity: AlertSeverity) => {
    switch (severity) {
      case 'CRITICAL':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-red-900/40 text-red-400 border border-red-800">CRITICAL</span>;
      case 'HIGH':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-orange-900/40 text-orange-400 border border-orange-800">HIGH</span>;
      case 'MEDIUM':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-yellow-900/40 text-yellow-400 border border-yellow-800">MEDIUM</span>;
      case 'LOW':
        return <span className="px-2 py-0.5 rounded text-xs font-bold bg-blue-900/40 text-blue-400 border border-blue-800">LOW</span>;
    }
  };

  return (
    <div className="bg-[#121212] text-white p-6 rounded-xl border border-gray-800">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Market Surveillance & Pattern Detection</h2>
          <p className="text-gray-400 text-sm mt-1">Real-time Wash Trading, Spoofing, Circular Networks & Layering Alarms</p>
        </div>
        <div className="flex gap-3">
          <select
            value={selectedCategory}
            onChange={e => setSelectedCategory(e.target.value as any)}
            className="bg-[#1A1A1A] border border-gray-700 text-sm rounded px-3 py-1.5 text-gray-300"
          >
            <option value="ALL">All Categories</option>
            <option value="WASH_TRADING">Wash Trading</option>
            <option value="CIRCULAR_TRADING">Circular Trading</option>
            <option value="SPOOFING">Spoofing</option>
            <option value="LAYERING">Layering</option>
            <option value="PUMP_AND_DUMP">Pump & Dump</option>
          </select>
          <select
            value={selectedSeverity}
            onChange={e => setSelectedSeverity(e.target.value as any)}
            className="bg-[#1A1A1A] border border-gray-700 text-sm rounded px-3 py-1.5 text-gray-300"
          >
            <option value="ALL">All Severities</option>
            <option value="CRITICAL">Critical</option>
            <option value="HIGH">High</option>
            <option value="MEDIUM">Medium</option>
            <option value="LOW">Low</option>
          </select>
        </div>
      </div>

      <div className="space-y-4">
        {filteredAlerts.length === 0 ? (
          <div className="text-center py-12 text-gray-500 bg-[#1A1A1A] rounded-lg border border-gray-800">
            No active surveillance alerts matching the current filters.
          </div>
        ) : (
          filteredAlerts.map(alert => (
            <div
              key={alert.alertId}
              className="bg-[#181818] border border-gray-800 rounded-lg p-4 hover:border-gray-700 transition"
            >
              <div className="flex justify-between items-start mb-2">
                <div className="flex items-center gap-3">
                  <span className="text-lg font-bold font-mono text-emerald-400">{alert.symbol}</span>
                  <span className="text-xs font-mono text-gray-400 bg-gray-800 px-2 py-0.5 rounded">{alert.category}</span>
                  {getSeverityBadge(alert.severity)}
                </div>
                <div className="text-right">
                  <span className="text-xs font-mono text-gray-400">{new Date(alert.detectedAt).toLocaleTimeString()}</span>
                </div>
              </div>

              <p className="text-sm text-gray-300 mb-3">{alert.description}</p>

              <div className="grid grid-cols-3 gap-4 text-xs font-mono bg-[#121212] p-3 rounded mb-3 border border-gray-850">
                <div>
                  <span className="text-gray-500 block">Confidence Score</span>
                  <span className="font-semibold text-emerald-400">{(alert.confidenceScore * 100).toFixed(1)}%</span>
                </div>
                <div>
                  <span className="text-gray-500 block">Traded Volume (INR)</span>
                  <span className="font-semibold text-white">₹{alert.volumeINR.toLocaleString('en-IN')}</span>
                </div>
                <div>
                  <span className="text-gray-500 block">Involved Accounts</span>
                  <span className="font-semibold text-white">{alert.participantIds.length} Nodes</span>
                </div>
              </div>

              {alert.circularParticipants && alert.circularParticipants.length > 0 && (
                <div className="mb-3">
                  <span className="text-xs font-semibold text-gray-400 block mb-1">Circular Network Participants</span>
                  <div className="flex flex-wrap gap-2">
                    {alert.circularParticipants.map(cp => (
                      <span key={cp.participantId} className="bg-gray-800 border border-gray-700 text-gray-300 text-xs px-2 py-1 rounded font-mono">
                        {cp.participantId} (Vol: ₹{cp.tradedVolume.toLocaleString('en-IN')}, IP: {cp.ipSubnet})
                      </span>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex justify-end gap-2 pt-2 border-t border-gray-800">
                <button
                  onClick={() => onReviewAlert(alert.alertId, 'DISMISS')}
                  className="px-3 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs rounded font-medium"
                >
                  Dismiss / False Positive
                </button>
                <button
                  onClick={() => onReviewAlert(alert.alertId, 'INVESTIGATE')}
                  className="px-3 py-1 bg-blue-700 hover:bg-blue-600 text-white text-xs rounded font-medium"
                >
                  Open Investigation Dossier
                </button>
                <button
                  onClick={() => onReviewAlert(alert.alertId, 'ESCALATE_FIU')}
                  className="px-3 py-1 bg-red-600 hover:bg-red-500 text-white text-xs rounded font-bold"
                >
                  Escalate to FIU-IND STR
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
