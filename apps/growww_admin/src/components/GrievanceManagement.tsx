import React, { useState } from 'react';
import { ComplaintDocket, ActionTakenReport, ResolutionDisposition, EscalationStage } from '../types/admin';

interface GrievanceManagementProps {
  dockets: ComplaintDocket[];
  onSubmitATR: (complaintId: string, atr: Partial<ActionTakenReport>) => void;
  onEscalate: (complaintId: string, targetStage: EscalationStage) => void;
}

export const GrievanceManagement: React.FC<GrievanceManagementProps> = ({
  dockets,
  onSubmitATR,
  onEscalate,
}) => {
  const [selectedDocket, setSelectedDocket] = useState<ComplaintDocket | null>(null);
  const [disposition, setDisposition] = useState<ResolutionDisposition>('RESOLVED_SATISFIED');
  const [summary, setSummary] = useState('');
  const [refundAmount, setRefundAmount] = useState<number>(0);
  const [filterStage, setFilterStage] = useState<number | 'ALL'>('ALL');

  const filteredDockets = filterStage === 'ALL'
    ? dockets
    : dockets.filter(d => d.currentEscalationStage === filterStage);

  const getSLAColor = (stage: EscalationStage) => {
    switch (stage) {
      case 0: return '#10B981'; // Green (Day 0-6)
      case 1: return '#F59E0B'; // Yellow (Day 7-13)
      case 2: return '#F97316'; // Orange (Day 14-17)
      case 3: return '#EF4444'; // Red P1 War Room (Day 18-19)
      case 4: return '#991B1B'; // Dark Red (Day 20)
      case 5: return '#7F1D1D'; // Breach (Day 21+)
      default: return '#6B7280';
    }
  };

  const handleSubmitATR = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDocket) return;

    onSubmitATR(selectedDocket.id, {
      complaintId: selectedDocket.id,
      disposition,
      actionTakenSummary: summary,
      refundAmount: refundAmount > 0 ? refundAmount : undefined,
      makerDraftedAt: new Date().toISOString(),
      regulatorySubmissionStatus: 'DRAFT',
    });

    setSelectedDocket(null);
    setSummary('');
    setRefundAmount(0);
  };

  return (
    <div className="bg-[#121212] text-white p-6 rounded-xl border border-gray-800">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">SEBI SCORES 2.0 & RBI Grievance Gateway</h2>
          <p className="text-gray-400 text-sm mt-1">21-Day Statutory SLA Escalation Matrix & ATR Resolution Desk</p>
        </div>
        <div className="flex gap-2">
          {([ 'ALL', 0, 1, 2, 3, 4, 5 ] as const).map(stage => (
            <button
              key={stage}
              onClick={() => setFilterStage(stage)}
              className={`px-3 py-1.5 rounded text-xs font-semibold transition ${
                filterStage === stage ? 'bg-emerald-600 text-white' : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
              }`}
            >
              {stage === 'ALL' ? 'All Dockets' : `Stage ${stage}`}
            </button>
          ))}
        </div>
      </div>

      {/* Dockets Table */}
      <div className="overflow-x-auto rounded-lg border border-gray-800">
        <table className="w-full text-left text-sm">
          <thead className="bg-[#1A1A1A] text-gray-400 font-mono text-xs uppercase">
            <tr>
              <th className="p-3">Reg. No</th>
              <th className="p-3">Regulator</th>
              <th className="p-3">Complainant</th>
              <th className="p-3">Category</th>
              <th className="p-3">Disputed Amount</th>
              <th className="p-3">SLA Stage</th>
              <th className="p-3">Status</th>
              <th className="p-3">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {filteredDockets.map(docket => (
              <tr key={docket.id} className="hover:bg-gray-850 transition">
                <td className="p-3 font-mono font-medium text-emerald-400">{docket.externalRegistrationNumber}</td>
                <td className="p-3">
                  <span className="bg-blue-900/40 text-blue-300 px-2 py-0.5 rounded text-xs border border-blue-800">
                    {docket.regulator}
                  </span>
                </td>
                <td className="p-3 font-mono">{docket.complainantNameMasked}</td>
                <td className="p-3 text-gray-300">{docket.category}</td>
                <td className="p-3 font-mono font-semibold">₹{docket.disputedAmount.toLocaleString('en-IN')}</td>
                <td className="p-3">
                  <span
                    className="px-2.5 py-1 rounded text-xs font-bold"
                    style={{ backgroundColor: `${getSLAColor(docket.currentEscalationStage)}25`, color: getSLAColor(docket.currentEscalationStage) }}
                  >
                    Stage {docket.currentEscalationStage} {docket.isSLABreached ? '(BREACHED)' : ''}
                  </span>
                </td>
                <td className="p-3">
                  <span className="text-xs bg-gray-800 text-gray-300 px-2 py-0.5 rounded">
                    {docket.currentState}
                  </span>
                </td>
                <td className="p-3">
                  <button
                    onClick={() => setSelectedDocket(docket)}
                    className="bg-emerald-600 hover:bg-emerald-500 text-white px-3 py-1 rounded text-xs font-semibold"
                  >
                    Draft ATR
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* ATR Modal */}
      {selectedDocket && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50">
          <div className="bg-[#1A1A1A] border border-gray-700 rounded-xl max-w-2xl w-full p-6 text-white shadow-2xl">
            <h3 className="text-xl font-bold mb-2">Draft Action Taken Report (ATR)</h3>
            <p className="text-gray-400 text-sm mb-4">Docket #{selectedDocket.externalRegistrationNumber} | Complainant: {selectedDocket.complainantNameMasked}</p>

            <form onSubmit={handleSubmitATR} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-300 mb-1">Resolution Disposition</label>
                <select
                  value={disposition}
                  onChange={e => setDisposition(e.target.value as ResolutionDisposition)}
                  className="w-full bg-[#121212] border border-gray-700 rounded p-2 text-sm text-white"
                >
                  <option value="RESOLVED_SATISFIED">RESOLVED_SATISFIED (Issue Fixed)</option>
                  <option value="RESOLVED_WITH_REFUND">RESOLVED_WITH_REFUND (Full/Partial Refund)</option>
                  <option value="REJECTED_UNFOUNDED">REJECTED_UNFOUNDED (No Exchange Fault)</option>
                  <option value="CLARIFICATION_PROVIDED">CLARIFICATION_PROVIDED</option>
                  <option value="SETTLED_VIA_ODR">SETTLED_VIA_ODR (SMART ODR Conciliation)</option>
                </select>
              </div>

              {disposition === 'RESOLVED_WITH_REFUND' && (
                <div>
                  <label className="block text-xs font-medium text-gray-300 mb-1">Refund Amount (₹)</label>
                  <input
                    type="number"
                    value={refundAmount}
                    onChange={e => setRefundAmount(Number(e.target.value))}
                    className="w-full bg-[#121212] border border-gray-700 rounded p-2 text-sm text-white font-mono"
                  />
                </div>
              )}

              <div>
                <label className="block text-xs font-medium text-gray-300 mb-1">Action Taken Redressal Summary (SEBI Form Format)</label>
                <textarea
                  rows={4}
                  value={summary}
                  onChange={e => setSummary(e.target.value)}
                  placeholder="Provide precise root cause analysis, corrective actions taken, and transaction reference logs..."
                  className="w-full bg-[#121212] border border-gray-700 rounded p-2 text-sm text-white font-sans"
                  required
                />
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setSelectedDocket(null)}
                  className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded text-sm"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-sm font-semibold"
                >
                  Submit ATR Package
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
