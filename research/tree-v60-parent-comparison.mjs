// Descriptive post-collection comparison with V59; no new acceptance gate.
import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const root = 'research/tree-v60-diagnostic', parent = 'research/tree-v59-diagnostic';
const stream = p => readline.createInterface({input: fs.createReadStream(p), crlfDelay: Infinity})[Symbol.asyncIterator]();
const a = stream(root + '/diagnostic.jsonl'), b = stream(parent + '/diagnostic.jsonl');
assert(!(await a.next()).done && !(await b.next()).done);
const modes = ['full', 'adaptive', 'no_pair', 'random', 'uncertainty', 'information', 'falsification', 'predictive'];
const totals = Object.fromEntries(modes.map(m => [m, {cells: 0, riskGain: 0, terminalGain: 0, utilityGain: 0, recoveryGain: 0, wins: 0, losses: 0}]));
const stripped = x => {const y = structuredClone(x); delete y.Costs; delete y.Breakdown; return y;};
let controls = 0;
for (let i = 0; i < 40; i++) {
  const aa = await a.next(), bb = await b.next();
  assert(!aa.done && !bb.done);
  const current = JSON.parse(aa.value), previous = JSON.parse(bb.value);
  assert.deepEqual(current.Population, previous.Population);
  assert.equal(current.Arms.length, 24); assert.equal(previous.Arms.length, 24);
  for (let j = 0; j < 24; j++) {
    const x = current.Arms[j], y = previous.Arms[j], t = totals[x.Mode];
    assert.equal(x.Mode, y.Mode); assert.equal(x.Mode, modes[j % 8]); assert.equal(x.Schedule, y.Schedule);
    if (j % 8 < 2) {assert.deepEqual(stripped(x), stripped(y)); controls++;}
    t.cells++;
    const gain = y.IssuedBrier - x.IssuedBrier;
    t.riskGain += gain / 120;
    t.terminalGain += (y.Snapshots.at(-1).Brier - x.Snapshots.at(-1).Brier) / 120;
    t.utilityGain += (x.Snapshots.at(-1).PacketUsefulness - y.Snapshots.at(-1).PacketUsefulness) / 120;
    t.recoveryGain += (y.Recovery - x.Recovery) / 120;
    t.wins += +(gain > 1e-12); t.losses += +(gain < -1e-12);
  }
}
assert((await a.next()).done && (await b.next()).done);
assert.equal(controls, 240);
const result = {stage: 'Descriptive V59 comparison, not a new gate', allPopulationsIdentical: true,
  controlsBitwiseEqual: controls, changes: totals, wholeGoals: Array(7).fill('OPEN'), goal: 'ACTIVE',
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync('research/tree-v60-parent-comparison.mjs')).digest('hex')};
fs.writeFileSync(root + '/comparison-v59.json', JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify(result, null, 2));
