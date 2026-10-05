// Consumed-data capacity diagnostic. Oracle weights see true rates and MUST
// NOT enter inference, nomination, serving, or a fresh-confirmation claim.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const clamp=v=>Math.max(0,Math.min(1,v));
function optimum(q,p,priority=Array(p.length).fill(1)){
  assert.equal(q.length,3);q.forEach(v=>assert.equal(v.length,p.length));
  assert.equal(priority.length,p.length);assert(priority.every(v=>Number.isFinite(v)&&v>0));
  const mass=priority.reduce((z,v)=>z+v,0),objective=w=>p.reduce((z,v,i)=>z+priority[i]*(q.reduce((s,row,k)=>s+w[k]*row[i],0)-v)**2/mass,0);
  const candidates=[[1,0,0],[0,1,0],[0,0,1]];
  for(const[a,b]of [[0,1],[0,2],[1,2]]){
    let numerator=0,denominator=0;for(let i=0;i<p.length;i++){const d=q[a][i]-q[b][i];numerator+=priority[i]*d*(p[i]-q[b][i]);denominator+=priority[i]*d*d;}
    const t=denominator>0?clamp(numerator/denominator):0,w=[0,0,0];w[a]=t;w[b]=1-t;candidates.push(w);
  }
  let a=0,b=0,c=0,u=0,v=0;
  for(let i=0;i<p.length;i++){const x=q[0][i]-q[2][i],y=q[1][i]-q[2][i],z=q[2][i]-p[i],w=priority[i];a+=w*x*x;b+=w*x*y;c+=w*y*y;u+=w*x*z;v+=w*y*z;}
  const det=a*c-b*b;
  if(det>1e-20){const x=(b*v-c*u)/det,y=(b*u-a*v)/det;if(x>=0&&y>=0&&x+y<=1)candidates.push([x,y,1-x-y]);}
  candidates.sort((a,b)=>objective(a)-objective(b));const weights=candidates[0];assert(weights.every(v=>v>=0&&v<=1));
  return{weights,excess:objective(weights),brier:objective(weights)+p.reduce((z,v,i)=>z+priority[i]*v*(1-v)/mass,0)};
}

// Exact interior solution, identical/collinear families, and independent grid
// upper bounds over deterministic random examples guard the QP arithmetic.
assert(optimum([[0,0],[1,0],[0,1]],[.2,.3]).excess<1e-25);
assert(Math.abs(optimum([[.2,.7],[.2,.7],[.2,.7]],[.2,.7]).excess)<1e-25);
let seed=71337;const random=()=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return seed/4294967296;};
for(let trial=0;trial<100;trial++){
  const q=Array.from({length:3},()=>Array.from({length:5},random)),p=Array.from({length:5},random),answer=optimum(q,p);let grid=Infinity;
  for(let a=0;a<=30;a++)for(let b=0;b<=30-a;b++){const w=[a/30,b/30,1-(a+b)/30],v=p.reduce((z,t,i)=>z+(q.reduce((s,row,k)=>s+w[k]*row[i],0)-t)**2/5,0);grid=Math.min(grid,v);}
  assert(answer.excess<=grid+1e-12);
}
const interval=values=>{const mean=values.reduce((z,v)=>z+v,0)/values.length,se=Math.sqrt(values.reduce((z,v)=>z+(v-mean)**2,0)/(values.length-1)/values.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};};
const output={study:'blend-v30-consumed-headroom',oracleOnly:true,newValidation:false,splits:{}};
for(const split of ['design','confirmation']){
  const raw=fs.readFileSync(`docs/experiments/mmm-blend-v30-${split}.jsonl`),rows=raw.toString().trim().split('\n').map(JSON.parse);rows.shift();const groups={};
  for(const w of rows){
    const get=m=>w.Arms.find(a=>a.Model===m&&a.Policy==='stratified_random'),a=get('blend'),children=['local','old','partition'].map(get),q=children.map(c=>c.Forecast);
    for(let i=0;i<150;i++){const expected=q.reduce((z,row,k)=>z+a.Weights[k]*row[i],0);assert(Math.abs(a.Forecast[i]-expected)<1e-12,'actual law not child-predictive mixture');}
    const oracle=optimum(q,w.Rates),priorityOracle=optimum(q,w.Rates,Array.from({length:150},(_,i)=>i<10?3:1));
    assert(oracle.brier<=Math.min(...children.map(c=>c.Brier))+1e-12);assert(oracle.brier<=a.Brier+1e-12);
    (groups[`${w.Geometry}/${w.Regime}`]??=[]).push({actual:a.Brier,oracle:oracle.brier,component:Math.min(...children.map(c=>c.Brier)),gap:a.Brier-oracle.brier,weights:a.Weights,oracleWeights:oracle.weights,priorityGap:a.PriorityBrier-priorityOracle.brier});
  }
  output.splits[split]={sha256:hash(raw),groups:Object.fromEntries(Object.entries(groups).map(([name,rows])=>[name,{actual:interval(rows.map(r=>r.actual)),oracle:interval(rows.map(r=>r.oracle)),bestComponent:interval(rows.map(r=>r.component)),gap:interval(rows.map(r=>r.gap)),priorityGap:interval(rows.map(r=>r.priorityGap)),meanWeights:[0,1,2].map(k=>rows.reduce((z,r)=>z+r.weights[k]/rows.length,0)),meanOracleWeights:[0,1,2].map(k=>rows.reduce((z,r)=>z+r.oracleWeights[k]/rows.length,0))}]))};
}
fs.writeFileSync('docs/experiments/mmm-blend-v30-headroom.json',JSON.stringify(output,null,2)+'\n');
for(const name of ['tight/curved','wide/curved','wide/independent','wide/reversed'])console.log(name,JSON.stringify(output.splits.confirmation.groups[name]));
