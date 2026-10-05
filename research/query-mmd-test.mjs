import assert from 'node:assert/strict';
import {mmdState,mmdDesignView} from './query-mmd.mjs';
const k=(a,b)=>Math.exp(-[...((a^b).toString(2))].filter(v=>v==='1').length/2);
const direct=(h,t)=>h.reduce((s,x)=>s+h.reduce((s,y)=>s+k(x,y),0),0)/h.length**2+t.reduce((s,x)=>s+t.reduce((s,y)=>s+k(x,y),0),0)/t.length**2-2*h.reduce((s,x)=>s+t.reduce((s,y)=>s+k(x,y),0),0)/(h.length*t.length);
const near=(a,b)=>assert(Math.abs(a-b)<1e-11);let checks=0;
for(let seed=0;seed<40;seed++){
  const h=Array.from({length:1+seed%10},(_,i)=>(i*i+seed*19)%512),t=Array.from({length:1+seed%17},(_,i)=>(i*37+seed)%512),state=mmdState(h,t);
  near(state.baseline,direct(h,t));
  for(const x of [0,1,63,255,511]){near(state.append(x).after,direct([...h,x],t));checks++;}
  near(mmdState(h.toReversed(),t.toReversed()).baseline,state.baseline);
  const perm=x=>parseInt(x.toString(2).padStart(9,'0').split('').reverse().join(''),2);near(mmdState(h.map(perm),t.map(perm)).baseline,state.baseline);
  near(mmdState([...h,...h],t).baseline,state.baseline);
  const before=state.append(1);h.fill(511);t.fill(0);assert.deepEqual(state.append(1),before);
}
assert.equal(mmdState([0,1],[0,1]).append(511).factor,1);
const h=Array(63).fill(0),t=[...Array(160).fill(0),1];assert(mmdState(h,t).append(511).factor<0);
for(const bad of [[],[-1],[512],[NaN]])assert.throws(()=>mmdState(bad,[0]));
const source={Initial:Array.from({length:16},()=>({Bits:1})),Steps:Array.from({length:256},(_,i)=>({X:i,Missing:i===159,Delay:0}))};
const original=mmdDesignView([-1,1,2],[159],source);
for(const s of [...source.Initial,...source.Steps])for(const key of ['Q','Y','teacher','Outcome'])Object.defineProperty(s,key,{get(){throw Error(key);}});
for(let i=161;i<256;i++)source.Steps[i]=new Proxy({},{get(){throw Error('future');}});
assert.deepEqual(mmdDesignView([-1,1,2],[159],source),original);
assert.throws(()=>mmdDesignView([159],[158],source));assert.throws(()=>mmdDesignView([-1],[158],source));
console.log(`PASS: ${checks} independent append-MMD comparisons, direct baseline, duplicates, permutation/bit symmetry, ownership, zero denominator, negative factor and invalid inputs`);
