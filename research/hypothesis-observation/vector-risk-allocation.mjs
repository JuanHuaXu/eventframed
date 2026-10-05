import assert from 'node:assert/strict';
export function vectorRisk(row,p,width) {
  let risk=row.constant;
  for(let i=0;i<p.length;i++)risk+=row.mass[Math.floor(i/width)]*p[i]*p[i]-2*row.joint[i]*p[i];
  return risk;
}

// The separable Lagrangian minimizer is a normalized nonnegative weighted
// joint law, so it already lies in each history's probability simplex.
export function allocateVectorRisk(objective,rows,base,epsilon=.01) {
  const n=objective.mass.length,width=base.length/n;
  assert(n>0&&Number.isInteger(width)&&width>=2&&rows.length>0);
  assert(Number.isFinite(epsilon)&&epsilon>0);
  for(let x=0;x<n;x++){
    assert(base.slice(x*width,(x+1)*width).every(p=>Number.isFinite(p)&&p>=0&&p<=1));
    assert(Math.abs(base.slice(x*width,(x+1)*width).reduce((s,p)=>s+p,0)-1)<1e-12);
  }
  for(const row of [objective,...rows]){
    assert(row.mass.length===n&&row.joint.length===base.length&&Number.isFinite(row.constant));
    assert(row.mass.every(m=>Number.isFinite(m)&&m>=0)&&row.joint.every(v=>Number.isFinite(v)&&v>=0));
    for(let x=0;x<n;x++)assert(Math.abs(row.joint.slice(x*width,(x+1)*width).reduce((s,v)=>s+v,0)-row.mass[x])<1e-12);
    assert(Math.abs(vectorRisk(row,base,width))<1e-10);
  }
  assert(objective.mass.every(m=>m>0));
  const mu=rows.map(()=>0);let last;
  for(let sweep=0;sweep<=5000;sweep++){
    let mass=objective.mass.slice(),joint=objective.joint.slice(),constant=objective.constant;
    rows.forEach((r,j)=>{constant+=mu[j]*(r.constant-epsilon);for(let x=0;x<n;x++)mass[x]+=mu[j]*r.mass[x];for(let i=0;i<joint.length;i++)joint[i]+=mu[j]*r.joint[i];});
    const p=joint.map((v,i)=>v/mass[Math.floor(i/width)]);
    const dual=vectorRisk({mass,joint,constant},p,width);
    let scale=1;
    for(const r of rows){
      let A=0,B=0;
      for(let i=0;i<p.length;i++){const d=p[i]-base[i],m=r.mass[Math.floor(i/width)];A+=m*d*d;B+=2*d*(m*base[i]-r.joint[i]);}
      if(A+B>epsilon){
        const root=A>0?(B>=0?2*epsilon/(B+Math.sqrt(B*B+4*A*epsilon)):(Math.sqrt(B*B+4*A*epsilon)-B)/(2*A)):epsilon/B;
        scale=Math.min(scale,root*(1-1e-14));
      }
    }
    const forecast=p.map((v,i)=>base[i]+scale*(v-base[i]));
    const primal=vectorRisk(objective,forecast,width),gap=primal-dual;
    const maxRisk=Math.max(...rows.map(r=>vectorRisk(r,forecast,width)));
    assert(gap>=-1e-9);last={sweep,primal,dual,gap,maxRisk,scale};
    if(gap<=1e-8&&maxRisk<=epsilon+1e-12){
      for(let x=0;x<n;x++)assert(Math.abs(forecast.slice(x*width,(x+1)*width).reduce((s,v)=>s+v,0)-1)<1e-12);
      assert(forecast.every(p=>p>=0&&p<=1));return {forecast,...last};
    }
    if(sweep===5000)break;
    for(let j=0;j<rows.length;j++){
      const r=rows[j],bm=mass.map((v,x)=>v-mu[j]*r.mass[x]),bj=joint.map((v,i)=>v-mu[j]*r.joint[i]);
      const riskAt=m=>{
        let risk=r.constant;
        for(let x=0;x<n;x++){
          const denominator=bm[x]+m*r.mass[x];
          for(let y=0;y<width;y++){const i=x*width+y,v=(bj[i]+m*r.joint[i])/denominator;risk+=r.mass[x]*v*v-2*r.joint[i]*v;}
        }
        return risk;
      };
      let next=0;
      if(riskAt(0)>epsilon){
        let lo=0,hi=Math.max(1,mu[j]);
        while(riskAt(hi)>epsilon){hi*=2;assert(hi<1e12,'Vector dual failed to bracket');}
        for(let k=0;k<45;k++){const mid=(lo+hi)/2;if(riskAt(mid)>epsilon)lo=mid;else hi=mid;}
        next=hi;
      }
      mu[j]=next;
      for(let x=0;x<n;x++)mass[x]=bm[x]+next*r.mass[x];
      for(let i=0;i<joint.length;i++)joint[i]=bj[i]+next*r.joint[i];
    }
  }
  throw new Error('Full-vector allocation did not meet tolerances: '+JSON.stringify(last));
}

