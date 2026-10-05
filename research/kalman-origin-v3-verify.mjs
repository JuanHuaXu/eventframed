import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');
const scenarios = ['stable05', 'shift128', 'gradual', 'delayed_missing', 'interaction'];
const noise = [0.0005, 0.01];
const dim = 11;

function assert(ok, message) {
  if (!ok) throw new Error(message);
}

function near(actual, expected, label, tolerance = 1e-10) {
  assert(Number.isFinite(actual) && Math.abs(actual - expected) <= tolerance,
    `${label}: ${actual} != ${expected}`);
}

function features(x, base) {
  const h = [1, 2 * (base - 0.5)];
  for (let i = 0; i < 9; i++) h.push((x & (1 << i)) !== 0 ? 1 / 3 : -1 / 3);
  return h;
}

function newFilter(q) {
  return { mean: Array(dim).fill(0), cov: Array.from({ length: dim }, (_, i) =>
    Array.from({ length: dim }, (_, j) => i === j ? 1 : 0)), origin: -1, q };
}

function advance(filter, clock) {
  assert(clock > filter.origin, 'nonmonotone origin');
  const cov = filter.cov.map((row) => [...row]);
  for (let i = 0; i < dim; i++) cov[i][i] += (clock - filter.origin) * filter.q;
  return cov;
}

function forecast(filter, clock, base, h) {
  advance(filter, clock);
  let p = base;
  for (let i = 0; i < dim; i++) p += h[i] * filter.mean[i];
  return Math.max(0.01, Math.min(0.99, p));
}

function observe(filter, origin, base, h, y) {
  const cov = advance(filter, origin);
  const ph = Array(dim).fill(0);
  let variance = 0.25;
  let mu = 0;
  for (let i = 0; i < dim; i++) {
    mu += h[i] * filter.mean[i];
    for (let j = 0; j < dim; j++) ph[i] += cov[i][j] * h[j];
    variance += h[i] * ph[i];
  }
  assert(Number.isFinite(variance) && variance > 0, 'invalid innovation variance');
  const innovation = Number(y) - base - mu;
  for (let i = 0; i < dim; i++) {
    filter.mean[i] += ph[i] / variance * innovation;
    for (let j = i; j < dim; j++) {
      const next = cov[i][j] - ph[i] * ph[j] / variance;
      cov[i][j] = next;
      cov[j][i] = next;
    }
  }
  filter.cov = cov;
  filter.origin = origin;
}

function brier(ticks, predictions, start, end) {
  let sum = 0;
  for (let i = start; i < end; i++) {
    const delta = predictions[i] - Number(ticks[i].Y);
    sum += delta * delta;
  }
  return sum / (end - start);
}

function paired(values) {
  const mean = values.reduce((a, b) => a + b, 0) / values.length;
  const variance = values.reduce((a, b) => a + (b - mean) ** 2, 0) / (values.length - 1);
  const margin = 3.5 * Math.sqrt(variance / values.length);
  return { mean, lower: mean - margin, upper: mean + margin };
}

function readSplit(split) {
  const path = resolve(root, `docs/experiments/mmm-kalman-origin-v3-${split}.jsonl`);
  const content = readFileSync(path);
  const [manifestLine, ...rowLines] = content.toString().trimEnd().split('\n');
  const manifest = JSON.parse(manifestLine);
  assert(manifest.kind === 'manifest' && manifest.split === split, `${split} manifest`);
  assert(manifest.seedBase === (split === 'design' ? 2026102201 : 2026102202), `${split} seed`);
  assert(manifest.fitOffset === (split === 'design' ? 400 : 500), `${split} fit offset`);
  for (const [name, wanted] of Object.entries(manifest.hashes)) {
    const actual = createHash('sha256').update(readFileSync(resolve(root, name))).digest('hex');
    assert(actual === wanted, `${split} source hash ${name}`);
  }
  assert(rowLines.length === 160, `${split} row count`);
  const seen = new Set();
  const cases = new Map(scenarios.map((name) => [name, []]));
  for (const line of rowLines) {
    const row = JSON.parse(line);
    const key = `${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert(!seen.has(key), `${split} duplicate ${key}`);
    seen.add(key);
    assert(row.Kind === 'trial' && row.Split === split && cases.has(row.Scenario), `${split} row identity`);
    assert(row.Fit >= 0 && row.Fit < 4 && row.Stream >= 0 && row.Stream < 8, `${split} row coordinate`);
    const ticks = row.Control.Ticks;
    assert(ticks.length === 512 && row.OriginPredictions.every((p) => p.length === 512), `${split} tick count`);
    const scenario = row.Scenario;
    const delay = scenario === 'delayed_missing' ? 16 : 0;
    const filters = noise.map(newFilter);
    let auditedDeliveries = 0;
    for (let clock = 0; clock < 512; clock++) {
      const tick = ticks[clock];
      const h = features(tick.X, tick.P[0]);
      for (let arm = 0; arm < 2; arm++) {
        const prediction = forecast(filters[arm], clock, tick.P[0], h);
        near(prediction, row.OriginPredictions[arm][clock], `${split} ${key} forecast ${clock}/${arm}`);
        assert(prediction >= 0.01 && prediction <= 0.99, `${split} prediction bound`);
        if (delay === 0) near(prediction, tick.P[arm === 0 ? 2 : 4], `${split} ${key} zero delay ${clock}/${arm}`);
      }
      for (const origin of tick.Delivered ?? []) {
        assert(origin >= 0 && origin <= clock && origin + delay === clock, `${split} delivery time ${key}`);
        const packet = ticks[origin];
        assert(!packet.Missing, `${split} delivered missing packet`);
        if (!packet.Audit) continue;
        auditedDeliveries++;
        const ph = features(packet.X, packet.P[0]);
        for (const filter of filters) observe(filter, origin, packet.P[0], ph, packet.Y);
      }
    }
    assert(auditedDeliveries === row.Control.Updates, `${split} audit count ${key}`);
    assert(row.OriginStateBytes <= 4096, `${split} state bytes ${key}`);
    assert(row.ForecastP99NS.every((n) => n < 50000) && row.UpdateP99NS.every((n) => n < 50000),
      `${split} isolated timing ${key}`);
    const change = scenario === 'shift128' || scenario === 'gradual' ? 128 :
      scenario === 'stable05' ? 0 : 256;
    const all = Array.from({ length: 7 }, (_, arm) => arm < 5 ? ticks.map((t) => t.P[arm]) :
      row.OriginPredictions[arm - 5]);
    for (let arm = 0; arm < 7; arm++) {
      near(brier(ticks, all[arm], 0, 512), row.Metric.Full[arm], `${split} ${key} full ${arm}`);
      near(brier(ticks, all[arm], change, change + 64), row.Metric.Early[arm], `${split} ${key} early ${arm}`);
      near(brier(ticks, all[arm], 384, 512), row.Metric.Tail[arm], `${split} ${key} tail ${arm}`);
      if (scenario === 'delayed_missing') {
        near(brier(ticks, all[arm], 256, 272), row.Metric.DelayedBlind[arm], `${split} ${key} blind ${arm}`);
        near(brier(ticks, all[arm], 272, 336), row.Metric.DelayedAfter[arm], `${split} ${key} after ${arm}`);
      }
    }
    cases.get(scenario).push(row);
  }
  for (const [name, rows] of cases) assert(rows.length === 32, `${split} ${name} cohort size`);
  const result = {};
  for (const [name, rows] of cases) {
    const measure = (metric, left, right) => paired(rows.map((r) => r.Metric[metric][left] - r.Metric[metric][right]));
    result[name] = {
      brier: Object.fromEntries(['Full', 'Early', 'Tail', 'DelayedBlind', 'DelayedAfter']
        .map((metric) => [metric, Array.from({ length: 7 }, (_, arm) =>
          rows.reduce((sum, r) => sum + r.Metric[metric][arm], 0) / rows.length)])),
      originGain: [measure(name === 'delayed_missing' ? 'DelayedAfter' : 'Early', 2, 5),
        measure(name === 'delayed_missing' ? 'DelayedAfter' : 'Early', 4, 6)],
      stableHarm: [measure('Full', 5, 0), measure('Full', 6, 0)],
      interactionHarm: [measure('Tail', 5, 1), measure('Tail', 6, 1)],
      shortGap: [measure(name === 'delayed_missing' ? 'DelayedAfter' : 'Tail', 5, 1),
        measure(name === 'delayed_missing' ? 'DelayedAfter' : 'Tail', 6, 1)],
      maxForecastP99NS: noise.map((_, arm) => Math.max(...rows.map((r) => r.ForecastP99NS[arm]))),
      maxUpdateP99NS: noise.map((_, arm) => Math.max(...rows.map((r) => r.UpdateP99NS[arm]))),
    };
  }
  return { split, dataSHA256: createHash('sha256').update(content).digest('hex'), result };
}

const design = readSplit('design');
const confirmation = readSplit('confirmation');
const verdict = noise.map((q, arm) => {
  const splitPass = ({ result }) => {
    const delayed = result.delayed_missing;
    const stable = result.stable05;
    const interaction = result.interaction;
    const gradual = result.gradual;
    return delayed.originGain[arm].mean >= 0.01 && delayed.originGain[arm].lower > 0 &&
      delayed.shortGap[arm].mean <= 0.005 && stable.stableHarm[arm].upper < 0.01 &&
      interaction.interactionHarm[arm].upper < 0.01 && gradual.shortGap[arm].mean <= 0.005;
  };
  return { q, designPass: splitPass(design), confirmationPass: splitPass(confirmation) };
});
console.log(JSON.stringify({ design, confirmation, verdict }, null, 2));
