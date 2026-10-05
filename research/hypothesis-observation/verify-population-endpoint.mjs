import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
const dir='research/hypothesis-observation/',path=dir+'population-endpoint-objective.mjs',input=dir+'uniform-target-noise.json';
const raw=readFileSync(dir+'population-endpoint-objective.json'),data=JSON.parse(raw);
const replay=execFileSync(process.execPath,[path,input],{maxBuffer:32*1024*1024});
assert(raw.equals(replay));
let code=readFileSync(path,'utf8');
for(const name of ['robust-transfer','noise-envelope','uniform-noise-target','population-risk-allocation-long'])code=code.replace("'./"+name+".mjs'","'./research/hypothesis-observation/"+name+".mjs'");
code=code.replace('process.argv[2]','process.argv[1]');
const direct='weights.reduce((s,w,h)=>s+w*p.reduce((sum,v,y)=>sum+(v-Number(y===h%4))**2,0),0)';
assert(code.includes(direct));
code=code.replace(direct,'(mass+p.reduce((s,v,y)=>s+mass*v*v-2*v*weights.reduce((a,w,h)=>a+(h%4===y?w:0),0),0))');
// Reuse archived lambda only for the independent scoring calculation. The
// full unmodified optimization was already replayed byte-for-byte above.
code=code.replace('const solution=allocateRisk(objective,rows);',"const solution=JSON.parse(readFileSync('"+dir+"population-endpoint-objective.json')).solves.find(s=>JSON.stringify(s.counts)===JSON.stringify(counts));");
const alt=JSON.parse(execFileSync(process.execPath,['--input-type=module','-e',code,input],{maxBuffer:32*1024*1024}));
let checks=0,maxError=0;
data.results.forEach((r,i)=>r.cells.forEach((c,j)=>c.brier.forEach((v,k)=>{const e=Math.abs(v-alt.results[i].cells[j].brier[k]);assert(e<1e-12);maxError=Math.max(maxError,e);checks++;})));
assert.deepEqual(data.gates.map(g=>g.passed),alt.gates.map(g=>g.passed));
const testSource=readFileSync(dir+'test-population-risk-allocation.mjs','utf8').replace('./population-risk-allocation.mjs','./research/hypothesis-observation/population-risk-allocation-long.mjs');
const componentTests=JSON.parse(execFileSync(process.execPath,['--input-type=module','-e',testSource]));
const summary=[.1,.15,.2,.25,.3].map(n=>{const gs=data.gates.filter(g=>g.worldNoise===n);return {noise:n,passed:gs.filter(g=>g.passed).length,failures:Object.fromEntries(['nonharm','gain','false_confidence'].map(k=>[k,gs.filter(g=>g.kind===k&&!g.passed).length])),worstHarm:Math.max(...gs.filter(g=>g.kind==='nonharm').map(g=>-g.gain)),genuineGainRange:[Math.min(...gs.filter(g=>g.kind==='gain').map(g=>g.gain)),Math.max(...gs.filter(g=>g.kind==='gain').map(g=>g.gain))]};});
const out={scope:'Full solver replay, alternate scoring with frozen lambdas, analytic and feasible-grid solver checks; not an independently implemented optimizer',inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-population-endpoint.mjs')).digest('hex'),replay:true,checks,maxError,componentTests,maxPointwiseHarm:data.maxPointwiseHarm,coefficientChecks:data.coefficientChecks,maxCoefficientError:data.maxCoefficientError,summary};
writeFileSync(dir+'population-endpoint-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));


