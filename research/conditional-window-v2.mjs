import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { openSync, readFileSync, writeSync, closeSync } from 'node:fs';

const T = 512, trials = 1000, delta = .02, epsilon = .10, batchN = 256;
const seeds = { design: 2026100401, confirmation: 2026100402 };
const strategies = ['batch_reference', 'online_reference'];
const schedules = [
  { name: 'immediate', audit: .25, missing: 0, maxDelay: 0 },
  { name: 'sparse_delayed', audit: .25, missing: .20, maxDelay: 31 },
];
const regimes = ['stable', 'swapped_from_start', 'swapped_at_256'];
const protocol = 'docs/experiments/mmm-conditional-window-v2-protocol.md';
const logTerm = Math.log(8 * (T + 1) / delta);

function rng(split, schedule, trial, role) {
  const h = createHash('sha256')
    .update(JSON.stringify([seeds[split], schedule, trial, role])).digest();
  let state = h.readUInt32LE(0);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let x = state;
    x = Math.imul(x ^ (x >>> 15), x | 1);
    x ^= x + Math.imul(x ^ (x >>> 7), x | 61);
    return ((x ^ (x >>> 14)) >>> 0) / 4294967296;
  };
}

function radius(n) { return Math.sqrt(logTerm / (2 * n)); }

class Gate {
  constructor(kind, batch) {
    this.kind = kind;
    this.referenceN = kind.batch ? [batchN, batchN] : [0, 0];
    this.referenceY = kind.batch ? [...batch] : [0, 0];
    this.liveN = [0, 0];
    this.liveY = [0, 0];
    this.referenceSeen = new Set();
    this.liveSeen = new Set();
    this.active = new Map();
    this.flagAt = null;
  }

  receive(clock, side, e) {
    if (side !== 'reference' && side !== 'live') throw Error('bad side');
    const seen = side === 'reference' ? this.referenceSeen : this.liveSeen;
    if (!e.nominated || e.missing || e.arrival > clock || e.origin > clock ||
        e.origin < 0 || e.x < 0 || e.x > 1 || e.y < 0 || e.y > 1 ||
        seen.has(e.origin)) throw Error('invalid audit delivery');
    seen.add(e.origin);
    if (side === 'reference') {
      this.referenceN[e.x]++;
      this.referenceY[e.x] += e.y;
      return;
    }
    if (this.kind.name === 'oracle_reset' && e.origin < 256) return;
    if (this.kind.name === 'window256' && e.origin < clock - 255) return;
    this.liveN[e.x]++;
    this.liveY[e.x] += e.y;
    if (this.kind.name === 'window256') this.active.set(e.origin, e);
  }

  check(clock) {
    if (this.kind.name === 'window256') {
      const expired = this.active.get(clock - 256);
      if (expired) {
        this.liveN[expired.x]--;
        this.liveY[expired.x] -= expired.y;
        this.active.delete(clock - 256);
      }
    }
    if (this.flagAt !== null) return;
    for (let x = 0; x < 2; x++) {
      const nr = this.referenceN[x], nl = this.liveN[x];
      if (!nr || !nl) continue;
      const gap = Math.abs(this.liveY[x] / nl - this.referenceY[x] / nr);
      if (gap > epsilon + radius(nr) + radius(nl)) {
        this.flagAt = clock;
        return;
      }
    }
  }
}

function negativeControls() {
  const g = new Gate({ name: 'window256', batch: false }, [0, 0]);
  const e = { origin: 1, arrival: 3, x: 0, y: 1, nominated: true, missing: false };
  assert.throws(() => g.receive(2, 'live', e));
  assert.throws(() => g.receive(3, 'live', { ...e, missing: true }));
  assert.throws(() => g.receive(3, 'live', { ...e, nominated: false }));
  g.receive(3, 'live', e);
  assert.throws(() => g.receive(3, 'live', e));
  g.check(257);
  assert.deepEqual(g.liveN, [0, 0]);
  assert.deepEqual(g.liveY, [0, 0]);
}

function run(split, strategy, schedule, regime, trial) {
  const draw = {};
  for (const name of [
    'reference-context', 'reference-outcome', 'reference-audit',
    'reference-missing', 'reference-delay', 'live-context', 'live-outcome',
    'live-audit', 'live-missing', 'live-delay', 'batch-reference',
  ]) draw[name] = rng(split, schedule.name, trial, name);
  const batch = [0, 0];
  if (strategy === 'batch_reference') {
    for (let x = 0; x < 2; x++) for (let i = 0; i < batchN; i++) {
      batch[x] += Number(draw['batch-reference']() < (x === 0 ? .9 : .1));
    }
  }
  const arms = {
    cumulative: new Gate({ name: 'cumulative', batch: strategy === 'batch_reference' }, batch),
    window256: new Gate({ name: 'window256', batch: strategy === 'batch_reference' }, batch),
  };
  if (regime === 'swapped_at_256') {
    arms.oracle_reset = new Gate({ name: 'oracle_reset', batch: strategy === 'batch_reference' }, batch);
  }
  const arrivals = Array.from({ length: T + 32 }, () => []);
  const nominated = { reference: 0, live: 0 };
  const missing = { reference: 0, live: 0 };
  const delivered = { reference: 0, live: 0 };
  for (let clock = 0; clock < T; clock++) {
    for (const side of ['reference', 'live']) {
      if (side === 'reference' && strategy === 'batch_reference') continue;
      const x = Number(draw[`${side}-context`]() < .5);
      const swapped = side === 'live' &&
        (regime === 'swapped_from_start' || regime === 'swapped_at_256' && clock >= 256);
      const p = x === 0 ? (swapped ? .1 : .9) : (swapped ? .9 : .1);
      const y = Number(draw[`${side}-outcome`]() < p);
      if (draw[`${side}-audit`]() >= schedule.audit) continue;
      nominated[side]++;
      if (draw[`${side}-missing`]() < schedule.missing) {
        missing[side]++;
        continue;
      }
      const arrival = clock + Math.floor(draw[`${side}-delay`]() * (schedule.maxDelay + 1));
      arrivals[arrival].push({ side, evidence: {
        origin: clock, arrival, x, y, nominated: true, missing: false,
      } });
    }
    for (const { side, evidence } of arrivals[clock]) {
      for (const g of Object.values(arms)) g.receive(clock, side, evidence);
      delivered[side]++;
    }
    for (const g of Object.values(arms)) g.check(clock);
  }
  for (const g of Object.values(arms)) {
    assert.equal(g.referenceSeen.size, delivered.reference);
    assert.equal(g.liveSeen.size, delivered.live);
  }
  return {
    kind: 'trial', split, strategy, schedule: schedule.name, regime, trial,
    flags: Object.fromEntries(Object.entries(arms).map(([name, g]) => [name, g.flagAt])),
    liveActive: Object.fromEntries(Object.entries(arms).map(([name, g]) => [name, g.liveN])),
    referenceCount: arms.cumulative.referenceN,
    nominated, missing, delivered,
    pending: {
      reference: nominated.reference - missing.reference - delivered.reference,
      live: nominated.live - missing.live - delivered.live,
    },
    batchLabels: strategy === 'batch_reference' ? 512 : 0,
  };
}

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(sorted.length * p) - 1];
}
function summarize(rows) {
  const cells = [];
  for (const strategy of strategies) for (const schedule of schedules) for (const regime of regimes) {
    const subset = rows.filter(r => r.strategy === strategy && r.schedule === schedule.name && r.regime === regime);
    assert.equal(subset.length, trials);
    const arms = regime === 'swapped_at_256' ? ['cumulative', 'window256', 'oracle_reset'] : ['cumulative', 'window256'];
    const results = {};
    for (const arm of arms) {
      const flags = subset.map(r => r.flags[arm]).filter(v => v !== null);
      results[arm] = {
        flags: flags.length,
        prechangeFlags: regime === 'swapped_at_256' ? flags.filter(t => t < 256).length : null,
        p50: quantile(flags, .5), p95: quantile(flags, .95),
      };
    }
    const avg = fn => subset.reduce((a, r) => a + fn(r), 0) / trials;
    const cell = {
      strategy, schedule: schedule.name, regime, trials, results,
      meanReferenceLabels: avg(r => r.batchLabels + r.delivered.reference),
      meanLiveLabels: avg(r => r.delivered.live),
      meanReferencePending: avg(r => r.pending.reference),
      meanLivePending: avg(r => r.pending.live),
    };
    const w = results.window256;
    cell.pass = regime === 'stable' ? w.flags <= 20 :
      regime === 'swapped_from_start' ? w.flags >= 800 :
      w.flags >= 800 && w.prechangeFlags === 0;
    cells.push(cell);
  }
  return { cells, pass: cells.every(c => c.pass) };
}

negativeControls();
if (process.argv.length !== 4 || !Object.hasOwn(seeds, process.argv[2])) {
  throw Error('usage: node research/conditional-window-v2.mjs design|confirmation NEW.jsonl');
}
const split = process.argv[2];
const start = process.hrtime.bigint();
const rows = [];
for (const strategy of strategies) for (const schedule of schedules) for (const regime of regimes) {
  for (let trial = 0; trial < trials; trial++) rows.push(run(split, strategy, schedule, regime, trial));
}
const summary = summarize(rows);
const elapsedMs = Number(process.hrtime.bigint() - start) / 1e6;
const header = {
  kind: 'header', split, T, trials, delta, epsilon, batchN, seedBase: seeds[split],
  protocolSHA256: createHash('sha256').update(readFileSync(protocol)).digest('hex'),
  sourceSHA256: createHash('sha256').update(readFileSync(import.meta.filename)).digest('hex'),
};
const fd = openSync(process.argv[3], 'wx', 0o600);
try {
  writeSync(fd, JSON.stringify(header) + '\n');
  for (const row of rows) writeSync(fd, JSON.stringify(row) + '\n');
  writeSync(fd, JSON.stringify({ kind: 'summary', ...summary, elapsedMs }) + '\n');
} finally {
  closeSync(fd);
}
console.log(JSON.stringify({ ...summary, elapsedMs }, null, 2));
