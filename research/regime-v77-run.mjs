// Freeze one isolated attempt before executing tests. Failed attempts stay saved.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync, spawn} from 'node:child_process';
const stage = process.argv[2];
assert(/^[a-z0-9-]+$/.test(stage));
const root = `research/regime-v77-${stage}`;
assert(!fs.existsSync(root));
async function hash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
const parentPath = 'research/checkpoint-2026-10-05-mean-anchor-v75-ratio-v76/manifest.json';
assert.equal(await hash(parentPath), 'acd27d5b83186333790ae27fa3c07a27e540b039e6c1e41313ea3f8147b33554');
const parent = JSON.parse(fs.readFileSync(parentPath));
for (const [p, h] of Object.entries(parent.copies)) assert.equal(await hash(path.dirname(parentPath) + '/saved/' + p), h, p);
const tracked = execFileSync('git', ['diff', '--name-only'], {encoding: 'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked, Object.keys(parent.trackedHashes).sort());
for (const [p, h] of Object.entries(parent.trackedHashes)) assert.equal(await hash(p), h, p);
const closure = JSON.parse(execFileSync('go', ['run', './cmd/research-go-list-closure', './internal/researchregime'], {maxBuffer: 64*1024*1024}));
const files = {};
for (const row of closure) for (const key of ['GoFiles','CgoFiles','EmbedFiles']) for (const n of row[key] ?? []) {
  const p = path.resolve(row.Dir, n);
  files[p] = await hash(p);
}
for (const p of ['go.mod','go.sum','research/regime-v77-run.mjs','docs/experiments/mmm-regime-v77-protocol.md']) files[path.resolve(p)] = await hash(p);
fs.mkdirSync(root, {mode: 0o700});
const save = (p, value) => fs.writeFileSync(root+'/'+p, JSON.stringify(value,null,2)+'\n', {flag:'wx', mode:0o600});
for (const [p,h] of Object.entries(files)) {
  const out = root+'/source/'+p.replace(/^\//,'');
  fs.mkdirSync(path.dirname(out), {recursive:true, mode:0o700});
  fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);
  assert.equal(await hash(out),h,p);
}
save('freeze.json', {time:new Date().toISOString(), stage, parentPath, parentSHA256:await hash(parentPath), files, closure, protectedFiles:parent.trackedHashes,
  toolchain:execFileSync('go',['version'],{encoding:'utf8'}).trim(), scope:'changing shared explanation, exact dense/run-length preflight and explicitly capped approximation; no empirical rescue or full-seven-goal proof'});
async function unchanged() {for (const [p,h] of Object.entries({...files,...parent.trackedHashes})) assert.equal(await hash(p),h,p)}
const checks=[];
for (const [name,args] of [
  ['unit',['test','-v','-count=1','./internal/researchregime','-timeout=20m']],
  ['race',['test','-race','-v','-count=1','./internal/researchregime','-timeout=30m']],
  ['vet',['vet','./internal/researchregime']],
  ['benchmark',['test','-run=^$','-bench=Benchmark','-benchtime=100ms','-count=2','./internal/researchregime','-timeout=20m']],
]) {
  await unchanged();
  const env={...process.env}; for(const k of Object.keys(env)) if(k.startsWith('EVENTFRAME_')) delete env[k];
  if(name==='unit') env.EVENTFRAME_REGIME_V77_REPORT=path.resolve(root+'/audit.json');
  const fd=fs.openSync(root+'/'+name+'.log','wx',0o600), begin=performance.now();
  console.log('START',name,new Date().toISOString());
  let code;
  try {code=await new Promise((resolve,reject)=>{
    const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});
    c.on('error',reject);
    for(const s of [c.stdout,c.stderr]) s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});
    c.on('close',x=>resolve(x??-1));
  })} finally {fs.fsyncSync(fd);fs.closeSync(fd)}
  const row={name,args,exitCode:code,seconds:(performance.now()-begin)/1000,logSHA256:await hash(root+'/'+name+'.log')};
  checks.push(row);save(name+'-command.json',row);await unchanged();console.log(JSON.stringify(row));
  if(code!==0) break;
}
const pass=checks.length===4&&checks.every(x=>x.exitCode===0);
save('completed.json',{time:new Date().toISOString(),stage,checks,allJobsTerminal:true,allChecksPass:pass,sourceUnchanged:true,
  scientificQualityRescueEstablished:false,scalableApproximationCertified:false,wholeCohortTested:false,loadedServingEstablished:false,
  goals:Array(7).fill('OPEN'),goal:'ACTIVE'});
if(!pass) process.exitCode=1;
