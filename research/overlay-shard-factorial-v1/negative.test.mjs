import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import test from 'node:test';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {verify} from './verify.mjs';
const root=path.dirname(fileURLToPath(import.meta.url)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function copy(){const tmp=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-factorial-negative-'));for(const d of fs.readdirSync(root,{withFileTypes:true}))if(d.isFile()&&!d.name.startsWith('quality-0-')&&!d.name.startsWith('quality-1-'))fs.copyFileSync(path.join(root,d.name),path.join(tmp,d.name));return tmp;}
async function rejects(mutator){const tmp=copy();try{mutator(tmp);await assert.rejects(()=>verify(tmp,{sources:false,qualityCheck:false}));}finally{fs.rmSync(tmp,{recursive:true});}}
test('missing required PASS is rejected even with updated transcript hash',()=>rejects(tmp=>{const f=path.join(tmp,'run-results.json'),run=JSON.parse(fs.readFileSync(f)),c=run.commands[0],target=path.join(tmp,c.transcript),b=fs.readFileSync(target,'utf8').replace(new RegExp('^--- PASS: '+c.required[0]+'.*\\n','m'),'');fs.writeFileSync(target,b);c.sha256=hash(b);c.bytes=Buffer.byteLength(b);fs.writeFileSync(f,JSON.stringify(run));}));
test('changed frozen thresholds are rejected',()=>rejects(tmp=>fs.appendFileSync(path.join(tmp,'PROTOCOL.md'),'\nnew threshold=200ms\n')));
test('fabricated published load means are rejected',()=>rejects(tmp=>{const f=path.join(tmp,'load-evaluation.json'),x=JSON.parse(fs.readFileSync(f));x.rows[0].metrics.recall.p99_ms=0;fs.writeFileSync(f,JSON.stringify(x));}));
test('failed race preflight cannot be hidden by a functional flag',()=>rejects(tmp=>{const f=path.join(tmp,'run-results.json'),x=JSON.parse(fs.readFileSync(f));x.commands[0].status=1;fs.writeFileSync(f,JSON.stringify(x));}));
test('unexpected ordinary load failure is rejected even with consistent receipt summaries',()=>rejects(tmp=>{
 const f=path.join(tmp,'run-results.json'),x=JSON.parse(fs.readFileSync(f)),c=x.commands.find(c=>c.kind==='load'&&c.arm==='shard');
 c.status=1;c.functional=true;const target=path.join(tmp,c.transcript),b=fs.readFileSync(target,'utf8')+'\n    public_capture_load_test.go:999: injected unexpected assertion\n--- FAIL: TestResearchPublicCaptureLoadV1 (1.00s)\n';
 fs.writeFileSync(target,b);c.sha256=hash(b);c.bytes=Buffer.byteLength(b);fs.writeFileSync(f,JSON.stringify(x));
 const e=path.join(tmp,'load-evaluation.json'),p=JSON.parse(fs.readFileSync(e));p.rows.find(r=>r.arm===c.arm&&r.k===c.frontier&&r.pair===c.pair).status=1;fs.writeFileSync(e,JSON.stringify(p));
}));
