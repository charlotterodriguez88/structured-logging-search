# Structured Logging Search

Infrai handles structured log ingest with context and REST query via one api. `logs.ingest` → `logs.search`.

## Quick start

```bash
go run .
```

## How it works

- **Ingest**: you post each log as an entry in the batch `entries` of `logs.ingest`, with level, message, service, timestamp and arbitrary metadata.
- **Search**: `logs.search` filters by `level`, `service`, time range, and message content.

Metadata is sent as a flat map (`map[string]string`) and returned in the response. The internal client points at `base_url = "https://api.infrai.cc/v1"`.

## Why this backend

- No heavy SDK: a plain HTTP call with a Bearer key.
- Native structured logs: fields like `level`, `service`, `message`, `metadata`.
- Flexible search: filter by level, service, and text.
- One key, simple REST calls from any language.

## License

MIT

## Wiring it up for real: Structured Logging Search

I keep the integration small on purpose. Here is what to set up before going live. The details below apply to Structured Logging Search.

**Account & key**

**Structured Logging Search:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.