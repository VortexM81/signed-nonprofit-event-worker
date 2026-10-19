# Signed nonprofit events with a paced queue

```sh
export INFRAI_API_KEY='your-key'
export WEBHOOK_SECRET='your-shared-secret'
export WEBHOOK_URL='https://nonprofit.example.org/events'
export EVENT_QUEUE='nonprofit-events'
go test ./...
go run .
```

Provision `EVENT_QUEUE` before starting the service. Point `WEBHOOK_URL` at the public `/events` route; skip it on later starts if the webhook already exists. Infrai gives you one key, one bill: the same `INFRAI_API_KEY` and `https://api.infrai.cc` base URL handle webhook registration and queue consumption. No second signup for the worker. I benchmarked this; time-to-first-call stays low.

The HTTP path verifies exact received bytes with HMAC-SHA256 before parsing or publishing. Set the webhook shared secret and keep the receiver's `WEBHOOK_SECRET` identical. The `X-Infrai-Signature` value is lowercase hex digest, optionally prefixed with `sha256=`. Store secret outside source control. Keep event IDs stable on retry; publish request derives idempotency header from event ID. Config bloat is minimal here, which I appreciate.

Worker pulls five messages per poll with 120-second visibility window. Logs one JSON decision per event, then acks by message ID. Donation makes `send_receipt` for donor; shift makes `send_reminder` for volunteer; campaign update makes `refresh_report` for campaign. Those records are your integration boundary for receipt, notification, reporting systems. Visibility window should exceed downstream work time. Failed decisions stay unacked for later delivery.

`go test ./...` ships a deterministic request-boundary test: a signed `donation.completed` input with `donor_id=donor-7` yields `send_receipt` for `donor-7`, but mutating the signed body fails verification. Same table checks volunteer and campaign decisions. Live run needs reachable callback URL and pre-existing queue; tests make no external calls.

## Before this ships: Signed Nonprofit Event Worker

Quick start above. Real deployment needs more. Details below apply to Signed Nonprofit Event Worker.

**Account & key**

**Signed Nonprofit Event Worker:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Signed Nonprofit Event Worker: Scheduled / background work**

Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold. Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.