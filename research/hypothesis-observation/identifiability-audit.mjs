import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';

// Exact information ceiling for the existing copied-report simulator. Knowing
// every first report dominates any adaptive policy that can only repeat them.
// This is an oracle diagnostic, not an implementable certificate of true mode.
function table(noise){
  const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];
  return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{
    const b=Array.from({length:4},(_,i)=>(h>>i)&1),values=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]];
    return values[t]?1-errors[t]:errors[t];
  }));
}
function audit(likelihood){
  const cache=new Map();let states=0;
  function state(code){
    if(cache.has(code))return cache.get(code);
    let c=code;const digits=Array.from({length:8},()=>{const x=c%3;c=Math.floor(c/3);return x;});
    const joint=Array.from({length:16},(_,h)=>digits.reduce((p,d,t)=>p*(d===0?1:d===1?1-likelihood[t][h]:likelihood[t][h]),1/16));
    const mass=joint.reduce((s,x)=>s+x,0);assert(mass>0);
    const classes=Array(4).fill(0);joint.forEach((w,h)=>classes[h%4]+=w/mass);
    const risk=1-classes.reduce((s,p)=>s+p*p,0);
    const out={digits,joint,mass,classes,risk,values:new Map()};cache.set(code,out);states++;return out;
  }
  function value(code,budget){
    const s=state(code);if(budget===0||s.digits.every(x=>x!==0))return s.risk;
    if(s.values.has(budget))return s.values.get(budget);
    let best=Infinity;
    for(let t=0;t<8;t++)if(s.digits[t]===0){
      const left=code+3**t,right=code+2*3**t,l=state(left),r=state(right);
      assert(Math.abs(l.mass+r.mass-s.mass)<1e-13);
      const risk=l.mass/s.mass*value(left,budget-1)+r.mass/s.mass*value(right,budget-1);
      best=Math.min(best,risk);
    }
    s.values.set(budget,best);return best;
  }
  const curve=Array.from({length:9},(_,b)=>value(0,b));
  for(let b=1;b<9;b++)assert(curve[b]<=curve[b-1]+1e-12);
  // Independent complete-vector enumeration verifies the dynamic program's
  // full-information endpoint, including normalization and classification error.
  let mass=0,brier=0,error=0;
  for(let bits=0;bits<256;bits++){
    const joint=Array.from({length:16},(_,h)=>Array.from({length:8},(_,t)=>((bits>>t)&1)?likelihood[t][h]:1-likelihood[t][h]).reduce((p,q)=>p*q,1/16));
    const p=joint.reduce((s,x)=>s+x,0),cls=Array(4).fill(0);joint.forEach((w,h)=>cls[h%4]+=w/p);
    mass+=p;brier+=p*(1-cls.reduce((s,q)=>s+q*q,0));error+=p*(1-Math.max(...cls));
  }
  assert(Math.abs(mass-1)<1e-12);assert(Math.abs(curve[8]-brier)<1e-12);
  // Distribution signatures concern structural identifiability, not whether
  // one finite report vector uniquely recovers the latent state.
  const signatures=new Set(Array.from({length:16},(_,h)=>JSON.stringify(likelihood.map(row=>row[h]))));
  return {states,distinctHypothesisLaws:signatures.size,optimalKnownModeBrierByUniqueReports:curve,fullReportBayesBrier:brier,fullReportBayesError:error,mass};
}
const results={copied05:audit(table(.05)),copied20:audit(table(.2)),null:audit(Array.from({length:8},()=>Array(16).fill(.5)))};
assert(results.copied05.distinctHypothesisLaws===16&&results.copied20.distinctHypothesisLaws===16);
assert(results.null.distinctHypothesisLaws===1);assert(Math.abs(results.null.fullReportBayesBrier-.75)<1e-12);
const paths=['research/hypothesis-observation/experiment.py','research/hypothesis-observation/dependence_v3.py','research/hypothesis-observation/identifiability-audit.mjs'];
console.log(JSON.stringify({scope:'Exact known-copy-mode information ceiling; not fitted or confirmation data',results,hashes:Object.fromEntries(paths.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')]))},null,2));
