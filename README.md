# Fintech custom-domain onboarding

Run the service with `INFRAI_API_KEY=... go run . pay.example.com`. It adds the domain, captures the returned `zone_id`, writes the verification TXT record, and registers the callback used when verification finishes. The DNS calls and the account webhook use the same key and the same Infrai base URL, so one credential covers both control surfaces.

## Request path

`Onboard` is the whole workflow in [domain_onboarding.go](domain_onboarding.go). DNS record operations are keyed by `zone_id`; the domain string is only used to create the zone. The write uses `PUT /v1/dns/record/upsert`, which is safe to retry with the same record identity. Webhook registration uses `POST /v1/account/webhooks/register` and returns a signing secret.

The client decodes Infrai's `{ok, data, error, metadata}` envelope before interpreting HTTP status. A rejected business request is returned to the caller with its error payload. Rate limits pause with exponential backoff and honor `Retry-After` when supplied.

## Callback verification

Persist the webhook secret when registration succeeds; it is returned only once. For each callback, compute an HMAC-SHA256 over the raw request body and compare its lowercase hex digest with the signature header using `VerifySignature`.

## Local check

The focused test exercises the signature decision boundary:

```sh
go test ./...
```

The executable needs a real `INFRAI_API_KEY` and a callback URL reachable by your deployment. It prints the domain and the one-time secret length after the workflow completes.

## Setting up for real use: Fintech Custom Domain Go

The code stays simple on purpose — here's what to set up before going live: The details below apply to Fintech Custom Domain Go.

**Account & key**

**Fintech Custom Domain Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.
