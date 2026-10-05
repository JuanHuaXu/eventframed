import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {weightedRisk,minimumMassRisk} from './query-mass.mjs';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const [populationPath,decisionPath,sourcePath,summaryPath]=process.argv.slice(2),reference=JSON.parse(fs.readFileSync(summaryPath));
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
// Attach all iterators before awaiting any flowing source stream.
const streams=[populationPath,decisionPath,sourcePath].map(input),headers=[];
for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
for(const [i,v]of ['population-query-v1','regime-query-disjoint-v1','soft-learners-v120'].entries())assert.equal(headers[i].Version,v);
const near=(a,b)=>assert(Math.abs(a-b)<1e-12),mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length,records=[];let identities=0,bounds=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(s=>s.done)){assert(next.every(s=>s.done));break;}
  const [r,d,source]=next.map(s=>JSON.parse(s.value)),ref=reference.records[records.length];assert(ref&&!r.Error&&!d.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[upper],ref[lower]);assert.equal(r[upper],source[upper]);}
  auditRegimeOutcome(d.Original,source);auditRegimeOutcome(d.Candidate,source,137);
  assert.deepEqual(r.Branches.map(b=>b.Origin),[-1,...(d.Original.Decision.Pool??[])]);
  assert.deepEqual(r.Branches.map(b=>b.Origin),ref.branches.map(b=>b.origin));
  const branches=r.Branches.map((b,i)=>{
    let p=.5;
    if(b.Origin>=0){
      p=d.Original.Decision.Values.find(v=>v.Origin===b.Origin).Mass[1];near(p,d.Candidate.Decision.Values.find(v=>v.Origin===b.Origin).Mass[1]);
      assert.equal(b.Q,source.Steps[b.Origin].Q);assert.equal(b.ActualY,source.Steps[b.Origin].Y);
      assert.equal(b.Redundant,!source.Steps[b.Origin].Missing&&b.Origin+source.Steps[b.Origin].Delay<=161);
    }
    const truth=weightedRisk(b.Population,b.Q),modeled=weightedRisk(b.Population,p),neutral=weightedRisk(b.Population,.5),sample=weightedRisk(b.Sample,b.Q);
    near(truth,ref.branches[i].population);near(sample,ref.branches[i].sample);
    const distortion=(p-b.Q)*(b.Population[1]-b.Population[0]);near(modeled-truth,distortion);identities++;
    if(b.Origin<0||b.Redundant){assert.equal(b.Population[0],b.Population[1]);near(modeled,truth);near(neutral,truth);}
    return {origin:b.Origin,losses:b.Population,q:b.Q,p,truth,modeled,neutral,sample,distortion,redundant:b.Redundant};
  });
  const choose=(field,abstain)=>{
    const actions=abstain||branches.length===1?branches:branches.slice(1);
    return minimumMassRisk(actions.map(b=>({origin:b.origin,losses:b.losses,probability:field==='neutral'?.5:b[field]}))).origin;
  };
  const byOrigin=new Map(branches.map(b=>[b.origin,b]));
  const selected=[...d.Original.Decision.Selected,d.Candidate.Decision.Selected[3],choose('q',false),choose('p',false),choose('neutral',false),choose('q',true),choose('p',true),choose('neutral',true)];
  const losses=selected.map(j=>byOrigin.get(j).truth),sampleLosses=selected.map(j=>byOrigin.get(j).sample),costs=selected.map(j=>Number(j>=0));
  for(let i=0;i<5;i++)near(losses[i],ref.population[i]);near(losses[5],ref.population[6]);near(losses[8],ref.population[7]);
  const regret=[];
  for(const [oracle,model]of [[5,6],[8,9]]){
    const a=byOrigin.get(selected[model]),b=byOrigin.get(selected[oracle]),value=a.truth-b.truth,bound=Math.abs(a.distortion)+Math.abs(b.distortion);
    assert(value>=-1e-12&&value<=bound+1e-12);regret.push({value,bound});bounds++;
  }
  if(r.Schedule===0){assert(selected.every(j=>j===-1));losses.forEach(x=>near(x,losses[0]));}
  const paid=branches.slice(1),probabilityMSE=paid.length?mean(paid.map(b=>(b.p-b.q)**2)):null;
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,branches,selected,losses,sampleLosses,costs,regret,probabilityMSE,paidChanged:Number(selected[5]!==selected[6]),abstainChanged:Number(selected[8]!==selected[9])});
}
assert.equal(records.length,2688);assert.equal(new Set(records.map(r=>`${r.phase}:${r.case}:${r.index}:${r.schedule}`)).size,2688);
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],reference.artifactHashes[0]);assert.equal(hashes[1],reference.artifactHashes[3]);assert.equal(hashes[2],reference.artifactHashes[1]);for(let i=0;i<2;i++)assert.equal(headers[i].InputSHA256,hashes[2]);
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const summarize=rs=>({records:rs.length,losses:Array.from({length:11},(_,a)=>mean(rs.map(r=>r.losses[a]))),costs:Array.from({length:11},(_,a)=>rs.reduce((s,r)=>s+r.costs[a],0)),paidChanged:rs.reduce((s,r)=>s+r.paidChanged,0),abstainChanged:rs.reduce((s,r)=>s+r.abstainChanged,0),probabilityMSE:rs.some(r=>r.probabilityMSE!==null)?mean(rs.filter(r=>r.probabilityMSE!==null).map(r=>r.probabilityMSE)):null});
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  groups.push({phase,case:c,schedule,...summarize(rs),modelVsEntropy:ci(rs.map(r=>r.losses[2]-r.losses[6])),modelRegret:ci(rs.map(r=>r.losses[6]-r.losses[5])),neutralRegret:ci(rs.map(r=>r.losses[7]-r.losses[5]))});
}
const delayed=summarize(records.filter(r=>r.schedule===1));delayed.entropyOracleGap=delayed.losses[2]-delayed.losses[5];delayed.massSubstitutionRegret=delayed.losses[6]-delayed.losses[5];delayed.gapFraction=delayed.massSubstitutionRegret/delayed.entropyOracleGap;
console.log(JSON.stringify({scope:'Offline branch-risk oracle with as-of candidate masses; branch risks still require teacher and future natural evidence',arms:['noquery','random','entropy','joint','disjoint','teacher-paid','model-paid','neutral-paid','teacher-abstain','model-abstain','neutral-abstain'],hashes:Object.fromEntries([[populationPath,hashes[0]],[decisionPath,hashes[1]],[sourcePath,hashes[2]],...[summaryPath,'research/query-mass.mjs','research/query-mass-test.mjs','research/query-mass-experiment.mjs','research/regime-query-outcome-audit.mjs','docs/experiments/mmm-query-mass-protocol.md'].map(p=>[p,sha(p)])]),identities,bounds,delayed,groups,records}));
