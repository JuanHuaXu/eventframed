import assert from 'node:assert/strict';
const minimizer=(a,b)=>a>0?Math.max(0,Math.min(1,-b/(2*a))):Number(b<0);
const evaluate=(a,b,x)=>x.reduce((s,v,i)=>s+a[i]*v*v+b[i]*v,0);

// Nonnegative dual multipliers give a lower bound even before convergence.
// Always test the repaired primal, not an infeasible intermediate minimizer.
export function allocateRisk(objective,rows,epsilon=.01) {
  const n=objective.a.length;
  assert(n>0&&objective.b.length===n&&rows.length>0&&Number.isFinite(epsilon)&&epsilon>0);
  for(const r of [objective,...rows]){
    assert(r.a.length===n&&r.b.length===n);
    assert(r.a.every(v=>Number.isFinite(v)&&v>=0)&&r.b.every(Number.isFinite));
  }
  const mu=rows.map(()=>0);
  let a=objective.a.slice(),b=objective.b.slice(),last=null;
  for(let sweep=0;sweep<=500;sweep++){
    // Rebuild to avoid accumulating subtract/add roundoff across sweeps.
    a=objective.a.slice();b=objective.b.slice();
    rows.forEach((r,j)=>{for(let i=0;i<n;i++){a[i]+=mu[j]*r.a[i];b[i]+=mu[j]*r.b[i];}});
    const x=a.map((v,i)=>minimizer(v,b[i]));
    const dual=evaluate(a,b,x)-epsilon*mu.reduce((s,v)=>s+v,0);
    let scale=1;
    for(const r of rows){
      let A=0,B=0;for(let i=0;i<n;i++){A+=r.a[i]*x[i]*x[i];B+=r.b[i]*x[i];}
      if(A+B>epsilon){
        const root=A>0?(B>=0?2*epsilon/(B+Math.sqrt(B*B+4*A*epsilon)):(Math.sqrt(B*B+4*A*epsilon)-B)/(2*A)):epsilon/B;
        scale=Math.min(scale,root*(1-1e-14));
      }
    }
    const repaired=x.map(v=>v*scale),primal=evaluate(objective.a,objective.b,repaired);
    const maxRisk=Math.max(...rows.map(r=>evaluate(r.a,r.b,repaired)));
    const gap=primal-dual;assert(gap>=-1e-9);
    last={sweep,gap,maxRisk,scale,primal,dual};
    if(gap<=1e-8&&maxRisk<=epsilon+1e-12)return {lambda:repaired,...last};
    if(sweep===500)break;
    for(let j=0;j<rows.length;j++){
      const r=rows[j],baseA=a.map((v,i)=>Math.max(0,v-mu[j]*r.a[i])),baseB=b.map((v,i)=>v-mu[j]*r.b[i]);
      const riskAt=m=>{let risk=0;for(let i=0;i<n;i++){const v=minimizer(baseA[i]+m*r.a[i],baseB[i]+m*r.b[i]);risk+=r.a[i]*v*v+r.b[i]*v;}return risk;};
      let next=0;
      if(riskAt(0)>epsilon){
        let lo=0,hi=Math.max(1,mu[j]);
        while(riskAt(hi)>epsilon){hi*=2;assert(hi<1e12,'Dual coordinate failed to bracket');}
        for(let k=0;k<45;k++){const mid=(lo+hi)/2;if(riskAt(mid)>epsilon)lo=mid;else hi=mid;}
        next=hi;
      }
      mu[j]=next;
      for(let i=0;i<n;i++){a[i]=baseA[i]+next*r.a[i];b[i]=baseB[i]+next*r.b[i];}
    }
  }
  throw new Error('Risk allocation did not meet tolerances: '+JSON.stringify(last));
}

