import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { closeSync, openSync, readFileSync, writeSync } from 'node:fs';

const T = 512, trials = 1000, delta = .02, epsilon = .10, width = 128;
const seeds = { design: 2026100501, confirmation: 2026100502 };
const cases = ['stable', 'live_shift256', 'reference_shift256', 'common_shift256', 'live_shift0'];
const schedules = ['complete', 'sparse_delayed'];
const arms = ['paired_window', 'paired_cumulative', 'scalar_window'];
const protocol = 'docs/experiments/mmm-paired-context-v3-protocol.md';

function rng(split, schedule, scenario, trial, role) {
  const h = createHash('sha256')
    .update(JSON.stringify([seeds[split], schedule, scenario, trial, role])).digest();
  let state = h.readUInt32LE(0);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let x = state;
    x = Math.imul(x ^ (x >>> 15), x | 1);
    x ^= x + Math.imul(x ^ (x >>> 7), x | 61);
    return ((x ^ (x >>> 14)) >>> 0) / 4294967296;
  };
}

function radius(n, cells) {
  return Math.sqrt(2 * Math.log(2 * cells * (T + 1) / delta) / n);
}

class Gate {
  constructor(kind) {
    this.kind = kind;
    this.count = [0, 0];
    this.sum = [0, 0];
    this.active = new Map();
    this.seen = new Set();
    this.flagAt = null;
  }

  receive(clock, pair) {
    if (!pair.nominated || pair.missing || pair.arrival > clock ||
        pair.origin > clock || pair.origin < 0 || pair.origin >= T ||
        pair.x !== 0 && pair.x !== 1 || ![-1, 0, 1].includes(pair.d) ||
        this.seen.has(pair.origin)) throw Error('invalid paired evidence');
    this.seen.add(pair.origin);
    if (this.kind !== 'paired_cumulative' && pair.origin < clock - width + 1) return;
    const cell = this.kind === 'scalar_window' ? 0 : pair.x;
    this.count[cell]++;
    this.sum[cell] += pair.d;
    if (this.kind !== 'paired_cumulative') this.active.set(pair.origin, { cell, d: pair.d });
  }

  check(clock) {
    if (this.kind !== 'paired_cumulative') {
      const expired = this.active.get(clock - width);
      if (expired) {
        this.count[expired.cell]--;
        this.sum[expired.cell] -= expired.d;
        this.active.delete(clock - width);
      }
    }
    if (this.flagAt !== null) return;
    const cells = this.kind === 'scalar_window' ? 1 : 2;
    for (let x = 0; x < cells; x++) {
      if (this.count[x] > 0 &&
          Math.abs(this.sum[x] / this.count[x]) > epsilon + radius(this.count[x], cells)) {
        this.flagAt = clock;
        return;
      }
    }
  }
}

function negativeControls() {
  const g = new Gate('paired_window');
  const p = { nominated: true, missing: false, origin: 1, arrival: 3, x: 0, d: 1 };
  assert.throws(() => g.receive(2, p));
  assert.throws(() => g.receive(3, { ...p, nominated: false }));
  assert.throws(() => g.receive(3, { ...p, missing: true }));
  assert.throws(() => g.receive(3, { ...p, origin: 4 }));
  g.receive(3, p);
  assert.throws(() => g.receive(3, p));
  g.check(129);
  assert.deepEqual(g.count, [0, 0]);
  assert.deepEqual(g.sum, [0, 0]);
  const h = new Gate('paired_window');
  h.receive(129, { ...p, arrival: 129 });
  assert.deepEqual(h.count, [0, 0]);
}

function shifted(scenario, side, clock) {
  return (scenario === 'live_shift0' && side === 'live') ||
    clock >= 256 && ((scenario === 'live_shift256' && side === 'live') ||
      (scenario === 'reference_shift256' && side === 'reference') ||
      scenario === 'common_shift256');
}

function run(split, schedule, scenario, trial) {
  const draw = {};
  for (const role of ['context', 'reference-outcome', 'live-outcome', 'nomination',
    'reference-missing', 'live-missing', 'reference-delay', 'live-delay']) {
    draw[role] = rng(split, schedule, scenario, trial, role);
  }
  const gates = Object.fromEntries(arms.map(name => [name, new Gate(name)]));
  const arrivals = Array.from({ length: T + 32 }, () => []);
  let nominated = 0, missingPairs = 0, missingReference = 0, missingLive = 0;
  let observedReference = 0, observedLive = 0, deliveredPairs = 0;
  for (let clock = 0; clock < T; clock++) {
    const x = Number(draw.context() < .5);
    const y = {};
    for (const side of ['reference', 'live']) {
      const swap = shifted(scenario, side, clock);
      const p = x === 0 ? (swap ? .1 : .9) : (swap ? .9 : .1);
      y[side] = Number(draw[`${side}-outcome`]() < p);
    }
    const nominate = schedule === 'complete' || draw.nomination() < .25;
    if (nominate) {
      nominated++;
      const mr = schedule === 'sparse_delayed' && draw['reference-missing']() < .2;
      const ml = schedule === 'sparse_delayed' && draw['live-missing']() < .2;
      missingReference += Number(mr);
      missingLive += Number(ml);
      missingPairs += Number(mr || ml);
      const dr = schedule === 'complete' ? 0 : Math.floor(draw['reference-delay']() * 32);
      const dl = schedule === 'complete' ? 0 : Math.floor(draw['live-delay']() * 32);
      if (!mr && clock + dr < T) arrivals[clock + dr].push({ origin: clock, side: 'reference', x, y: y.reference });
      if (!ml && clock + dl < T) arrivals[clock + dl].push({ origin: clock, side: 'live', x, y: y.live });
    }
    // Only paired observations sharing an origin and context may affect a gate.
    for (const item of arrivals[clock]) {
      const slot = arrivals.pairs ??= new Map();
      const pair = slot.get(item.origin) ?? { origin: item.origin, x: item.x };
      assert.equal(pair.x, item.x);
      assert(!Object.hasOwn(pair, item.side));
      pair[item.side] = item.y;
      slot.set(item.origin, pair);
      if (item.side === 'reference') observedReference++;
      else observedLive++;
      if (Object.hasOwn(pair, 'reference') && Object.hasOwn(pair, 'live')) {
        const evidence = { nominated: true, missing: false, origin: pair.origin,
          arrival: clock, x: pair.x, d: pair.reference - pair.live };
        for (const gate of Object.values(gates)) gate.receive(clock, evidence);
        deliveredPairs++;
        slot.delete(item.origin);
      }
    }
    for (const gate of Object.values(gates)) gate.check(clock);
  }
  for (const gate of Object.values(gates)) assert.equal(gate.seen.size, deliveredPairs);
  assert(observedReference <= nominated - missingReference);
  assert(observedLive <= nominated - missingLive);
  return { kind: 'trial', split, schedule, scenario, trial,
    flags: Object.fromEntries(arms.map(name => [name, gates[name].flagAt])),
    counts: Object.fromEntries(arms.map(name => [name, gates[name].count])),
    nominated, missingPairs, missingReference, missingLive,
    observedReference, observedLive, deliveredPairs,
    pendingPairs: nominated - missingPairs - deliveredPairs,
    contextReadings: 2 * nominated,
    observedLabels: observedReference + observedLive,
    usableLabels: 2 * deliveredPairs };
}

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(values.length * p) - 1];
}

function summarize(rows) {
  const cells = [];
  for (const schedule of schedules) for (const scenario of cases) {
    const subset = rows.filter(r => r.schedule === schedule && r.scenario === scenario);
    assert.equal(subset.length, trials);
    const results = {};
    for (const arm of arms) {
      const flags = subset.map(r => r.flags[arm]).filter(v => v !== null);
      results[arm] = { flags: flags.length,
        prechangeFlags: scenario.includes('shift256') ? flags.filter(t => t < 256).length : null,
        p50: quantile(flags, .5), p95: quantile(flags, .95) };
    }
    const avg = key => subset.reduce((s, r) => s + r[key], 0) / trials;
    const row = { schedule, scenario, trials, results,
      meanNominated: avg('nominated'), meanMissingPairs: avg('missingPairs'),
      meanPendingPairs: avg('pendingPairs'), meanDeliveredPairs: avg('deliveredPairs'),
      meanContextReadings: avg('contextReadings'), meanObservedLabels: avg('observedLabels'),
      meanUsableLabels: avg('usableLabels') };
    const candidate = results.paired_window;
    row.pass = scenario === 'stable' || scenario === 'common_shift256' ? candidate.flags <= 20 :
      scenario === 'live_shift0' ? candidate.flags >= 800 :
      candidate.flags >= 800 && candidate.prechangeFlags === 0 &&
      (schedule !== 'complete' || candidate.p50 - 256 <= 192);
    cells.push(row);
  }
  return { kind: 'summary', cells, pass: cells.every(c => c.pass) };
}

negativeControls();
if (process.argv.length !== 4 || !Object.hasOwn(seeds, process.argv[2])) {
  throw Error('usage: node research/paired-context-v3.mjs design|confirmation NEW.jsonl');
}
const split = process.argv[2], start = process.hrtime.bigint();
const rows = [];
for (const schedule of schedules) for (const scenario of cases) {
  for (let trial = 0; trial < trials; trial++) rows.push(run(split, schedule, scenario, trial));
}
const summary = summarize(rows);
const header = { kind: 'header', split, T, trials, delta, epsilon, width,
  seedBase: seeds[split],
  protocolSHA256: createHash('sha256').update(readFileSync(protocol)).digest('hex'),
  sourceSHA256: createHash('sha256').update(readFileSync(import.meta.filename)).digest('hex') };
const fd = openSync(process.argv[3], 'wx', 0o600);
try {
  for (const row of [header, ...rows, { ...summary,
    elapsedMs: Number(process.hrtime.bigint() - start) / 1e6 }]) {
    writeSync(fd, JSON.stringify(row) + '\n');
  }
} finally { closeSync(fd); }
console.log(JSON.stringify({ split, pass: summary.pass, elapsedMs: Number(process.hrtime.bigint() - start) / 1e6,
  cells: summary.cells.map(c => [c.schedule, c.scenario, c.results, c.pass]) }, null, 2));
