import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const protocolPath = 'docs/experiments/mmm-conditional-gate-v1-protocol.md';
const T = 512, trials = 1000, referencePerCell = 256;
const delta = 0.02, epsilon = 0.10, baseSeed = 2026100101;
const schedules = [
  { name: 'complete-immediate', audit: 1, missing: 0, maxDelay: 0 },
  { name: 'sparse-delayed', audit: 0.25, missing: 0.20, maxDelay: 31 },
];
const regimes = ['stable', 'swapped'];

function seed(...parts) {
  return crypto.createHash('sha256').update(JSON.stringify([baseSeed, ...parts])).digest().readUInt32LE(0);
}

function random(seedValue) {
  let state = seedValue >>> 0;
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let z = state;
    z = Math.imul(z ^ (z >>> 15), z | 1);
    z ^= z + Math.imul(z ^ (z >>> 7), z | 61);
    return ((z ^ (z >>> 14)) >>> 0) / 2 ** 32;
  };
}

function radius(k, n) {
  assert(n > 0);
  return Math.sqrt(Math.log(4 * k * (T + 1) / delta) / (2 * n));
}

class AuditMonitor {
  constructor(kind, reference) {
    this.kind = kind;
    this.reference = reference;
    this.count = [0, 0];
    this.success = [0, 0];
    this.ids = new Set();
    this.firstFlag = null;
    this.flagLabels = null;
  }

  observe(clock, event) {
    if (!event.nominated || event.missing || event.origin > clock ||
        event.arrival > clock || this.ids.has(event.origin) ||
        ![0, 1].includes(event.x) || ![0, 1].includes(event.y)) {
      throw Error('unauthorized or duplicate audit evidence');
    }
    this.ids.add(event.origin);
    this.count[event.x]++;
    this.success[event.x] += event.y;
  }

  test(clock) {
    if (this.firstFlag !== null) return;
    if (this.kind === 'conditional') {
      for (let x = 0; x < 2; x++) {
        if (!this.count[x]) continue;
        const live = this.success[x] / this.count[x];
        const reference = this.reference.success[x] / referencePerCell;
        if (Math.abs(live - reference) > epsilon + radius(2, referencePerCell) + radius(2, this.count[x])) {
          this.firstFlag = clock;
          this.flagLabels = this.count[0] + this.count[1];
          return;
        }
      }
      return;
    }
    const n = this.count[0] + this.count[1];
    if (!n) return;
    const live = (this.success[0] + this.success[1]) / n;
    const reference = (this.reference.success[0] + this.reference.success[1]) / (2 * referencePerCell);
    if (Math.abs(live - reference) > epsilon + radius(1, 2 * referencePerCell) + radius(1, n)) {
      this.firstFlag = clock;
      this.flagLabels = n;
    }
  }
}

function makeReference(trial, schedule, regime) {
  const rng = random(seed('reference', schedule, regime, trial));
  const success = [0, 0];
  for (let x = 0; x < 2; x++) {
    for (let i = 0; i < referencePerCell; i++) success[x] += +(rng() < (x === 0 ? .9 : .1));
  }
  return { success };
}

function runTrial(trial, schedule, regime) {
  const reference = makeReference(trial, schedule.name, regime);
  const rng = random(seed('live', schedule.name, regime, trial));
  const scalar = new AuditMonitor('scalar', reference);
  const conditional = new AuditMonitor('conditional', reference);
  const arrivals = Array.from({ length: T + 32 }, () => []);
  let nominated = 0, missing = 0, delivered = 0;
  for (let clock = 0; clock < T; clock++) {
    const x = +(rng() < .5);
    const p = regime === 'stable' ? (x === 0 ? .9 : .1) : (x === 0 ? .1 : .9);
    const y = +(rng() < p);
    if (rng() < schedule.audit) {
      nominated++;
      if (rng() < schedule.missing) {
        missing++;
      } else {
        const delay = Math.floor(rng() * (schedule.maxDelay + 1));
        const arrival = clock + delay;
        arrivals[arrival].push({ origin: clock, arrival, x, y, nominated: true, missing: false });
      }
    }
    for (const event of arrivals[clock]) {
      scalar.observe(clock, event);
      conditional.observe(clock, event);
      delivered++;
    }
    scalar.test(clock);
    conditional.test(clock);
    assert.equal(scalar.ids.size, conditional.ids.size);
    assert.deepEqual(scalar.count, conditional.count);
    assert.deepEqual(scalar.success, conditional.success);
    assert(scalar.ids.size <= nominated - missing);
  }
  assert.equal(delivered, scalar.ids.size);
  return {
    trial, schedule: schedule.name, regime,
    reference: reference.success,
    nominated, missing, delivered, pending: nominated - missing - delivered,
    scalarClock: scalar.firstFlag, conditionalClock: conditional.firstFlag,
    scalarLabels: scalar.flagLabels, conditionalLabels: conditional.flagLabels,
    observedByCell: conditional.count,
  };
}

function negativeControls() {
  const m = new AuditMonitor('conditional', { success: [230, 26] });
  const e = { origin: 0, arrival: 10, x: 0, y: 0, nominated: true, missing: false };
  assert.throws(() => m.observe(0, e), /unauthorized/);
  assert.equal(m.ids.size, 0);
  assert.throws(() => m.observe(10, { ...e, nominated: false }), /unauthorized/);
  assert.throws(() => m.observe(10, { ...e, missing: true }), /unauthorized/);
  m.observe(10, e);
  assert.throws(() => m.observe(10, e), /duplicate/);
  assert.equal(m.ids.size, 1);
  assert.equal((.9 + .1) / 2, (.1 + .9) / 2);
  assert.equal((Math.abs(.9 - .1) + Math.abs(.1 - .9)) / 2, .8);
}

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(p * sorted.length) - 1];
}

function summarize(rows) {
  const cells = [];
  for (const schedule of schedules) for (const regime of regimes) {
    const rs = rows.filter(r => r.schedule === schedule.name && r.regime === regime);
    assert.equal(rs.length, trials);
    const scalar = rs.filter(r => r.scalarClock !== null);
    const conditional = rs.filter(r => r.conditionalClock !== null);
    const mean = key => rs.reduce((v, r) => v + r[key], 0) / rs.length;
    const cell = {
      schedule: schedule.name, regime, trials: rs.length,
      scalarFlags: scalar.length, conditionalFlags: conditional.length,
      meanNominated: mean('nominated'), meanMissing: mean('missing'),
      meanDelivered: mean('delivered'), meanPending: mean('pending'),
      scalarClockP50: quantile(scalar.map(r => r.scalarClock), .5),
      conditionalClockP50: quantile(conditional.map(r => r.conditionalClock), .5),
      conditionalClockP95: quantile(conditional.map(r => r.conditionalClock), .95),
      conditionalLabelsP50: quantile(conditional.map(r => r.conditionalLabels), .5),
    };
    cell.pass = regime === 'stable'
      ? scalar.length <= 20 && conditional.length <= 20
      : scalar.length <= 20 && conditional.length >= 900;
    cells.push(cell);
  }
  return { cells, pass: cells.every(c => c.pass) };
}

negativeControls();
const rows = [];
const start = process.hrtime.bigint();
for (const schedule of schedules) for (const regime of regimes) {
  for (let trial = 0; trial < trials; trial++) rows.push(runTrial(trial, schedule, regime));
}
const result = summarize(rows);
const elapsedMs = Number(process.hrtime.bigint() - start) / 1e6;
const header = {
  kind: 'header', protocolSHA256: crypto.createHash('sha256').update(fs.readFileSync(protocolPath)).digest('hex'),
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),
  T, trials, referencePerCell, delta, epsilon, baseSeed, schedules,
};
if (process.argv[2]) {
  const fd = fs.openSync(process.argv[2], 'wx', 0o600);
  try {
    fs.writeSync(fd, JSON.stringify(header) + '\n');
    for (const row of rows) fs.writeSync(fd, JSON.stringify({ kind: 'trial', ...row }) + '\n');
    fs.writeSync(fd, JSON.stringify({ kind: 'summary', ...result, elapsedMs }) + '\n');
  } finally {
    fs.closeSync(fd);
  }
}
console.log(JSON.stringify({ ...result, elapsedMs }, null, 2));
