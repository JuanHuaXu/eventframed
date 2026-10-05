import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {projectedTransfer} from './projected-transfer.mjs';
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



const likelihood=table(.2),results=[],gates=[];let checks=0,clipped=0,maxConditionalRegret=-Infinity,minMovementRatio=1,maxCycles=0,maxGap=0;
const risk=(p,q)=>p.reduce((s,x,i)=>s+x*x-2*x*q[i]+q[i],0);
for(const worldNoise of [.1,.15,.2,.25,.3])for(const allocation of data.results){
  const schedule=data.types.map(t=>[0,t,0]);data.types.forEach((t,i)=>{for(let slot=0;slot<allocation.counts[i];slot++)schedule.push([1,t,slot]);});
  const sums=Array.from({length:16},(_,mask)=>({mask,mass:0,brier:[0,0,0,0],wrong:[0,0,0,0],movementRatio:0}));
  for(let bits=0;bits<1024;bits++){
    const history=schedule.map((action,i)=>({action,outcome:Boolean((bits>>(9-i))&1)})),predictions=expected(history,likelihood).ps;
    const scenarios=Array.from({length:16},(_,mask)=>{
      const weights=Array.from({length:16},(_,h)=>{
        const roots=new Map();let w=1/16;
        for(const s of history){const [kind,t]=s.action,q=likelihood[t][h];if(kind===0){roots.set(t,s.outcome);w*=s.outcome?q:1-q;}else{const copied=mask&(1<<data.types.indexOf(t));w*=copied?Number(s.outcome===roots.get(t)):s.outcome?q:1-q;}}
        return w;
      });
      const mass=weights.reduce((s,x)=>s+x,0),q=Array(4).fill(0);if(mass>0)weights.forEach((w,h)=>q[h%4]+=w/mass);
      return {weights,mass,q};
    });
    const laws=Array.from({length:4},(_,i)=>Array.from({length:4},(_,j)=>Number(i===j))),guard=projectedTransfer(predictions[1],predictions[0],laws,.01);
    predictions[0].forEach((p,i)=>assert(Math.abs(p-scenarios[0].q[i])<1e-12));
    for(const q of laws){const regret=risk(guard.forecast,q)-risk(predictions[1],q);assert(regret<=.01+1e-12);maxConditionalRegret=Math.max(maxConditionalRegret,regret);checks++;}
    const norm=(a,b)=>Math.sqrt(a.reduce((s,x,i)=>s+(x-b[i])**2,0)),denom=norm(predictions[1],predictions[0]);
    const ratio=denom>0?norm(predictions[1],guard.forecast)/denom:1;
    if(norm(predictions[0],guard.forecast)>1e-12)clipped++;minMovementRatio=Math.min(minMovementRatio,ratio);maxCycles=Math.max(maxCycles,guard.cycles);maxGap=Math.max(maxGap,guard.gap);
    predictions.push(guard.forecast);
    const worldLikelihood=table(worldNoise);
    const actualScenarios=Array.from({length:16},(_,mask)=>{
      const weights=Array.from({length:16},(_,h)=>{
        const roots=new Map();let w=1/16;
        for(const s of history){const [kind,t]=s.action,q=worldLikelihood[t][h];if(kind===0){roots.set(t,s.outcome);w*=s.outcome?q:1-q;}else{const copied=mask&(1<<data.types.indexOf(t));w*=copied?Number(s.outcome===roots.get(t)):s.outcome?q:1-q;}}
        return w;
      });
      const mass=weights.reduce((s,x)=>s+x,0),q=Array(4).fill(0);if(mass>0)weights.forEach((w,h)=>q[h%4]+=w/mass);
      return {weights,mass,q};
    });
    actualScenarios.forEach(({weights,mass},mask)=>{
      const out=sums[mask];out.mass+=mass;out.movementRatio+=mass*ratio;if(mass===0)return;
      predictions.forEach((p,i)=>{out.brier[i]+=weights.reduce((sum,w,h)=>sum+w*p.reduce((s,x,c)=>s+(x-Number(c===h%4))**2,0),0);const chosen=p.indexOf(Math.max(...p));if(Math.max(...p)>=.9)out.wrong[i]+=weights.reduce((s,w,h)=>s+(h%4===chosen?0:w),0);});
    });
  }
  for(let mask=0;mask<16;mask++){
    const out=sums[mask],old=allocation.cells[mask];assert(Math.abs(out.mass-1)<1e-12);
    if(worldNoise===.2)for(let i=0;i<3;i++){assert(Math.abs(out.brier[i]-old.brier[i])<1e-11);assert(Math.abs(out.wrong[i]-old.wrong[i])<1e-11);}
    const gain=out.brier[1]-out.brier[3];gates.push({worldNoise,counts:allocation.counts,mask,kind:'nonharm',gain,passed:gain>=-.01-1e-12});
    if(mask===0)gates.push({worldNoise,counts:allocation.counts,mask,kind:'gain',gain,passed:gain>=.005});
    if(mask===15){const reduction=out.wrong[0]-out.wrong[3];gates.push({worldNoise,counts:allocation.counts,mask,kind:'false_confidence',gain:reduction,passed:reduction>=.05});}
  }
  results.push({worldNoise,counts:allocation.counts,cells:sums});
}
assert.equal(gates.length,900);
const sources=['research/hypothesis-observation/projected-transfer.mjs','research/hypothesis-observation/test-projected-transfer.mjs','research/hypothesis-observation/robust-transfer.mjs','research/hypothesis-observation/test-robust-transfer.mjs','research/hypothesis-observation/simplex-guard-noise.mjs','research/hypothesis-observation/SIMPLEX_GUARD_PROTOCOL.md'];
console.log(JSON.stringify({scope:'Outcome-simplex guard: predictor fixed at .2, actual worlds .1/.15/.2/.25/.3; per-outcome regret bounded without a measurement-likelihood premise',inputSHA256:createHash('sha256').update(input).digest('hex'),hashes:Object.fromEntries(sources.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),checks,clipped,totalVectors:51200,minMovementRatio,maxCycles,maxGap,maxConditionalRegret,results,gates,screen_passed:gates.every(g=>g.passed)},null,2));





