import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

// Frozen consumed-data decomposition. Counterfactual feedback changes only the
// mixer, holding all issued expert forecasts fixed. It is NOT a counterfactual
// retraining experiment, deployable oracle, or fresh confirmation.
const raw=process.argv[2],comparisonPath=process.argv[3];
const comparisonBytes=readFileSync(comparisonPath),comparison=JSON.parse(comparisonBytes);
const arms=[0,1,2,3,10,11],prior=[.95,.01,.01,.01,.01,.01];
function weights(steps,t,mode){
  const w=prior.slice();
  for(let j=0;j<t;j++){
    const s=steps[j];
    const eligible=mode===2||(!s.Missing&&(mode===1||j+s.Delay<=t));
    if(eligible){
      let z=0;for(let k=0;k<6;k++){w[k]*=s.Y?s.P[arms[k]]:1-s.P[arms[k]];z+=w[k];}
      assert(z>0);for(let k=0;k<6;k++)w[k]/=z;
    }
    for(let k=0;k<6;k++)w[k]=.999*w[k]+.001*prior[k];
  }
  return w;
}
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const digest=createHash('sha256'),stream=createReadStream(raw);stream.on('data',b=>digest.update(b));
let header,identities=0,asOfChecks=0,maxIdentityError=0,maxReferenceError=0;
const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),metrics=Array.from({length:5},()=>({actual:0,delayFree:0,completeFeedback:0,hull:0,noise:0,delayCost:0,missingCost:0,selectionCost:0,hullCost:0,segmentWeight:0,completeSegmentWeight:0}));
  assert.equal(r.Steps.length,256);
  for(let t=0;t<256;t++){
    const s=r.Steps[t],ws=[0,1,2].map(mode=>weights(r.Steps,t,mode));
    const ps=ws.map(w=>w.reduce((a,v,k)=>a+v*s.P[arms[k]],0));
    const min=Math.min(...arms.map(k=>s.P[k])),max=Math.max(...arms.map(k=>s.P[k]));
    assert(ps.every(p=>p>=min-1e-12&&p<=max+1e-12));
    const h=Math.max(min,Math.min(max,s.Q)),noise=s.Q*(1-s.Q);
    const loss=p=>(p-s.Q)**2+noise,ls=ps.map(loss),hl=loss(h);
    const values={actual:ls[0],delayFree:ls[1],completeFeedback:ls[2],hull:hl,noise,
      delayCost:ls[0]-ls[1],missingCost:ls[1]-ls[2],selectionCost:ls[2]-hl,hullCost:hl-noise,
      segmentWeight:ws[0][4]+ws[0][5],completeSegmentWeight:ws[2][4]+ws[2][5]};
    assert(values.selectionCost>=-1e-12&&values.hullCost>=-1e-12);
    const error=Math.abs(ls[0]-noise-values.delayCost-values.missingCost-values.selectionCost-values.hullCost);
    maxIdentityError=Math.max(maxIdentityError,error);assert(error<1e-12);identities++;
    for(const stage of [0,1+Math.floor(t/64)])for(const [k,v] of Object.entries(values))metrics[stage][k]+=v/(stage===0?256:64);
  }
  if(r.Index===0)for(const t of [0,32,160,255]){
    const changed=r.Steps.map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
    assert.deepEqual(weights(r.Steps,t,0),weights(changed,t,0));asOfChecks++;
    for(const mode of [1,2]){
      const future=r.Steps.map((s,j)=>({...s,Q:1-s.Q,Y:j>=t?!s.Y:s.Y}));
      assert.deepEqual(weights(r.Steps,t,mode),weights(future,t,mode));asOfChecks++;
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,metrics});
}
assert.equal(records.length,2688);assert.equal(identities,688128);assert.equal(asOfChecks,1008);
const rawHash=digest.digest('hex');assert.equal(rawHash,comparison.artifactSHA256);
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  assert.deepEqual(rs.map(r=>r.index),Array.from({length:32},(_,i)=>i));
  for(let stage=0;stage<5;stage++){
    const metrics=Object.fromEntries(Object.keys(rs[0].metrics[stage]).map(k=>[k,interval(rs.map(r=>r.metrics[stage][k]))]));
    groups.push({phase,case:c,schedule,stage,metrics});
    if(stage===0||stage===4){
      const ref=comparison.groups.find(g=>g.phase===phase&&g.scenario===c&&g.schedule===schedule&&g.segment===(stage===0?0:1));
      assert(ref);const error=Math.abs(metrics.actual.mean-ref.meanBrier[2]);maxReferenceError=Math.max(maxReferenceError,error);assert(error<1e-12);
    }
  }
}
console.log(JSON.stringify({scope:'Consumed fixed-tape feedback/selection/convex-hull decomposition; oracle comparisons are not deployable',artifactSHA256:rawHash,comparisonSHA256:createHash('sha256').update(comparisonBytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),arms,prior,transition:.001,records:records.length,identities,asOfChecks,maxIdentityError,maxReferenceError,groups},null,2));
