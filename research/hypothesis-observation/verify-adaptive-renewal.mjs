import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const bytes=readFileSync(process.argv[2]),data=JSON.parse(bytes);
for(const [p,h]of Object.entries(data.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
function table(noise){const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{const b=Array.from({length:4},(_,i)=>(h>>i)&1),v=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]];return v[t]?1-errors[t]:errors[t];}));}
function posterior(history,likelihood,uncertain){
  const weights=Array.from({length:16},(_,h)=>{
    let weight=1/16;
    for(let t=0;t<8;t++){
      const q=likelihood[t][h],reports=history.filter(s=>s.action[1]===t);let local=0;
      for(let ordinary=0;ordinary<2;ordinary++)for(let fresh=0;fresh<2;fresh++)for(let root=0;root<2;root++){
        let p=.5*(uncertain?.5:fresh)*(root?q:1-q);
        for(const s of reports){const [kind,,slot]=s.action,independent=kind?fresh:ordinary&&slot>0,emission=independent?q:root;p*=s.outcome?emission:1-emission;}
        local+=p;
      }
      weight*=local;
    }
    return weight;
  });
  const total=weights.reduce((s,x)=>s+x,0),cls=Array(4).fill(0);weights.forEach((w,h)=>cls[h%4]+=w/total);return cls;
}
let checks=0,maxError=0,armCount=0;const brier=(p,y)=>p.reduce((s,x,i)=>s+(x-Number(i===y))**2,0);
assert.equal(data.records.length,640);assert.equal(new Set(data.records.map(r=>r.seed)).size,640);
for(const r of data.records){
  const likelihood=table(r.case.endsWith('05')?.05:.2);
  assert.equal(Object.keys(r.arms).length,6);
  for(const [arm,a]of Object.entries(r.arms)){
    const history=[],counts=[Array(8).fill(0),Array(8).fill(0)];let area=0,credits=0;
    const check=p=>{const expected=posterior(history,likelihood,arm.startsWith('uncertain_'));p.forEach((v,k)=>{const e=Math.abs(v-expected[k]);assert(e<1e-11);maxError=Math.max(maxError,e);});checks++;return expected;};
    for(const s of a.trace){
      const p=check(s.forecast),[kind,t,slot]=s.action;
      assert(kind===0||kind===1);assert(Number.isInteger(t)&&t>=0&&t<8);assert.equal(slot,counts[kind][t]++);assert.equal(s.cost,kind+1);assert.equal(s.credit,credits);
      assert(counts[kind][t]<=(kind===0?4:8));if(arm.endsWith('regular'))assert.equal(kind,0);
      area+=s.cost*brier(p,r.truth%4)/16;credits+=s.cost;history.push(s);
    }
    const p=check(a.final);assert.equal(credits,16);assert.equal(a.credits,16);assert.equal(a.calls,history.length);assert.equal(a.renewals,counts[1].reduce((s,x)=>s+x,0));
    assert(Math.abs(area-a.credit_brier)<1e-11);assert(Math.abs(brier(p,r.truth%4)-a.final_brier)<1e-11);armCount++;
  }
}
console.log(JSON.stringify({scope:'Independent original and uncertain joint-likelihood reconstruction on adaptive traces',checks,armCount,maxError,artifactSHA256:createHash('sha256').update(bytes).digest('hex'),sourceSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex')},null,2));
