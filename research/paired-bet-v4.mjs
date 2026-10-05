import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { closeSync, openSync, readFileSync, writeSync } from 'node:fs';

const T = 512, N = 1000, width = 128, epsilon = .10, delta = .02, lambda = .80;
const starts = [0, 64, 128, 192, 256, 320, 384, 448];
const cases = ['stable', 'boundary', 'common_shift256', 'live_shift256',
  'reference_shift256', 'live_shift224', 'live_shift0'];
const schedules = ['complete', 'sparse_delayed'];
const arms = ['paired_bet', 'conditional_window', 'scalar_window'];
const seedBase = { design: 2026100601, confirmation: 2026100602 };
const protocol = 'docs/experiments/mmm-paired-bet-v4-protocol.md';

function rng(split, schedule, scenario, trial, role) {
  const hash = createHash('sha256')
    .update(JSON.stringify([seedBase[split], schedule, scenario, trial, role])).digest();
  let state = hash.readUInt32LE(0);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let x = state;
    x = Math.imul(x ^ (x >>> 15), x | 1);
    x ^= x + Math.imul(x ^ (x >>> 7), x | 61);
    return ((x ^ (x >>> 14)) >>> 0) / 4294967296;
  };
}

function validPair(clock, p, seen) {
  if (p.complete !== true || p.nominated !== true || p.missing !== false ||
      !Number.isInteger(p.origin) || p.origin < 0 || p.origin >= T || p.origin > clock ||
      !Number.isInteger(p.arrival) || p.arrival > clock || p.arrival < p.origin ||
      (p.x !== 0 && p.x !== 1) || ![-1, 0, 1].includes(p.d) || seen.has(p.origin)) {
    throw Error('invalid or incomplete paired evidence');
  }
  seen.add(p.origin);
}

class BetGate {
  constructor() {
    this.wealth = starts.map(() => [[1, 1], [1, 1]]);
    this.seen = new Set();
    this.flagAt = null;
    this.factorUpdates = 0;
  }

  receive(clock, p) {
    validPair(clock, p, this.seen);
    for (let j = 0; j < starts.length; j++) {
      if (p.origin < starts[j]) continue;
      for (const [k, sign] of [[0, -1], [1, 1]]) {
        const factor = 1 + lambda * (sign * p.d - epsilon);
        assert(factor >= .12 - 1e-12 && factor <= 1.72 + 1e-12);
        this.wealth[j][p.x][k] *= factor;
        this.factorUpdates++;
      }
    }
  }

  check(clock) {
    if (this.flagAt !== null) return;
    let sum = 0;
    for (const start of this.wealth) for (const cell of start) for (const w of cell) sum += w;
    if (sum / (starts.length * 2 * 2) >= 1 / delta) this.flagAt = clock;
  }
}

class WindowGate {
  constructor(scalar) {
    this.scalar = scalar;
    this.count = [0, 0];
    this.sum = [0, 0];
    this.active = new Map();
    this.seen = new Set();
    this.flagAt = null;
  }

  receive(clock, p) {
    validPair(clock, p, this.seen);
    if (p.origin < clock - width + 1) return;
    const x = this.scalar ? 0 : p.x;
    this.count[x]++;
    this.sum[x] += p.d;
    this.active.set(p.origin, { x, d: p.d });
  }

  check(clock) {
    const old = this.active.get(clock - width);
    if (old) {
      this.count[old.x]--;
      this.sum[old.x] -= old.d;
      this.active.delete(clock - width);
    }
    if (this.flagAt !== null) return;
    const cells = this.scalar ? 1 : 2;
    for (let x = 0; x < cells; x++) {
      const n = this.count[x];
      const r = n ? Math.sqrt(2 * Math.log(2 * cells * (T + 1) / delta) / n) : Infinity;
      if (n && Math.abs(this.sum[x] / n) > epsilon + r) {
        this.flagAt = clock;
        return;
      }
    }
  }
}

function negativeControls() {
  const g = new BetGate();
  const p = { complete: true, nominated: true, missing: false,
    origin: 1, arrival: 3, x: 0, d: 1 };
  const before = JSON.stringify(g.wealth);
  for (const bad of [{ ...p, complete: false }, { ...p, nominated: false },
    { ...p, missing: true }, { ...p, origin: 4 }]) {
    assert.throws(() => g.receive(3, bad));
    assert.equal(JSON.stringify(g.wealth), before);
  }
  assert.throws(() => g.receive(2, p));
  g.receive(3, p);
  assert.throws(() => g.receive(3, p));
  assert.equal(g.factorUpdates, 2);
  assert.equal(g.wealth[0][0][1], 1 + lambda * (1 - epsilon));
  assert.equal(g.wealth[1][0][1], 1);
  const w = new WindowGate(false);
  w.receive(3, p);
  w.check(129);
  assert.deepEqual(w.count, [0, 0]);
}

function eventProbability(scenario, side, x, clock) {
  if (scenario === 'boundary') return side === 'reference' ? (x === 0 ? .55 : .45) :
    (x === 0 ? .45 : .55);
  const shift = (scenario === 'live_shift0' && side === 'live') ||
    (scenario === 'live_shift224' && side === 'live' && clock >= 224) ||
    (scenario === 'live_shift256' && side === 'live' && clock >= 256) ||
    (scenario === 'reference_shift256' && side === 'reference' && clock >= 256) ||
    (scenario === 'common_shift256' && clock >= 256);
  return x === 0 ? (shift ? .1 : .9) : (shift ? .9 : .1);
}

function run(split, schedule, scenario, trial) {
  const draw = {};
  for (const role of ['context', 'reference-outcome', 'live-outcome', 'nomination',
    'reference-missing', 'live-missing', 'reference-delay', 'live-delay']) {
    draw[role] = rng(split, schedule, scenario, trial, role);
  }
  const gates = { paired_bet: new BetGate(), conditional_window: new WindowGate(false),
    scalar_window: new WindowGate(true) };
  const arrivals = Array.from({ length: T }, () => []);
  const partial = new Map();
  let nominated = 0, missingPairs = 0, missingReference = 0, missingLive = 0;
  let observedReference = 0, observedLive = 0, deliveredPairs = 0;
  for (let clock = 0; clock < T; clock++) {
    const x = Number(draw.context() < .5);
    const y = {};
    for (const side of ['reference', 'live']) {
      y[side] = Number(draw[`${side}-outcome`]() < eventProbability(scenario, side, x, clock));
    }
    if (schedule === 'complete' || draw.nomination() < .25) {
      nominated++;
      const mr = schedule === 'sparse_delayed' && draw['reference-missing']() < .2;
      const ml = schedule === 'sparse_delayed' && draw['live-missing']() < .2;
      missingReference += Number(mr);
      missingLive += Number(ml);
      missingPairs += Number(mr || ml);
      for (const side of ['reference', 'live']) {
        if (side === 'reference' && mr || side === 'live' && ml) continue;
        const delay = schedule === 'complete' ? 0 : Math.floor(draw[`${side}-delay`]() * 32);
        if (clock + delay < T) arrivals[clock + delay].push({ origin: clock, x, side, y: y[side] });
      }
    }
    for (const item of arrivals[clock]) {
      if (item.side === 'reference') observedReference++;
      else observedLive++;
      const pair = partial.get(item.origin) ?? { origin: item.origin, x: item.x };
      assert.equal(pair.x, item.x);
      assert(!Object.hasOwn(pair, item.side));
      pair[item.side] = item.y;
      partial.set(item.origin, pair);
      if (Object.hasOwn(pair, 'reference') && Object.hasOwn(pair, 'live')) {
        const p = { complete: true, nominated: true, missing: false, origin: pair.origin,
          arrival: clock, x: pair.x, d: pair.reference - pair.live };
        for (const gate of Object.values(gates)) gate.receive(clock, p);
        deliveredPairs++;
        partial.delete(item.origin);
      }
    }
    for (const gate of Object.values(gates)) gate.check(clock);
  }
  for (const gate of Object.values(gates)) assert.equal(gate.seen.size, deliveredPairs);
  assert(observedReference <= nominated - missingReference);
  assert(observedLive <= nominated - missingLive);
  return { kind: 'trial', split, schedule, scenario, trial,
    flags: Object.fromEntries(arms.map(a => [a, gates[a].flagAt])),
    factorUpdates: gates.paired_bet.factorUpdates,
    nominated, missingPairs, missingReference, missingLive,
    observedReference, observedLive, deliveredPairs,
    pendingPairs: nominated - missingPairs - deliveredPairs,
    contextReadings: 2 * nominated,
    observedLabels: observedReference + observedLive,
    usableLabels: 2 * deliveredPairs };
}

function quantile(values, p) {
  if (!values.length) return null;
  const a = [...values].sort((x, y) => x - y);
  return a[Math.ceil(a.length * p) - 1];
}

function changeAt(scenario) {
  if (scenario === 'live_shift224') return 224;
  if (scenario.endsWith('shift256')) return 256;
  return null;
}

function summarize(rows) {
  const cells = [];
  for (const schedule of schedules) for (const scenario of cases) {
    const subset = rows.filter(r => r.schedule === schedule && r.scenario === scenario);
    assert.equal(subset.length, N);
    const results = {};
    for (const arm of arms) {
      const flags = subset.map(r => r.flags[arm]).filter(v => v !== null);
      const onset = changeAt(scenario);
      results[arm] = { flags: flags.length,
        prechangeFlags: onset === null ? null : flags.filter(t => t < onset).length,
        p50: quantile(flags, .5), p95: quantile(flags, .95) };
    }
    const mean = key => subset.reduce((s, r) => s + r[key], 0) / N;
    const c = { schedule, scenario, trials: N, results,
      meanNominated: mean('nominated'), meanMissingPairs: mean('missingPairs'),
      meanPendingPairs: mean('pendingPairs'), meanDeliveredPairs: mean('deliveredPairs'),
      meanContextReadings: mean('contextReadings'), meanObservedLabels: mean('observedLabels'),
      meanUsableLabels: mean('usableLabels'), meanFactorUpdates: mean('factorUpdates') };
    const candidate = results.paired_bet, onset = changeAt(scenario);
    c.pass = ['stable', 'boundary', 'common_shift256'].includes(scenario) ?
      candidate.flags <= 20 :
      candidate.flags >= 800 && (onset === null || candidate.prechangeFlags === 0 &&
        candidate.p50 - onset <= (schedule === 'complete' ? 192 : 224));
    cells.push(c);
  }
  return { kind: 'summary', cells, pass: cells.every(c => c.pass) };
}

negativeControls();
if (process.argv.length !== 4 || !Object.hasOwn(seedBase, process.argv[2])) {
  throw Error('usage: node research/paired-bet-v4.mjs design|confirmation NEW.jsonl');
}
const split = process.argv[2], start = process.hrtime.bigint();
const rows = [];
for (const schedule of schedules) for (const scenario of cases) {
  for (let trial = 0; trial < N; trial++) rows.push(run(split, schedule, scenario, trial));
}
const summary = summarize(rows);
const header = { kind: 'header', split, T, trials: N, width, epsilon, delta, lambda,
  starts, seedBase: seedBase[split],
  protocolSHA256: createHash('sha256').update(readFileSync(protocol)).digest('hex'),
  sourceSHA256: createHash('sha256').update(readFileSync(import.meta.filename)).digest('hex') };
const fd = openSync(process.argv[3], 'wx', 0o600);
try {
  for (const row of [header, ...rows,
    { ...summary, elapsedMs: Number(process.hrtime.bigint() - start) / 1e6 }]) {
    writeSync(fd, JSON.stringify(row) + '\n');
  }
} finally { closeSync(fd); }
console.log(JSON.stringify({ split, pass: summary.pass, elapsedMs: Number(process.hrtime.bigint() - start) / 1e6,
  cells: summary.cells.map(c => [c.schedule, c.scenario, c.results, c.pass]) }, null, 2));
