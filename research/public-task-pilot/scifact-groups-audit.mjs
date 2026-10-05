// Independent breadth-first audit of the preparation-only evidence split.
// It does not import the union-find producer or expose labels to serving.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
export function audit(plan, train, test) {
  const graph = new Map(), official = new Map();
  const add = (a, b) => {
    if (!graph.has(a)) graph.set(a, new Set());
    graph.get(a).add(b);
  };
  for (const [rows, split] of [[train, 'train'], [test, 'test']]) {
    const seen = new Set();
    for (const row of rows) {
      assert.equal(typeof row.query, 'string');
      assert.equal(typeof row.document, 'string');
      assert(row.query && row.document && row.score === 1);
      const edge = JSON.stringify([row.query, row.document]);
      assert(!seen.has(edge), 'duplicate relevance relation');
      seen.add(edge);
      assert(!official.has(row.query) || official.get(row.query) === split, 'official split overlap');
      official.set(row.query, split);
      add('q:' + row.query, 'd:' + row.document);
      add('d:' + row.document, 'q:' + row.query);
    }
  }
  const visited = new Set(), expected = [];
  for (const start of graph.keys()) {
    if (visited.has(start)) continue;
    const queue = [start], queries = [], documents = [];
    visited.add(start);
    for (let i = 0; i < queue.length; i++) {
      const node = queue[i];
      (node.startsWith('q:') ? queries : documents).push(node.slice(2));
      for (const neighbor of graph.get(node)) {
        if (!visited.has(neighbor)) { visited.add(neighbor); queue.push(neighbor); }
      }
    }
    queries.sort(); documents.sort();
    expected.push({id: hash(JSON.stringify({queries, documents})), queries, documents,
      containsOfficialTest: queries.some(q => official.get(q) === 'test')});
  }
  expected.sort((a, b) => a.id.localeCompare(b.id));
  assert.deepEqual(plan.components, expected, 'exact connected evidence families');
  const splits = {fit: [], calibration: [], confirmation: [], excludedOfficialTrain: []};
  const units = {fit: [], calibration: [], confirmation: []};
  for (const component of expected) {
    if (component.containsOfficialTest) {
      units.confirmation.push(component.id);
      for (const q of component.queries) splits[official.get(q) === 'test' ? 'confirmation' : 'excludedOfficialTrain'].push(q);
    } else {
      const split = Number.parseInt(component.id.slice(0, 8), 16) % 3 === 0 ? 'calibration' : 'fit';
      units[split].push(component.id);
      splits[split].push(...component.queries);
    }
  }
  for (const values of Object.values(splits)) values.sort();
  assert.deepEqual(plan.splits, splits, 'frozen identity partition');
  assert.deepEqual(plan.units, units, 'independent evaluation units');
  const counts = {components: expected.length,
    queries: Object.fromEntries(Object.entries(splits).map(([k, v]) => [k, v.length])),
    units: Object.fromEntries(Object.entries(units).map(([k, v]) => [k, v.length]))};
  assert.deepEqual(plan.counts, counts, 'reported denominators');
  assert.equal(plan.allOfficialTestQueriesRetained, true);
  assert.equal(plan.exactEvidenceOverlapPrevented, true);
  assert.equal(plan.topicIndependenceUnproven, true);
  assert.equal(plan.partitionUsesIdentitiesNotModelPerformance, true);
  assert.equal(plan.noModelRuns, true);
  assert.equal(plan.noFitting, true);
  return counts;
}

export function controls(plan, train, test) {
  audit(plan, train, test);
  audit(plan, [...train].reverse(), [...test].reverse());
  const mutations = [
    p => { p.components[0].documents.pop(); },
    p => { p.components[0].queries.push('invented-query'); },
    p => { p.components[0].containsOfficialTest = !p.components[0].containsOfficialTest; },
    p => { p.splits.confirmation.pop(); },
    p => { p.splits.fit.push(p.splits.confirmation[0]); },
    p => { p.splits.excludedOfficialTrain.pop(); },
    p => { p.units.confirmation.pop(); },
    p => { p.counts.queries.fit++; },
    p => { p.topicIndependenceUnproven = false; },
    p => { p.noFitting = false; },
  ];
  for (const mutate of mutations) {
    const bad = structuredClone(plan); mutate(bad);
    assert.throws(() => audit(bad, train, test), 'corruption must fail');
  }
  assert.throws(() => audit(plan, [...train, train[0]], test), 'duplicate relation must fail');
  assert.throws(() => audit(plan, train, [...test, {...test[0], query: train[0].query}]), 'official overlap must fail');
  return {positiveChecks: 2, rejectedCorruptions: mutations.length + 2};
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const root = 'research/public-task-pilot/scifact-v1';
  const sourceBytes = fs.readFileSync(root + '/source.json');
  const source = JSON.parse(sourceBytes), planBytes = fs.readFileSync(root + '/split-plan.json');
  const plan = JSON.parse(planBytes);
  assert.equal(plan.sourceManifestSHA256, hash(sourceBytes));
  assert.equal(plan.scriptSHA256, hash(fs.readFileSync('research/public-task-pilot/scifact-groups.mjs')));
  for (const [name, digest] of Object.entries(source.artifacts)) assert.equal(hash(fs.readFileSync(root + '/' + name)), digest);
  const train = JSON.parse(fs.readFileSync(root + '/labels-train.json'));
  const test = JSON.parse(fs.readFileSync(root + '/labels-test.json'));
  const report = {time: new Date().toISOString(), sourceManifestSHA256: hash(sourceBytes),
    splitPlanSHA256: hash(planBytes), auditorSHA256: hash(fs.readFileSync(fileURLToPath(import.meta.url))),
    counts: audit(plan, train, test), controls: controls(plan, train, test),
    independentBreadthFirstReconstruction: true, noModelRuns: true, noFitting: true,
    topicIndependenceUnproven: true, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'};
  fs.writeFileSync(root + '/split-audit.json', JSON.stringify(report, null, 2) + '\n', {flag: 'wx', mode: 0o600});
  console.log(JSON.stringify(report, null, 2));
}
