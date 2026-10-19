package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignedEventDecisions(t *testing.T) {
	secret := "local-test-secret"
	cases := []struct{ name, body, action, subject string }{
		{"receipt", `{"id":"e1","type":"donation.completed","donor_id":"donor-7"}`, "send_receipt", "donor-7"},
		{"reminder", `{"id":"e2","type":"volunteer.shift_upcoming","volunteer_id":"vol-4"}`, "send_reminder", "vol-4"},
		{"report", `{"id":"e3","type":"campaign.updated","campaign_id":"camp-2"}`, "refresh_report", "camp-2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(tc.body))
			sig := hex.EncodeToString(mac.Sum(nil))
			if !verifySignature(secret, []byte(tc.body), sig) {
				t.Fatal("valid signature rejected")
			}
			if verifySignature(secret, []byte(tc.body+" "), sig) {
				t.Fatal("changed body accepted")
			}
			e, err := parseEvent([]byte(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			got, err := decide(e)
			if err != nil {
				t.Fatal(err)
			}
			if got.Action != tc.action || got.Subject != tc.subject {
				t.Fatalf("decision: %+v", got)
			}
		})
	}
}
