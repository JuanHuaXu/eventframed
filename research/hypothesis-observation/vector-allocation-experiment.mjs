import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {guardedTransfer} from './robust-transfer.mjs';
import {noiseEnvelope,multiplyAffine} from './noise-envelope.mjs';
import {uniformNoiseTarget} from './uniform-noise-target.mjs';
const input=readFileSync(process.argv[2]),data=JSON.parse(input);
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
  const hard=Array(4).fill(0),hardZ=ws[1].reduce((s,x)=>s+x,0);ws[1].forEach((w,h)=>hard[h%4]+=w/hardZ);
  return {ps:[hard,local,mix],components:masses.map(m=>m/total),evidence:Math.log(total)};
}



import {allocateVectorRisk,vectorRisk} from './vector-risk-allocation.mjs';

const types=[0,1,2,7],likelihood=table(.2),results=[],gates=[],solves=[];
const allocations=data.results.filter(r=>r.worldNoise===.2).map(r=>r.counts);
const dot=(a,b)=>a.reduce((s,x,i)=>s+x*b[i],0);
const elevate=c=>{while(c.length<11)c=multiplyAffine(c,1,1);return c;};
function basis(k,u){let c=1;for(let i=1;i<=k;i++)c*=(11-i)/i;return c*u**k*(1-u)**(10-k);}
let coefficientChecks=0,maxCoefficientError=0,maxPointwiseHarm=-Infinity;
for(const counts of allocations){
  const schedule=types.map(t=>[0,t,0]);types.forEach((t,i)=>{for(let slot=0;slot<counts[i];slot++)schedule.push([1,t,slot]);});
  const rows=Array.from({length:176},()=>({mass:Array(1024).fill(0),joint:Array(4096).fill(0),constant:0}));
  const records=Array.from({length:1024},(_,bits)=>{
    const history=schedule.map((action,i)=>({action,outcome:!!(bits&(1<<(9-i)))}));
    const predictions=expected(history,likelihood).ps,models=noiseEnvelope(history),target=uniformNoiseTarget(models);
    const base=predictions[1],d=target.map((v,i)=>v-base[i]),a=dot(d,d),db=dot(d,base);
    for(const model of models){
      const masses=elevate(model.total),classes=model.classCoefficients.map(elevate);
      for(let k=0;k<11;k++){
        const row=rows[model.mask*11+k],mass=masses[k];
        row.mass[bits]=mass;
        for(let y=0;y<4;y++){row.joint[bits*4+y]=classes[y][k];row.constant+=2*classes[y][k]*base[y]-mass*base[y]*base[y];}
      }
    }
    return {history,predictions,d};
  });
  rows.forEach(r=>assert(Math.abs(r.mass.reduce((s,m)=>s+m,0)-1)<1e-12));
  const objective={mass:Array(1024).fill(0),joint:Array(4096).fill(0),constant:0};
  for(let k=0;k<11;k++){objective.constant+=rows[k].constant/11;for(let x=0;x<1024;x++)objective.mass[x]+=rows[k].mass[x]/11;for(let i=0;i<4096;i++)objective.joint[i]+=rows[k].joint[i]/11;}
  const base=records.flatMap(r=>r.predictions[1]);
  const solution=allocateVectorRisk(objective,rows,base);
  solves.push({counts,...solution});
  records.forEach((r,x)=>{
    const base=r.predictions[1],p=solution.forecast.slice(x*4,x*4+4);
    r.predictions.push(p);
    for(let y=0;y<4;y++){const risk=p.reduce((s,v,i)=>s+(v-Number(i===y))**2-(base[i]-Number(i===y))**2,0);maxPointwiseHarm=Math.max(maxPointwiseHarm,risk);}
  });
  for(const worldNoise of [.1,.15,.2,.25,.3]){
    const worldLikelihood=table(worldNoise);
    const cells=Array.from({length:16},(_,mask)=>({mask,mass:0,brier:[0,0,0,0],wrong:[0,0,0,0]}));
    records.forEach(({history,predictions})=>{
      for(let mask=0;mask<16;mask++){
        const weights=Array.from({length:16},(_,h)=>{
          const roots=new Map();let w=1/16;
          for(const s of history){const [kind,t]=s.action,q=worldLikelihood[t][h];if(kind===0){roots.set(t,s.outcome);w*=s.outcome?q:1-q;}else{const copied=mask&(1<<types.indexOf(t));w*=copied?Number(s.outcome===roots.get(t)):s.outcome?q:1-q;}}
          return w;
        });
        const mass=weights.reduce((s,w)=>s+w,0),out=cells[mask];out.mass+=mass;
        predictions.forEach((p,i)=>{
          out.brier[i]+=weights.reduce((s,w,h)=>s+w*p.reduce((sum,v,y)=>sum+(v-Number(y===h%4))**2,0),0);
          const chosen=p.indexOf(Math.max(...p));
          if(Math.max(...p)>=.9)out.wrong[i]+=weights.reduce((s,w,h)=>s+(chosen===h%4?0:w),0);
        });
      }
    });
    const old=data.results.find(r=>r.worldNoise===worldNoise&&JSON.stringify(r.counts)===JSON.stringify(counts));
    cells.forEach((out,mask)=>{
      assert(Math.abs(out.mass-1)<1e-12);
      for(let i=0;i<3;i++){assert(Math.abs(out.brier[i]-old.cells[mask].brier[i])<1e-11);assert(Math.abs(out.wrong[i]-old.cells[mask].wrong[i])<1e-11);}
      const regret=out.brier[3]-out.brier[1],u=(worldNoise-.1)/.2;
      const polynomial=Array.from({length:11},(_,k)=>{
        const r=rows[mask*11+k];return basis(k,u)*vectorRisk(r,solution.forecast,4);
      }).reduce((s,v)=>s+v,0);
      const error=Math.abs(regret-polynomial);assert(error<1e-11);coefficientChecks++;maxCoefficientError=Math.max(maxCoefficientError,error);
      const gain=-regret;gates.push({counts,worldNoise,mask,kind:'nonharm',gain,passed:gain>=-.01-1e-12});
      if(mask===0)gates.push({counts,worldNoise,mask,kind:'gain',gain,passed:gain>=.005});
      if(mask===15){const reduction=out.wrong[0]-out.wrong[3];gates.push({counts,worldNoise,mask,kind:'false_confidence',gain:reduction,passed:reduction>=.05});}
    });
    results.push({counts,worldNoise,cells});
  }
}
assert.equal(gates.length,900);
const sources=['vector-risk-allocation.mjs','test-vector-risk-allocation.mjs','vector-allocation-experiment.mjs','noise-envelope.mjs','uniform-noise-target.mjs','VECTOR_ALLOCATION_PROTOCOL.md'].map(p=>'research/hypothesis-observation/'+p);
console.log(JSON.stringify({scope:'Full-vector finite-model population risk allocation across histories; continuous common-noise interval assumed, not empirical coverage',inputSHA256:createHash('sha256').update(input).digest('hex'),hashes:Object.fromEntries(sources.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),solves,coefficientChecks,maxCoefficientError,maxPointwiseHarm,results,gates,screen_passed:gates.every(g=>g.passed)},null,2));



