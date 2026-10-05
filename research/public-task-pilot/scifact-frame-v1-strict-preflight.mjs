import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import assert from 'node:assert/strict';

const root = 'research/public-task-pilot/scifact-frame-v1-strict-preflight';
fs.mkdirSync(root, {mode:0o700});
const sources = {};
for (const p of ['internal/researchpublicframe/document.go','internal/researchpublicframe/document_test.go',
  'cmd/research-public-frames/main.go','docs/experiments/scifact-frame-v1-protocol.md']) {
  const b = fs.readFileSync(p), out = root+'/source/'+p;
  fs.mkdirSync(path.dirname(out), {recursive:true,mode:0o700});
  fs.writeFileSync(out,b,{flag:'wx',mode:0o600});
  sources[p] = crypto.createHash('sha256').update(b).digest('hex');
}
const args = ['test','-race','-count=1','-v','./internal/researchpublicframe','-run','^TestPublicFrameV1StrictJSONIdentity$'];
const r = spawnSync('go',args,{encoding:'utf8',maxBuffer:8<<20});
const raw = (r.stdout??'')+(r.stderr??'');
fs.writeFileSync(root+'/test.log',raw,{flag:'wx',mode:0o600});
fs.writeFileSync(root+'/result.json',JSON.stringify({command:['go',...args],code:r.status,signal:r.signal,
  sources,logSHA256:crypto.createHash('sha256').update(raw).digest('hex'),expectedRegression:true},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(raw);assert.equal(r.status,1,'must reproduce actual strict decoding defect');
