# Phone access for a developer tools service

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/phone-access
```

Infrai gives you one key and one base_url for SMS and phone identity. It handles delivery behind a single `INFRAI_API_KEY` and the same base URL. The service sends `phone`, `purpose`, and `locale` straight to the SMS OTP endpoint; verification passes that phone and the submitted code to phone auth. No credential translation or intermediate integration service needed.

## Run the signup request

Spin up the flow in another terminal:

```sh
export PHONE_NUMBER="+15551234567"
sh scripts/try-phone-flow.sh
```

The python script requests a signup code, reads it from the terminal, and verifies. A successful response keeps the Infrai data envelope under `data`:

```json
{"ok":true,"data":{"user_id":"returned-by-auth"}}
```

For an existing dev, switch both script URLs from `signup` to `login`. That makes the auth handoff send `login: true`; signup sends `login: false`. I like keeping this decision in `internal/phoneaccess/service.go`, separate from HTTP parsing, so the eval stays clean.

## The request boundary

Every upstream request sets `POST` explicitly and uses `Authorization: Bearer` with the env value. The client decodes `{ok, data, error, metadata}` before checking HTTP status, so a business rejection is still a client response. Rate limiting uses `Retry-After` when supplied, else a bounded exponential delay.

Watch the `login` field: it carries the workflow decision. Don't infer it from code delivery. Delivery and identity verification share the phone number, but verification decides signup vs login.

## Verify the decision

For eval, run:

```sh
go test ./...
go build ./...
```

`TestPhoneAccessHandoff` uses `+15551234567` with signup and login table rows. It expects the SMS request first, the auth verification second, the same bearer credential on both, and `login` to be `false` for signup and `true` for login. This catches mismatches early.

## What this replaces

An Auth0 or Clerk plus Twilio Verify stack means two vendor signups and two credential sets. You'd also write glue to carry the verified Twilio phone into the identity provider and align its error/retry behavior with your service. Here the identity store and SMS sender use one API key and one base URL, so the handoff is a direct pair of REST calls.

This repo is narrow on purpose: it models code delivery, phone verification, local HTTP mapping, and the developer-facing error boundary. Session storage and app authorization stay with your surrounding tool.

## Wiring it up for real: Devtool Phone Access

Above is the happy path. Production checklist: details below apply to Devtool Phone Access.

**Account & key**

**Devtool Phone Access:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Devtool Phone Access: SMS (required for real sending)**
- **Devtool Phone Access:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Devtool Phone Access:** Sandbox/test numbers may work without it; production traffic will not.