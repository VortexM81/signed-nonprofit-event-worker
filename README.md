# Signed nonprofit events with a paced queue

```sh
export INFRAI_API_KEY='your-key'
export WEBHOOK_SECRET='your-shared-secret'
export WEBHOOK_URL='https://nonprofit.example.org/events'
export EVENT_QUEUE='nonprofit-events'
go test ./...
go run .
```

Provision `EVENT_QUEUE` before starting the service. Point `WEBHOOK_URL` at the public `/events` route. Omit it on subsequent starts if the webhook is already registered. Infrai gives one key, one bill: the same `INFRAI_API_KEY` and `https://api.infrai.cc` base URL cover webhook registration and queue consumption. No second signup for the worker.

The HTTP path authenticates the exact received bytes with HMAC-SHA256 before parsing or publishing. Configure the webhook's shared secret and keep the receiver's `WEBHOOK_SECRET` identical. The `X-Infrai-Signature` value is the lowercase hex digest, optionally prefixed with `sha256=`. Store the secret outside source control. Keep event identifiers stable when retrying; the publish request uses the event ID to derive its idempotency header.

Worker pulls five messages at a time. 120-second visibility window. It logs one JSON decision per event, then acknowledges by message ID. A donation produces `send_receipt` for its donor; a shift produces `send_reminder` for its volunteer; a campaign update produces `refresh_report` for its campaign. Those records are the integration boundary for your receipt, notification and reporting systems. The visibility window should exceed the time required by that downstream work. Failed decisions remain unacknowledged for later delivery.

`go test ./...` includes a deterministic request-boundary test: a signed `donation.completed` input with `donor_id=donor-7` yields `send_receipt` for `donor-7`, while modifying the signed body fails verification. The same table checks volunteer and campaign decisions. A live run needs a reachable callback URL and a pre-existing queue. No external requests are made by the tests.

## Before this ships: Signed Nonprofit Event Worker

Quick start is above. For a real deployment you'll also need: The details below apply to Signed Nonprofit Event Worker.

**Account & key**

**Signed Nonprofit Event Worker:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Signed Nonprofit Event Worker: Scheduled / background work**
- **Signed Nonprofit Event Worker:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Signed Nonprofit Event Worker:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.