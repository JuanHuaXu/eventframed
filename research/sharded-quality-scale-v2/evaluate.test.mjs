import test from "node:test";
import assert from "node:assert/strict";
import {summarize,compare,maxDifference} from "./evaluate.mjs";

const forecast=id=>({model_kind:"cold",rank_score:.5,horizon_key:"test",base_law:{useful:.5,not_useful:.5},pre_residual_law:{useful:.5,not_useful:.5},corrected_law:{useful:.5,not_useful:.5},template:{event_id:id,predicted_useful:true,confidence:.5},residual_applied:false});
function fixture(){
  const ids=Array.from({length:150},(_,i)=>"past-"+String(i).padStart(4,"0"));
  const scores=ids.map(id=>({id,native:.5,exact:.5,baseline:.5}));
  const packet={ids:ids.slice(0,10),exact_baseline:Array(10).fill(.5),forecasts:ids.slice(0,10).map(forecast),frontier:ids.slice(0,50).map(id=>({id,baseline:.5,forecast:forecast(id)})),confidence:.5,answer_certainty:.5,ns:100};
  return {phase:"past",query:0,k:50,nomination:structuredClone(scores),exact_top:scores,actual:structuredClone(packet),reference:packet};
}
test("ties differ in ID without an exact-score deficit",()=>{
  const r=fixture();r.nomination.at(-1).id="past-0150";
  const x=summarize(r);assert.equal(x.tieRecall,1);assert.equal(x.strictRecall,149/150);assert.equal(x.regret,0);
});
test("missed high-score neighbor incurs regret and lost tie recall",()=>{
  const r=fixture();r.nomination.at(-1).exact=0;
  const x=summarize(r);assert.equal(x.tieRecall,149/150);assert.ok(Math.abs(x.regret-.5/150)<1e-12);
});
test("boundary tie cannot replace a required strictly better neighbor",()=>{
  const r=fixture();r.exact_top[0].native=r.exact_top[0].exact=.9;
  r.nomination[0]={id:"past-0150",native:.5,exact:.5,baseline:.5};
  assert.equal(summarize(r).tieRecall,149/150);
});
test("future and duplicate nominations fail validation",()=>{
  for(const id of["future-0000","past-0000"]){const r=fixture();r.nomination.at(-1).id=id;assert.throws(()=>summarize(r));}
});
test("law motion cannot hide behind matching packet IDs",()=>{
  const a=summarize(fixture()),b=structuredClone(a);const f=b.actual.frontier[0].forecast;
  f.corrected_law={useful:.6,not_useful:.4};const c=compare(a,b,"actual");assert.equal(c.packetChanged,false);assert.equal(c.frontierChanged,false);assert.equal(c.strictEqual,false);assert.ok(Math.abs(c.maxLawDelta-.1)<1e-12);
});
test("packet confidence motion and frontier order are independently visible",()=>{
  const a=summarize(fixture());for(const change of["confidence","order"]){const b=structuredClone(a);if(change==="confidence")b.actual.confidence+=.01;else b.actual.frontier.reverse();assert.equal(compare(a,b,"actual").strictEqual,false);}
});
test("invalid and nonnormalized laws are rejected",()=>{
  for(const law of[{useful:NaN,not_useful:.5},{useful:.5,not_useful:.7}]){const r=fixture();r.actual.frontier[0].forecast.corrected_law=law;assert.throws(()=>summarize(r));}
});
test("field deletion is not numerical equality",()=>{
  const a=forecast("past-0000"),b=structuredClone(a);delete b.template;assert.equal(maxDifference(a,b),Infinity);
});
test("binary Brier difference is bounded by 2 absolute probability motion",()=>{
  for(let i=0;i<=100;i++)for(let j=0;j<=100;j++)for(const y of[0,1]){const p=i/100,q=j/100;assert.ok(Math.abs((p-y)**2-(q-y)**2)<=Math.min(1,2*Math.abs(p-q))+1e-12);}
});
