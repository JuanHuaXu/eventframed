import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';
import {createInterface} from 'node:readline';
async function* rows(path) {
  for await(const line of createInterface({input:fs.createReadStream(path),crlfDelay:Infinity}))
    if(line.trim()) yield JSON.parse(line);
}
const dir='docs/experiments/';
const paths=['mmm-soft-learners-v120.jsonl','mmm-learned-degree-v1-forecasts.jsonl','mmm-degree-composition-v1-forecasts.jsonl'];
const iter=paths.map(p=>rows(dir+p));
assert.equal((await iter[0].next()).value.Version,'soft-learners-v120');
// Independently evaluate in log space, from the original prior at every audit
// clock. No settled-prefix checkpoint or probability-space update is reused.
const logsum=xs=>{const m=Math.max(...xs);return m+Math.log(xs.reduce((s,x)=>s+Math.exp(x-m),0));};
function reference(s,r,t,arm) {
  const uniform=arm>=4;
  const pi=Array.from({length:5},(_,j)=>uniform?.2:j===0?.95:.0125);
  let logs=pi.map(Math.log);
  for(let j=0;j<t;j++) {
    const e=s.Steps[j];
    assert(Number.isInteger(e.Delay)&&e.Delay>=0&&e.Delay<=31);
    if(!e.Missing&&j+e.Delay<=t) {
      const p=[...e.P.slice(0,4),r.P[j][2+arm%2]];
      logs=logs.map((x,k)=>x+(e.Y?Math.log(p[k]):Math.log1p(-p[k])));
      const z=logsum(logs);logs=logs.map(x=>x-z);
    }
    logs=logs.map((x,k)=>logsum([Math.log(.999)+x,Math.log(.001*pi[k])]));
  }
  const p=[...s.Steps[t].P.slice(0,4),r.P[t][2+arm%2]];
  return logs.reduce((v,x,k)=>v+Math.exp(x)*p[k],0);
}
let records=0,independent=0,maxError=0,linear=0;
for(;;) {
  const a=await iter[0].next(),b=await iter[1].next(),c=await iter[2].next();
  assert.equal(a.done,b.done);assert.equal(a.done,c.done);if(a.done) break;
  const s=a.value,r=b.value,out=c.value;
  for(const key of ['Phase','Case','Index','Schedule']) {assert.equal(s[key],r[key]);assert.equal(s[key],out[key]);}
  for(let t=0;t<256;t++) for(let arm=0;arm<2;arm++) {assert.equal(out.P[t][arm],r.P[t][arm]);linear++;}
  if(s.Index===0) for(const t of [0,1,32,160,255]) for(let arm=2;arm<6;arm++) {
    const p=reference(s,r,t,arm),error=Math.abs(p-out.P[t][arm]);
    maxError=Math.max(maxError,error);assert(error<1e-12);independent++;
  }
  records++;
}
assert.equal(records,2688);assert.equal(independent,1680);
const checks=JSON.parse(fs.readFileSync(dir+'mmm-degree-composition-v1-forecasts.jsonl.checks.json'));
const summary=JSON.parse(fs.readFileSync(dir+'mmm-degree-composition-v1-summary.json'));
const groups=new Map();
for(const r of checks.contrasts) {
  const k=[r.Phase,r.Case,r.Schedule].join(':');if(!groups.has(k)) groups.set(k,[]);groups.get(k).push(r);
}
const comparisons=[];
for(const [key,rs] of groups) {
  assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.Index)).size,32);
  for(let candidate=2;candidate<6;candidate++) for(let span=0;span<2;span++) {
    const control=candidate<4?0:1;
    const xs=rs.map(r=>r.loss[control][span]-r.loss[candidate][span]);
    const mean=xs.reduce((a,b)=>a+b,0)/32;
    const se=Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/(31*32));
    comparisons.push({key,candidate,control,span,mean,lower:mean-3.5*se,upper:mean+3.5*se});
  }
}
const paired=[];
for(let candidate=2;candidate<6;candidate++) {
  const cs=comparisons.filter(c=>c.candidate===candidate);
  paired.push({candidate,control:candidate<4?0:1,nonharm:cs.filter(c=>c.lower>=-.01).length,total:cs.length,
    gains:cs.filter(c=>c.mean>=.005&&c.lower>0).length,
    meanGain:cs.reduce((s,c)=>s+c.mean,0)/cs.length});
}
const failures=summary.gates.filter(g=>!g.pass);
const nonharmFailures=failures.filter(g=>g.type==='nonharm');
const failCounts=[];
for(let arm=2;arm<6;arm++) for(const control of ['linear','logistic','generic','boolean','markov'])
  failCounts.push({arm,control,count:nonharmFailures.filter(g=>g.arm===arm&&g.control===control).length});
const hashes={};
for(const p of [...paths,'mmm-degree-composition-v1-summary.json']) {
  const hash=crypto.createHash('sha256');for await(const chunk of fs.createReadStream(dir+p)) hash.update(chunk);
  hashes[p]=hash.digest('hex');
}
console.log(JSON.stringify({records,independent,maxError,linear,paired,failCounts,
  candidate5NonharmFailures:nonharmFailures.filter(g=>g.arm===5),comparisons,hashes},null,2));
