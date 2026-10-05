import assert from 'node:assert/strict';

const sigmoid=x=>x>=0?1/(1+Math.exp(-x)):Math.exp(x)/(1+Math.exp(x));
export const excess=(p,b,y)=>(p-b)*(p+b-2*Number(y));
const worst=(p,b)=>Math.max(excess(p,b,0),excess(p,b,1));

// Fixed-rate expert feedback and worst-case pending reserves share issue-time
// records, but the expert learner always scores the original expert forecasts.
export function runBudget(rows, proposedWeights) {
  if(proposedWeights!==undefined){assert.equal(proposedWeights.length,rows.length);assert(proposedWeights.every(w=>Number.isFinite(w)&&w>=0&&w<=1));}
  const out=[];
  for(let i=0;i<rows.length;i++) {
    let ledger=0,odds=0;
    for(let j=0;j<i;j++) {
      const r=rows[j], arrived=!r.missing&&j+r.delay<=i;
      ledger+=arrived?excess(out[j].p,r.b,r.y):worst(out[j].p,r.b);
      if(arrived) odds+=(r.b-r.c)*(r.b+r.c-2*Number(r.y));
    }
    const r=rows[i], proposed=proposedWeights===undefined?sigmoid(odds):proposedWeights[i], d=r.c-r.b, a=d*d;
    const k=Math.max(2*d*r.b,2*d*(r.b-1)), allowance=.01*(i+1)-ledger;
    assert(allowance>=-1e-12);
    let w=proposed;
    if(a>0 && a*w*w+k*w>allowance) {
      const room=Math.max(0,allowance);
      w=room===0?0:Math.min(w,2*room/(k+Math.sqrt(k*k+4*a*room)));
    }
    const p=(1-w)*r.b+w*r.c;
    const reserved=ledger+worst(p,r.b);
    assert(reserved<=.01*(i+1)+1e-12);
    out.push({p,w,proposed,reserved,ledger,clipped:w<proposed});
  }
  return out;
}
