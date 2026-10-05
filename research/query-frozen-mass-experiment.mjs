import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {minimumMassRisk,weightedRisk} from './query-mass.mjs';
const [rawPath,coveragePath,parentPath]=process.argv.slice(2);
const parent=JSON.parse(fs.readFileSync(parentPath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[rawPath,coveragePath].map(input),headers=[];
for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
assert.equal(headers[0].Version,'query-publication-v1');assert.equal(headers[1].Version,'query-coverage-v1');
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-12);
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const records=[];let identities=0,bounds=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());
  if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [r,c]=next.map(x=>JSON.parse(x.value)),ref=parent.records[records.length];assert(ref&&!r.Error&&!c.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[upper],c[upper]);assert.equal(r[upper],ref[lower]);}
  assert.deepEqual((r.Branches??[]).map(b=>b.Origin),c.Pool??[]);
  const paid=(r.Branches??[]).map((b,i)=>{
    assert.equal(b.Origin,c.Branches[i].Origin);
    const p=c.Branches[i].Mass[1],q=b.Q,losses=b.Risk.map(x=>x.Population);
    const truth=weightedRisk(losses,q),modeled=weightedRisk(losses,p),distortion=(p-q)*(losses[1]-losses[0]);
    near(modeled-truth,distortion);identities++;
    return {origin:b.Origin,losses,p,q,truth,modeled,distortion,actual:losses[Number(b.ActualY)],sample:weightedRisk(b.Risk.map(x=>x.Sample),q)};
  });
  const states={};
  for(const [si,state]of ['A','B'].entries()){
    const base=r.Baseline[si];
    const actions=[{origin:-1,losses:[base.Population,base.Population],p:.5,q:.5,truth:base.Population,modeled:base.Population,distortion:0,actual:base.Population,sample:base.Sample},...paid];
    const choose=(field,abstain)=>minimumMassRisk((abstain||!paid.length?actions:paid).map(b=>({origin:b.origin,losses:b.losses,probability:field==='neutral'?.5:b[field]}))).origin;
    const selected=[...ref.selected,...[false,true].flatMap(abstain=>['q','p','neutral'].map(f=>choose(f,abstain)))];
    const map=new Map(actions.map(b=>[b.origin,b]));
    const losses=selected.map(j=>map.get(j).truth),actual=selected.map(j=>map.get(j).actual),sample=selected.map(j=>map.get(j).sample),costs=selected.map(j=>Number(j>=0));
    for(let a=0;a<10;a++){near(losses[a],ref.states[state].integrated[a]);near(actual[a],ref.states[state].actualPopulation[a]);near(sample[a],ref.states[state].expected[a]);}
    const regret=[];
    for(const [oracle,model]of [[10,11],[13,14]]){
      const a=map.get(selected[model]),b=map.get(selected[oracle]),value=a.truth-b.truth,bound=Math.abs(a.distortion)+Math.abs(b.distortion);
      assert(value>=-1e-12&&value<=bound+1e-12);bounds++;regret.push({value,bound});
    }
    states[state]={baseline:actions[0],selected,losses,actual,sample,costs,regret,paidChanged:Number(selected[10]!==selected[11]),abstainChanged:Number(selected[13]!==selected[14])};
  }
  assert.deepEqual(states.A.selected.slice(10,13),states.B.selected.slice(10,13));
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,paid,states,probabilityMSE:paid.length?mean(paid.map(b=>(b.p-b.q)**2)):null});
}
assert.equal(records.length,2688);assert.equal(new Set(records.map(r=>`${r.phase}:${r.case}:${r.index}:${r.schedule}`)).size,2688);
const hashes=streams.map(s=>s.hash.digest('hex'));for(let i=0;i<2;i++)assert.equal(hashes[i],parent.hashes[[rawPath,coveragePath][i]]);assert.equal(headers[0].CoverageSHA256,hashes[1]);
const summarize=rs=>({records:rs.length,probabilityMSE:rs.some(r=>r.probabilityMSE!==null)?mean(rs.filter(r=>r.probabilityMSE!==null).map(r=>r.probabilityMSE)):null,states:Object.fromEntries(['A','B'].map(s=>[s,{...Object.fromEntries(['losses','actual','sample'].map(f=>[f,Array.from({length:16},(_,a)=>mean(rs.map(r=>r.states[s][f][a])))])),costs:Array.from({length:16},(_,a)=>rs.reduce((n,r)=>n+r.states[s].costs[a],0)),paidChanged:rs.reduce((n,r)=>n+r.states[s].paidChanged,0),abstainChanged:rs.reduce((n,r)=>n+r.states[s].abstainChanged,0)}]))});
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  groups.push({phase,case:c,schedule,...summarize(rs),contrasts:Object.fromEntries(['A','B'].map(s=>[s,{modelVsEntropy:ci(rs.map(r=>r.states[s].losses[2]-r.states[s].losses[11])),modelRegret:ci(rs.map(r=>r.states[s].losses[11]-r.states[s].losses[10]))}]))});
}
console.log(JSON.stringify({scope:'Fixed-support offline branch-risk oracle, not an online acquisition policy or fresh confirmation',arms:[...parent.arms,'teacher-paid','model-paid','neutral-paid','teacher-abstain','model-abstain','neutral-abstain'],hashes:Object.fromEntries([[rawPath,hashes[0]],[coveragePath,hashes[1]],...[parentPath,import.meta.filename,'research/query-mass.mjs','research/query-mass-test.mjs','docs/experiments/mmm-query-frozen-mass-protocol.md'].map(p=>[p,sha(p)])]),identities,bounds,delayed:summarize(records.filter(r=>r.schedule===1)),phase1Delayed:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),groups,records}));
