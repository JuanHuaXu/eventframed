import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const T = 512, trials = 1000, referenceN = 256, delta = .02, epsilon = .10;
const seedBase = 2026100191;
const protocol = 'docs/experiments/mmm-conditional-gate-v2-protocol.md';
const schedules = [
  { name: 'complete-immediate', audit: 1, missing: 0, maxDelay: 0 },
  { name: 'sparse-delayed', audit: .25, missing: .20, maxDelay: 31 },
];

function rng(...parts) {
  const h = crypto.createHash('sha256').update(JSON.stringify([seedBase, ...parts])).digest();
  let state = h.readUInt32LE(0);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let z = state;
    z = Math.imul(z ^ (z >>> 15), z | 1);
    z ^= z + Math.imul(z ^ (z >>> 7), z | 61);
    return ((z ^ (z >>> 14)) >>> 0) / 4294967296;
  };
}

function rad(k, n) {
  if (n <= 0) throw Error('empty prefix');
  return Math.sqrt(Math.log(4 * k * (T + 1) / delta) / (2 * n));
}

class Gate {
  constructor(kind, reference) {
    this.kind = kind;
    this.reference = reference;
    this.n = [0, 0];
    this.s = [0, 0];
    this.origins = new Set();
    this.flagAt = null;
    this.flagLabels = null;
  }

  forecast(x) {
    if (x !== 0 && x !== 1) throw Error('invalid context');
    if (this.flagAt === null) return (1 + this.reference[x]) / (2 + referenceN);
    return (1 + this.s[x]) / (2 + this.n[x]);
  }

  receive(clock, e) {
    if (!e.nominated || e.missing || e.arrival > clock || e.origin > clock ||
        this.origins.has(e.origin) || (e.x !== 0 && e.x !== 1) ||
        (e.y !== 0 && e.y !== 1)) throw Error('unauthorized delayed evidence');
    this.origins.add(e.origin);
    this.n[e.x]++;
    this.s[e.x] += e.y;
  }

  check(clock) {
    if (this.flagAt !== null) return;
    let flag = false;
    if (this.kind === 'conditional') {
      for (let x = 0; x < 2; x++) {
        if (!this.n[x]) continue;
        const gap = Math.abs(this.s[x] / this.n[x] - this.reference[x] / referenceN);
        flag ||= gap > epsilon + rad(2, referenceN) + rad(2, this.n[x]);
      }
    } else {
      const n = this.n[0] + this.n[1];
      if (n) {
        const gap = Math.abs((this.s[0] + this.s[1]) / n -
          (this.reference[0] + this.reference[1]) / (2 * referenceN));
        flag = gap > epsilon + rad(1, 2 * referenceN) + rad(1, n);
      }
    }
    if (flag) {
      this.flagAt = clock;
      this.flagLabels = this.origins.size;
    }
  }
}

function referenceSample(schedule, regime, trial) {
  const draw = rng('reference', schedule, regime, trial);
  const out = [0, 0];
  for (let x = 0; x < 2; x++) {
    for (let i = 0; i < referenceN; i++) out[x] += +(draw() < (x === 0 ? .9 : .1));
  }
  return out;
}

function run(schedule, regime, trial) {
  const ref = referenceSample(schedule.name, regime, trial);
  const scalar = new Gate('scalar', ref), conditional = new Gate('conditional', ref);
  const input = rng('input', schedule.name, regime, trial);
  const outcome = rng('outcome', schedule.name, regime, trial);
  const audit = rng('audit', schedule.name, regime, trial);
  const missingDraw = rng('missing', schedule.name, regime, trial);
  const delayDraw = rng('delay', schedule.name, regime, trial);
  const arrivals = Array.from({ length: T + 32 }, () => []);
  let nominated = 0, missing = 0, delivered = 0;
  const brier = [0, 0], expectedBrier = [0, 0], correct = [0, 0];

  for (let clock = 0; clock < T; clock++) {
    const x = +(input() < .5);
    // Both forecasts are fixed before this event's outcome or same-tick
    // zero-delay audit can be delivered.
    const predictions = [scalar.forecast(x), conditional.forecast(x)];
    const q = regime === 'stable' ? (x === 0 ? .9 : .1) : (x === 0 ? .1 : .9);
    const y = +(outcome() < q);
    for (let arm = 0; arm < 2; arm++) {
      const p = predictions[arm];
      assert(p > 0 && p < 1);
      brier[arm] += (p - y) ** 2 / T;
      expectedBrier[arm] += ((p - q) ** 2 + q * (1 - q)) / T;
      correct[arm] += +((p >= .5) === !!y) / T;
    }
    if (audit() < schedule.audit) {
      nominated++;
      if (missingDraw() < schedule.missing) missing++;
      else {
        const arrival = clock + Math.floor(delayDraw() * (schedule.maxDelay + 1));
        arrivals[arrival].push({ origin: clock, arrival, x, y, nominated: true, missing: false });
      }
    }
    for (const event of arrivals[clock]) {
      scalar.receive(clock, event);
      conditional.receive(clock, event);
      delivered++;
    }
    scalar.check(clock);
    conditional.check(clock);
    assert.equal(scalar.origins.size, conditional.origins.size);
    assert.deepEqual(scalar.n, conditional.n);
    assert.deepEqual(scalar.s, conditional.s);
  }
  assert.equal(delivered, scalar.origins.size);
  assert(delivered <= nominated - missing);
  return {
    kind: 'trial', schedule: schedule.name, regime, trial,
    reference: ref, nominated, missing, delivered,
    pending: nominated - missing - delivered,
    scalarClock: scalar.flagAt, conditionalClock: conditional.flagAt,
    scalarFlagLabels: scalar.flagLabels, conditionalFlagLabels: conditional.flagLabels,
    observedByCell: conditional.n, brier, expectedBrier, accuracy: correct,
  };
}

function negativeControls() {
  const g = new Gate('conditional', [230, 25]);
  const e = { origin: 0, arrival: 5, x: 0, y: 1, nominated: true, missing: false };
  assert.throws(() => g.receive(0, e));
  assert.throws(() => g.receive(5, { ...e, nominated: false }));
  assert.throws(() => g.receive(5, { ...e, missing: true }));
  assert.equal(g.origins.size, 0);
  g.receive(5, e);
  assert.throws(() => g.receive(5, e));
  assert.equal(g.origins.size, 1);
}

function stats(values) {
  assert.equal(values.length, trials);
  const mean = values.reduce((a, b) => a + b, 0) / trials;
  const sumsq = values.reduce((a, b) => a + (b - mean) ** 2, 0);
  const se = Math.sqrt(sumsq / (trials - 1) / trials);
  return { mean, lower: mean - 3.5 * se, upper: mean + 3.5 * se };
}

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(p * sorted.length) - 1];
}

function summarize(rows) {
  const cells = [];
  for (const schedule of schedules) for (const regime of ['stable', 'swapped']) {
    const rs = rows.filter(r => r.schedule === schedule.name && r.regime === regime);
    assert.equal(rs.length, trials);
    const scalar = rs.filter(r => r.scalarClock !== null);
    const conditional = rs.filter(r => r.conditionalClock !== null);
    const gain = stats(rs.map(r => r.brier[0] - r.brier[1]));
    const expectedGain = stats(rs.map(r => r.expectedBrier[0] - r.expectedBrier[1]));
    const average = f => rs.reduce((a, r) => a + f(r), 0) / trials;
    const cell = {
      schedule: schedule.name, regime, trials,
      scalarFlags: scalar.length, conditionalFlags: conditional.length,
      gain, expectedGain,
      scalarBrier: average(r => r.brier[0]),
      conditionalBrier: average(r => r.brier[1]),
      scalarAccuracy: average(r => r.accuracy[0]),
      conditionalAccuracy: average(r => r.accuracy[1]),
      meanNominated: average(r => r.nominated),
      meanMissing: average(r => r.missing),
      meanDelivered: average(r => r.delivered),
      meanPending: average(r => r.pending),
      conditionalClockP50: quantile(conditional.map(r => r.conditionalClock), .5),
      conditionalClockP95: quantile(conditional.map(r => r.conditionalClock), .95),
    };
    cell.pass = regime === 'stable'
      ? scalar.length <= 20 && conditional.length <= 20 && -gain.lower <= .01
      : scalar.length <= 20 && conditional.length >= 900 && gain.mean >= .05 && gain.lower > 0;
    cells.push(cell);
  }
  return { cells, pass: cells.every(c => c.pass) };
}

negativeControls();
const start = process.hrtime.bigint();
const rows = [];
for (const schedule of schedules) for (const regime of ['stable', 'swapped']) {
  for (let trial = 0; trial < trials; trial++) rows.push(run(schedule, regime, trial));
}
const summary = summarize(rows);
const elapsedMs = Number(process.hrtime.bigint() - start) / 1e6;
const header = {
  kind: 'header',
  protocolSHA256: crypto.createHash('sha256').update(fs.readFileSync(protocol)).digest('hex'),
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),
  T, trials, referenceN, delta, epsilon, seedBase, schedules,
};
if (process.argv[2]) {
  const fd = fs.openSync(process.argv[2], 'wx', 0o600);
  try {
    fs.writeSync(fd, JSON.stringify(header) + '\n');
    for (const row of rows) fs.writeSync(fd, JSON.stringify(row) + '\n');
    fs.writeSync(fd, JSON.stringify({ kind: 'summary', ...summary, elapsedMs }) + '\n');
  } finally {
    fs.closeSync(fd);
  }
}
console.log(JSON.stringify({ ...summary, elapsedMs }, null, 2));
