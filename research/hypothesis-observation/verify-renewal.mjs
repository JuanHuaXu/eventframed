import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const path=process.argv[2],bytes=readFileSync(path),data=JSON.parse(bytes);
for(const [p,h]of Object.entries(data.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
function table(noise){
  const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];
  return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{const b=Array.from({length:4},(_,i)=>(h>>i)&1),v=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]];return v[t]?1-errors[t]:errors[t];}));
}
function posterior(history,likelihood){
  const weights=Array.from({length:16},(_,h)=>{
    let w=1/16;
    for(let t=0;t<8;t++){
      const ys=history.filter(s=>s.action[0]===0&&s.action[1]===t).map(s=>s.outcome),q=likelihood[t][h];
      if(ys.length){const independent=ys.reduce((p,y)=>p*(y?q:1-q),1),copied=ys.every(y=>y===ys[0])?(ys[0]?q:1-q):0;w*=.5*independent+.5*copied;}
    }
    for(const s of history)if(s.action[0]===1){const q=likelihood[s.action[1]][h];w*=s.outcome?q:1-q;}
    return w;
  });
  const z=weights.reduce((s,x)=>s+x,0),cls=Array(4).fill(0);weights.forEach((w,h)=>cls[h%4]+=w/z);return cls;
}
const brier=(p,y)=>p.reduce((s,x,i)=>s+(x-Number(i===y))**2,0);
let forecastChecks=0,maxError=0,arms=0;
assert.equal(data.records.length,640);
assert.equal(new Set(data.records.map(r=>r.seed)).size,640);
for(const r of data.records){
  const likelihood=table(r.case.endsWith('05')?.05:.2);
  for(const [name,a]of Object.entries(r.arms)){
    const history=[],counts=[Array(8).fill(0),Array(8).fill(0)];let spent=0,area=0;
    function check(actual){const expected=posterior(history,likelihood);for(let i=0;i<4;i++){maxError=Math.max(maxError,Math.abs(actual[i]-expected[i]));assert(Math.abs(actual[i]-expected[i])<1e-12);}forecastChecks++;return expected;}
    for(const s of a.trace){
      assert.equal(s.credit,spent);const p=check(s.forecast),[kind,t,slot]=s.action;
      assert(kind===0||kind===1);assert(t>=0&&t<8);assert.equal(slot,counts[kind][t]++);assert.equal(s.cost,kind+1);assert(counts[kind][t]<=(kind===0?4:8));
      if(name==='regular')assert.equal(kind,0);
      area+=s.cost*brier(p,r.truth%4)/16;spent+=s.cost;history.push(s);
    }
    const final=check(a.final);assert.equal(spent,16);assert.equal(a.credits,16);assert.equal(a.calls,history.length);
    assert.equal(a.renewals,history.filter(s=>s.action[0]===1).length);
    assert(Math.abs(area-a.credit_brier)<1e-12);assert(Math.abs(brier(final,r.truth%4)-a.final_brier)<1e-12);arms++;
  }
}
console.log(JSON.stringify({scope:'Independent batch marginalization of copy modes and fresh measurements; no online model import',arms,forecastChecks,maxError,artifactSHA256:createHash('sha256').update(bytes).digest('hex'),sourceSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex')},null,2));
