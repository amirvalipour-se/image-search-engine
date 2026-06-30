import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// ── Metrics ──────────────────────────────────────────────────────────────────
const errorRate = new Rate('errors');
const searchLatency = new Trend('search_latency', true);
const searchSuccess = new Counter('search_success');
const searchFail = new Counter('search_fail');

// ── Config ───────────────────────────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const IMAGES_DIR = __ENV.IMAGES_DIR || './images';
const K = __ENV.K || '500';

// Pre-load image pool
const IMAGE_POOL = [];
for (let i = 0; i < 50; i++) {
  try {
    const path = `${IMAGES_DIR}/${i}.jpg`;
    IMAGE_POOL.push({ path, data: open(path, 'b') });
  } catch (_) {}
}

function randomImage() {
  return IMAGE_POOL[randomIntBetween(0, IMAGE_POOL.length - 1)];
}

// ── Scenario configs ─────────────────────────────────────────────────────────
const SCENARIOS = {
  baseline: {
    baseline: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
    },
  },
  concurrency: {
    concurrency: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 10 },
        { duration: '20s', target: 25 },
        { duration: '20s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 },
      ],
    },
  },
  spike: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '5s', target: 0 },
      ],
    },
  },
  soak: {
    soak: {
      executor: 'constant-vus',
      vus: 30,
      duration: '10m',
    },
  },
};

const selected = __ENV.SCENARIO || 'baseline';

export const options = {
  scenarios: SCENARIOS[selected] || SCENARIOS.baseline,
  thresholds: {
    search_latency: ['p(95)<5000'],
    errors: ['rate<0.05'],
    http_req_duration: ['p(99)<10000'],
  },
};

// ── Search function ──────────────────────────────────────────────────────────
function doSearch() {
  const img = randomImage();
  const res = http.post(
    `${BASE_URL}/api/search?k=${K}`,
    { file: http.file(img.data, img.path, 'image/jpeg') },
    { timeout: '30s' }
  );

  const ok = check(res, {
    'status is 200': (r) => r.status === 200,
    'has results': (r) => {
      try { return JSON.parse(r.body).results !== undefined; }
      catch { return false; }
    },
  });

  errorRate.add(!ok);
  searchLatency.add(res.timings.duration);
  ok ? searchSuccess.add(1) : searchFail.add(1);

  sleep(randomIntBetween(500, 2000) / 1000);
}

export default function () {
  doSearch();
}

// ── Summary ──────────────────────────────────────────────────────────────────
function fmt(v, dec = 1) {
  return v != null ? v.toFixed(dec) : 'n/a';
}

export function handleSummary(data) {
  const ts = new Date().toISOString().replace(/[:.]/g, '-');
  const lines = [];
  lines.push('');
  lines.push('═══════════════════════════════════════════════════════');
  lines.push(`  LOAD TEST RESULTS — ${selected}`);
  lines.push('═══════════════════════════════════════════════════════');

  const m = data.metrics;

  const reqs = m.http_reqs && m.http_reqs.values;
  if (reqs) {
    lines.push(`  Total requests:    ${reqs.count}`);
    lines.push(`  Requests/sec:      ${fmt(reqs.rate, 2)}`);
  }

  const sl = m.search_latency && m.search_latency.values;
  if (sl) {
    lines.push('');
    lines.push('  Search Latency (ms):');
    lines.push(`    min:   ${fmt(sl.min)}`);
    lines.push(`    avg:   ${fmt(sl.avg)}`);
    lines.push(`    med:   ${fmt(sl.med)}`);
    lines.push(`    p90:   ${fmt(sl['p(90)'])}`);
    lines.push(`    p95:   ${fmt(sl['p(95)'])}`);
    lines.push(`    p99:   ${fmt(sl['p(99)'])}`);
    lines.push(`    max:   ${fmt(sl.max)}`);
  }

  const hd = m.http_req_duration && m.http_req_duration.values;
  if (hd) {
    lines.push('');
    lines.push('  HTTP Duration (ms):');
    lines.push(`    avg:   ${fmt(hd.avg)}`);
    lines.push(`    p95:   ${fmt(hd['p(95)'])}`);
    lines.push(`    p99:   ${fmt(hd['p(99)'])}`);
  }

  const er = m.errors && m.errors.values;
  if (er) {
    lines.push('');
    lines.push(`  Error rate:        ${fmt(er.rate * 100, 2)}%`);
  }
  if (m.search_success) lines.push(`  Successful:        ${m.search_success.values.count}`);
  if (m.search_fail) lines.push(`  Failed:            ${m.search_fail.values.count}`);

  const passed = data.thresholds && Object.values(data.thresholds).every(t => t.ok);
  lines.push('');
  lines.push(`  Thresholds:        ${passed ? 'PASSED ✓' : 'FAILED ✗'}`);
  lines.push('═══════════════════════════════════════════════════════');

  return {
    [`results/${selected}-${ts}.json`]: JSON.stringify(data, null, 2),
    stdout: lines.join('\n'),
  };
}
