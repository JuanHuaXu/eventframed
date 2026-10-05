import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {readEnvelope} from './eager-load-v44-stream.mjs';

const root = fs.mkdtempSync(path.join(os.tmpdir(), 'eventframe-v44-stream-'));
const header = {Type: 'header', Trials: 2}, footer = {Type: 'footer', Trials: 2};
const a = {Type: 'trial', value: 1}, b = {Type: 'trial', value: 2};
const good = [header, a, b, footer];
const text = rows => rows.map(row => typeof row === 'string' ? row : JSON.stringify(row)).join('\n') + '\n';
const file = (name, rows) => { const p = path.join(root, name); fs.writeFileSync(p, text(rows), {flag: 'wx'}); return p; };
const p = file('good.ndjson', good), seen = [];
const report = await readEnvelope(p, async row => { await Promise.resolve(); seen.push(row.value); }, 2);
assert.deepEqual(seen, [1, 2]);
assert.equal(report.rawSHA256, crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'));
const bad = [good.slice(1), good.slice(0, -1), [header, a, footer], [...good, a],
  [header, a, b, b, footer], [...good, footer], [{...header, Trials: 3}, a, b, footer],
  [header, a, b, {...footer, Trials: 3}], [header, '', a, b, footer], [header, '{bad', b, footer]];
for (let i = 0; i < bad.length; i++) await assert.rejects(readEnvelope(file('bad-' + i, bad[i]), () => {}, 2));
await assert.rejects(readEnvelope(p, () => { throw Error('scientific checker rejection'); }, 2));
console.log(JSON.stringify({positive: 1, byteHashMatchesWholeRead: true, corruptionControls: bad.length + 1,
  tempFixtureRoot: root, noRealRowsDropped: true}));
