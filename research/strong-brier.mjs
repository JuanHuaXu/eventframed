import assert from 'node:assert/strict';

// Binary specialization of Vovk/Zhdanov Algorithm 1. Their summed two-class
// Brier loss is 2*(p-y)^2, so their eta=1 means eta=2 on our scalar loss.
export function strongBrier(b,c,w) {
  assert([b,c,w].every(x=>Number.isFinite(x)&&x>=0&&x<=1));
  const z0=(1-w)*Math.exp(-2*b*b)+w*Math.exp(-2*c*c);
  const z1=(1-w)*Math.exp(-2*(1-b)**2)+w*Math.exp(-2*(1-c)**2);
  return Math.max(0,Math.min(1,.5+(Math.log(z1)-Math.log(z0))/4));
}

// Reference refiltering, not a production constant-time implementation.
// Missing/current/future labels are excluded; each eligible factor enters once
// per reconstructed path. Sharing/delay do not inherit the static regret bound.
export function brierWeights(rows,eta=2,sharing=true) {
  assert(Number.isFinite(eta)&&eta>0);
  return rows.map((_,t)=>{
    let w=.5;
    for(let j=0;j<=t;j++) {
      const a=sharing&&j>0?1/(j+1):0;w=(1-a)*w+a*.5;
      const r=rows[j];
      if(j<t&&!r.missing&&j+r.delay<=t) {
        const u=w*Math.exp(-eta*(r.c-Number(r.y))**2);
        const v=(1-w)*Math.exp(-eta*(r.b-Number(r.y))**2);
        w=u/(u+v);
      }
    }
    return w;
  });
}
