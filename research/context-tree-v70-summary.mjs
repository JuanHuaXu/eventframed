import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';

const root = process.cwd();
const filename = path.join(root, 'docs/experiments/mmm-context-tree-v70.jsonl');
const raw = fs.readFileSync(filename);
const [header, ...rows] = raw.toString().trim().split('\n').map(JSON.parse);
if (header.Version !== 'v70' || rows.length !== 512) throw Error('incomplete artifact');
for (const [file, expected] of Object.entries(header.Sources)) {
  const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(root, file))).digest('hex');
  if (actual !== expected) throw Error(`source changed: ${file}`);
}
const mean = xs => xs.reduce((s, x) => s + x, 0) / xs.length;
const groups = Map.groupBy(rows, r => JSON.stringify([r.Phase, r.Family, r.Noise, r.NewLabels]));
let primaryFailures = 0, protectionFailures = 0;
const cells = [];
for (const [key, rs] of groups) {
  if (rs.length !== 8 || new Set(rs.map(r => r.Seed)).size !== 8) throw Error('bad cell');
  const delta = rs.map(r => r.TreeBrier - r.SubsetBrier);
  if (delta.some(x => !Number.isFinite(x))) throw Error('nonfinite loss');
  const d = mean(delta);
  const half = 2.364624251 * Math.sqrt(mean(delta.map(x => (x-d)**2)) / 7);
  const [phase, family, noise, newLabels] = JSON.parse(key);
  const primary = family === 1 && newLabels >= 32;
  const primaryPass = !primary || d <= -.005;
  const protectionPass = d <= .005;
  if (!primaryPass) primaryFailures++;
  if (!protectionPass) protectionFailures++;
  cells.push({phase, family, noise, newLabels, subset:mean(rs.map(r=>r.SubsetBrier)), tree:mean(rs.map(r=>r.TreeBrier)), delta:d, descriptive95:[d-half,d+half], primary, primaryPass, protectionPass});
}
console.log(JSON.stringify({sha256:crypto.createHash('sha256').update(raw).digest('hex'), sourceCount:Object.keys(header.Sources).length, pairedFits:rows.length, primaryFailures, protectionFailures, finiteScreen:primaryFailures+protectionFailures===0?'PASS':'FAIL', cells}, null, 2));
