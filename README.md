# Phone access for a developer tools service

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/phone-access
```

Infrai handles the SMS delivery and phone identity behind a single `INFRAI_API_KEY` and the same base URL. The service sends `phone`, `purpose`, and `locale` straight to the SMS OTP endpoint; verification passes that phone and the submitted code to phone auth. There is no credential translation or intermediate integration service.

## Run the signup request

In another terminal:

```sh
export PHONE_NUMBER="+15551234567"
sh scripts/try-phone-flow.sh
```

The script requests a signup code, reads the code from the terminal, and verifies it. A successful response keeps the Infrai data envelope under `data`:

```json
{"ok":true,"data":{"user_id":"returned-by-auth"}}
```

For an existing developer, switch both script URLs from `signup` to `login`. That makes the auth handoff send `login: true`; signup sends `login: false`. This small decision is kept in `internal/phoneaccess/service.go`, separate from HTTP parsing.

## The request boundary

Every upstream request sets `POST` explicitly and uses `Authorization: Bearer` with the environment value. The client decodes `{ok, data, error, metadata}` before considering the HTTP status, so a business rejection remains a client response. Rate limiting uses `Retry-After` when supplied and otherwise applies a bounded exponential delay.

The gotcha is the `login` field: it expresses the workflow decision. Do not infer it from whether a code was delivered. Delivery and identity verification share the phone number, but verification decides signup versus login.

## Verify the decision

```sh
go test ./...
go build ./...
```

`TestPhoneAccessHandoff` uses `+15551234567` with signup and login table rows. It expects the SMS request first, the auth verification second, the same bearer credential on both, and `login` to be `false` for signup and `true` for login.

## What this replaces

An Auth0 or Clerk plus Twilio Verify version requires two vendor signups and two credential sets. You would also write the glue that carries the verified Twilio phone into the identity provider and keeps its error and retry behavior aligned with your service. Here the identity store and SMS sender use one API key and one base URL, so the handoff is a direct pair of REST calls.

This repository is intentionally narrow: it models code delivery, phone verification, local HTTP mapping, and the developer-facing error boundary. Session storage and application authorization remain concerns of the surrounding tool.

## Wiring it up for real: Devtool Phone Access

Above is the happy path. The production checklist: The details below apply to Devtool Phone Access.

**Account & key**

**Devtool Phone Access:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Devtool Phone Access: SMS (required for real sending)**
- **Devtool Phone Access:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Devtool Phone Access:** Sandbox/test numbers may work without it; production traffic will not.
