// Mechanical diagnostic clone; alters only the retained suffix, not the model
// prior. This is not a new observer policy, prospective study or adoption gate.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const source = 'research/tree-v60-branch-audit.mjs', dest = 'research/tree-v60-window-probe.mjs';
assert(!fs.existsSync(dest));
const raw = fs.readFileSync(source), hash = b => crypto.createHash('sha256').update(b).digest('hex');
let s = raw.toString();
function change(from, to) {assert(s.includes(from), from); s = s.replace(from, to);}
change("const root = 'research/tree-v60-diagnostic',", "const window = Number(process.argv[2]); assert([1200, 2400].includes(window));\nconst root = 'research/tree-v60-diagnostic',");
change('.filter(o => o.Available && !o.Expired)', '.filter(o => o.Available)');
change('for (let k = 1800; k < 2400; k++)', 'for (let k = 2400 - window; k < 2400; k++)');
change('near(q, a.Snapshots.at(-1).Forecast[i]); forecastChecks++;', 'assert(q >= 0 && q <= 1 && Number.isFinite(q)); forecastChecks++;');
change('near(actual.brier, a.Snapshots.at(-1).Brier);', 'const published600Brier = a.Snapshots.at(-1).Brier;');
change('actual, independentCounterfactual:', 'actual, published600Brier, retainedRiskGain: published600Brier - actual.brier, independentCounterfactual:');
change("{cells: c.length, rootPoolPosteriorMean:", "{cells: c.length, retainedRiskGain: mean(c, 'retainedRiskGain'), retentionWins: c.filter(x => x.retainedRiskGain > 1e-12).length, retentionLosses: c.filter(x => x.retainedRiskGain < -1e-12).length, rootPoolPosteriorMean:");
change("stage: 'Retrospective final-suffix reconstruction and fixed-selected-stream alternatives',", "stage: 'Retrospective suffix probe on fixed old nominations, not a prospective policy', retainedIssuedPositions: window,");
change("root + '/branch-audit.json'", "root + '/window-probe-' + window + '.json'");
s = s.replaceAll('research/tree-v60-branch-audit.mjs', dest);
change('const result = {stage:', 'const result = {forecastChecksAreProbabilityShapesNotPublished600Equality: true, oldExpiredRepliesIncludedOnlyIfWithinExpandedSuffix: true, stage:');
change('// The final issued suffix is [1800,2399]. All first measurements have', '// A larger final issued suffix is being probed. All first measurements have');
change('// arrived; paired factors exist ONLY for requested, available, unexpired W2.', '// arrived. Requested available W2 is valid if its original slot is retained\n    // under the counterfactual window, even if the OLD 600-position model expired it.');
fs.writeFileSync(dest, s, {flag: 'wx', mode: 0o600});
fs.writeFileSync('research/tree-v60-window-probe-generation.json', JSON.stringify({source, sourceSHA256: hash(raw),
  dest, outputSHA256: hash(s), plannedWindows: [1200, 2400], notProspectiveOrPolicyEvidence: true}, null, 2) + '\n', {flag: 'wx', mode: 0o600});
