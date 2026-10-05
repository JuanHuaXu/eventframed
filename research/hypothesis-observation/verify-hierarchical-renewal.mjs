import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const input=readFileSync(process.argv[2]),bytes=readFileSync(process.argv[3]),old=JSON.parse(input),data=JSON.parse(bytes);
assert.equal(createHash('sha256').update(input).digest('hex'),data.inputSHA256);
for(const [p,h]of Object.entries(data.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
function table(noise){const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{const b=Array.from({length:4},(_,i)=>(h>>i)&1),v=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]];return v[t]?1-errors[t]:errors[t];}));}
function component(history,likelihood,freshPrior){
  return Array.from({length:16},(_,h)=>{
    let weight=1/16;
    for(let t=0;t<8;t++){
      const q=likelihood[t][h],reports=history.filter(s=>s.action[1]===t);let local=0;
      for(let ordinary=0;ordinary<2;ordinary++)for(let fresh=0;fresh<2;fresh++)for(let root=0;root<2;root++){
        let p=.5*(fresh?freshPrior:1-freshPrior)*(root?q:1-q);
        for(const s of reports){const [kind,,slot]=s.action,independent=kind?fresh:ordinary&&slot>0,emission=independent?q:root;p*=s.outcome?emission:1-emission;}
        local+=p;
      }
      weight*=local;
    }
    return weight;
  });
}
function expected(history,likelihood){
  const ws=[0,1,.5].map(p=>component(history,likelihood,p)),priors=[.25,.25,.5];
  const masses=ws.map((w,i)=>w.reduce((s,x)=>s+x,0)*priors[i]),total=masses.reduce((s,x)=>s+x,0);
  const localTotal=ws[2].reduce((s,x)=>s+x,0),local=Array(4).fill(0),mix=Array(4).fill(0);
  for(let h=0;h<16;h++){local[h%4]+=ws[2][h]/localTotal;mix[h%4]+=ws.reduce((s,w,i)=>s+priors[i]*w[h],0)/total;}
  return {ps:[local,mix],components:masses.map(m=>m/total),evidence:Math.log(total)};
}
const brier=(p,y)=>p.reduce((s,x,i)=>s+(x-Number(i===y))**2,0);
let checks=0,maxError=0,maxEvidenceError=0;
assert.equal(data.records.length,640);
for(let i=0;i<640;i++){
  const r=old.records[i],s=data.records[i];assert.deepEqual([r.seed,r.case,r.split],[s.seed,s.case,s.split]);const likelihood=table(r.case.endsWith('05')?.05:.2);
  for(const [arm,a]of Object.entries(s.arms)){
    const trace=r.arms[arm].trace,history=[],area=[0,0];assert.equal(a.forecasts.length,trace.length);
    const check=ps=>{const want=expected(history,likelihood);for(let k=0;k<2;k++)for(let h=0;h<4;h++){const e=Math.abs(ps[k][h]-want.ps[k][h]);maxError=Math.max(maxError,e);assert(e<1e-11);}checks+=2;return want;};
    for(let j=0;j<trace.length;j++){const want=check(a.forecasts[j]);for(let k=0;k<2;k++)area[k]+=trace[j].cost*brier(want.ps[k],r.truth%4)/16;history.push(trace[j]);}
    const want=check(a.final);for(let k=0;k<2;k++){assert(Math.abs(area[k]-a.credit_brier[k])<1e-11);assert(Math.abs(brier(want.ps[k],r.truth%4)-a.final_brier[k])<1e-11);}
    want.components.forEach((v,j)=>assert(Math.abs(v-a.components[j])<1e-11));maxEvidenceError=Math.max(maxEvidenceError,Math.abs(want.evidence-a.log_evidence));assert(Math.abs(want.evidence-a.log_evidence)<1e-10);
  }
}
console.log(JSON.stringify({scope:'Independent batch model/latent-state marginalization on all fixed traces',checks,maxError,maxEvidenceError,artifactSHA256:createHash('sha256').update(bytes).digest('hex'),sourceSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex')},null,2));
