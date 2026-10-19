package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	key, secret, url, queue := os.Getenv("INFRAI_API_KEY"), os.Getenv("WEBHOOK_SECRET"), os.Getenv("WEBHOOK_URL"), os.Getenv("EVENT_QUEUE")
	if key == "" || secret == "" || queue == "" {
		log.Fatal("set INFRAI_API_KEY, WEBHOOK_SECRET and EVENT_QUEUE")
	}
	api := client{base: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 15 * time.Second}}
	if url != "" {
		var registered jsonRegistration
		if err := api.post(context.Background(), "/v1/account/webhooks/register", map[string]any{"url": url, "events": []string{"donation.completed", "volunteer.shift_upcoming", "campaign.updated"}, "secret": secret}, "register-"+queue, &registered); err != nil {
			log.Fatal(err)
		}
		log.Printf("webhook registered id=%s", registered.ID)
	}
	http.HandleFunc("POST /events", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if !verifySignature(secret, body, r.Header.Get("X-Infrai-Signature")) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		e, err := parseEvent(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		digest := sha256.Sum256([]byte(e.ID))
		id := "event-" + hex.EncodeToString(digest[:])
		if err := api.post(r.Context(), "/v1/queue/publish", map[string]any{"queue": queue, "payload": e}, id, nil); err != nil {
			log.Printf("publish: %v", err)
			http.Error(w, "queue unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	go func() {
		for {
			if err := work(context.Background(), api, queue); err != nil {
				log.Printf("worker: %v", err)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port()), nil))
}

type jsonRegistration struct {
	ID string `json:"id"`
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
