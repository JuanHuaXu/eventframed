import assert from 'node:assert/strict';

// Cache the unchanged prefix, then replay only from the earliest newly
// delivered origin. Late outcomes revise internal messages, never issued laws.
// Batch research implementation: O(T*K) cache, not a bounded daemon journal.
export function incrementalPoolWeights(rows,prior,sharing=true) {
  assert(prior.length>=2&&prior.every(w=>Number.isFinite(w)&&w>=0));
  assert(Math.abs(prior.reduce((s,w)=>s+w,0)-1)<1e-12);
  const cache=[],issued=[],arrivals=Array.from({length:rows.length},()=>[]),delivered=[];
  let transitions=0;
  for(let t=0;t<rows.length;t++) {
    let earliest=t;
    for(const origin of arrivals[t]) {delivered[origin]=true;earliest=Math.min(earliest,origin);}
    let weights=earliest===0?prior.slice():cache[earliest-1].slice();
    for(let j=earliest;j<=t;j++) {
      const a=sharing&&j>0?1/(j+1):0;
      weights=weights.map((w,k)=>(1-a)*w+a*prior[k]);transitions++;
      if(delivered[j]) {
        const r=rows[j];assert(typeof r.y==='boolean'||r.y===0||r.y===1);
        weights=weights.map((w,k)=>w*Math.exp(-2*(r.p[k]-Number(r.y))**2));
        const sum=weights.reduce((s,w)=>s+w,0);assert(sum>0);weights=weights.map(w=>w/sum);
      }
      cache[j]=weights;
    }
    issued.push(weights.slice());
    const r=rows[t];
    assert(r.p.length===prior.length&&r.p.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
    assert(Number.isInteger(r.delay)&&r.delay>=0);
    const ready=Math.max(t+1,t+r.delay);
    if(!r.missing&&ready<rows.length)arrivals[ready].push(t);
  }
  return {weights:issued,transitions};
}
