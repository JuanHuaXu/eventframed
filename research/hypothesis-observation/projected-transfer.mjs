import assert from 'node:assert/strict';
import {guardedTransfer} from './robust-transfer.mjs';
const dot=(a,b)=>a.reduce((s,x,i)=>s+x*b[i],0);
const square=(a,b)=>a.reduce((s,x,i)=>s+(x-b[i])**2,0);
function activeDual(x,proposal,laws,radii){
  const n=x.length,normals=[],owners=[];
  laws.forEach((q,i)=>{if(Math.abs(square(x,q)-radii[i]**2)<1e-8){normals.push(x.map((v,j)=>v-q[j]));owners.push(i);}});
  x.forEach((v,i)=>{if(v<1e-10){normals.push(x.map((_,j)=>1/n-Number(i===j)));owners.push(laws.length);}});
  const coefficients=normals.map(()=>0),residual=proposal.map((v,i)=>v-x[i]);
  // Nonnegative coordinate descent only proposes dual multipliers. The
  // resulting dual bound is checked independently; convergence is not assumed.
  for(let sweep=0;sweep<200;sweep++){
    for(let i=0;i<normals.length;i++){
      const normal=normals[i],norm=dot(normal,normal);if(norm===0)continue;
      const next=Math.max(0,coefficients[i]+dot(normal,residual)/norm),delta=next-coefficients[i];
      residual.forEach((_,j)=>residual[j]-=delta*normal[j]);coefficients[i]=next;
    }
    if(dot(residual,residual)<1e-24)break;
  }
  const ys=Array.from({length:laws.length+1},()=>Array(n).fill(0));
  normals.forEach((v,i)=>v.forEach((x,j)=>ys[owners[i]][j]+=coefficients[i]*x));
  const sum=Array(n).fill(0);let support=0;
  ys.forEach((y,i)=>{y.forEach((v,j)=>sum[j]+=v);support+=i<laws.length?dot(laws[i],y)+radii[i]*Math.sqrt(dot(y,y)):Math.max(...y);});
  return dot(proposal,sum)-dot(sum,sum)/2-support;
}
export function simplexProjection(v){
  const sorted=v.slice().sort((a,b)=>b-a);let sum=0,theta=0;
  for(let j=0;j<sorted.length;j++){sum+=sorted[j];const candidate=(sum-1)/(j+1);if(sorted[j]>candidate)theta=candidate;}
  return v.map(x=>Math.max(0,x-theta));
}

// Dykstra corrections retain the original projection objective. Ordinary
// alternating projections need not find the closest feasible forecast.
export function projectedTransfer(base,proposal,laws,epsilon=.01){
  const line=guardedTransfer(base,proposal,laws,epsilon);assert(epsilon>0);
  const n=base.length,radii=laws.map(q=>Math.sqrt(square(base,q)+epsilon));
  const corrections=Array.from({length:laws.length+1},()=>Array(n).fill(0));
  let x=proposal.slice(),lastGap=Infinity,lastViolation=Infinity;
  for(let cycle=0;cycle<=10000;cycle++){
    const violation=Math.max(...laws.map((q,i)=>square(x,q)-radii[i]**2));
    lastViolation=violation;
    // Fenchel dual lower bound using support functions of the balls/simplex.
    const sum=Array(n).fill(0);let support=0;
    corrections.forEach((y,i)=>{y.forEach((v,j)=>sum[j]+=v);support+=i<laws.length?dot(laws[i],y)+radii[i]*Math.sqrt(dot(y,y)):Math.max(...y);});
    let dual=dot(proposal,sum)-dot(sum,sum)/2-support;
    if(violation<=1e-10){
      // Repair roundoff toward the known feasible baseline, then evaluate the
      // objective gap of the REPAIRED point, not the unrepaired iterate.
      const repaired=guardedTransfer(base,x,laws,epsilon).forecast;
      if(cycle%10===0)dual=Math.max(dual,activeDual(repaired,proposal,laws,radii));
      const gap=square(repaired,proposal)/2-dual;
      lastGap=gap;
      assert(gap>=-1e-10);
      if(gap<=1e-12){
        assert(square(repaired,proposal)<=square(line.forecast,proposal)+1e-10);
        return {forecast:repaired,cycles:cycle,gap,violation,repairDistance:Math.sqrt(square(repaired,x))};
      }
    }
    if(cycle===10000)break;
    for(let i=0;i<corrections.length;i++){
      const y=x.map((v,j)=>v+corrections[i][j]);let next;
      if(i===laws.length)next=simplexProjection(y);
      else {const q=laws[i],norm=Math.sqrt(square(y,q)),scale=norm>radii[i]?radii[i]/norm:1;next=y.map((v,j)=>q[j]+scale*(v-q[j]));}
      corrections[i]=y.map((v,j)=>v-next[j]);x=next;
    }
  }
  throw new Error('Projection did not meet tolerances: '+JSON.stringify({base,proposal,laws,x,lastGap,lastViolation}));
}
