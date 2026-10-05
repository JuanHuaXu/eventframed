import assert from 'node:assert/strict';

function distribution(weights) {
  assert(weights.length>=2&&weights.every(w=>Number.isFinite(w)&&w>=0));
  assert(Math.abs(weights.reduce((s,w)=>s+w,0)-1)<1e-12);
}
export function strongPool(ps,weights) {
  distribution(weights);
  assert(ps.length===weights.length&&ps.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
  let z0=0,z1=0;
  for(let k=0;k<ps.length;k++) {z0+=weights[k]*Math.exp(-2*ps[k]**2);z1+=weights[k]*Math.exp(-2*(1-ps[k])**2);}
  return Math.max(0,Math.min(1,.5+(Math.log(z1)-Math.log(z0))/4));
}

// Same origin-order finite-horizon reference as the two-expert study. The
// prior is shared over hypotheses, not a count of independent evidence sources.
export function poolWeights(rows,prior,sharing=true) {
  distribution(prior);
  return rows.map((_,t)=>{
    let weights=prior.slice();
    for(let j=0;j<=t;j++) {
      const a=sharing&&j>0?1/(j+1):0;
      weights=weights.map((w,k)=>(1-a)*w+a*prior[k]);
      const r=rows[j];
      if(j<t&&!r.missing&&j+r.delay<=t) {
        assert(typeof r.y==='boolean'||r.y===0||r.y===1);
        assert(r.p.length===prior.length&&r.p.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
        weights=weights.map((w,k)=>w*Math.exp(-2*(r.p[k]-Number(r.y))**2));
        const sum=weights.reduce((s,w)=>s+w,0);assert(sum>0);
        weights=weights.map(w=>w/sum);
      }
    }
    distribution(weights);return weights;
  });
}
