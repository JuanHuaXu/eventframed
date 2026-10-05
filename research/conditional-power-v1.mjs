import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { openSync, readFileSync, writeSync, closeSync } from 'node:fs';

const T = 512;
const trials = 1000;
const delta = .02;
const epsilon = .10;
const batchN = 256;
const strategies = ['batch_reference', 'online_reference'];
const schedules = [
  { name: 'immediate', audit: .25, missing: 0, maxDelay: 0 },
  { name: 'sparse_delayed', audit: .25, missing: .20, maxDelay: 31 },
];
const cases = ['stable', 'swapped_from_start', 'swapped_at_256'];
const seeds = { design: 2026100301, confirmation: 2026100302 };
const protocol = 'docs/experiments/mmm-conditional-power-v1-protocol.md';
const radLog = Math.log(8 * (T + 1) / delta);

function rng(split, schedule, trial, role) {
  const hash = createHash('sha256')
    .update(JSON.stringify([seeds[split], schedule, trial, role])).digest();
  let state = hash.readUInt32LE(0);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let x = state;
    x = Math.imul(x ^ (x >>> 15), x | 1);
    x ^= x + Math.imul(x ^ (x >>> 7), x | 61);
    return ((x ^ (x >>> 14)) >>> 0) / 4294967296;
  };
}

function rad(n) {
  return Math.sqrt(radLog / (2 * n));
}

class Gate {
  constructor(strategy, batch) {
    this.reference = {
      count: strategy === 'batch_reference' ? [batchN, batchN] : [0, 0],
      yes: strategy === 'batch_reference' ? [...batch] : [0, 0],
      origins: new Set(),
    };
    this.live = { count: [0, 0], yes: [0, 0], origins: new Set() };
    this.flagAt = null;
  }

  receive(clock, side, e) {
    if (side !== 'reference' && side !== 'live') throw Error('invalid side');
    const s = this[side];
    if (!e.nominated || e.missing || e.arrival > clock || e.origin > clock ||
        e.origin < 0 || e.x < 0 || e.x > 1 || e.y < 0 || e.y > 1 ||
        s.origins.has(e.origin)) throw Error('invalid audit delivery');
    s.origins.add(e.origin);
    s.count[e.x]++;
    s.yes[e.x] += e.y;
  }

  check(clock) {
    if (this.flagAt !== null) return;
    for (let x = 0; x < 2; x++) {
      const nr = this.reference.count[x], nl = this.live.count[x];
      if (!nr || !nl) continue;
      const gap = Math.abs(this.reference.yes[x] / nr - this.live.yes[x] / nl);
      if (gap > epsilon + rad(nr) + rad(nl)) {
        this.flagAt = clock;
        return;
      }
    }
  }
}

function negativeControls() {
  const gate = new Gate('online_reference', [0, 0]);
  const e = { origin: 0, arrival: 2, x: 0, y: 1, nominated: true, missing: false };
  assert.throws(() => gate.receive(1, 'live', e));
  assert.throws(() => gate.receive(2, 'live', { ...e, nominated: false }));
  assert.throws(() => gate.receive(2, 'live', { ...e, missing: true }));
  gate.receive(2, 'live', e);
  assert.throws(() => gate.receive(2, 'live', e));
  assert.deepEqual(gate.live.count, [1, 0]);
}

function run(split, strategy, schedule, regime, trial) {
  const role = name => rng(split, schedule.name, trial, name);
  const draw = {
    referenceX: role('reference-context'), referenceY: role('reference-outcome'),
    referenceAudit: role('reference-audit'), referenceMissing: role('reference-missing'),
    referenceDelay: role('reference-delay'), liveX: role('live-context'),
    liveY: role('live-outcome'), liveAudit: role('live-audit'),
    liveMissing: role('live-missing'), liveDelay: role('live-delay'),
    batch: role('batch-reference'),
  };
  const batch = [0, 0];
  if (strategy === 'batch_reference') {
    for (let x = 0; x < 2; x++) {
      for (let i = 0; i < batchN; i++) batch[x] += Number(draw.batch() < (x === 0 ? .9 : .1));
    }
  }
  const gate = new Gate(strategy, batch);
  const arrivals = Array.from({ length: T + 32 }, () => []);
  const nominated = { reference: 0, live: 0 };
  const missing = { reference: 0, live: 0 };
  const delivered = { reference: 0, live: 0 };

  for (let clock = 0; clock < T; clock++) {
    for (const side of ['reference', 'live']) {
      if (side === 'reference' && strategy === 'batch_reference') continue;
      const x = Number(draw[`${side}X`]() < .5);
      const swapped = side === 'live' &&
        (regime === 'swapped_from_start' || regime === 'swapped_at_256' && clock >= 256);
      const p = x === 0 ? (swapped ? .1 : .9) : (swapped ? .9 : .1);
      const y = Number(draw[`${side}Y`]() < p);
      if (draw[`${side}Audit`]() >= schedule.audit) continue;
      nominated[side]++;
      if (draw[`${side}Missing`]() < schedule.missing) {
        missing[side]++;
        continue;
      }
      const arrival = clock + Math.floor(draw[`${side}Delay`]() * (schedule.maxDelay + 1));
      arrivals[arrival].push({ side, evidence: {
        origin: clock, arrival, x, y, nominated: true, missing: false,
      } });
    }
    for (const { side, evidence } of arrivals[clock]) {
      gate.receive(clock, side, evidence);
      delivered[side]++;
    }
    gate.check(clock);
  }
  assert.equal(delivered.reference, gate.reference.origins.size);
  assert.equal(delivered.live, gate.live.origins.size);
  for (const side of ['reference', 'live']) {
    assert(delivered[side] <= nominated[side] - missing[side]);
  }
  return {
    kind: 'trial', split, strategy, schedule: schedule.name, regime, trial,
    flagAt: gate.flagAt, referenceCount: gate.reference.count,
    liveCount: gate.live.count, referenceYes: gate.reference.yes,
    liveYes: gate.live.yes, nominated, missing, delivered,
    pending: {
      reference: nominated.reference - missing.reference - delivered.reference,
      live: nominated.live - missing.live - delivered.live,
    },
    batchLabels: strategy === 'batch_reference' ? 2 * batchN : 0,
  };
}

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(sorted.length * p) - 1];
}
function summarize(rows) {
  const cells = [];
  for (const strategy of strategies) for (const schedule of schedules) for (const regime of cases) {
    const subset = rows.filter(r => r.strategy === strategy && r.schedule === schedule.name && r.regime === regime);
    assert.equal(subset.length, trials);
    const flags = subset.map(r => r.flagAt).filter(v => v !== null);
    const average = fn => subset.reduce((a, r) => a + fn(r), 0) / trials;
    const cell = {
      strategy, schedule: schedule.name, regime, trials,
      flags: flags.length,
      prechangeFlags: regime === 'swapped_at_256' ? flags.filter(t => t < 256).length : null,
      detectionP50: quantile(flags, .5), detectionP95: quantile(flags, .95),
      meanReferenceLabels: average(r => r.batchLabels + r.delivered.reference),
      meanLiveLabels: average(r => r.delivered.live),
      meanReferenceNominated: average(r => r.nominated.reference),
      meanLiveNominated: average(r => r.nominated.live),
      meanReferenceMissing: average(r => r.missing.reference),
      meanLiveMissing: average(r => r.missing.live),
      meanReferencePending: average(r => r.pending.reference),
      meanLivePending: average(r => r.pending.live),
    };
    cell.pass = regime === 'stable' ? flags.length <= 20 :
      regime === 'swapped_from_start' ? flags.length >= 800 :
      flags.length >= 800 && cell.prechangeFlags === 0;
    cells.push(cell);
  }
  return { cells, pass: cells.every(c => c.pass) };
}

negativeControls();
if (process.argv.length !== 4 || !Object.hasOwn(seeds, process.argv[2])) {
  throw Error('usage: node research/conditional-power-v1.mjs design|confirmation NEW.jsonl');
}
const split = process.argv[2];
const started = process.hrtime.bigint();
const rows = [];
for (const strategy of strategies) for (const schedule of schedules) for (const regime of cases) {
  for (let trial = 0; trial < trials; trial++) rows.push(run(split, strategy, schedule, regime, trial));
}
const summary = summarize(rows);
const elapsedMs = Number(process.hrtime.bigint() - started) / 1e6;
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
