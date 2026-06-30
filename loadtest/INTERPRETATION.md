# Load Test Interpretation

## Run

```bash
cd loadtest
./run.sh                    # all scenarios
./run.sh baseline           # single scenario
./run.sh concurrency
```

## Metrics

| Metric | What it measures |
|--------|-----------------|
| `search_latency` | Go → Python → FAISS → response (end-to-end) |
| `http_req_duration` | Full HTTP round-trip |
| `errors` | Non-200 responses or timeouts |

## Pass/Fail Thresholds

| Condition | Threshold |
|-----------|-----------|
| p95 latency | < 5s |
| Error rate | < 5% |
| p99 latency | < 10s |

## Bottleneck Diagnosis

| Symptom | Bottleneck | Fix |
|---------|-----------|-----|
| Go CPU maxed, Python idle | Go handler blocking | Check multipart parsing, JSON serialization |
| Python CPU maxed, Go idle | CLIP inference bound | Add `--workers 4`, use GPU, or switch to IndexIVF |
| Latency scales with image size | Network/upload overhead | Resize images before upload, use Docker network |
| Latency drifts upward over time | Memory leak / GC | Check Go heap, Python process RSS |

## Scenario Purpose

| Scenario | VUs | Duration | Tests |
|----------|-----|----------|-------|
| baseline | 1 | 30s | Minimum latency, no contention |
| concurrency | 0→200 | 2m | Scaling behavior under load |
| spike | 0→100 | 30s | Recovery from traffic burst |
| soak | 30 | 10m | Memory leaks, GC pressure |

## What "Good" Looks Like

| Metric | Acceptable | Good | Excellent |
|--------|-----------|------|-----------|
| p50 latency | < 2s | < 1s | < 500ms |
| p95 latency | < 5s | < 3s | < 1.5s |
| Error rate | < 5% | < 1% | < 0.1% |

## Monitor During Tests

```bash
top -pid $(pgrep -f "cmd/api")     # Go
top -pid $(pgrep -f uvicorn)       # Python
docker stats                       # Docker
```
