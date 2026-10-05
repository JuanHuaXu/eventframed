import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { gunzipSync } from 'node:zlib';

const [mode, input, output] = process.argv.slice(2);
if (!['uniform', 'latent'].includes(mode) || !input || !output || existsSync(output)) {
  throw new Error('usage: node research/check-ridge-v10.mjs uniform|latent input.json.gz new-output.json');
}
const hash = (path) => createHash('sha256').update(readFileSync(path)).digest('hex');
const data = JSON.parse(gunzipSync(readFileSync(input)));
if (data.Mode !== mode || data.Records.length !== 480 ||
    data.ProtocolSHA256 !== hash('docs/experiments/mmm-ridge-v10-component-protocol.md') ||
    data.SourceSHA256 !== hash('internal/observationlearners/research_ridge_v10_run_test.go')) {
  throw new Error('source or corpus identity mismatch');
}
const expectedJournal = {
  uniform: 'c9a404fbe5c3b0243563e48f229615ae3d3423cb3b17962f8ff1e9df15e5751b',
  latent: '3b8ea4d05c6158613d0e169e6e93f1afe4add6efa767676c43746749d86b0066'
};
if (data.JournalSHA256 !== expectedJournal[mode]) throw new Error('journal hash mismatch');

const scenarios = ['stable05', 'stable20', 'shift128', 'shift256', 'shift384', 'gradual', 'recurring', 'delayed_missing', 'interaction', 'null'];
const splitIndex = { design: 0, confirmation: 1 };
const seen = new Set();
const cells = new Map();
let attempts = 0, publications = 0, failures = 0, fitNS = 0, compileNS = 0, forecastNS = 0, readyFrames = 0;
for (const row of data.Records) {
  if (!(row.Split in splitIndex) || !scenarios.includes(row.Scenario) || row.Fit < 0 || row.Fit >= 3 || row.Stream < 0 || row.Stream >= 8 || row.Ticks.length !== 512) {
    throw new Error('bad row identity');
  }
  const identity = `${row.Split}/${row.Scenario}/${row.Fit}/${row.Stream}`;
  if (seen.has(identity)) throw new Error(`duplicate ${identity}`);
  seen.add(identity);
  if (row.PublicationsLog.length !== row.Publications + row.PublicationsLog.filter(p => p.Failure).length) throw new Error(`publication count ${identity}`);
  attempts += row.PublicationsLog.length;
  publications += row.Publications;
  failures += row.PublicationsLog.filter(p => p.Failure).length;
  fitNS += row.FitNS;
  compileNS += row.CompileNS;
  forecastNS += row.ForecastNS;
  let previousClock = -1, previousAudits = 0;
  for (const p of row.PublicationsLog) {
    if (p.Clock <= previousClock || p.Audits <= previousAudits || p.Audits < 32 || p.Audits % 16 !== 0 || p.LastOrigin > p.Clock) throw new Error(`publication order ${identity}`);
    previousClock = p.Clock;
    previousAudits = p.Audits;
  }
  const change = Number(row.Scenario.slice(5));
  const post = ['shift128', 'shift256', 'shift384'].includes(row.Scenario) ? change : ['gradual', 'recurring'].includes(row.Scenario) ? 128 : 256;
  const windows = [['full', 0, 512], ['post', post, 512], ['early64', post, Math.min(512, post + 64)]];
  for (let t = 0; t < 512; t++) {
    const tick = row.Ticks[t];
    const readyExpected = row.PublicationsLog.some(p => !p.Failure && p.Clock < t);
    if (tick.ready !== readyExpected) throw new Error(`as-of readiness ${identity}/${t}`);
    if (tick.ready) readyFrames++;
    for (let arm = 0; arm < 4; arm++) {
      const incumbent = tick.incumbent[arm], candidate = tick.candidate[arm];
      if (!(incumbent > 0 && incumbent < 1 && candidate > 0 && candidate < 1)) throw new Error(`probability ${identity}/${t}/${arm}`);
      if (!tick.ready && candidate !== incumbent) throw new Error(`fallback ${identity}/${t}/${arm}`);
      const y = Number(tick.y);
      const iLoss = (incumbent - y) ** 2, cLoss = (candidate - y) ** 2;
      for (const [window, start, end] of windows) {
        if (t < start || t >= end) continue;
        const key = `${row.Split}/${row.Scenario}/${window}/${arm}`;
        if (!cells.has(key)) cells.set(key, { Split: row.Split, Scenario: row.Scenario, Window: window, Arm: arm, N: 0, ReadyN: 0, IncumbentSum: 0, CandidateSum: 0, ReadyIncumbentSum: 0, ReadyCandidateSum: 0, RowDiffs: new Map() });
        const cell = cells.get(key);
        cell.N++;
        cell.IncumbentSum += iLoss;
        cell.CandidateSum += cLoss;
        if (tick.ready) {
          cell.ReadyN++;
          cell.ReadyIncumbentSum += iLoss;
          cell.ReadyCandidateSum += cLoss;
        }
        const diff = cell.RowDiffs.get(`${row.Fit}/${row.Stream}`) ?? { fit: row.Fit, n: 0, sum: 0 };
        diff.n++;
        diff.sum += iLoss - cLoss;
        cell.RowDiffs.set(`${row.Fit}/${row.Stream}`, diff);
      }
    }
  }
}
if (seen.size !== 480 || readyFrames === 0) throw new Error('incomplete replay');
const summary = [...cells.values()].map(c => {
  const rowDiffs = [...c.RowDiffs.values()].map(r => r.sum / r.n);
  if (rowDiffs.length !== 24) throw new Error('cell lacks 24 trajectories');
  const mean = rowDiffs.reduce((a, b) => a + b, 0) / 24;
  const variance = rowDiffs.reduce((a, b) => a + (b - mean) ** 2, 0) / 23;
  const fitMeans = [0, 1, 2].map(fit => [...c.RowDiffs.values()].filter(r => r.fit === fit).reduce((a, r) => a + r.sum / r.n, 0) / 8);
  return {
    Split: c.Split, Scenario: c.Scenario, Window: c.Window, Arm: c.Arm,
    N: c.N, ReadyN: c.ReadyN,
    IncumbentBrier: c.IncumbentSum / c.N,
    CandidateBrier: c.CandidateSum / c.N,
    Gain: mean, Lower36SE: mean - 3.6 * Math.sqrt(variance / 24),
    ReadyIncumbentBrier: c.ReadyN ? c.ReadyIncumbentSum / c.ReadyN : null,
    ReadyCandidateBrier: c.ReadyN ? c.ReadyCandidateSum / c.ReadyN : null,
    FitMeans: fitMeans
  };
});
writeFileSync(output, JSON.stringify({
  Mode: mode, SourceSHA256: hash(input), JournalSHA256: data.JournalSHA256,
  Records: data.Records.length, Attempts: attempts, Publications: publications, Failures: failures,
  ReadyFrames: readyFrames, FitNS: fitNS, CompileNS: compileNS, ForecastNS: forecastNS,
  ReplayWallNS: data.WallNS, Cells: summary
}, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify({ mode, records: seen.size, attempts, publications, failures, readyFrames, replayWallSeconds: data.WallNS / 1e9 }));
