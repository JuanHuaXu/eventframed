import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';

const source='research/hypothesis-observation/identifiability-audit.json';
const oracle=JSON.parse(readFileSync(source)).results;
function table(noise,isNull){
  const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];
  return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{
    const b=Array.from({length:4},(_,i)=>(h>>i)&1),values=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]],e=isNull?.5:errors[t];return values[t]?1-e:e;
  }));
}
function evaluate(likelihood,reference){
  const states=new Map(),planCache=new Map(),riskCache=new Map();let branches=0,giniChecks=0;
  function state(code){
    if(states.has(code))return states.get(code);
    let c=code;const digits=Array.from({length:8},()=>{const d=c%3;c=Math.floor(c/3);return d;});
    const w=Array.from({length:16},(_,h)=>digits.reduce((p,d,t)=>p*(d===0?1:d===1?1-likelihood[t][h]:likelihood[t][h]),1/16));
    const mass=w.reduce((s,x)=>s+x,0),cls=Array(4).fill(0);w.forEach((x,h)=>cls[h%4]+=x/mass);
    const out={digits,mass,risk:1-cls.reduce((s,x)=>s+x*x,0),available:digits.flatMap((d,t)=>d===0?[t]:[])};states.set(code,out);return out;
  }
  function children(code,t){const l=code+3**t,r=code+2*3**t,s=state(code),p=state(r).mass/s.mass;assert(Math.abs(state(l).mass/s.mass+p-1)<1e-12);return {l,r,p};}
  function plan(code,depth){
    const s=state(code);if(depth===0||!s.available.length)return {risk:s.risk,action:null};
    const key=`${code}:${depth}`;if(planCache.has(key))return planCache.get(key);
    let best={risk:Infinity,action:null};
    for(const t of s.available){const {l,r,p}=children(code,t);branches++;const v=(1-p)*plan(l,depth-1).risk+p*plan(r,depth-1).risk;if(v<best.risk-1e-14)best={risk:v,action:t};}
    planCache.set(key,best);return best;
  }
  function action(code,budget,policy){
    const s=state(code);if(policy==='random')return s.available.map(t=>[t,1/s.available.length]);
    if(policy==='entropy'){
      let chosen=null,best=-Infinity;
      for(const t of s.available){const {p}=children(code,t),h=-p*Math.log(p)-(1-p)*Math.log1p(-p);if(h>best+1e-14){chosen=t;best=h;}}
      return [[chosen,1]];
    }
    const depth=policy==='optimal'?budget:Math.min(budget,Number(policy));
    const chosen=plan(code,depth).action;
    if(policy==='1'){
      let direct=null,best=-Infinity;
      for(const t of s.available){const {l,r,p}=children(code,t),gain=s.risk-(1-p)*state(l).risk-p*state(r).risk;if(gain>best+1e-14){best=gain;direct=t;}}
      assert.equal(chosen,direct);giniChecks++;
    }
    return [[chosen,1]];
  }
  function risk(code,budget,policy){
    const s=state(code);if(budget===0||!s.available.length)return s.risk;
    const key=`${code}:${budget}:${policy}`;if(riskCache.has(key))return riskCache.get(key);
    let v=0;for(const [t,weight]of action(code,budget,policy)){const {l,r,p}=children(code,t);v+=weight*((1-p)*risk(l,budget-1,policy)+p*risk(r,budget-1,policy));}
    riskCache.set(key,v);return v;
  }
  const curves=Object.fromEntries(['random','entropy','1','2','3','optimal'].map(policy=>[policy,Array.from({length:9},(_,b)=>risk(0,b,policy))]));
  for(let b=0;b<=8;b++){
    assert(Math.abs(curves.optimal[b]-reference.optimalKnownModeBrierByUniqueReports[b])<1e-12);
    for(const curve of Object.values(curves))assert(curve[b]>=curves.optimal[b]-1e-12);
  }
  for(const curve of Object.values(curves))assert(Math.abs(curve[8]-reference.fullReportBayesBrier)<1e-12);
  return {curves,states:states.size,planningSubproblems:planCache.size,planningBranches:branches,evaluatedPolicyStates:riskCache.size,giniChecks};
}
const results={};for(const [name,noise,isNull]of[['copied05',.05,false],['copied20',.2,false],['null',.5,true]])results[name]=evaluate(table(noise,isNull),oracle[name]);
for(const curve of Object.values(results.null.curves))for(const x of curve)assert(Math.abs(x-.75)<1e-12);
const paths=[source,'research/hypothesis-observation/planning-ceiling.mjs','research/hypothesis-observation/PLANNING_CEILING_PROTOCOL.md'];
console.log(JSON.stringify({scope:'Exact known-copy-mode planning comparison; not unknown-mode validation',results,hashes:Object.fromEntries(paths.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')]))},null,2));
