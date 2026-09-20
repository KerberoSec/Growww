package main

import (
	"strings"
	"testing"
	"time"
)

func TestIAMRBAC_RolePermissions(t *testing.T) {
	engine := NewIAMRBACEngine()

	// Compliance Officer should have PermComplianceRead and PermComplianceFreeze, but NOT PermTradeAdjust
	authCtx := AuthContext{
		ActorID:             "OFFICER-01",
		AssignedRoles:       []Role{RoleComplianceOfficer},
		IPAddress:           "10.0.1.50",
		AllowedCIDRs:        []string{"10.0.0.0/16"},
		HardwareMFAVerified: true,
		RequestedPermission: PermComplianceRead,
		ResourceID:          "INVESTOR-100",
		TimeNow:             time.Now().UTC(),
	}

	allowed, reason, audit := engine.Authorize(authCtx)
	if !allowed {
		t.Fatalf("expected PermComplianceRead to be allowed for Compliance Officer, got reason: %s", reason)
	}
	if audit.Decision != "ALLOW" {
		t.Errorf("expected audit decision ALLOW, got %s", audit.Decision)
	}

	// Compliance Officer trying Trade Adjust must be denied
	authCtx.RequestedPermission = PermTradeAdjust
	allowed, reason, _ = engine.Authorize(authCtx)
	if allowed {
		t.Fatalf("expected PermTradeAdjust to be denied for Compliance Officer")
	}
	if !strings.Contains(reason, "insufficient permissions") {
		t.Errorf("unexpected denial reason: %s", reason)
	}

	// Developer SRE trying to freeze compliance account must be denied
	devCtx := AuthContext{
		ActorID:             "DEV-01",
		AssignedRoles:       []Role{RoleDeveloperSRE},
		RequestedPermission: PermComplianceFreeze,
		ResourceID:          "INVESTOR-200",
		TimeNow:             time.Now().UTC(),
	}
	allowed, _, _ = engine.Authorize(devCtx)
	if allowed {
		t.Fatalf("expected Developer to be denied PermComplianceFreeze")
	}
}

func TestIAMRBAC_SegregationOfDuties(t *testing.T) {
	engine := NewIAMRBACEngine()

	// Developer + Trade Ops is SoD violation
	err := engine.CheckSoDConflict([]Role{RoleDeveloperSRE, RoleTradeOpsLead})
	if err == nil {
		t.Fatal("expected SoD violation for Developer + TradeOpsLead")
	}

	// AML Analyst + Trade Ops is SoD violation
	err = engine.CheckSoDConflict([]Role{RoleAMLAnalyst, RoleTradeOpsLead})
	if err == nil {
		t.Fatal("expected SoD violation for AMLAnalyst + TradeOpsLead")
	}

	// Compliance Officer + AML Analyst is allowed
	err = engine.CheckSoDConflict([]Role{RoleComplianceOfficer, RoleAMLAnalyst})
	if err != nil {
		t.Fatalf("unexpected SoD conflict for ComplianceOfficer + AMLAnalyst: %v", err)
	}
}

func TestIAMRBAC_DualControlFourEyes(t *testing.T) {
	engine := NewIAMRBACEngine()

	requesterID := "OFFICER-ALICE"
	approverID := "LEAD-BOB"
	investorUUID := "INV-999-FREEZE"

	// 1. Alice requests account freeze
	req, err := engine.RequestDualControlAction(
		ActionAccountFreeze,
		investorUUID,
		requesterID,
		RoleComplianceOfficer,
		"Hit on MHA UAPA watchlist, freezing investor account per Section 12 PMLA",
	)
	if err != nil {
		t.Fatalf("failed to create dual control request: %v", err)
	}
	if req.Status != DualApprovalPending {
		t.Fatalf("expected status PENDING, got %s", req.Status)
	}

	// 2. Alice attempts to approve her own request -> must fail (SoD self-approval check)
	err = engine.ApproveDualControlAction(req.RequestID, requesterID, RoleComplianceLead, "sig-alice")
	if err == nil || !strings.Contains(err.Error(), "four-eyes violation") {
		t.Fatalf("expected four-eyes self-approval rejection, got: %v", err)
	}

	// 3. Unauthorized role (e.g. AML Analyst) attempts to approve -> must fail
	err = engine.ApproveDualControlAction(req.RequestID, "ANALYST-CHARLIE", RoleAMLAnalyst, "sig-charlie")
	if err == nil || !strings.Contains(err.Error(), "insufficient authority") {
		t.Fatalf("expected rejection for unauthorized approver role, got: %v", err)
	}

	// 4. Bob (Compliance Lead) approves with valid signature
	err = engine.ApproveDualControlAction(req.RequestID, approverID, RoleComplianceLead, "RSA_SIGNATURE_BOB_LEAD_OK")
	if err != nil {
		t.Fatalf("lead approval failed: %v", err)
	}

	// 5. Execution attempt without hardware MFA -> must fail
	authCtx := AuthContext{
		ActorID:             requesterID,
		AssignedRoles:       []Role{RoleComplianceOfficer},
		HardwareMFAVerified: false, // Missing MFA
		RequestedPermission: PermComplianceFreeze,
		HighImpactOp:        ActionAccountFreeze,
		ActionID:            req.RequestID,
		ResourceID:          investorUUID,
		TimeNow:             time.Now().UTC(),
	}
	allowed, reason, _ := engine.Authorize(authCtx)
	if allowed || !strings.Contains(reason, "FIDO2 hardware token") {
		t.Fatalf("expected MFA requirement denial, got allowed=%v, reason=%s", allowed, reason)
	}

	// 6. Execution with Hardware MFA verified -> must succeed
	authCtx.HardwareMFAVerified = true
	allowed, reason, audit := engine.Authorize(authCtx)
	if !allowed {
		t.Fatalf("expected dual-control action to be authorized, got reason: %s", reason)
	}
	if audit.DualApprovalID != req.RequestID {
		t.Errorf("expected audit record to capture action ID %s", req.RequestID)
	}

	// 7. Four-Eyes violation: Approver (Bob) cannot be the sole executor
	approverExecuteCtx := AuthContext{
		ActorID:             approverID, // Bob is approver
		AssignedRoles:       []Role{RoleComplianceLead},
		HardwareMFAVerified: true,
		RequestedPermission: PermComplianceFreeze,
		HighImpactOp:        ActionAccountFreeze,
		ActionID:            req.RequestID,
		ResourceID:          investorUUID,
		TimeNow:             time.Now().UTC(),
	}
	allowed, reason, _ = engine.Authorize(approverExecuteCtx)
	if allowed || !strings.Contains(reason, "approver cannot be executor") {
		t.Fatalf("expected approver cannot be executor rejection, got: allowed=%v, reason=%s", allowed, reason)
	}
}

func TestIAMRBAC_EphemeralJITAccess(t *testing.T) {
	engine := NewIAMRBACEngine()

	userID := "DEV-RAMESH"
	approverID := "LEAD-SRE"

	// 1. Initially Ramesh has only RoleDeveloperSRE and cannot read compliance dossiers
	authCtx := AuthContext{
		ActorID:             userID,
		AssignedRoles:       []Role{RoleDeveloperSRE},
		RequestedPermission: PermComplianceRead,
		ResourceID:          "COMPLIANCE-DOSSIER-1",
		TimeNow:             time.Now().UTC(),
	}
	allowed, _, _ := engine.Authorize(authCtx)
	if allowed {
		t.Fatal("developer should not have PermComplianceRead initially")
	}

	// 2. Request JIT elevation to RoleComplianceOfficer for 15 minutes
	grant, err := engine.RequestJITAccess(userID, RoleComplianceOfficer, 15*time.Minute, "Investigating production KYC bug ticket INC-402", approverID)
	if err != nil {
		t.Fatalf("failed to grant JIT access: %v", err)
	}

	// 3. Now Ramesh has active JIT role -> authorization should succeed
	allowed, _, _ = engine.Authorize(authCtx)
	if !allowed {
		t.Fatal("expected developer to have temporary PermComplianceRead via JIT grant")
	}

	// 4. Test revocation
	err = engine.RevokeJITAccess(grant.GrantID)
	if err != nil {
		t.Fatalf("failed to revoke JIT access: %v", err)
	}

	// 5. After revocation -> denied again
	allowed, _, _ = engine.Authorize(authCtx)
	if allowed {
		t.Fatal("expected access denied after JIT grant revocation")
	}

	// 6. Test excessive duration rejection (> 60 mins)
	_, err = engine.RequestJITAccess(userID, RoleSuperAdmin, 120*time.Minute, "Too long duration", approverID)
	if err == nil {
		t.Fatal("expected error for JIT duration exceeding 60 minutes limit")
	}
}

func TestIAMRBAC_ABAC_CIDR(t *testing.T) {
	engine := NewIAMRBACEngine()

	allowedCIDRs := []string{"10.200.0.0/16", "172.16.0.0/12"}

	// IP inside CIDR
	authCtx := AuthContext{
		ActorID:             "ADMIN-01",
		AssignedRoles:       []Role{RoleSuperAdmin},
		IPAddress:           "10.200.4.15",
		AllowedCIDRs:        allowedCIDRs,
		RequestedPermission: PermAuditRead,
		ResourceID:          "SYSTEM-AUDIT",
		TimeNow:             time.Now().UTC(),
	}
	allowed, reason, _ := engine.Authorize(authCtx)
	if !allowed {
		t.Fatalf("expected IP 10.200.4.15 to be allowed, got: %s", reason)
	}

	// IP outside CIDR
	authCtx.IPAddress = "192.168.1.100"
	allowed, reason, _ = engine.Authorize(authCtx)
	if allowed {
		t.Fatal("expected IP outside CIDR to be rejected")
	}
	if !strings.Contains(reason, "untrusted source IP") {
		t.Errorf("unexpected reason: %s", reason)
	}
}
