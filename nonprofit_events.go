package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type event struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	DonorID     string `json:"donor_id,omitempty"`
	VolunteerID string `json:"volunteer_id,omitempty"`
	CampaignID  string `json:"campaign_id,omitempty"`
}

type decision struct {
	EventID string `json:"event_id"`
	Action  string `json:"action"`
	Subject string `json:"subject"`
}

func verifySignature(secret string, body []byte, signature string) bool {
	provided, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil || len(provided) != sha256.Size || secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), provided)
}

func decide(e event) (decision, error) {
	if e.ID == "" {
		return decision{}, errors.New("missing event id")
	}
	switch e.Type {
	case "donation.completed":
		if e.DonorID != "" {
			return decision{e.ID, "send_receipt", e.DonorID}, nil
		}
	case "volunteer.shift_upcoming":
		if e.VolunteerID != "" {
			return decision{e.ID, "send_reminder", e.VolunteerID}, nil
		}
	case "campaign.updated":
		if e.CampaignID != "" {
			return decision{e.ID, "refresh_report", e.CampaignID}, nil
		}
	default:
		return decision{}, fmt.Errorf("unhandled event type %q", e.Type)
	}
	return decision{}, fmt.Errorf("missing subject for %s", e.Type)
}

func parseEvent(body []byte) (event, error) {
	var e event
	if err := json.Unmarshal(body, &e); err != nil {
		return e, err
	}
	_, err := decide(e)
	return e, err
}
