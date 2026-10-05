import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {prepare} from './audit-joint-policy.mjs';
import {actualJoint} from './exact-regime.mjs';
import {auditLikelihood} from './noisy-audit.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'audit-joint-game.json'),d=JSON.parse(raw),modelRaw=readFileSync(dir+'noisy-audit-evaluation.json'),model=JSON.parse(modelRaw);
assert.equal(createHash('sha256').update(modelRaw).digest('hex'),d.inputSHA256);
for(const [f,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(dir+f)).digest('hex'),h);
const worlds=model.results.map(r=>({noise:r.noise,mask:r.mask})),graphs=prepare(model,worlds);
let graphChecks=0;
for(const graph of graphs){
  assert.equal(graph.states[0].audit,null);assert.equal(graph.level[0],0);
  for(const pair of graph.children[0])assert.deepEqual(pair,graph.children[0][0]);
  for(let i=0;i<graph.level.length;i++){
    assert.equal(graph.level[i],i===0?0:1+graph.states[i].counts.reduce((a,n)=>a+n,0));
    for(const pair of graph.children[i]){
      for(const j of pair)assert.equal(graph.level[j],graph.level[i]+1);
      for(let g=0;g<worlds.length;g++)for(let y=0;y<4;y++){
        const child=pair.reduce((v,j)=>v+graph.joints[(j*80+g)*4+y],0),parent=graph.joints[(i*80+g)*4+y];
        assert(Math.abs(child-parent)<1e-12);graphChecks++;
      }
    }
  }
}
const witness=d.rounds.reduce((a,b)=>a.lower>b.lower?a:b),fw=Array(80).fill(0),aw=Array(80).fill(0);let constant=0;
assert(Math.abs(witness.weights.reduce((a,v)=>a+v,0)-1)<1e-12);
d.rows.forEach((r,i)=>{const q=witness.weights[i];assert(q>=0);(r.metric==='final'?fw:aw)[r.g]+=q;constant+=q*r.threshold;
  assert.equal(r.threshold,model.results[r.g].scores[r.control][r.metric==='final'?'finalBrier':'areaBrier']+(r.metric==='final'?.01:0));
});
let objective=0,statesChecked=0;
for(let pattern=0;pattern<16;pattern++){
  const cost=(counts,audit,terminal)=>{
    const weights=terminal?fw:aw,joint=[0,0,0,0];
    worlds.forEach((world,g)=>{
      const factor=audit===null?1:auditLikelihood(world.mask,audit),row=actualJoint(pattern,counts,world.noise,world.mask);
      row.forEach((v,y)=>joint[y]+=weights[g]*factor*v);
    });
    const m=joint.reduce((a,v)=>a+v,0);return (m?m-joint.reduce((a,v)=>a+v*v,0)/m:0)/(terminal?1:6);
  };
  const zero=Array(8).fill(0);
  // A single forecast is scored before branching on the future audit outcome.
  objective+=cost(zero,null,false);statesChecked++;
  for(const audit of [0,1]){
    const memo=new Map();
    function solve(counts){
      const key=counts.join(',');if(memo.has(key))return memo.get(key);
      const terminal=counts.reduce((a,v)=>a+v,0)===5;let value=cost(counts,audit,terminal);
      if(!terminal){let best=Infinity;for(let a=0;a<4;a++){
        let sum=0;for(let y=0;y<2;y++){const c=counts.slice();c[2*a+y]++;sum+=solve(c);}best=Math.min(best,sum);
      }value+=best;}
      memo.set(key,value);statesChecked++;return value;
    }
    objective+=solve(zero);
  }
}
const lower=objective-constant;assert(Math.abs(lower-witness.lower)<1e-11);
let mixtureChecks=0;
d.rows.forEach((r,i)=>{const mean=d.rounds.reduce((v,t)=>v+(t[r.metric][r.g]-r.threshold)/d.rounds.length,0);assert(Math.abs(mean-d.violations[i].meanPayoff)<1e-12);mixtureChecks++;});
assert(raw.equals(execFileSync(process.execPath,[dir+'audit-joint-game.mjs'],{maxBuffer:64*1024*1024})));
const out={scope:'Forced audit graph, probability/cost layers, separate pre-audit recursive oracle, mixture reconstruction and full replay; no outward rounding',graphChecks,statesChecked,witnessRound:witness.round,independentLower:lower,reportedLower:witness.lower,mixtureChecks,replay:true,inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-audit-joint.mjs')).digest('hex')};
writeFileSync(dir+'audit-joint-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
