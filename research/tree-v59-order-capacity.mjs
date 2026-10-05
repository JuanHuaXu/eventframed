// Retrospective structural audit. Hidden rates are auditor-only; no training,
// observation choice, confirmation dispatch, or old gate is changed.
import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';

const root = 'research/tree-v59-diagnostic';
const near = (a, b) => assert(Math.abs(a - b) < 2e-10, `${a} != ${b}`);
function isotonic(rows) {
  // Collapse equal public scores first: their entire forecast path is identical.
  const groups = [];
  for (const row of rows) {
    const last = groups.at(-1);
    if (last && last.base === row.base) {
      last.sum += row.rate;
      last.ids.push(row.id);
    } else groups.push({base: row.base, sum: row.rate, ids: [row.id]});
  }
  const blocks = [];
  for (const g of groups) {
    blocks.push({sum: g.sum, count: g.ids.length, ids: g.ids.slice()});
    while (blocks.length > 1) {
      const b = blocks.at(-1), a = blocks.at(-2);
      if (a.sum / a.count <= b.sum / b.count) break;
      blocks.splice(-2, 2, {sum: a.sum + b.sum, count: a.count + b.count, ids: a.ids.concat(b.ids)});
    }
  }
  const q = new Map();
  for (const b of blocks) for (const i of b.ids) q.set(i, b.sum / b.count);
  return q;
}
const rows = rates => rates.map((rate, id) => ({base: id, rate, id}));
near(isotonic(rows([.2, .8])).get(0), .2);
near(isotonic(rows([.8, .2])).get(0), .5);
near(isotonic([{base: 1, id: 0, rate: .2}, {base: 1, id: 1, rate: .8}]).get(1), .5);

const cells = [];
let first = true, snapshotChecks = 0, pairChecks = 0, equalityChecks = 0;
for await (const line of readline.createInterface({input: fs.createReadStream(root + '/diagnostic.jsonl'), crlfDelay: Infinity})) {
  const w = JSON.parse(line);
  if (first) { first = false; continue; }
  const p = w.Population.World;
  const ordered = p.Base.map((base, id) => ({base, id})).sort((a, b) => a.base - b.base || a.id - b.id);
  const firstRank = new Map();
  ordered.forEach((x, i) => { if (!firstRank.has(x.base)) firstRank.set(x.base, i); });
  const buckets = new Map();
  for (const x of ordered) {
    const leaf = Math.floor(firstRank.get(x.base) * 128 / 150);
    if (!buckets.has(leaf)) buckets.set(leaf, []);
    buckets.get(leaf).push(x);
  }
  const roundFloors = p.Rates.map(rates => {
    let variance = 0, excess = 0, inversions = 0;
    for (const bucket of buckets.values()) {
      const rr = bucket.map(x => ({...x, rate: rates[x.id]}));
      const oracle = isotonic(rr);
      for (const x of rr) {
        variance += x.rate * (1 - x.rate) / 150;
        excess += (oracle.get(x.id) - x.rate) ** 2 / 150;
      }
      for (let i = 1; i < rr.length; i++) inversions += +(rr[i - 1].rate > rr[i].rate + 2e-10);
    }
    return {variance, excess, floor: variance + excess, inversions};
  });
  for (const a of w.Arms) {
    if (a.Mode === 'full' || a.Mode === 'adaptive') continue;
    for (const s of a.Snapshots) {
      snapshotChecks++;
      for (const b of buckets.values()) for (let j = 1; j < b.length; j++) {
        const x = b[j - 1], y = b[j];
        assert(s.Forecast[x.id] <= s.Forecast[y.id] + 2e-10, 'forecast violates proven within-leaf order');
        pairChecks++;
        if (x.base === y.base) { near(s.Forecast[x.id], s.Forecast[y.id]); equalityChecks++; }
      }
      const round = Math.min(15, Math.floor(s.Tick / 150));
      let risk = 0;
      for (let i = 0; i < 150; i++) {
        const q = s.Forecast[i], r = p.Rates[round][i];
        risk += (r * (1 - r) + (q - r) ** 2) / 150;
      }
      near(risk, s.Brier);
      assert(risk >= roundFloors[round].floor - 2e-10, 'risk below relaxed isotonic oracle floor');
    }
  }
  cells.push({geometry: p.Geometry, regime: p.Regime, uniqueBase: firstRank.size,
    occupiedLeaves: buckets.size, multiMemberLeaves: [...buckets.values()].filter(b => b.length > 1).length,
    meanExcess: roundFloors.reduce((s, r) => s + r.excess / 16, 0),
    terminal: roundFloors.at(-1), roundFloors});
}
assert.equal(cells.length, 40);
assert(snapshotChecks > 10000 && pairChecks > 0 && equalityChecks > 0);
const result = {stage: 'Retrospective monotone-leaf capacity audit, not a new gate',
  proof: 'All conditional atoms increase with public b. Members in the same terminal leaf share all mixture coefficients, so their simultaneous forecasts are ordered by b, and equal b gives identical forecasts. Isotonic projection is a relaxed oracle lower bound.',
  scope: 'Simultaneous snapshots only; sequential issued forecasts use different histories. This floor and the pointwise range floor cannot be added.',
  oracleUsedOnlyByAuditor: true, cells, snapshotChecks, pairChecks, equalityChecks,
  meanRelaxedOrderExcess: cells.reduce((s, c) => s + c.meanExcess / 40, 0),
  terminalRelaxedOrderExcess: cells.reduce((s, c) => s + c.terminal.excess / 40, 0),
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync('research/tree-v59-order-capacity.mjs')).digest('hex'),
  wholeGoals: Array(7).fill('OPEN')};
fs.writeFileSync(root + '/order-capacity.json', JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({snapshotChecks, pairChecks, equalityChecks,
  meanRelaxedOrderExcess: result.meanRelaxedOrderExcess, terminalRelaxedOrderExcess: result.terminalRelaxedOrderExcess,
  mostBinding: cells.slice().sort((a, b) => b.meanExcess - a.meanExcess).slice(0, 6).map(({roundFloors, ...c}) => c)}, null, 2));
