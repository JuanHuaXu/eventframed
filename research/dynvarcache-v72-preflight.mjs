// Preserve each attempted preflight before executing it; no production writes.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync, spawnSync} from 'node:child_process';

const stage = process.argv[2];
assert(/^[a-z0-9-]+$/.test(stage));
const root = `research/dynvarcache-v72-${stage}`;
assert(!fs.existsSync(root), 'attempt archives are immutable');
const parentPath = 'research/checkpoint-2026-10-04-dynvariance-v71/manifest.json';
const hash = p => crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
assert.equal(hash(parentPath), '9d6cb6dcb7fc8281b407ebc9cb5035c27a2e8b29bcf9230013517e2da9a2b297');
const parent = JSON.parse(fs.readFileSync(parentPath));
const tracked = execFileSync('git', ['diff','--name-only'], {encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked, Object.keys(parent.trackedHashes).sort());
for (const [p,h] of Object.entries(parent.trackedHashes)) assert.equal(hash(p),h,p);
fs.mkdirSync(root,{mode:0o700});
const packages = ['./internal/researchdynvarcache','./internal/researchdynvarcachecheck'];
const listing = execFileSync('go',['list','-deps','-test','-f','{{.Dir}}|{{join .GoFiles ","}}|{{join .TestGoFiles ","}}|{{join .XTestGoFiles ","}}',...packages],{encoding:'utf8'});
const files = {};
for (const row of listing.trim().split('\n')) {
  const [dir,...sets]=row.split('|');
  if (!dir) continue;
  for (const name of sets.flatMap(s=>s.split(',')).filter(Boolean)) {
    const p=path.join(dir,name);
    if (!fs.existsSync(p)) continue; // generated test-main is not a saved source file
    files[p]=hash(p);
  }
}
for (const p of ['go.mod','go.sum','research/dynvarcache-v72-preflight.mjs']) files[path.resolve(p)]=hash(p);
for (const [p,h] of Object.entries(files)) {
  const out=root+'/source/'+p.replace(/^\//,'');
  fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});
  fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);
  assert.equal(hash(out),h,p);
}
const freeze={stage,time:new Date().toISOString(),parent:{path:parentPath,sha256:hash(parentPath)},files,protected:parent.trackedHashes,dependencyListing:listing,goVersion:execFileSync('go',['version'],{encoding:'utf8'}).trim(),packages,generatedTestMainNotArchived:true};
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const jobs=[['unit',['test','-count=1','-v',...packages]],['race',['test','-race','-count=1',...packages]],['vet',['vet',...packages]]];
const checks=[];
for (const [name,args] of jobs) {
  const t=performance.now();
  const proc=spawnSync('go',args,{encoding:'utf8',maxBuffer:64*1024*1024});
  const output=(proc.stdout??'')+(proc.stderr??'');
  fs.writeFileSync(root+'/'+name+'.log',output,{flag:'wx',mode:0o600});
  checks.push({name,args,exitCode:proc.status,signal:proc.signal,error:proc.error?.message??null,seconds:(performance.now()-t)/1000,logSHA256:hash(root+'/'+name+'.log')});
  console.log(JSON.stringify(checks.at(-1)));
}
for (const [p,h] of Object.entries(files)) assert.equal(hash(p),h,p);
for (const [p,h] of Object.entries(parent.trackedHashes)) assert.equal(hash(p),h,p);
const completed={stage,time:new Date().toISOString(),checks,allJobsTerminal:checks.every(x=>x.exitCode!==null),allChecksPass:checks.every(x=>x.exitCode===0),sourceUnchanged:true,scientificExperimentNotRun:true,allSevenGoalsOpen:true};
fs.writeFileSync(root+'/completed.json',JSON.stringify(completed,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,...completed}));
if (!completed.allChecksPass) process.exitCode=1;
