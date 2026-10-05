import assert from 'node:assert/strict';
import fs from 'node:fs';
import {createInterface} from 'node:readline';

async function* rows(path) {
  for await (const line of createInterface({input: fs.createReadStream(path), crlfDelay: Infinity})) {
    if (line.trim()) yield JSON.parse(line);
  }
}
const pop = x => { let n = 0; for (; x; x &= x - 1) n++; return n; };
const masks = Array.from({length: 512}, (_, i) => i).filter(x => pop(x) <= 4);
const phi = x => masks.map(m => pop(m & ~x) % 2 ? -1 : 1);
assert.equal(masks.length, 256);
const dot = (a, b) => a.reduce((s, v, i) => s + v * b[i], 0);
const mv = (a, b) => a.map(r => dot(r, b));
const sigmoid = x => 1 / (1 + Math.exp(-x));
const errors = {};
function close(name, a, b, tol = 1e-7) {
  assert(Number.isFinite(a) && Number.isFinite(b), name);
  const e = Math.abs(a - b);
  errors[name] = Math.max(errors[name] ?? 0, e);
  assert(e <= tol, `${name}: ${a} vs ${b}, error ${e}`);
}

// Independent unscaled sample-space inverse with partial-pivot Gauss-Jordan.
// The implementation uses scaled Cholesky factors instead.
function inverse(a) {
  const n = a.length;
  const v = a.map((r, i) => [...r, ...Array.from({length: n}, (_, j) => +(i === j))]);
  let logdet = 0, sign = 1;
  for (let c = 0; c < n; c++) {
    let p = c;
    for (let i = c + 1; i < n; i++) if (Math.abs(v[i][c]) > Math.abs(v[p][c])) p = i;
    if (p !== c) { [v[c], v[p]] = [v[p], v[c]]; sign *= -1; }
    const pivot = v[c][c];
    assert(Number.isFinite(pivot) && Math.abs(pivot) > 1e-12);
    sign *= Math.sign(pivot); logdet += Math.log(Math.abs(pivot));
    for (let j = 0; j < 2 * n; j++) v[c][j] /= pivot;
    for (let i = 0; i < n; i++) if (i !== c) {
      const f = v[i][c];
      for (let j = 0; j < 2 * n; j++) v[i][j] -= f * v[c][j];
    }
  }
  assert.equal(sign, 1);
  return {inv: v.map(r => r.slice(n)), logdet};
}

// Fixed midpoint normal integration, independent of adaptive Simpson in Go.
const panels = 32768, dz = 20 / panels;
const grid = Array.from({length: panels}, (_, i) => -10 + (i + .5) * dz);
const weights = grid.map(z => Math.exp(-z * z / 2) * dz / Math.sqrt(2 * Math.PI));
function integrate(mu, variance) {
  const sd = Math.sqrt(variance);
  return grid.reduce((s, z, i) => s + weights[i] * sigmoid(mu + sd * z), 0);
}

const src = rows('docs/experiments/mmm-soft-learners-v120.jsonl');
const pilot = rows(process.argv[2]);
const convergence = process.argv[4] === 'convergence';
const capped = new Map();
if (convergence) {
  for await (const r of rows('docs/experiments/mmm-sparse-vrvm-v1-pilot.jsonl')) {
    if (r.Clock === 128 && r.Window === 64) capped.set([r.Phase, r.Case, r.Schedule].join('/'), r);
  }
  assert.equal(capped.size, 84);
}
assert.equal((await src.next()).value.Version, 'soft-learners-v120');
let fits = 0, forecasts = 0;
const stops = {}, descriptive = new Map();
const comparisons = [];
for await (const s of src) {
  if (s.Index !== 0) continue;
  for (const clock of convergence ? [128] : [0, 128, 224]) for (const window of convergence ? [64] : [64, 32]) {
    const next = await pilot.next(); assert(!next.done);
    const r = next.value;
    for (const key of ['Phase', 'Case', 'Index', 'Schedule']) assert.equal(r[key], s[key]);
    assert.equal(r.Clock, clock); assert.equal(r.Window, window);
    let reference;
    if (convergence) {
      reference = capped.get([r.Phase, r.Case, r.Schedule].join('/'));
      assert(reference);
      assert.deepEqual(r.Trace.slice(0, 64), reference.Trace);
      assert.deepEqual(r.Origins, reference.Origins);
      assert.deepEqual(r.X, reference.X); assert.deepEqual(r.Q, reference.Q);
      assert.deepEqual(r.Y, reference.Y); assert.deepEqual(r.Control, reference.Control);
      assert(r.Trace.at(-1)[2] >= reference.Trace.at(-1)[2]);
    }
    let origins = Array.from({length: 16}, (_, i) => i - 16);
    for (let i = 0; i < clock; i++) if (!s.Steps[i].Missing && i + s.Steps[i].Delay <= clock) origins.push(i);
    origins = origins.slice(-window);
    assert.deepEqual(r.Origins, origins);
    assert.deepEqual(origins, s.Fits[clock / 32].Origins[window === 64 ? 0 : 1]);
    const samples = origins.map(i => i < 0 ? {x: s.Initial[i + 16].Bits, y: +s.Initial[i + 16].Outcome} : {x: s.Steps[i].X, y: +s.Steps[i].Y});
    const ph = samples.map(v => phi(v.x)), d = r.PriorVariance;
    assert.equal(r.UsedXi.length, samples.length);
    assert.equal(r.FinalXi.length, samples.length);
    assert.equal(d.length, 256); assert(d.every(x => Number.isFinite(x) && x > 0));
    const w = r.UsedXi.map(x => { assert(Number.isFinite(x) && x >= 0); return x < 1e-8 ? .25 : Math.tanh(x / 2) / (2 * x); });
    const k = (a, b) => a.reduce((sum, v, j) => sum + d[j] * v * b[j], 0);
    const t = ph.map((a, i) => ph.map((b, j) => k(a, b) + (i === j ? 1 / w[i] : 0)));
    const {inv, logdet} = inverse(t);
    const b = masks.map((_, j) => samples.reduce((sum, v, i) => sum + ph[i][j] * (v.y - .5), 0));
    const db = b.map((v, j) => v * d[j]), correction = mv(inv, mv(ph, db));
    const mean = db.map((v, j) => v - d[j] * ph.reduce((sum, row, i) => sum + row[j] * correction[i], 0));
    for (let j = 0; j < 256; j++) {
      close('mean', mean[j], r.Mean[j]);
      const col = ph.map(row => row[j]);
      const diagonal = d[j] - d[j] ** 2 * dot(col, mv(inv, col));
      close('diagonal', diagonal, r.Diagonal[j]);
      if (j > 0) close('gammaRate', r.Rates[j], 1e-6 + (mean[j] ** 2 + diagonal) / 2);
    }
    close('logdet', d.reduce((sum, v) => sum + Math.log(v), 0) - logdet - w.reduce((sum, v) => sum + Math.log(v), 0), r.Logdet);
    const moments = x => {
      const p = phi(x), cross = ph.map(row => k(row, p));
      const variance = d.reduce((sum, v) => sum + v, 0) - dot(cross, mv(inv, cross));
      assert(variance >= -1e-8);
      return [dot(mean, p), Math.max(0, variance)];
    };
    for (let i = 0; i < samples.length; i++) {
      const [mu, variance] = moments(samples[i].x);
      close('finalXi', Math.sqrt(mu * mu + variance), r.FinalXi[i]);
    }
    assert(r.Trace.length >= 1 && r.Trace.length <= (convergence ? 1024 : 64));
    const trace = r.Trace.flat(); assert(trace.every(Number.isFinite));
    for (let i = 1; i < trace.length; i++) assert(trace[i] >= trace[i - 1] - 1e-8 * Math.max(1, Math.abs(trace[i - 1])));
    assert(['iteration-cap', 'bound-change'].includes(r.Stop), r.Stop);
    if (r.Stop === 'iteration-cap') assert.equal(r.Trace.length, convergence ? 1024 : 64);
    if (r.Stop === 'bound-change') assert(Math.abs(r.Trace.at(-1)[2] - r.Trace.at(-2)[2]) <= 1e-6);
    stops[r.Stop] = (stops[r.Stop] ?? 0) + 1;
    assert(r.Residual <= 1e-8);
    for (let i = 0; i < 32; i++) {
      const step = s.Steps[clock + i];
      assert.equal(r.X[i], step.X); assert.equal(r.Q[i], step.Q); assert.equal(r.Y[i], step.Y);
      assert.deepEqual(r.Control[i], step.P);
      const [mu, variance] = moments(step.X);
      close('logitMean', mu, r.LogitMean[i]); close('logitVariance', variance, r.LogitVariance[i]);
      close('integratedProbability', integrate(mu, variance), r.P[i][0], 2e-8);
      close('pluginProbability', Math.max(1e-12, Math.min(1 - 1e-12, sigmoid(mu))), r.P[i][1]);
      const key = [r.Phase, r.Schedule, r.Window, r.Clock].join('/');
      const cell = descriptive.get(key) ?? {n: 0, brier: Array(17).fill(0)};
      const ps = [...r.P[i], ...r.Control[i]];
      ps.forEach((p, a) => { assert(p >= 0 && p <= 1); cell.brier[a] += (p - step.Q) ** 2 + step.Q * (1 - step.Q); });
      cell.n++; descriptive.set(key, cell); forecasts++;
    }
    if (reference) {
      const loss = p => p.reduce((sum, v, i) => sum + (v[0] - r.Q[i]) ** 2 + r.Q[i] * (1 - r.Q[i]), 0) / 32;
      comparisons.push({phase: r.Phase, scenario: r.Case, schedule: r.Schedule,
        iterations: r.Trace.length, stop: r.Stop,
        boundGain: r.Trace.at(-1)[2] - reference.Trace.at(-1)[2],
        lastIncrement: r.Trace.at(-1)[2] - r.Trace.at(-2)[2],
        cappedBrier: loss(reference.P), extendedBrier: loss(r.P)});
    }
    fits++;
  }
}
assert((await pilot.next()).done); assert.equal(fits, convergence ? 84 : 504); assert.equal(forecasts, fits * 32);
const result = {pass: true, fits, forecasts, stops, errors, midpointPanels: panels,
  interpretation: 'Consumed index-0 feasibility pilot; descriptive means only, no trajectory confidence intervals or goal validation.',
  descriptive: [...descriptive].map(([key, c]) => ({key, n: c.n, brier: c.brier.map(v => v / c.n)}))};
if (convergence) result.comparisons = comparisons;
if (process.argv[3]) fs.writeFileSync(process.argv[3], JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({...result, descriptive: result.descriptive.length, ...(convergence ? {comparisons: comparisons.length} : {})}, null, 2));
