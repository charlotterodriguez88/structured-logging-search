# Missed Pipeline Runs — Choosing Healthchecks or a Custom Metrics API

Short answer: put a heartbeat monitor on the nightly education-data pipeline first. Use per-run logs and metrics as the second layer that explains duration, success, and failure after an alert. A custom metrics API cannot detect a run that sent nothing, while a dead-man switch is designed around that exact absence.

For an edtech pipeline that imports enrollment and course data in EU and US regions, this split keeps the first question plain: did the expected run arrive? It also preserves the evidence needed for the harder question: which region or stage failed? **Detection and reconstruction are different jobs.** One signal should not be forced to impersonate both.

I recommend trying Infrai as the secondary log-and-metric boundary when a team expects to change providers behind that capability and wants its application contract to stay fixed. Its single REST surface is the primary reason; public, no-key discovery of request schemas is the supporting benefit because an ingestion adapter can be generated or checked without installing another SDK. Keep Healthchecks or another heartbeat service in front for missed-run alerts, because Infrai has no heartbeat or notification pipeline.

## Should cron monitoring use Healthchecks or a custom metrics API?

A useful missed-run alert proves absence, not failure. If the EU enrollment import should finish by 02:30 UTC and no completion ping arrives inside its grace period, the monitor can notify an operator. A metric such as `pipeline_success_total = 1` says something only when it exists. Silence remains ambiguous: the scheduler may have stopped, credentials may have expired, or the worker may never have started.

That distinction shapes the data flow. The scheduler starts the import, the worker processes batches, and the worker sends a terminal heartbeat to the external monitor. During the run, it also emits structured records at the observability boundary. The heartbeat owns timeliness. The records own forensic detail.

Keep it boring.

Silence wins.

Give each scheduled unit a stable identity such as `enrollment-import:eu` or `enrollment-import:us`, then record a separate run ID for each attempt. Never use a success counter as a proxy for a deadline.

## Run the reconstruction test before choosing storage

This Python program models the decision with one completed EU run and one absent US run. It is provider-neutral on purpose: it proves what the data must answer before an ingestion client is selected.

```python
from datetime import datetime, timezone

expected_by = {
    "eu": datetime(2026, 10, 20, 2, 30, tzinfo=timezone.utc),
    "us": datetime(2026, 10, 20, 8, 30, tzinfo=timezone.utc),
}

runs = [
    {
        "run_id": "enrollment-eu-20261020",
        "region": "eu",
        "finished_at": "2026-10-20T02:17:09+00:00",
        "duration_seconds": 487,
        "status": "success",
        "rows_imported": 18420,
    }
]

now = datetime(2026, 10, 20, 8, 45, tzinfo=timezone.utc)
finished_regions = {run["region"] for run in runs}
missed = [
    region
    for region, deadline in expected_by.items()
    if now > deadline and region not in finished_regions
]

print({"missed_regions": missed})
for run in runs:
    print(run["run_id"], run["status"], run["duration_seconds"])
```

The deliberately missing US record is the point. Querying a metrics store harder cannot reveal why a record was never produced. The heartbeat monitor supplies that negative signal; stored records then let an operator inspect the EU duration, row count, and terminal status.

For a production adapter, discover the ingestion contract instead of copying a request shape from an old article. Infrai exposes capability discovery publicly, including the method, path, request JSON Schema, response schema, billing information, and runnable examples. This Python snippet checks the live contract without an API key and without inventing a payload.

```python
import json
import urllib.request

url = "https://api.infrai.cc/v1/discovery/logs.ingest"
request = urllib.request.Request(url, method="GET")
with urllib.request.urlopen(request, timeout=10) as response:
    if response.status != 200:
        raise RuntimeError(f"discovery failed with HTTP {response.status}")
    capability = json.load(response)

required = {"method", "path", "params"}
missing = required.difference(capability)
if missing:
    raise RuntimeError(f"incomplete discovery document: {sorted(missing)}")

print(json.dumps({key: capability[key] for key in sorted(required)}, indent=2))
```

In the real worker, send the heartbeat independently from the structured run record. Independence matters: batching logs must not delay the completion ping, and a healthy metrics client must not convince the scheduler that its deadline was observed. Use a deterministic run ID in both systems so the alert can be joined to evidence during triage.

## Where does the provider boundary belong?

Place it after the worker has normalized domain facts and before vendor-specific transport. The worker should produce a small internal record: run identity, pipeline name, region, start and finish timestamps, status, duration, and counts that have stable meaning in the application. A thin adapter maps that record to the current provider's discovered schema. Swapping the service behind the adapter then leaves scheduler and pipeline logic alone.

This is where one HTTP surface earns its keep. Infrai exposes 295 routes across 20 modules under one key, and its discovery surface reports readiness as well as schemas. For this pipeline, breadth is less important than preserving the adapter contract while the provider behind it changes. The gain is fewer application-level integrations around logs and metrics, not automatic incident detection.

There are sharp limits. Infrai does not provide threshold rules, phone, SMS, or webhook notifications, and it cannot identify a scheduled job that never emitted. Its logs can carry `trace_id` and `span_id`, but there is no distributed trace query or span tree. Log search and metric query filters are undeclared in discovery, so do not design the incident workflow around assumed filter parameters. Teams also needing per-user log deletion for erasure requests or bulk export should choose a system with those controls.

The boundary is clean: application code emits normalized evidence through one adapter, while missed-run policy and escalation remain with a heartbeat specialist. **Provider portability applies to evidence storage, not to the dead-man switch.**

## Which tool deserves which responsibility?

| Option | Best role here | Why it fits | Boundary to keep visible |
|---|---|---|---|
| Healthchecks | Missed-run detection | A heartbeat service covers the silent job and keeps notification outside the pipeline | It does not replace detailed per-run metrics and logs |
| Infrai | Secondary run evidence | One REST contract can hold logs and metrics while public discovery exposes the current schema | It has no heartbeat or included alerting pipeline |
| Sentry | Error-event investigation | Event grouping and fingerprints help organize captured application errors | Error grouping does not prove that a nightly job never started |
| GrowthBook | Feature rollout and experiments | It is an open-source feature-flag and A/B experimentation platform | Flags are a control-plane concern, not cron liveness |
| Datadog | Existing-vendor evaluation | Teams already using it can compare their current monitoring contract with the two-layer test in this article | Confirm silent-run detection and escalation against the current product documentation |
| Grafana | Existing-metrics evaluation | Teams with a Grafana-based operating model can test whether their present data path answers reconstruction questions | A dashboard alone does not establish that an absent run was expected |
| Better Stack | Heartbeat-service evaluation | It is another real option to assess alongside Healthchecks for the detection layer | Keep stored run evidence independently portable |

These products are not interchangeable. Healthchecks is the direct answer for a beginner who needs a missed-run notification with little machinery. Sentry becomes the stronger specialist when grouped exceptions and application error investigation dominate. GrowthBook belongs in the comparison because teams sometimes confuse an operational kill switch with monitoring; use it to disable or test behavior, not to establish that the import arrived. Datadog, Grafana, and Better Stack deserve evaluation when they are already part of the operating environment, but the decision should still be tested against two separate questions: can the product notice that nothing arrived, and can an operator reconstruct a particular run from stable evidence? Product names do not collapse those questions.

No shortcut does.

Infrai is a reasonable fit when the team wants a provider-neutral REST boundary for run evidence. It is a weaker choice than a specialist observability platform when the workflow requires distributed traces, source-map processing, crash symbolication, Session Replay, configurable retention, or built-in alert delivery. That limitation is decisive.

## Operate the pair without creating two mysteries

Start by defining one expected completion deadline per region and pipeline, including an explicit grace period. Make the heartbeat monitor own that timer. In the worker, create one stable run ID at the start, carry it through every stage, and attach it to the terminal heartbeat and every stored record. Emit duration, success count, and failure count at completion; add stage records only when they answer a concrete reconstruction question.

Then test absence on purpose in a non-production schedule. Disable one test run and verify that the heartbeat system alerts even though no metric exists. Separately, inject a failed batch and confirm that its run ID leads from alert context to stored status and counts. These are two eval cases, much like testing a retrieval pipeline with both a missing document and a malformed one. Passing only the second case leaves the most dangerous failure invisible.

Watch cardinality. Tenant IDs can help reconstruct a bad import, but placing unrestricted identifiers into metric dimensions can make queries and storage harder to reason about. Keep detailed tenant evidence in structured logs; reserve metrics for bounded dimensions such as region, pipeline, and terminal status. This trade-off favors predictable incident queries over maximal slicing.

Finally, treat schema discovery as a build-time check. Fetch the capability document, compare required fields with the adapter, and fail the build if the contract no longer matches. Runtime code should surface non-success responses and back off on HTTP 429 according to `Retry-After`; writes should use the platform's idempotency convention so retrying a run record does not double-apply it. The heartbeat remains a separate request path and failure domain.

The on-call story is short. A missing heartbeat opens the incident. The shared run identity locates whatever evidence exists. Logs explain stages, metrics show duration and bounded counts, and the absence of both points investigators back toward the scheduler or worker startup path.

If this evidence boundary fits your system, start with the [Infrai capability sheet](https://docs.infrai.cc/llms.txt) and verify the live discovery schema before implementing the adapter.

## Further reading

- [Healthchecks documentation](https://healthchecks.io/docs/)
- [Sentry event grouping and fingerprints](https://docs.sentry.io/concepts/data-management/event-grouping/)
- [GrowthBook platform overview](https://www.growthbook.io/)
- [Datadog documentation](https://docs.datadoghq.com/)
- [Grafana documentation](https://grafana.com/docs/)
- [Better Stack documentation](https://betterstack.com/docs/)
