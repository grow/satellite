package auth

import (
	"testing"

	"appengine"
	"appengine/aetest"
)

func TestBasicAuth(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	a := NewBasicAuth()
	if a == nil {
		t.Fatalf("Failed to create BasicAuth")
	}

	// Create new users: alice and bob.
	req1, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}
	c1 := appengine.NewContext(req1)
	a.AddUser(c1, "alice", "alicerocks")
	a.AddUser(c1, "bob", "alicesucks")

	// Verify alice login success.
	req2, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req2: %v", err)
	}
	req2.SetBasicAuth("alice", "alicerocks")
	if !a.IsAuthorized(req2) {
		t.Fatalf("Failed to log in as alice")
	}

	// Verify bob login success.
	req3, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req3: %v", err)
	}
	req3.SetBasicAuth("bob", "alicesucks")
	if !a.IsAuthorized(req3) {
		t.Fatalf("Failed to log in as bob")
	}

	// Verify alice login fail.
	req4, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req4: %v", err)
	}
	req4.SetBasicAuth("alice", "bobsucks")
	if a.IsAuthorized(req4) {
		t.Fatalf("Failed: alice logged in with incorrect password")
	}

	// Verify unknown user fail.
	req5, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req5: %v", err)
	}
	req5.SetBasicAuth("alice", "bobsucks")
	if a.IsAuthorized(req5) {
		t.Fatalf("Failed: alice logged in with incorrect password")
	}
}
