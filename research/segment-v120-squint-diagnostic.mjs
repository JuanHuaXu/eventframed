import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

// Finite-grid Squint, not Squint-CE. The randomized expert loss supplies regret;
// the proper forecast scored below is its convex probability average. Delayed
// feedback uses ORIGINAL issued weights, never retrospectively revised ones.
// No immediate-feedback regret theorem is claimed for the delayed adaptation.
const variants=[[0,1,2,3],[0,1,2,3,8,9,10,11]],rates=[.5,.25,.125,.0625];
function stateWeights(R,V,prior){
  const logs=R.map((r,i)=>rates.map(eta=>Math.log(prior[i]/rates.length)+eta*r-eta*eta*V[i]));
  const max=Math.max(...logs.flat()),w=logs.map(row=>row.reduce((s,l,j)=>s+rates[j]*Math.exp(l-max),0)),z=w.reduce((s,x)=>s+x,0);
  return w.map(x=>x/z);
}
function run(steps,arms){
  const prior=arms.map((_,i)=>i===0?.95:.05/(arms.length-1)),R=arms.map(()=>0),V=arms.map(()=>0);
  const issued=[],predictions=[],seen=new Set();let maxPotential=0;
  for(let t=0;t<steps.length;t++){
    for(let j=0;j<t;j++)if(!seen.has(j)&&!steps[j].Missing&&j+steps[j].Delay<=t){
      const s=steps[j],loss=arms.map(k=>(s.P[k]-Number(s.Y))**2),average=loss.reduce((a,l,k)=>a+l*issued[j][k],0);
      for(let k=0;k<arms.length;k++){const r=average-loss[k];assert(Math.abs(r)<=1+1e-12);R[k]+=r;V[k]+=r*r;}
      seen.add(j);
    }
    const w=stateWeights(R,V,prior);issued.push(w);
    const p=w.reduce((a,v,k)=>a+v*steps[t].P[arms[k]],0);assert(p>0&&p<1);predictions.push(p);
    const potential=prior.reduce((a,pi,k)=>a+pi*rates.reduce((b,eta)=>b+Math.exp(eta*R[k]-eta*eta*V[k])/rates.length,0),0);
    maxPotential=Math.max(maxPotential,potential);
  }
  return {predictions,maxPotential};
}
// Exhaust every binary outcome sequence of length8 with fixed diverse experts.
// Full immediate feedback must obey the finite prior-potential bound.
let potentialChecks=0,maxImmediatePotential=0;
for(let bits=0;bits<256;bits++){
  const steps=Array.from({length:8},(_,t)=>({Y:Boolean(bits&(1<<t)),Delay:0,Missing:false,P:[.1,.4,.6,.9]}));
  const result=run(steps,[0,1,2,3]);assert(result.maxPotential<=1+1e-12);
  maxImmediatePotential=Math.max(maxImmediatePotential,result.maxPotential);potentialChecks++;
}
const raw=process.argv[2],comparisonBytes=readFileSync(process.argv[3]),comparison=JSON.parse(comparisonBytes);
const stream=createReadStream(raw),digest=createHash('sha256');stream.on('data',b=>digest.update(b));
let header,forecastChecks=0,asOfChecks=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),scores=[];
  for(const arms of variants){
    const result=run(r.Steps,arms),metric=[0,0];
    for(let t=0;t<256;t++){const q=r.Steps[t].Q,p=result.predictions[t],l=(p-q)**2+q*(1-q);metric[0]+=l/256;if(t>=192)metric[1]+=l/64;forecastChecks++;}
    scores.push(metric);
    if(r.Index===0)for(const t of [0,32,160,255]){
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
      assert.deepEqual(run(changed,arms).predictions,result.predictions.slice(0,t+1));asOfChecks++;
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores});
}
assert.equal(records.length,2688);assert.equal(forecastChecks,1376256);assert.equal(asOfChecks,672);
const artifactSHA256=digest.digest('hex');assert.equal(artifactSHA256,comparison.artifactSHA256);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  const ref=comparison.groups.find(g=>g.phase===phase&&g.scenario===c&&g.schedule===schedule&&g.segment===segment);assert(ref);
  groups.push({phase,case:c,schedule,segment,meanBrier:variants.map((_,v)=>mean(rs.map(r=>r.scores[v][segment]))),incumbentMeans:ref.meanBrier,fourToEightGain:interval(rs.map(r=>r.scores[0][segment]-r.scores[1][segment]))});
}
console.log(JSON.stringify({scope:'Consumed finite-grid Squint diagnostic, not fresh confirmation or Squint-CE',artifactSHA256,comparisonSHA256:createHash('sha256').update(comparisonBytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),variants,rates,prior:'generic64=.95, remainder uniform',potentialChecks,maxImmediatePotential,forecastChecks,asOfChecks,records,groups},null,2));
