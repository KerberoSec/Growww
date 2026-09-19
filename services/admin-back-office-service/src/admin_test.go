package adminbackoffice

import (
	"testing"
)

func TestCreateAndGetUser(t *testing.T) {
	svc := NewAdminService()

	u, err := svc.CreateUser("u1", "alice@example.com", "Alice Smith", RoleViewer)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if u.ID != "u1" || u.Email != "alice@example.com" {
		t.Errorf("unexpected user data: %+v", u)
	}
	if u.KYC != KYCPending {
		t.Errorf("expected KYC PENDING, got %s", u.KYC)
	}
	if !u.Enabled {
		t.Error("expected user to be enabled by default")
	}

	// Duplicate should fail
	_, err = svc.CreateUser("u1", "bob@example.com", "Bob", RoleViewer)
	if err == nil {
		t.Error("expected error on duplicate user creation")
	}

	// GetUser
	fetched, err := svc.GetUser("u1")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if fetched.FullName != "Alice Smith" {
		t.Errorf("unexpected full name: %s", fetched.FullName)
	}

	// GetUser not found
	_, err = svc.GetUser("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestDisableEnableUser(t *testing.T) {
	svc := NewAdminService()
	svc.CreateUser("u1", "a@b.com", "Test", RoleSupport)

	if err := svc.DisableUser("u1"); err != nil {
		t.Fatalf("DisableUser failed: %v", err)
	}
	u, _ := svc.GetUser("u1")
	if u.Enabled {
		t.Error("expected user to be disabled")
	}

	if err := svc.EnableUser("u1"); err != nil {
		t.Fatalf("EnableUser failed: %v", err)
	}
	u, _ = svc.GetUser("u1")
	if !u.Enabled {
		t.Error("expected user to be enabled")
	}

	// Disable nonexistent
	if err := svc.DisableUser("no-user"); err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestKYCApprovalQueue(t *testing.T) {
	svc := NewAdminService()
	svc.CreateUser("u1", "a@b.com", "Test", RoleViewer)

	// Submit KYC
	if err := svc.SubmitKYC("u1", "https://docs.example.com/pan.pdf"); err != nil {
		t.Fatalf("SubmitKYC failed: %v", err)
	}
	if count := svc.PendingKYCCount(); count != 1 {
		t.Errorf("expected 1 pending KYC, got %d", count)
	}

	// Approve KYC
	if err := svc.ApproveKYC("u1", "admin1"); err != nil {
		t.Fatalf("ApproveKYC failed: %v", err)
	}
	u, _ := svc.GetUser("u1")
	if u.KYC != KYCApproved {
		t.Errorf("expected APPROVED, got %s", u.KYC)
	}
	if count := svc.PendingKYCCount(); count != 0 {
		t.Errorf("expected 0 pending KYC after approval, got %d", count)
	}

	// Submit again should fail since approved
	if err := svc.SubmitKYC("u1", "https://docs.example.com/pan2.pdf"); err == nil {
		t.Error("expected error when KYC already approved")
	}
}

func TestKYCRejection(t *testing.T) {
	svc := NewAdminService()
	svc.CreateUser("u2", "b@b.com", "Bob", RoleViewer)
	svc.SubmitKYC("u2", "https://docs.example.com/doc.pdf")

	if err := svc.RejectKYC("u2", "admin1", "blurry document"); err != nil {
		t.Fatalf("RejectKYC failed: %v", err)
	}
	u, _ := svc.GetUser("u2")
	if u.KYC != KYCRejected {
		t.Errorf("expected REJECTED, got %s", u.KYC)
	}

	// Reject already rejected
	if err := svc.RejectKYC("u2", "admin1", "again"); err == nil {
		t.Error("expected error rejecting non-PENDING KYC")
	}
}

func TestTradeAlerts(t *testing.T) {
	svc := NewAdminService()

	alert := &TradeAlert{
		ID:        "alert-001",
		UserID:    "u1",
		Symbol:    "BTC-INR",
		Side:      "BUY",
		Quantity:  100,
		Price:     5000000,
		AlertType: "WASH_TRADE",
		Severity:  "HIGH",
	}

	if err := svc.RaiseTradeAlert(alert); err != nil {
		t.Fatalf("RaiseTradeAlert failed: %v", err)
	}

	unresolved := svc.UnresolvedAlerts()
	if len(unresolved) != 1 {
		t.Fatalf("expected 1 unresolved alert, got %d", len(unresolved))
	}
	if unresolved[0].AlertType != "WASH_TRADE" {
		t.Errorf("unexpected alert type: %s", unresolved[0].AlertType)
	}

	// Resolve
	if err := svc.ResolveTradeAlert("alert-001"); err != nil {
		t.Fatalf("ResolveTradeAlert failed: %v", err)
	}
	if len(svc.UnresolvedAlerts()) != 0 {
		t.Error("expected 0 unresolved alerts after resolution")
	}

	// Resolve nonexistent
	if err := svc.ResolveTradeAlert("no-alert"); err == nil {
		t.Error("expected error resolving nonexistent alert")
	}
}

func TestSystemConfig(t *testing.T) {
	svc := NewAdminService()

	if err := svc.SetConfig("max_order_size", "1000000", "admin1"); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	cfg, err := svc.GetConfig("max_order_size")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if cfg.Value != "1000000" {
		t.Errorf("expected 1000000, got %s", cfg.Value)
	}

	// Overwrite
	svc.SetConfig("max_order_size", "2000000", "admin2")
	cfg, _ = svc.GetConfig("max_order_size")
	if cfg.Value != "2000000" {
		t.Errorf("expected 2000000 after overwrite, got %s", cfg.Value)
	}
	if cfg.UpdatedBy != "admin2" {
		t.Errorf("expected updater admin2, got %s", cfg.UpdatedBy)
	}

	// Not found
	_, err = svc.GetConfig("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent config")
	}

	// AllConfigs
	svc.SetConfig("maintenance_mode", "false", "admin1")
	all := svc.AllConfigs()
	if len(all) != 2 {
		t.Errorf("expected 2 configs, got %d", len(all))
	}
}

func TestListUsersFilter(t *testing.T) {
	svc := NewAdminService()
	svc.CreateUser("u1", "a@b.com", "A", RoleViewer)
	svc.CreateUser("u2", "b@b.com", "B", RoleViewer)
	svc.SubmitKYC("u1", "doc1")
	svc.ApproveKYC("u1", "admin1")

	// All users
	all := svc.ListUsers(nil)
	if len(all) != 2 {
		t.Errorf("expected 2 users, got %d", len(all))
	}

	// Only approved
	approved := KYCApproved
	filtered := svc.ListUsers(&approved)
	if len(filtered) != 1 {
		t.Errorf("expected 1 approved user, got %d", len(filtered))
	}

	// Only pending
	pending := KYCPending
	filteredPending := svc.ListUsers(&pending)
	if len(filteredPending) != 1 {
		t.Errorf("expected 1 pending user, got %d", len(filteredPending))
	}
}
