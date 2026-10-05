// Mechanical streaming traversal. Metric definitions and gates remain identical
// to the preregistered readback; never retain both entire corpora in RAM.
import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const src='research/retention-v64c-readback.mjs',dest='research/retention-v64d-readback.mjs',raw=fs.readFileSync(src);let s=raw.toString();
function replace(a,b){assert(s.includes(a),a);s=s.replace(a,b)}
replace("async function load(p){const out=[];for await(const line of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))out.push(JSON.parse(line));return out}","async function* load(p){for await(const line of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(line)}");
replace("const rows=await load('research/retention-v64b-diagnostic/diagnostic.jsonl'),old=await load('research/tree-v60-diagnostic/diagnostic.jsonl');assert.equal(rows.length,41);assert.equal(old.length,41);","const rows=load('research/retention-v64b-diagnostic/diagnostic.jsonl'),old=load('research/tree-v60-diagnostic/diagnostic.jsonl');const manifest=(await rows.next()).value,oldManifest=(await old.next()).value;assert.equal(oldManifest.Worlds,40);assert.equal(manifest.Worlds,40);");
s=s.replaceAll('rows[0]','manifest');
replace('for(let w=1;w<rows.length;w++) {','for(let w=1;w<=40;w++) {\n const current=(await rows.next()).value,previous=(await old.next()).value;assert(current&&previous);');
s=s.replaceAll('rows[w]','current').replaceAll('old[w]','previous');
replace('const summaries={};',"assert((await rows.next()).done);assert((await old.next()).done);\nconst summaries={};");
fs.writeFileSync(dest,s,{flag:'wx',mode:0o600});const hash=b=>crypto.createHash('sha256').update(b).digest('hex');fs.writeFileSync('research/retention-v64d-stream-generation.json',JSON.stringify({source:src,sourceSHA256:hash(raw),output:dest,outputSHA256:hash(s),change:'stream two worlds at a time; no metric equations, cohort, data or gates changed'},null,2)+'\n',{flag:'wx',mode:0o600});
