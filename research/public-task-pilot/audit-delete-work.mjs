import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';

// Static replay of the neighbor-set construction in removeAllConnections and
// reconnectNeighborsOptimized, not measured backend CPU or a concurrency trace.
const input = 'research/public-task-pilot/hnsw-touch-results-v2.json';
const source = 'research/public-task-pilot/candidate-libravdb-v1.6.13/internal/index/hnsw/delete.go';
const hash = p => crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const rows = JSON.parse(fs.readFileSync(input)).filter(r => r.Kind === 'delete');
assert.equal(rows.length, 16);
const reports = rows.map(r => {
  const entry = Object.entries(r.Before.Nodes).find(([,v]) => v.ID === r.ID);
  assert(entry); const [key, target] = entry; const id = Number(key);
  const levels = target.Links.map((links, level) => {
    // The implementation begins with the outgoing list, then appends unique
    // incoming nodes whose outgoing list actually contains the deleted ordinal.
    const neighbors = [...links];
    for (const incoming of target.Backlinks[level]) {
      const node = r.Before.Nodes[incoming];
      if (!node || node.Level < level || !node.Links[level].includes(id)) continue;
      if (!neighbors.includes(incoming)) neighbors.push(incoming);
    }
    const valid = neighbors.filter(n => r.Before.Nodes[n] !== undefined);
    const d = valid.length;
    return { level, neighbors: d, pairDistanceCalls: d * (d - 1) / 2,
      matrixFloat32Bytes: d < 2 ? 0 : 4 * d * d };
  });
  return { n: r.N, id: r.ID, changedRecords: r.Changed, levels,
    pairDistanceCalls: levels.reduce((s,x)=>s+x.pairDistanceCalls,0),
    matrixFloat32Bytes: levels.reduce((s,x)=>s+x.matrixFloat32Bytes,0),
    globalChanged: r.Before.Global !== r.After.Global,
    registryScanLowerBoundIfEntryReplacement: Math.max(...Object.keys(r.Before.Nodes).map(Number)) + 1 };
});
const result = { inputSHA256: hash(input), sourceSHA256: hash(source),
  harnessSHA256: hash(new URL(import.meta.url)),
  scope: 'Derived from captured state and single-writer source; excludes extra heuristic distance work, arena alignment and metadata scans',
  reports };
fs.writeFileSync('research/public-task-pilot/delete-work-results.json', JSON.stringify(result,null,2)+'\n');
for (const r of reports) console.log(JSON.stringify({ n:r.n,id:r.id,changed:r.changedRecords,pairs:r.pairDistanceCalls,matrixBytes:r.matrixFloat32Bytes,maxNeighbors:Math.max(...r.levels.map(x=>x.neighbors)),globalChanged:r.globalChanged }));
