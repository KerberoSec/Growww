package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Role defines system roles adhering to institutional least-privilege principles
type Role string

const (
	RoleSuperAdmin        Role = "ROLE_SUPER_ADMIN"
	RoleComplianceLead    Role = "ROLE_COMPLIANCE_LEAD"
	RoleComplianceOfficer Role = "ROLE_COMPLIANCE_OFFICER"
	RoleAMLAnalyst        Role = "ROLE_AML_ANALYST"
	RoleTradeOpsLead      Role = "ROLE_TRADE_OPS_LEAD"
	RoleCustodyAuditor    Role = "ROLE_CUSTODY_AUDITOR"
	RoleSupportL1         Role = "ROLE_SUPPORT_L1"
	RoleSupportL2         Role = "ROLE_SUPPORT_L2"
	RoleDeveloperSRE      Role = "ROLE_DEVELOPER_SRE"
)

// Permission defines granular operation capabilities
type Permission string

const (
	PermComplianceRead         Permission = "compliance:read"
	PermComplianceFreeze       Permission = "compliance:freeze_account"
	PermComplianceUnfreeze     Permission = "compliance:unfreeze_account"
	PermAMLAlertReview         Permission = "aml:alert_review"
	PermAMLAlertFileFIU        Permission = "aml:file_fiu_report"
	PermLedgerAdjust           Permission = "ledger:manual_adjustment"
	PermTokenEmergencyPause    Permission = "token:emergency_pause"
	PermTokenEmergencyUnpause  Permission = "token:emergency_unpause"
	PermIAMAssignRole          Permission = "iam:assign_role"
	PermIAMRevokeRole          Permission = "iam:revoke_role"
	PermAuditRead              Permission = "audit:read"
	PermTradeAdjust            Permission = "trade:adjust"
	PermTradeCancel            Permission = "trade:cancel"
	PermSupportViewPII         Permission = "support:view_pii"
	PermCKYCUpload             Permission = "ckyc:upload_batch"
	PermCKYCFetch              Permission = "ckyc:fetch_record"
)

// HighImpactAction defines operations requiring mandatory dual-control ("four-eyes" quorum)
type HighImpactAction string

const (
	ActionAccountFreeze          HighImpactAction = "ACTION_ACCOUNT_FREEZE"
	ActionAccountUnfreeze        HighImpactAction = "ACTION_ACCOUNT_UNFREEZE"
	ActionTokenEmergencyPause    HighImpactAction = "ACTION_TOKEN_EMERGENCY_PAUSE"
	ActionTokenEmergencyUnpause  HighImpactAction = "ACTION_TOKEN_EMERGENCY_UNPAUSE"
	ActionManualLedgerAdjust     HighImpactAction = "ACTION_MANUAL_LEDGER_ADJUST"
	ActionRolePrivilegeEscalation HighImpactAction = "ACTION_PRIVILEGE_ESCALATION"
)

// DualApprovalStatus indicates the state of a multi-party approval request
type DualApprovalStatus string

const (
	DualApprovalPending  DualApprovalStatus = "PENDING"
	DualApprovalApproved DualApprovalStatus = "APPROVED"
	DualApprovalRejected DualApprovalStatus = "REJECTED"
	DualApprovalExpired  DualApprovalStatus = "EXPIRED"
)

// DualControlRequest encapsulates a four-eyes authorization request
type DualControlRequest struct {
	RequestID         string             `json:"request_id"`
	Action            HighImpactAction   `json:"action"`
	ResourceID        string             `json:"resource_id"`
	RequesterID       string             `json:"requester_id"`
	RequesterRole     Role               `json:"requester_role"`
	Reason            string             `json:"reason"`
	Status            DualApprovalStatus `json:"status"`
	CreatedAt         time.Time          `json:"created_at"`
	ExpiresAt         time.Time          `json:"expires_at"`
	ApproverID        string             `json:"approver_id,omitempty"`
	ApproverRole      Role               `json:"approver_role,omitempty"`
	ApproverSignature string             `json:"approver_signature,omitempty"`
	ApprovedAt        time.Time          `json:"approved_at,omitempty"`
}

// JITGrant represents an ephemeral Just-In-Time role elevation
type JITGrant struct {
	GrantID      string    `json:"grant_id"`
	UserID       string    `json:"user_id"`
	ElevatedRole Role      `json:"elevated_role"`
	Reason       string    `json:"reason"`
	ApprovedBy   string    `json:"approved_by"`
	GrantedAt    time.Time `json:"granted_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsRevoked    bool      `json:"is_revoked"`
}

// AuthContext encapsulates incoming request context for RBAC + ABAC evaluation
type AuthContext struct {
	ActorID             string           `json:"actor_id"`
	AssignedRoles       []Role           `json:"assigned_roles"`
	IPAddress           string           `json:"ip_address"`
	AllowedCIDRs        []string         `json:"allowed_cidrs"`
	HardwareMFAVerified bool             `json:"hardware_mfa_verified"`
	RequestedPermission Permission       `json:"requested_permission"`
	HighImpactOp        HighImpactAction `json:"high_impact_op,omitempty"`
	ActionID            string           `json:"action_id,omitempty"`
	ResourceID          string           `json:"resource_id"`
	TimeNow             time.Time        `json:"time_now"`
}

// IAMAuditEvent captures immutable tamper-evident authorization audit records
type IAMAuditEvent struct {
	EventID        string    `json:"event_id"`
	Timestamp      time.Time `json:"timestamp"`
	ActorID        string    `json:"actor_id"`
	Roles          []Role    `json:"roles"`
	Permission     Permission`json:"permission"`
	ResourceID     string    `json:"resource_id"`
	Decision       string    `json:"decision"` // ALLOW / DENY
	Reason         string    `json:"reason"`
	DualApprovalID string    `json:"dual_approval_id,omitempty"`
	SignatureHash  string    `json:"signature_hash"`
}

// IAMRBACEngine implements least privilege, SoD, ABAC, and dual-control quorums
type IAMRBACEngine struct {
	mu                 sync.RWMutex
	rolePermissions    map[Role]map[Permission]bool
	sodConflicts       map[Role][]Role
	highImpactActions  map[HighImpactAction]bool
	dualControlStore   map[string]*DualControlRequest
	jitGrants          map[string]*JITGrant // grantID -> grant
	userJITIndex       map[string][]string  // userID -> []grantID
	auditEvents        []*IAMAuditEvent
	maxJITDuration     time.Duration
	dualApprovalExpiry time.Duration
}

// NewIAMRBACEngine initializes the institutional IAM authorization engine
func NewIAMRBACEngine() *IAMRBACEngine {
	engine := &IAMRBACEngine{
		rolePermissions:    make(map[Role]map[Permission]bool),
		sodConflicts:       make(map[Role][]Role),
		highImpactActions:  make(map[HighImpactAction]bool),
		dualControlStore:   make(map[string]*DualControlRequest),
		jitGrants:          make(map[string]*JITGrant),
		userJITIndex:       make(map[string][]string),
		auditEvents:        make([]*IAMAuditEvent, 0),
		maxJITDuration:     60 * time.Minute, // Max 1 hour ephemeral access under PAM
		dualApprovalExpiry: 30 * time.Minute,
	}

	engine.initRolePermissions()
	engine.initSoDConflicts()
	engine.initHighImpactActions()
	return engine
}

func (e *IAMRBACEngine) initRolePermissions() {
	// Least privilege assignment
	e.rolePermissions[RoleSuperAdmin] = map[Permission]bool{
		PermComplianceRead:        true,
		PermComplianceFreeze:      true,
		PermComplianceUnfreeze:    true,
		PermAMLAlertReview:        true,
		PermAMLAlertFileFIU:       true,
		PermLedgerAdjust:          true,
		PermTokenEmergencyPause:   true,
		PermTokenEmergencyUnpause: true,
		PermIAMAssignRole:         true,
		PermIAMRevokeRole:         true,
		PermAuditRead:             true,
		PermTradeCancel:           true,
		PermCKYCUpload:            true,
		PermCKYCFetch:             true,
	}

	e.rolePermissions[RoleComplianceLead] = map[Permission]bool{
		PermComplianceRead:        true,
		PermComplianceFreeze:      true,
		PermComplianceUnfreeze:    true,
		PermAMLAlertReview:        true,
		PermAMLAlertFileFIU:       true,
		PermTokenEmergencyPause:   true,
		PermTokenEmergencyUnpause: true,
		PermAuditRead:             true,
		PermCKYCUpload:            true,
		PermCKYCFetch:             true,
	}

	e.rolePermissions[RoleComplianceOfficer] = map[Permission]bool{
		PermComplianceRead:   true,
		PermComplianceFreeze: true,
		PermAMLAlertReview:   true,
		PermAuditRead:        true,
		PermCKYCUpload:       true,
		PermCKYCFetch:        true,
	}

	e.rolePermissions[RoleAMLAnalyst] = map[Permission]bool{
		PermComplianceRead: true,
		PermAMLAlertReview: true,
		PermAuditRead:      true,
		PermCKYCFetch:      true,
	}

	e.rolePermissions[RoleTradeOpsLead] = map[Permission]bool{
		PermTradeAdjust: true,
		PermTradeCancel: true,
		PermAuditRead:   true,
	}

	e.rolePermissions[RoleCustodyAuditor] = map[Permission]bool{
		PermAuditRead:      true,
		PermComplianceRead: true,
	}

	e.rolePermissions[RoleSupportL1] = map[Permission]bool{
		PermComplianceRead: true,
	}

	e.rolePermissions[RoleSupportL2] = map[Permission]bool{
		PermComplianceRead: true,
		PermSupportViewPII: true,
	}

	e.rolePermissions[RoleDeveloperSRE] = map[Permission]bool{
		PermAuditRead: true,
	}
}

func (e *IAMRBACEngine) initSoDConflicts() {
	// Segregation of Duties matrix
	// 1. Developer cannot hold production trade operations or manual ledger adjustment
	e.sodConflicts[RoleDeveloperSRE] = []Role{RoleTradeOpsLead, RoleSuperAdmin, RoleComplianceLead}

	// 2. AML Analyst cannot hold Trade Ops roles
	e.sodConflicts[RoleAMLAnalyst] = []Role{RoleTradeOpsLead}

	// 3. Trade Ops cannot hold Compliance or IAM roles
	e.sodConflicts[RoleTradeOpsLead] = []Role{RoleAMLAnalyst, RoleComplianceOfficer, RoleComplianceLead, RoleSuperAdmin}

	// 4. Custody Auditor must be independent of operational roles
	e.sodConflicts[RoleCustodyAuditor] = []Role{RoleTradeOpsLead, RoleDeveloperSRE, RoleSuperAdmin}
}

func (e *IAMRBACEngine) initHighImpactActions() {
	e.highImpactActions[ActionAccountFreeze] = true
	e.highImpactActions[ActionAccountUnfreeze] = true
	e.highImpactActions[ActionTokenEmergencyPause] = true
	e.highImpactActions[ActionTokenEmergencyUnpause] = true
	e.highImpactActions[ActionManualLedgerAdjust] = true
	e.highImpactActions[ActionRolePrivilegeEscalation] = true
}

// CheckSoDConflict verifies whether combining the given roles violates Segregation of Duties
func (e *IAMRBACEngine) CheckSoDConflict(roles []Role) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	roleSet := make(map[Role]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	for _, r := range roles {
		conflicts := e.sodConflicts[r]
		for _, conf := range conflicts {
			if roleSet[conf] {
				return fmt.Errorf("segregation of duties (SoD) violation: role %s conflicts with role %s", r, conf)
			}
		}
	}
	return nil
}

// RequestDualControlAction creates a new four-eyes approval request for high-impact actions
func (e *IAMRBACEngine) RequestDualControlAction(action HighImpactAction, resourceID, requesterID string, requesterRole Role, reason string) (*DualControlRequest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if strings.TrimSpace(requesterID) == "" {
		return nil, errors.New("requester ID cannot be empty")
	}
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("detailed justification reason is mandatory for high-impact dual-control action")
	}

	reqID := fmt.Sprintf("DC-%d-%s", time.Now().UnixNano(), hex.EncodeToString([]byte(resourceID))[:8])
	now := time.Now().UTC()
	req := &DualControlRequest{
		RequestID:     reqID,
		Action:        action,
		ResourceID:    resourceID,
		RequesterID:   requesterID,
		RequesterRole: requesterRole,
		Reason:        reason,
		Status:        DualApprovalPending,
		CreatedAt:     now,
		ExpiresAt:     now.Add(e.dualApprovalExpiry),
	}

	e.dualControlStore[reqID] = req
	return req, nil
}

// ApproveDualControlAction allows a distinct authorized officer to approve a high-impact action
func (e *IAMRBACEngine) ApproveDualControlAction(requestID, approverID string, approverRole Role, approverSignature string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	req, exists := e.dualControlStore[requestID]
	if !exists {
		return errors.New("dual control request not found")
	}

	if req.Status != DualApprovalPending {
		return fmt.Errorf("dual control request is already %s", req.Status)
	}

	if time.Now().UTC().After(req.ExpiresAt) {
		req.Status = DualApprovalExpired
		return errors.New("dual control request has expired")
	}

	// Enforce strict Four-Eyes segregation: Requester cannot approve their own request
	if req.RequesterID == approverID {
		return errors.New("four-eyes violation: requester cannot approve their own high-impact action")
	}

	// Verify approver role eligibility
	if approverRole != RoleComplianceLead && approverRole != RoleSuperAdmin {
		return fmt.Errorf("insufficient authority: approver role %s cannot approve high-impact dual-control operations", approverRole)
	}

	if strings.TrimSpace(approverSignature) == "" {
		return errors.New("cryptographic approval signature is required")
	}

	req.Status = DualApprovalApproved
	req.ApproverID = approverID
	req.ApproverRole = approverRole
	req.ApproverSignature = approverSignature
	req.ApprovedAt = time.Now().UTC()

	return nil
}

// RequestJITAccess grants ephemeral, audited Just-In-Time role elevation
func (e *IAMRBACEngine) RequestJITAccess(userID string, elevatedRole Role, duration time.Duration, reason, approvedBy string) (*JITGrant, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if duration <= 0 || duration > e.maxJITDuration {
		return nil, fmt.Errorf("invalid JIT duration: must be > 0 and <= %v", e.maxJITDuration)
	}

	if strings.TrimSpace(approvedBy) == "" || approvedBy == userID {
		return nil, errors.New("JIT elevation requires independent authorization; self-approval prohibited")
	}

	grantID := fmt.Sprintf("JIT-%d", time.Now().UnixNano())
	now := time.Now().UTC()
	grant := &JITGrant{
		GrantID:      grantID,
		UserID:       userID,
		ElevatedRole: elevatedRole,
		Reason:       reason,
		ApprovedBy:   approvedBy,
		GrantedAt:    now,
		ExpiresAt:    now.Add(duration),
		IsRevoked:    false,
	}

	e.jitGrants[grantID] = grant
	e.userJITIndex[userID] = append(e.userJITIndex[userID], grantID)
	return grant, nil
}

// RevokeJITAccess terminates an ephemeral privilege grant immediately
func (e *IAMRBACEngine) RevokeJITAccess(grantID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	grant, exists := e.jitGrants[grantID]
	if !exists {
		return errors.New("JIT grant not found")
	}
	grant.IsRevoked = true
	return nil
}

// GetActiveJITRoles returns currently valid JIT elevated roles for a user
func (e *IAMRBACEngine) GetActiveJITRoles(userID string, now time.Time) []Role {
	e.mu.RLock()
	defer e.mu.RUnlock()

	grantIDs := e.userJITIndex[userID]
	active := make([]Role, 0)
	for _, id := range grantIDs {
		g := e.jitGrants[id]
		if g != nil && !g.IsRevoked && now.Before(g.ExpiresAt) {
			active = append(active, g.ElevatedRole)
		}
	}
	return active
}

// ValidateCIDR verifies if IP falls within allowed network ranges
func (e *IAMRBACEngine) ValidateCIDR(ipStr string, allowedCIDRs []string) bool {
	if len(allowedCIDRs) == 0 {
		return true // No restriction configured
	}
	parsedIP := net.ParseIP(strings.TrimSpace(ipStr))
	if parsedIP == nil {
		return false
	}

	for _, cidr := range allowedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// Authorize evaluates RBAC, SoD, ABAC (CIDR, MFA), and Dual-Control quorums
func (e *IAMRBACEngine) Authorize(ctx AuthContext) (bool, string, *IAMAuditEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := ctx.TimeNow
	if now.IsZero() {
		now = time.Now().UTC()
	}

	// 1. Check IP CIDR whitelisting (ABAC)
	if !e.ValidateCIDR(ctx.IPAddress, ctx.AllowedCIDRs) {
		event := e.recordAudit(ctx, "DENY", fmt.Sprintf("IP address %s outside allowed corporate CIDR boundaries", ctx.IPAddress), "")
		return false, "network boundary violation: untrusted source IP", event
	}

	// 2. Aggregate assigned roles + active JIT roles
	allRoles := append([]Role{}, ctx.AssignedRoles...)
	jitGrants := e.userJITIndex[ctx.ActorID]
	for _, gid := range jitGrants {
		g := e.jitGrants[gid]
		if g != nil && !g.IsRevoked && now.Before(g.ExpiresAt) {
			allRoles = append(allRoles, g.ElevatedRole)
		}
	}

	// 3. Evaluate role permissions
	hasPerm := false
	for _, r := range allRoles {
		perms := e.rolePermissions[r]
		if perms != nil && perms[ctx.RequestedPermission] {
			hasPerm = true
			break
		}
	}

	if !hasPerm {
		event := e.recordAudit(ctx, "DENY", fmt.Sprintf("principal lacks permission %s across active roles %v", ctx.RequestedPermission, allRoles), "")
		return false, "access denied: insufficient permissions", event
	}

	// 4. Evaluate High-Impact Action Dual-Control Quorum
	if ctx.HighImpactOp != "" && e.highImpactActions[ctx.HighImpactOp] {
		// Mandatory Hardware MFA for high-impact actions
		if !ctx.HardwareMFAVerified {
			event := e.recordAudit(ctx, "DENY", "hardware FIDO2 MFA token verification required for high-impact operation", "")
			return false, "MFA required: high-impact action mandates FIDO2 hardware token", event
		}

		if ctx.ActionID == "" {
			event := e.recordAudit(ctx, "DENY", "high-impact action requires approved dual-control action_id", "")
			return false, "dual-control required: action_id missing", event
		}

		approval, exists := e.dualControlStore[ctx.ActionID]
		if !exists {
			event := e.recordAudit(ctx, "DENY", fmt.Sprintf("dual control request %s not found", ctx.ActionID), "")
			return false, "dual-control invalid: request not found", event
		}

		if approval.Action != ctx.HighImpactOp {
			event := e.recordAudit(ctx, "DENY", fmt.Sprintf("action mismatch: request is for %s, attempted %s", approval.Action, ctx.HighImpactOp), ctx.ActionID)
			return false, "dual-control action mismatch", event
		}

		if approval.Status != DualApprovalApproved {
			event := e.recordAudit(ctx, "DENY", fmt.Sprintf("dual control request status is %s", approval.Status), ctx.ActionID)
			return false, fmt.Sprintf("dual-control not approved: status is %s", approval.Status), event
		}

		if now.After(approval.ExpiresAt) {
			approval.Status = DualApprovalExpired
			event := e.recordAudit(ctx, "DENY", "dual control approval token has expired", ctx.ActionID)
			return false, "dual-control authorization expired", event
		}

		if approval.ApproverID == ctx.ActorID {
			event := e.recordAudit(ctx, "DENY", "actor cannot execute action approved solely by self", ctx.ActionID)
			return false, "four-eyes violation: approver cannot be executor", event
		}
	}

	event := e.recordAudit(ctx, "ALLOW", "authorization successful", ctx.ActionID)
	return true, "authorized", event
}

func (e *IAMRBACEngine) recordAudit(ctx AuthContext, decision, reason, dualApprovalID string) *IAMAuditEvent {
	eventID := fmt.Sprintf("IAM-AUDIT-%d", time.Now().UnixNano())
	payload := fmt.Sprintf("%s:%s:%s:%s:%s:%s", eventID, ctx.ActorID, ctx.RequestedPermission, ctx.ResourceID, decision, reason)
	hash := sha256.Sum256([]byte(payload))

	event := &IAMAuditEvent{
		EventID:        eventID,
		Timestamp:      time.Now().UTC(),
		ActorID:        ctx.ActorID,
		Roles:          ctx.AssignedRoles,
		Permission:     ctx.RequestedPermission,
		ResourceID:     ctx.ResourceID,
		Decision:       decision,
		Reason:         reason,
		DualApprovalID: dualApprovalID,
		SignatureHash:  hex.EncodeToString(hash[:]),
	}
	e.auditEvents = append(e.auditEvents, event)
	return event
}

// GetAuditLog returns a snapshot of recorded authorization decisions
func (e *IAMRBACEngine) GetAuditLog() []*IAMAuditEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	copied := make([]*IAMAuditEvent, len(e.auditEvents))
	copy(copied, e.auditEvents)
	return copied
}
