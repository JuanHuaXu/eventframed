import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import test from 'node:test';
import {fileURLToPath} from 'node:url';
import {verify} from './service-verify.mjs';
const source=path.dirname(fileURLToPath(import.meta.url)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function clone() {
 const root=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-service-verifier-negative-'));
 for(const e of fs.readdirSync(source,{withFileTypes:true}))if(e.isFile()) {
  if(e.name.endsWith('.json')||e.name==='SERVICE_PROTOCOL.md'||e.name.startsWith('adjacent-'))fs.copyFileSync(path.join(source,e.name),path.join(root,e.name));
  else fs.symlinkSync(path.join(source,e.name),path.join(root,e.name));
 }
 return root;
}
test('service verifier rejects a hidden failed preflight',async()=>{
 const root=clone(),file=path.join(root,'service-run-results.json'),r=JSON.parse(fs.readFileSync(file));r.commands[0].status=1;fs.writeFileSync(file,JSON.stringify(r));await assert.rejects(verify(root,{sources:false}),assert.AssertionError);
});
test('service verifier requires every named adjacent case',async()=>{
 const root=clone(),file=path.join(root,'service-run-results.json'),r=JSON.parse(fs.readFileSync(file)),c=r.commands[0],target=path.join(root,c.transcript);
 const text=fs.readFileSync(target,'utf8').replace('--- PASS: TestDeleteRetentionCompactAndBackup','--- PASS: RemovedRequiredCase');fs.writeFileSync(target,text);c.sha256=hash(text);c.bytes=Buffer.byteLength(text);fs.writeFileSync(file,JSON.stringify(r));await assert.rejects(verify(root,{sources:false}),assert.AssertionError);
});
test('service verifier rejects changed frozen protocol',async()=>{
 const root=clone(),file=path.join(root,'SERVICE_PROTOCOL.md');fs.appendFileSync(file,'\npost hoc gate change\n');await assert.rejects(verify(root,{sources:false}),assert.AssertionError);
});
test('service verifier rejects fabricated aggregate verdict',async()=>{
 const root=clone(),file=path.join(root,'service-evaluation.json'),r=JSON.parse(fs.readFileSync(file));r.packetQualityPass=!r.packetQualityPass;fs.writeFileSync(file,JSON.stringify(r));await assert.rejects(verify(root,{sources:false}),assert.AssertionError);
});
