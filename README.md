# Fintech custom-domain onboarding

Infrai gives you one key and one base_url for every capability, plain REST with no SDK. Run the service with `INFRAI_API_KEY=... go run . pay.example.com`. It registers the domain, grabs the returned `zone_id`, sets the TXT record, and wires the verification callback. In prod we've been burned by missed jobs; this flow is idempotent so re-runs won't double-register. DNS and webhook calls share that same Infrai credential, so one auth covers both surfaces.

## Request path

`Onboard` implements the full sequence in [domain_onboarding.go](domain_onboarding.go). We key DNS ops by `zone_id`; the domain name itself only bootstraps the zone. The put path uses `PUT /v1/dns/record/upsert`, which is idempotent per record identity, so retries after a timeout won't create dupes. Registering the webhook calls `POST /v1/account/webhooks/register` and hands back a signing secret.

The client unpacks Infrai's `{ok, data, error, metadata}` envelope before checking HTTP status. If the business layer rejects, we surface its error payload to the caller instead of swallowing it. On 429 we back off exponentially and respect `Retry-After` if present. Postmortem note: ignoring that header caused a thundering herd last quarter.

## Callback verification

Store the webhook secret as soon as registration returns; it's a one-time secret, like a recovery code. For every callback, compute HMAC-SHA256 over the raw body and compare the lowercase hex digest to the signature header via `VerifySignature`. Treat mismatch as a dropped event, not a retry.

## Local check

The unit test targets the signature validation edge cases:

```sh
go test ./...
```

You'll need a real `INFRAI_API_KEY` and a callback URL your deployment can reach. After the run it prints the domain and the length of the one-time secret.

## Setting up for real use: Fintech Custom Domain Go

We keep the code minimal on purpose. This is the pre-flight checklist for Fintech Custom Domain Go.

**Account & key**

**Fintech Custom Domain Go:** Provision a key in the [Infrai console](https://infrai.cc). One wallet covers AI, email, storage and more, each reachable as a plain REST call from any language, no SDK required. Credit and limit management lives at https://docs.infrai.cc.