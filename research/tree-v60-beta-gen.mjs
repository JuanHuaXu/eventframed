// Mechanical clone only. The alternative prior is patched explicitly afterward.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const source = 'research/tree-v60-identities.mjs', dest = 'research/tree-v60-beta-identities.mjs';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
assert(!fs.existsSync(dest));
const original = fs.readFileSync(source), cloned = original.toString().replaceAll('tree-v60-identities', 'tree-v60-beta-identities');
fs.writeFileSync(dest, cloned, {flag: 'wx', mode: 0o600});
fs.writeFileSync('research/tree-v60-beta-generation.json', JSON.stringify({source, sourceSHA256: hash(original),
  dest, initialCloneSHA256: hash(cloned), semanticPriorPatchSeparate: true}, null, 2) + '\n', {flag: 'wx', mode: 0o600});
