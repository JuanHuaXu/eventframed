import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {actualJoint} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'global-policy-game.json'),d=JSON.parse(raw);
for(const[f,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(dir+f)).digest('hex'),h);
assert(raw.equals(execFileSync(process.execPath,[dir+'global-policy-game.mjs'],{maxBuffer:64*1024*1024})));
execFileSync(process.execPath,[dir+'test-global-policy.mjs']);
const original=JSON.parse(readFileSync(dir+'risk-budget-evaluation.json')),model=JSON.parse(readFileSync(dir+'count-planning-exact.json'));
const witness=d.rounds.reduce((a,b)=>a.lower>b.lower?a:b),q=witness.weights;
assert(q.every(v=>Number.isFinite(v)&&v>=0));assert(Math.abs(q.reduce((a,v)=>a+v,0)-1)<1e-12);
const fw=Array(80).fill(0),aw=Array(80).fill(0);let constant=0;
d.rows.forEach((r,i)=>{(r.metric==='final'?fw:aw)[r.g]+=q[i];constant+=q[i]*r.threshold;
  const expected=original.results[r.g].scores[r.control][r.metric==='final'?'finalBrier':'areaBrier']+(r.metric==='final'?.01:0);
  assert.equal(expected,r.threshold);
});
let optimum=0,statesChecked=0;
for(const root of model.roots){
  const states=new Map(root.states.map(s=>[s.counts.join(','),s])),memo=new Map();
  function recurse(key){
    if(memo.has(key))return memo.get(key);
    const s=states.get(key),terminal=s.counts.reduce((a,n)=>a+n,0)===6,weights=terminal?fw:aw;
    let current=0;
    for(let g=0;g<80;g++){
      const world=original.results[g],joint=actualJoint(root.pattern,s.counts,world.noise,world.mask),mass=joint.reduce((a,v)=>a+v,0);
      const loss=mass*(1+s.forecast.reduce((a,p)=>a+p*p,0))-2*joint.reduce((a,v,y)=>a+v*s.forecast[y],0);
      current+=weights[g]*loss/(terminal?1:6);
    }
    if(!terminal){
      let best=Infinity;
      for(let a=0;a<4;a++){
        let sum=0;for(let y=0;y<2;y++){const child=s.counts.slice();child[2*a+y]++;sum+=recurse(child.join(','));}
        best=Math.min(best,sum);
      }
      current+=best;
    }
    statesChecked++;memo.set(key,current);return current;
  }
  optimum+=recurse('0,0,0,0,0,0,0,0');
}
const lower=optimum-constant;assert(Math.abs(lower-witness.lower)<1e-11);
let mixtureChecks=0,maxMixtureError=0;
for(let i=0;i<d.rows.length;i++){
  const r=d.rows[i],payoff=d.rounds.reduce((a,t)=>a+(t[r.metric][r.g]-r.threshold)/d.rounds.length,0),error=Math.abs(payoff-d.violations[i].meanPayoff);
  assert(error<1e-12);maxMixtureError=Math.max(maxMixtureError,error);mixtureChecks++;
}
const out={scope:'Replay, tiny exhaustive policies, independent recursive weighted oracle with alternate loss expression and mixture score reconstruction; floating-point witness, not outward-rounded proof',witnessRound:witness.round,independentLower:lower,reportedLower:witness.lower,statesChecked,mixtureChecks,maxMixtureError,replay:true,inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-global-policy.mjs')).digest('hex')};
writeFileSync(dir+'global-policy-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
