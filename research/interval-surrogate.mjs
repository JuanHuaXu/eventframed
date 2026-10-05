import assert from 'node:assert/strict';

const normalize=xs=>{const m=Math.max(...xs),e=xs.map(x=>Math.exp(x-m)),z=e.reduce((a,x)=>a+x,0);return e.map(x=>x/z);};
export function coveringIntervals(horizon){
  assert(Number.isInteger(horizon)&&horizon>=1&&horizon<=256);
  const out=[[0,horizon-1]];
  // Source-style one-based geometric intervals, plus one persistent anchor.
  // Timing is declared from the horizon, never from observed changes.
  for(let length=1;length<=horizon;length*=2)for(let start=length;start+length-1<=horizon;start+=length){
    const pair=[start-1,start+length-2];
    if(!out.some(x=>x[0]===pair[0]&&x[1]===pair[1]))out.push(pair);
  }
  return out;
}

export function createIntervalSurrogate(horizon,prior,intervals=coveringIntervals(horizon)){
  assert(Number.isInteger(horizon)&&horizon>=1&&horizon<=256);
  assert(Array.isArray(prior)&&prior.length>=2&&prior.length<=8&&prior.every(p=>Number.isFinite(p)&&p>0));
  assert(Math.abs(prior.reduce((a,p)=>a+p,0)-1)<1e-12);
  assert(Array.isArray(intervals)&&intervals.length>0&&intervals.length<=512);
  const pi=prior.slice(),rates=[.5,.25,.125,.0625],K=pi.length;
  const bounds=intervals.map(pair=>{assert(Array.isArray(pair)&&pair.length===2);const [a,b]=pair;assert(Number.isInteger(a)&&Number.isInteger(b)&&a>=0&&a<=b&&b<horizon);return [a,b];});
  assert.equal(new Set(bounds.map(x=>x.join('/'))).size,bounds.length);
  const activeAt=Array.from({length:horizon},(_,t)=>bounds.flatMap(([a,b],i)=>a<=t&&t<=b?[i]:[]));
  assert(activeAt.every(a=>a.length>0));
  const states=bounds.map(()=>({R:Array(K).fill(0),V:Array(K).fill(0),relativeG:0}));
  const journal=[];let next=0;
  function issue(probabilities){
    assert(next<horizon);assert(Array.isArray(probabilities)&&probabilities.length===K&&probabilities.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
    const ids=activeAt[next],meta=normalize(ids.map(id=>-states[id].relativeG));
    const active=ids.map((id,j)=>{
      const state=states[id],logs=[];
      for(const eta of rates)for(let k=0;k<K;k++)logs.push(Math.log(pi[k]/rates.length)+eta*state.R[k]-eta*eta*state.V[k]);
      return {id,q:meta[j],P:normalize(logs)};
    });
    const w=Array(K).fill(0);
    // Each P has its own partition function. Removing that normalization
    // changes the interval mixture whenever evidence differs across intervals.
    for(const a of active)for(let j=0;j<rates.length;j++)for(let k=0;k<K;k++)w[k]+=a.q*a.P[j*K+k]*rates[j];
    const z=w.reduce((s,x)=>s+x,0);for(let k=0;k<K;k++)w[k]/=z;
    const p=w.reduce((s,x,k)=>s+x*probabilities[k],0),id=next++;
    journal.push({active,w,probabilities:probabilities.slice(),status:'pending'});
    return {id,p,weights:w.slice(),active:active.length};
  }
  function deliver(id,outcome){
    assert(Number.isInteger(id)&&id>=0&&id<next&&typeof outcome==='boolean');
    const entry=journal[id];
    if(entry.status==='delivered'){assert.equal(entry.outcome,outcome,'conflicting outcome');return {duplicate:true};}
    assert.equal(entry.status,'pending','expired evidence');
    const losses=entry.probabilities.map(p=>(p-Number(outcome))**2),avg=entry.w.reduce((s,w,k)=>s+w*losses[k],0),r=losses.map(l=>avg-l);
    const factors=rates.flatMap(eta=>r.map(x=>Math.exp(eta*x-eta*eta*x*x)));
    const zs=entry.active.map(a=>a.P.reduce((s,p,j)=>s+p*factors[j],0));
    const ghat=-Math.log(entry.active.reduce((s,a,j)=>s+a.q*zs[j],0));
    const gs=zs.map(z=>-Math.log(z));
    assert(Number.isFinite(ghat)&&gs.every(Number.isFinite));
    // Common inactive loss cancels from future meta weights. Retain only the
    // active-minus-common increment, including for late origin intervals.
    for(let j=0;j<entry.active.length;j++){
      const state=states[entry.active[j].id];state.relativeG+=gs[j]-ghat;
      for(let k=0;k<K;k++){state.R[k]+=r[k];state.V[k]+=r[k]*r[k];}
    }
    entry.status='delivered';entry.outcome=outcome;
    // Once accepted, only a small duplicate/conflict receipt is needed.
    delete entry.active;delete entry.w;delete entry.probabilities;
    return {duplicate:false,ghat};
  }
  function expireBefore(cutoff){
    assert(Number.isInteger(cutoff)&&cutoff>=0&&cutoff<=next);
    for(let i=0;i<cutoff;i++)if(journal[i].status==='pending')journal[i]={status:'expired'};
  }
  return {issue,deliver,expireBefore,snapshot:()=>structuredClone({next,states,bounds,journal})};
}
