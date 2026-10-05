import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const T = 512;
const trials = 1000;
const delta = 0.02;
const epsilon = 0.10;
const protocol = 'docs/experiments/mmm-conditional-transport-v1-protocol.md';
const splits = [{ name: 'design', seed: 2026100131 }, { name: 'confirmation', seed: 2026100132 }];
const schedules = [
  { name: 'immediate', missing: 0, maxDelay: 0 },
  { name: 'sparse_delayed', missing: 0.20, maxDelay: 31 },
];
const cases = ['stable', 'live_shift256', 'common_shift256', 'live_shift0'];

function rng(split, schedule, scenario, trial, role) {
  const h = crypto.createHash('sha256').update(JSON.stringify([split.seed, schedule.name, scenario, trial, role])).digest();
  let state = h.readUInt32LE(0);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let z = state;
    z = Math.imul(z ^ (z >>> 15), z | 1);
    z ^= z + Math.imul(z ^ (z >>> 7), z | 61);
    return ((z ^ (z >>> 14)) >>> 0) / 4294967296;
  };
}

function rad(n) {
  return Math.sqrt(Math.log(8 * (T + 1) / delta) / (2 * n));
}

function probability(x, shifted) {
  const parity = ((x >> 6) ^ (x >> 7) ^ (x >> 8)) & 1;
  const bit = shifted ? (x >> 2) & 1 : parity;
  return bit ? 0.95 : 0.05;
}

function run(split, schedule, scenario, trial) {
  const xRNG = [rng(split, schedule, scenario, trial, 'live_context'), rng(split, schedule, scenario, trial, 'reference_context')];
  const yRNG = [rng(split, schedule, scenario, trial, 'live_outcome'), rng(split, schedule, scenario, trial, 'reference_outcome')];
  const auditRNG = rng(split, schedule, scenario, trial, 'audit');
  const missingRNG = [rng(split, schedule, scenario, trial, 'live_missing'), rng(split, schedule, scenario, trial, 'reference_missing')];
  const delayRNG = [rng(split, schedule, scenario, trial, 'live_delay'), rng(split, schedule, scenario, trial, 'reference_delay')];
  const arrivals = Array.from({ length: T + 32 }, () => []);
  const counts = [[0, 0], [0, 0]];
  const successes = [[0, 0], [0, 0]];
  let audits = 0, missing = 0, delivered = 0, flagAt = null;

  for (let clock = 0; clock < T; clock++) {
    const requested = auditRNG() < 0.25;
    if (requested) audits++;
    for (let side = 0; side < 2; side++) {
      const x = Math.floor(xRNG[side]() * 512);
      const shifted = side === 0
        ? scenario === 'live_shift0' || (clock >= 256 && scenario !== 'stable')
        : scenario === 'common_shift256' && clock >= 256;
      const y = +(yRNG[side]() < probability(x, shifted));
      if (!requested) continue;
      if (missingRNG[side]() < schedule.missing) {
        missing++;
        continue;
      }
      const due = clock + Math.floor(delayRNG[side]() * (schedule.maxDelay + 1));
      arrivals[due].push({ side, cell: (x >> 2) & 1, y, origin: clock });
    }
    for (const packet of arrivals[clock]) {
      assert(packet.origin <= clock);
      counts[packet.side][packet.cell]++;
      successes[packet.side][packet.cell] += packet.y;
      delivered++;
    }
    if (flagAt !== null) continue;
    for (let cell = 0; cell < 2; cell++) {
      const nl = counts[0][cell], nr = counts[1][cell];
      if (nl === 0 || nr === 0) continue;
      const gap = Math.abs(successes[0][cell] / nl - successes[1][cell] / nr);
      if (gap > epsilon + rad(nl) + rad(nr)) {
        flagAt = clock;
        break;
      }
    }
  }
  const pending = 2 * audits - missing - delivered;
  assert(pending >= 0);
  assert.equal(counts.flat().reduce((a, b) => a + b, 0), delivered);
  return { kind: 'trial', split: split.name, schedule: schedule.name, scenario, trial,
    flagAt, audits, missing, delivered, pending, counts, successes };
}

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(p * sorted.length) - 1];
}

function summarize(rows) {
  const cells = [];
  for (const split of splits) for (const schedule of schedules) for (const scenario of cases) {
    const rs = rows.filter(r => r.split === split.name && r.schedule === schedule.name && r.scenario === scenario);
    assert.equal(rs.length, trials);
    const flags = rs.filter(r => r.flagAt !== null);
    const average = f => rs.reduce((a, r) => a + f(r), 0) / trials;
    const nullCase = scenario === 'stable' || scenario === 'common_shift256';
    cells.push({ split: split.name, schedule: schedule.name, scenario,
      flags: flags.length, flagP50: quantile(flags.map(r => r.flagAt), 0.5),
      meanAudits: average(r => r.audits), meanMissing: average(r => r.missing),
      meanDelivered: average(r => r.delivered), meanPending: average(r => r.pending),
      meanCounts: [0, 1].map(side => [0, 1].map(cell => average(r => r.counts[side][cell]))),
      pass: nullCase ? flags.length <= 20 : flags.length >= 800 });
  }
  return { cells, pass: cells.every(c => c.pass) };
}

const rows = [];
for (const split of splits) for (const schedule of schedules) for (const scenario of cases) {
  for (let trial = 0; trial < trials; trial++) rows.push(run(split, schedule, scenario, trial));
}
const header = { kind: 'header', version: 'conditional-transport-v1', T, trials, delta, epsilon,
  splits, schedules, cases,
  protocolSHA256: crypto.createHash('sha256').update(fs.readFileSync(protocol)).digest('hex'),
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex') };
const summary = { kind: 'summary', ...summarize(rows) };
const bytes = [header, ...rows, summary].map(row => JSON.stringify(row)).join('\n') + '\n';
if (process.argv[2] === '--replay') {
  assert.equal(bytes, fs.readFileSync(process.argv[3], 'utf8'), 'full replay mismatch');
} else if (process.argv[2]) {
  fs.writeFileSync(process.argv[2], bytes, { flag: 'wx', mode: 0o600 });
}
console.log(JSON.stringify(summary, null, 2));
