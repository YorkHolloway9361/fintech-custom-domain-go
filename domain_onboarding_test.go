package main

import "testing"

func TestVerifySignature(t *testing.T) {
	payload := `{"event":"verification.finished"}`
	if VerifySignature("secret", payload, "bad") {
		t.Fatal("invalid signature accepted")
	}
	if !VerifySignature("secret", payload, "ed45701d1ebdfc2c059f08e4af5e1813897cecf23e520dbc11565365378e74a2") {
		t.Fatal("valid signature rejected")
	}
}
