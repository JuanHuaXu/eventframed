import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const original=readFileSync(process.argv[2]),bytes=readFileSync(process.argv[3]),old=JSON.parse(original),data=JSON.parse(bytes);
assert.equal(createHash('sha256').update(original).digest('hex'),data.inputSHA256);
for(const [p,h]of Object.entries(data.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
function table(noise){const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{const b=Array.from({length:4},(_,i)=>(h>>i)&1),v=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]];return v[t]?1-errors[t]:errors[t];}));}
function posterior(history,likelihood){
  const weights=Array.from({length:16},(_,h)=>{
    let weight=1/16;
    for(let t=0;t<8;t++){
      const q=likelihood[t][h],reports=history.filter(s=>s.action[1]===t);let local=0;
      for(let ordinary=0;ordinary<2;ordinary++)for(let fresh=0;fresh<2;fresh++)for(let root=0;root<2;root++){
        let p=.25*(root?q:1-q);
        for(const s of reports){const [kind,,slot]=s.action,independent=kind?fresh:ordinary&&slot>0,emission=independent?q:root;p*=s.outcome?emission:1-emission;}
        local+=p;
      }
      weight*=local;
    }
    return weight;
  });
  const total=weights.reduce((s,x)=>s+x,0),cls=Array(4).fill(0);weights.forEach((w,h)=>cls[h%4]+=w/total);return cls;
}
let checks=0,maxError=0;const brier=(p,y)=>p.reduce((s,x,i)=>s+(x-Number(i===y))**2,0);
assert.equal(data.records.length,640);
for(let i=0;i<640;i++){
  const r=old.records[i],s=data.records[i];assert.deepEqual([r.seed,r.split,r.case],[s.seed,s.split,s.case]);const likelihood=table(r.case.endsWith('05')?.05:.2);
  for(const [arm,a]of Object.entries(s.arms)){
    const trace=r.arms[arm].trace,history=[];let area=0;
    assert.equal(trace.length,a.forecasts.length);
    const check=p=>{const expected=posterior(history,likelihood);p.forEach((v,k)=>{const e=Math.abs(v-expected[k]);assert(e<1e-12);maxError=Math.max(maxError,e);});checks++;return expected;};
    for(let j=0;j<trace.length;j++){const p=check(a.forecasts[j]);area+=trace[j].cost*brier(p,r.truth%4)/16;history.push(trace[j]);}
    const p=check(a.final);assert(Math.abs(area-a.credit_brier)<1e-12);assert(Math.abs(brier(p,r.truth%4)-a.final_brier)<1e-12);
  }
}
console.log(JSON.stringify({scope:'Independent full-trace enumeration of shared roots and both modes',checks,maxError,artifactSHA256:createHash('sha256').update(bytes).digest('hex'),sourceSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex')},null,2));
