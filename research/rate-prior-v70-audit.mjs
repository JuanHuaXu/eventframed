// Independent readback for the finite prior experiment, not a stream evaluator.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const base = 'research/rate-prior-v70-recheck';
const root = 'research/rate-prior-v70-audit';
const source = 'research/rate-prior-v70-audit.mjs';
assert(!fs.existsSync(root));
const digest = p => crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const read = p => JSON.parse(fs.readFileSync(p));
const frozen = read(base + '/freeze.json'), complete = read(base + '/completed.json');
const original = read('research/rate-prior-v70-preflight/freeze.json');
assert.equal(digest(frozen.source), frozen.sourceSHA256);
assert.equal(digest(base + '/source.mjs'), frozen.sourceSHA256);
assert.equal(digest(original.source), original.sourceSHA256);
assert.equal(digest('research/rate-prior-v70-preflight/source.mjs'), original.sourceSHA256);
assert.equal(complete.exitCode, 0);
assert(complete.allJobsTerminal && complete.sourceUnchanged);
assert.equal(digest(base + '/results.json'), complete.resultsSHA256);
assert.equal(read('research/rate-prior-v70-preflight/failure.json').exitCode, 1);
const paths = [source, base + '/freeze.json', base + '/results.json', base + '/completed.json', original.source, 'research/rate-prior-v70-preflight/failure.json'];
const files = Object.fromEntries(paths.map(p => [p, digest(p)]));
fs.mkdirSync(root, {mode: 0o700});
fs.copyFileSync(source, root + '/source.mjs', fs.constants.COPYFILE_EXCL);
fs.chmodSync(root + '/source.mjs', 0o600);
const save = (p, x) => fs.writeFileSync(root + '/' + p, JSON.stringify(x, null, 2) + '\n', {flag: 'wx', mode: 0o600});
save('freeze.json', {time: new Date().toISOString(), files, scope: 'independent finite-prior readback and corruption controls only'});
let checks = 0, maxDefect = 0;
const near = (a, b) => {
  const d = Math.abs(a - b);
  maxDefect = Math.max(maxDefect, d);
  checks++;
  assert(Number.isFinite(a) && Number.isFinite(b) && d < 2e-12);
};
function verify(x) {
  assert.equal(x.rows.length, 63);
  assert.equal(x.ratios.length, 42);
  assert.equal(x.checks, 2079);
  assert.equal(x.unsupportedPairBranches, 42);
  assert.equal(x.repeatedEvidenceNegativeControls, 42);
  assert(!x.qualityRescueEstablished && !x.uniqueCauseOfStreamFailureEstablished && !x.equalTotalCostEstablished);
  assert.deepEqual(x.goals, Array(7).fill('OPEN'));
  const keys = new Set();
  for (const r of x.rows) {
    const key = [r.baseline, r.config, r.eta].join('/');
    assert(!keys.has(key)); keys.add(key);
    const cfg = x.configurations.find(z => z.name === r.config);
    assert(cfg);
    const b = r.baseline, q = r.eta + (1 - 2 * r.eta) * b;
    // Compute E[P^2] directly from the urn moment, then its variance.
    const a = cfg.strength * b, c = cfg.strength * (1 - b);
    const second = (a * (a + 1) / ((a + c) * (a + c + 1))) * 19 / 20 + b / 20;
    const v = cfg.spike * b * b + (1 - cfg.spike) * second - b * b;
    near(r.variance, v);
    near(q * r.yes + (1 - q) * r.no, b);
    near(r.yes, (r.eta * b + (1 - 2 * r.eta) * (v + b * b)) / q);
    near(r.no, ((1 - r.eta) * b - (1 - 2 * r.eta) * (v + b * b)) / (1 - q));
    near(r.stationaryExcess, q * (r.yes - b) ** 2 + (1 - q) * (r.no - b) ** 2);
    assert(r.variance > 0 && r.variance <= b * (1 - b));
  }
  for (const r of x.ratios) {
    const old = x.rows.find(z => z.config === 'current' && z.baseline === r.baseline && z.eta === r.eta);
    const next = x.rows.find(z => z.config === r.config && z.baseline === r.baseline && z.eta === r.eta);
    assert(old && next);
    near(r.responseRatio, next.variance / old.variance);
    near(r.stationaryHarmRatio, next.stationaryExcess / old.stationaryExcess);
    assert(r.responseRatio > 1 && r.stationaryHarmRatio > 1);
  }
}
const data = read(base + '/results.json');
verify(data);
const originalChecks = checks, originalDefect = maxDefect;
const mutate = [
  x => x.rows[0].variance += .01,
  x => x.rows[0].yes += .01,
  x => x.rows[0].no += .01,
  x => x.rows[0].stationaryExcess += .01,
  x => x.ratios[0].responseRatio += .1,
  x => x.ratios[0].stationaryHarmRatio += .1,
  x => x.rows[1] = structuredClone(x.rows[0]),
  x => x.qualityRescueEstablished = true,
  x => x.goals[0] = 'COMPLETE',
];
let rejected = 0;
for (const alter of mutate) {
  const x = structuredClone(data); alter(x);
  assert.throws(() => verify(x)); rejected++;
}
for (const [p, h] of Object.entries(files)) assert.equal(digest(p), h);
save('results.json', {checks: originalChecks, maxDefect: originalDefect, corruptionControlsRejected: rejected, originalFailureRetained: true, sourceUnchanged: true, broaderPriorAloneValidated: false, streamOrRuntimeClaim: false, goal: 'ACTIVE', goals: Array(7).fill('OPEN')});
save('completed.json', {exitCode: 0, allJobsTerminal: true, resultsSHA256: digest(root + '/results.json')});
console.log(JSON.stringify({checks: originalChecks, maxDefect: originalDefect, corruptionControlsRejected: rejected, componentOnly: true}));
