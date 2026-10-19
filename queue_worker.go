package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type queuedMessage struct {
	MessageID string `json:"message_id"`
	Payload   event  `json:"payload"`
}

func work(ctx context.Context, api client, queue string) error {
	var batch struct {
		Items []queuedMessage `json:"items"`
	}
	if err := api.post(ctx, "/v1/queue/consume", map[string]any{"queue": queue, "max_messages": 5, "visibility_timeout": 120}, "consume-"+queue, &batch); err != nil {
		return err
	}
	for _, msg := range batch.Items {
		result, err := decide(msg.Payload)
		if err != nil {
			log.Printf("message=%s decision=%v", msg.MessageID, err)
			continue
		}
		line, err := json.Marshal(result)
		if err != nil {
			return err
		}
		// Replace this durable output with a receipt/reminder/report sink in your application.
		log.Print(string(line))
		if msg.MessageID == "" {
			return fmt.Errorf("missing message_id")
		}
		if err := api.post(ctx, "/v1/queue/ack", map[string]string{"queue": queue, "message_id": msg.MessageID}, "ack-"+msg.MessageID, nil); err != nil {
			return err
		}
	}
	return nil
}
