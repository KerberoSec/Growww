// Domain Types for Growww Admin & Back-Office Workstation

export type RegulatorType = 'SEBI_SCORES' | 'RBI_CMS' | 'SMART_ODR';

export type GrievanceState =
  | 'INGESTED'
  | 'MAPPED_TO_USER'
  | 'DIAGNOSTIC_TRIAGE'
  | 'INVESTIGATION_PENDING'
  | 'AUDIT_DOSSIER_COLLECTED'
  | 'ATR_DRAFTED_MAKER'
  | 'ATR_APPROVED_CHECKER'
  | 'ATR_SUBMITTED_TO_REGULATOR'
  | 'FIRST_REVIEW_ESCALATED'
  | 'ODR_CONCILIATION'
  | 'RESOLVED_CLOSED'
  | 'REJECTED_INVALID';

export type ResolutionDisposition =
  | 'RESOLVED_SATISFIED'
  | 'RESOLVED_WITH_REFUND'
  | 'REJECTED_UNFOUNDED'
  | 'CLARIFICATION_PROVIDED'
  | 'SETTLED_VIA_ODR';

export type EscalationStage =
  | 0 // Normal (Day 0-6)
  | 1 // Warning (Day 7-13)
  | 2 // PCO Escalation (Day 14-17)
  | 3 // Critical P1 War Room (Day 18-19)
  | 4 // Submission Freeze (Day 20)
  | 5; // Statutory SLA Breach (Day 21+)

export interface ComplaintDocket {
  id: string;
  externalRegistrationNumber: string;
  regulator: RegulatorType;
  currentState: GrievanceState;
  complainantPANHash: string;
  complainantNameMasked: string;
  complainantEmailMasked: string;
  complainantMobileMasked: string;
  userId?: string;
  category: string;
  subCategory?: string;
  disputedAmount: number;
  complaintDescription: string;
  statutorySLADays: number;
  intakeTimestamp: string;
  statutoryDeadline: string;
  currentEscalationStage: EscalationStage;
  isSLABreached: boolean;
  assignedOfficerId?: string;
  atrId?: string;
  besuTxHash?: string;
  resolvedAt?: string;
}

export interface ActionTakenReport {
  id: string;
  complaintId: string;
  disposition: ResolutionDisposition;
  actionTakenSummary: string;
  detailedRedressalText: string;
  refundAmount?: number;
  settlementBankUTR?: string;
  supportingDocuments: string[];
  makerUserId: string;
  makerDraftedAt: string;
  checkerUserId?: string;
  checkerApprovedAt?: string;
  checkerNotes?: string;
  atrPackageHash?: string;
  isDSCSigned: boolean;
  dscSignerDN?: string;
  dscSignedAt?: string;
  regulatorySubmissionStatus: 'DRAFT' | 'APPROVED' | 'SUBMITTED' | 'ACKNOWLEDGED' | 'REJECTED';
  regulatoryAckNumber?: string;
  besuTxHash?: string;
}

// Surveillance Alert Types
export type SurveillanceCategory =
  | 'WASH_TRADING'
  | 'CIRCULAR_TRADING'
  | 'PUMP_AND_DUMP'
  | 'SPOOFING'
  | 'LAYERING'
  | 'EXCESSIVE_OTR'
  | 'MARKING_THE_CLOSE';

export type AlertSeverity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export interface CircularParticipant {
  participantId: string;
  orderCount: number;
  tradedVolume: number;
  ipSubnet: string;
}

export interface SurveillanceAlert {
  alertId: string;
  isin: string;
  symbol: string;
  category: SurveillanceCategory;
  severity: AlertSeverity;
  participantIds: string[];
  description: string;
  confidenceScore: number; // 0.0 to 1.0 (or 0-100%)
  volumeINR: number;
  detectedAt: string;
  status: 'OPEN' | 'UNDER_REVIEW' | 'ESCALATED_FIU' | 'DISMISSED';
  circularParticipants?: CircularParticipant[];
}

// Limit Up / Limit Down (LULD) & Circuit Breaker Types
export type SecurityLiquidityTier = 'TIER_1_LARGE_CAP' | 'TIER_2_MID_CAP' | 'TIER_3_SMALL_CAP';

export type MarketLULDState =
  | 'REGULAR'
  | 'STRADDLE'
  | 'LIMIT_STATE'
  | 'PAUSED_CALL_AUCTION'
  | 'HALTED'
  | 'RESUMING';

export interface PriceBandSnapshot {
  isin: string;
  symbol: string;
  tier: SecurityLiquidityTier;
  referencePricePaise: number;
  upperPriceBandPaise: number;
  lowerPriceBandPaise: number;
  currentPricePaise: number;
  bandPercentageBps: number; // e.g., 500 = 5.00%
  currentState: MarketLULDState;
  haltStartTimeUnix?: number;
  haltDurationSeconds?: number;
  updatedAt: string;
}

export interface ManualHaltRequest {
  isin: string;
  symbol: string;
  reason: string;
  durationMinutes: number;
  makerAdminId: string;
  checkerAdminId?: string;
}
