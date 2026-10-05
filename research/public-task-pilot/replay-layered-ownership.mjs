import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';

// Diagnostic persistent-state model, not an ANN update algorithm. Vector hashes
// stand for immutable payload handles; actual float-vector ownership is untested.
const input = 'research/public-task-pilot/hnsw-touch-results-v2.json';
const bytes = fs.readFileSync(input);
const rows = JSON.parse(bytes);
const equal = (a, b) => JSON.stringify(a) === JSON.stringify(b);
function freeze(value) {
  if (value && typeof value === 'object') {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}
function lookup(tree, id) {
  for (let bit = 31; bit >= 0 && tree; bit--) tree = tree[(id >>> bit) & 1];
  return tree?.record;
}
function replace(tree, id, bit, record, stats) {
  stats.paths++;
  if (bit < 0) return Object.freeze({ record });
  const side = (id >>> bit) & 1;
  return Object.freeze({ ...tree, [side]: replace(tree?.[side], id, bit - 1, record, stats) });
}
function prepare(root, edits, global, cap = 128) {
  if (edits.length > cap) throw new Error('edit capacity');
  const seen = new Set();
  const stats = { paths: 0, records: 0, links: 0, reusedVectors: 0 };
  let tree = root.tree;
  for (const [key, raw] of edits) {
    const id = Number(key);
    assert(Number.isInteger(id) && id >= 0 && id <= 0xffffffff);
    assert(!seen.has(id)); seen.add(id);
    let record;
    if (raw !== undefined) {
      const copied = structuredClone(raw);
      const links = copied.Links.flat().length + copied.Backlinks.flat().length;
      // Explicit diagnostic bounds; overflow rejects the whole unpublished edit.
      if (links > 1024 || copied.Links.length > 32 || copied.Backlinks.length > 32) {
        throw new Error('adjacency capacity');
      }
      const old = lookup(root.tree, id);
      const payload = old?.payload.hash === copied.VectorHash
        ? old.payload : Object.freeze({ hash: copied.VectorHash });
      if (payload === old?.payload) stats.reusedVectors++;
      record = Object.freeze({ data: freeze(copied), payload });
      stats.records++; stats.links += links;
    }
    tree = replace(tree, id, 31, record, stats);
  }
  return { root: Object.freeze({ tree, global }), stats };
}
function verify(root, snapshot, allIDs) {
  assert.equal(root.global, snapshot.Global);
  for (const id of allIDs) assert.deepEqual(lookup(root.tree, Number(id))?.data, snapshot.Nodes[id]);
}
const reports = [];
for (const n of [800, 6400]) {
  const group = rows.filter(r => r.N === n);
  assert.equal(group.length, 16);
  const allIDs = new Set(group.flatMap(r => [...Object.keys(r.Before.Nodes), ...Object.keys(r.After.Nodes)]));
  let root = prepare({ tree: undefined }, Object.entries(group[0].Before.Nodes), group[0].Before.Global, n).root;
  const history = [[root, group[0].Before]];
  for (const row of group) {
    verify(root, row.Before, allIDs);
    const edits = [...allIDs].filter(id => !equal(row.Before.Nodes[id], row.After.Nodes[id]))
      .map(id => [id, row.After.Nodes[id]]);
    assert.equal(edits.length, row.Changed);
    assert.throws(() => prepare(root, edits, row.After.Global, edits.length - 1), /edit capacity/);
    verify(root, row.Before, allIDs);
    const bad = structuredClone(edits.find(([, value]) => value !== undefined));
    bad[1].Links = [Array(1025).fill(0)];
    assert.throws(() => prepare(root, [bad], row.After.Global), /adjacency capacity/);
    const candidate = prepare(root, edits, row.After.Global);
    verify(root, row.Before, allIDs); // Preparing must not publish or mutate.
    for (const id of allIDs) {
      if (equal(row.Before.Nodes[id], row.After.Nodes[id])) {
        assert.equal(lookup(root.tree, Number(id)), lookup(candidate.root.tree, Number(id)));
      }
    }
    // Mutating caller buffers after preparation cannot alter the candidate.
    const owned = structuredClone(edits);
    const independent = prepare(root, owned, row.After.Global);
    for (const [, value] of owned) if (value) { value.ID = 'mutated'; value.Links.length = 0; }
    verify(independent.root, row.After, allIDs);
    root = candidate.root;
    history.push([root, row.After]);
    verify(root, row.After, allIDs);
    assert.equal(candidate.stats.paths, 33 * edits.length);
    reports.push({ n, kind: row.Kind, changed: edits.length, ...candidate.stats });
  }
  for (const [old, expected] of history) verify(old, expected, allIDs);
}
const result = {
  inputSHA256: crypto.createHash('sha256').update(bytes).digest('hex'),
  harnessSHA256: crypto.createHash('sha256').update(fs.readFileSync(new URL(import.meta.url))).digest('hex'),
  passed: true, operations: reports.length, retainedSnapshotsChecked: 34,
  scope: 'Captured logical-state replay; no dynamic ANN preparation, vector payload, durability, or throughput claim',
  reports,
};
fs.writeFileSync('research/public-task-pilot/layered-ownership-results.json', JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify(result));
