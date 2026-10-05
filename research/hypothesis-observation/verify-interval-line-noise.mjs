import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
const dir='research/hypothesis-observation/',path=dir+'interval-line-noise.mjs',input=dir+'renewal-allocation-exact.json';
const archived=readFileSync(dir+'interval-line-noise.json'),data=JSON.parse(archived);
const replay=execFileSync(process.execPath,[path,input],{maxBuffer:32*1024*1024});
assert(archived.equals(replay));
const old=JSON.parse(readFileSync(dir+'optimistic-noise-mismatch.json'));
data.results.forEach((r,i)=>{assert.equal(r.worldNoise,old.results[i].worldNoise);assert.deepEqual(r.counts,old.results[i].counts);r.cells.forEach((c,j)=>{assert.equal(c.mass,old.results[i].cells[j].mass);assert.deepEqual(c.brier.slice(0,3),old.results[i].cells[j].brier.slice(0,3));assert.deepEqual(c.wrong.slice(0,3),old.results[i].cells[j].wrong.slice(0,3));});});
let code=readFileSync(path,'utf8').replace("'./robust-transfer.mjs'","'./research/hypothesis-observation/robust-transfer.mjs'").replace("'./noise-envelope.mjs'","'./research/hypothesis-observation/noise-envelope.mjs'").replace('process.argv[2]','process.argv[1]');
code=code.replace('actualScenarios.forEach(({weights,mass},mask)=>{','actualScenarios.forEach(({weights,mass,q},mask)=>{');
const direct='weights.reduce((sum,w,h)=>sum+w*p.reduce((s,x,c)=>s+(x-Number(c===h%4))**2,0),0)';
assert(code.includes(direct));code=code.replace(direct,'mass*risk(p,q)');
const alternative=JSON.parse(execFileSync(process.execPath,['--input-type=module','-e',code,input],{maxBuffer:32*1024*1024}));
let checks=0,maxError=0;
data.results.forEach((r,i)=>r.cells.forEach((c,j)=>c.brier.forEach((b,k)=>{const error=Math.abs(b-alternative.results[i].cells[j].brier[k]);assert(error<1e-12);maxError=Math.max(maxError,error);checks++;})));
assert.deepEqual(data.gates.map(g=>g.passed),alternative.gates.map(g=>g.passed));
const componentTests=JSON.parse(execFileSync(process.execPath,[dir+'test-noise-envelope.mjs']));
const summary=[.1,.15,.2,.25,.3].map(n=>{const gs=data.gates.filter(g=>g.worldNoise===n);return {noise:n,passed:gs.filter(g=>g.passed).length,failures:Object.fromEntries(['nonharm','gain','false_confidence'].map(k=>[k,gs.filter(g=>g.kind===k&&!g.passed).length])),worstHarm:Math.max(...gs.filter(g=>g.kind==='nonharm').map(g=>-g.gain)),genuineGainRange:[Math.min(...gs.filter(g=>g.kind==='gain').map(g=>g.gain)),Math.max(...gs.filter(g=>g.kind==='gain').map(g=>g.gain))]};});
const out={scope:'Byte-exact replay, all-world control parity, alternate conditional-risk identity; not an independent inference implementation',inputSHA256:createHash('sha256').update(archived).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-interval-line-noise.mjs')).digest('hex'),replay:true,allWorldControlsExact:true,checks,maxError,componentTests,summary};
writeFileSync(dir+'interval-line-noise-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));




