package main

import (
	"testing"
	"time"
)

func TestRedlock_AcquireAndRelease(t *testing.T) {
	mgr := NewRedlockManager()
	token, err := mgr.Acquire("order_mutex_user123", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Double acquire should fail
	_, err2 := mgr.Acquire("order_mutex_user123", 5*time.Second)
	if err2 == nil {
		t.Error("expected second acquire to fail while locked")
	}

	// Release with wrong token should fail
	if mgr.Release("order_mutex_user123", "wrong_token") {
		t.Error("release with wrong token should return false")
	}

	// Release with correct token
	if !mgr.Release("order_mutex_user123", token) {
		t.Error("release with correct token should succeed")
	}

	// Can acquire again
	newToken, err3 := mgr.Acquire("order_mutex_user123", 5*time.Second)
	if err3 != nil {
		t.Fatal("should be able to re-acquire after release")
	}
	mgr.Release("order_mutex_user123", newToken)
}
