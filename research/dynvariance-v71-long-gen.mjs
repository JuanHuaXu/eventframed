// Isolated audit runner clone; the full experiment's frozen runner is untouched.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const source='research/dynvariance-v71-preflight.mjs',raw=fs.readFileSync(source);
let s=raw.toString().replaceAll('research/dynvariance-v71-preflight.mjs','research/dynvariance-v71-long-audit.mjs');
const from="const packages = ['./internal/researchdynvariance','./internal/researchdynvarianceref'];";
assert.equal(s.split(from).length-1,1);
s=s.replace(from,"const packages = ['./internal/researchdynvariancecheck'];");
const output='research/dynvariance-v71-long-audit.mjs';fs.writeFileSync(output,s,{flag:'wx',mode:0o600});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.writeFileSync('research/dynvariance-v71-long-generation.json',JSON.stringify({source,sourceSHA256:hash(raw),output,outputSHA256:hash(s),changes:'test only the external full-journal audit; original experiment sources unchanged'},null,2)+'\n',{flag:'wx',mode:0o600});
