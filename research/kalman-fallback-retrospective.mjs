import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) {
  throw new Error('usage: node research/kalman-fallback-retrospective.mjs V1.jsonl');
}
const rows = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const [manifest, ...records] = rows;
for (const [path, expected] of Object.entries(manifest.hashes)) {
  const hash = createHash('sha256').update(readFileSync(path)).digest('hex');
  if (hash !== expected) throw new Error(`v1 source mismatch: ${path}`);
}

const prior = [.7, .1, .1, .1];
function predictive(w) {
  return w === null ? [...prior] : w.map((v, i) => .998 * v + .002 * prior[i]);
}
function score(ticks, predictions, start, end) {
  let loss = 0;
  for (let t = start; t < end; t++) {
    loss += (predictions[t] - Number(ticks[t].Outcome)) ** 2;
  }
  return loss / (end - start);
}
function replay(r) {
  let weights = null;
  let surprise = 0;
  let alarm = -1;
  let recent = [];
  const mix = [];
  const fallback = [];
  const experts = [];
  for (let t = 0; t < r.Ticks.length; t++) {
    const tick = r.Ticks[t];
    const e = [tick.Predictions[0], tick.Predictions[1], tick.Predictions[3], .5];
    experts.push(e);
    const w = predictive(weights);
    mix.push(w.reduce((a, wi, i) => a + wi * e[i], 0));
    fallback.push(alarm < 0 ? .8 * e[0] + .1 :
      recent.length >= 8 && recent.reduce((a, b) => a + b, 0) < -.08 ? e[2] : e[1]);

    // Only already-issued forecasts for delivered origins may change state.
    for (const origin of tick.Delivered ?? []) {
      if (origin > t || r.Ticks[origin].Missing) throw new Error('future or missing feedback');
      const y = Number(r.Ticks[origin].Outcome);
      const old = experts[origin];
      if (r.Ticks[origin].Audit) {
        const next = predictive(weights).map((wi, i) => wi * (y ? old[i] : 1 - old[i]));
        const total = next.reduce((a, b) => a + b, 0);
        weights = next.map(v => v / total);
      }
      if (alarm < 0) {
        surprise = Math.max(0, surprise + (old[0] - y) ** 2 - .28);
        if (surprise >= 1) {
          alarm = t;
          recent = [];
        }
      } else {
        recent.push((old[2] - y) ** 2 - (old[1] - y) ** 2);
        if (recent.length > 16) recent.shift();
      }
    }
  }
  const onset = r.Scenario === 'shift128' || r.Scenario === 'gradual' ? 128 :
    r.Scenario === 'stable05' ? 0 : 256;
  return {
    Scenario: r.Scenario,
    Fit: r.Fit,
    Stream: r.Stream,
    alarm,
    mix: {
      full: score(r.Ticks, mix, 0, 512),
      early: score(r.Ticks, mix, onset, onset + 64),
      tail: score(r.Ticks, mix, 384, 512),
    },
    fallback: {
      full: score(r.Ticks, fallback, 0, 512),
      early: score(r.Ticks, fallback, onset, onset + 64),
      tail: score(r.Ticks, fallback, 384, 512),
    },
  };
}
const out = records.map(replay);
const summary = { split: manifest.split, records: out.length, cases: {} };
for (const name of ['stable05', 'shift128', 'gradual', 'delayed_missing', 'interaction']) {
  const cases = out.filter(r => r.Scenario === name);
  const mean = (policy, period) => cases.reduce((a, r) => a + r[policy][period], 0) / cases.length;
  summary.cases[name] = {
    alarms: cases.filter(r => r.alarm >= 0).length,
    mix: { full: mean('mix', 'full'), early: mean('mix', 'early'), tail: mean('mix', 'tail') },
    fallback: { full: mean('fallback', 'full'), early: mean('fallback', 'early'), tail: mean('fallback', 'tail') },
  };
}
console.log(JSON.stringify(summary, null, 2));
