import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import readline from "node:readline";
import {createGunzip} from "node:zlib";
import {fileURLToPath} from "node:url";

export const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const q=(a,p)=>[...a].sort((a,b)=>a-b)[Math.ceil(a.length*p)-1];
export function maxDifference(a,b){
  if(typeof a==="number"&&typeof b==="number")return Math.abs(a-b);
  if(a===b)return 0;
  if(a===null||b===null||typeof a!=="object"||typeof b!=="object")return Infinity;
  const aa=Object.keys(a).sort(),bb=Object.keys(b).sort();if(JSON.stringify(aa)!==JSON.stringify(bb))return Infinity;
  return Math.max(0,...aa.map(k=>maxDifference(a[k],b[k])));
}
function lawCheck(f){for(const k of["base_law","pre_residual_law","corrected_law"]){const p=f[k];assert.ok(Number.isFinite(p.useful)&&p.useful>=0&&p.useful<=1);assert.ok(Number.isFinite(p.not_useful)&&Math.abs(p.useful+p.not_useful-1)<1e-12);}
  assert.equal(f.belief_law,undefined);assert.equal(f.residual_applied,false);assert.ok(!f.expert_mixture?.enabled);
}
function unique(a){assert.equal(new Set(a).size,a.length);assert.ok(a.every(id=>/^past-\d{4}$/.test(id)&&Number(id.slice(5))<4000));}
export function summarize(row){
  const {nomination:n,exact_top:e,k}=row;assert.equal(n.length,3*k);assert.equal(e.length,3*k);unique(n.map(x=>x.id));unique(e.map(x=>x.id));
  assert.ok(n.every(x=>Number.isFinite(x.exact)&&Number.isFinite(x.native)&&Number.isFinite(x.baseline)));
  assert.ok(e.every(x=>x.exact===x.native));for(let i=1;i<e.length;i++){assert.ok(e[i-1].exact>=e[i].exact);if(e[i-1].exact===e[i].exact)assert.ok(e[i-1].id<e[i].id);}
  const cutoff=e.at(-1).exact,ids=new Set(e.map(x=>x.id));
  const required=new Set(e.filter(x=>x.exact>cutoff+1e-6).map(x=>x.id));
  const requiredHits=n.filter(x=>required.has(x.id)).length;
  const tieSlots=e.length-required.size,tieHits=n.filter(x=>Math.abs(x.exact-cutoff)<=1e-6).length;
  const tieRecall=(requiredHits+Math.min(tieSlots,tieHits))/n.length;
  const strictRecall=n.filter(x=>ids.has(x.id)).length/n.length;
  const regret=Math.max(0,mean(e.map(x=>x.exact))-mean(n.map(x=>x.exact)));
  const a=row.actual,r=row.reference;
  for(const packet of[a,r]){
    assert.equal(packet.frontier.length,k);unique(packet.frontier.map(x=>x.id));unique(packet.ids);
    assert.ok(packet.ids.length>0&&packet.ids.length<=10);assert.equal(packet.ids.length,packet.exact_baseline.length);assert.equal(packet.ids.length,packet.forecasts.length);
    const frontierIDs=new Set(packet.frontier.map(x=>x.id));assert.ok(packet.ids.every(id=>frontierIDs.has(id)));
    assert.ok(Number.isFinite(packet.confidence)&&Number.isFinite(packet.answer_certainty));
    for(const f of [...packet.forecasts,...packet.frontier.map(x=>x.forecast)])lawCheck(f);
  }
  const refIDs=new Set(r.ids),refFrontier=new Map(r.frontier.map(x=>[x.id,x.forecast]));
  const common=a.frontier.filter(x=>refFrontier.has(x.id));
  const lawDelta=Math.max(0,...common.flatMap(x=>["base_law","pre_residual_law","corrected_law"].map(k=>Math.abs(x.forecast[k].useful-refFrontier.get(x.id)[k].useful))));
  const compact=p=>({ids:p.ids,frontier:p.frontier,confidence:p.confidence,answer_certainty:p.answer_certainty});
  return {phase:row.phase,query:row.query,k,tieRecall,strictRecall,regret,packetDeficit:Math.max(0,mean(r.exact_baseline)-mean(a.exact_baseline)),
    packedCount:a.ids.length,referencePackedCount:r.ids.length,packetIDOverlap:a.ids.filter(id=>refIDs.has(id)).length/a.ids.length,
    commonReferenceLaws:common.length,maxReferenceLawDelta:lawDelta,maxBinaryBrierDifferenceBound:Math.min(1,2*lawDelta),recallNS:a.ns,
    actual:compact(a),reference:compact(r),native:n.map(x=>({id:x.id,score:x.native})),exactDigest:hash(JSON.stringify(e))};
}
export function compare(a,b,key){
  const x=a[key],y=b[key],ym=new Map(y.frontier.map(c=>[c.id,c.forecast]));
  const common=x.frontier.filter(c=>ym.has(c.id));
  const maxForecastDelta=Math.max(0,...common.map(c=>maxDifference(c.forecast,ym.get(c.id))));
  const maxLawDelta=Math.max(0,...common.flatMap(c=>["base_law","pre_residual_law","corrected_law"].map(k=>Math.abs(c.forecast[k].useful-ym.get(c.id)[k].useful))));
  const packetChanged=JSON.stringify(x.ids)!==JSON.stringify(y.ids),frontierChanged=JSON.stringify(x.frontier.map(c=>c.id))!==JSON.stringify(y.frontier.map(c=>c.id));
  const confidenceDelta=Math.max(Math.abs(x.confidence-y.confidence),Math.abs(x.answer_certainty-y.answer_certainty));
  const nominationChanged=key==="actual"&&JSON.stringify(a.native.map(x=>x.id))!==JSON.stringify(b.native.map(x=>x.id));
  const nativeB=new Map(b.native.map(x=>[x.id,x.score]));
  const maxNativeScoreDelta=key==="actual"?Math.max(0,...a.native.filter(x=>nativeB.has(x.id)).map(x=>Math.abs(x.score-nativeB.get(x.id)))):0;
  return {packetChanged,frontierChanged,nominationChanged,maxNativeScoreDelta,commonLaws:common.length,maxForecastDelta,maxLawDelta,confidenceDelta,
    strictEqual:!packetChanged&&!frontierChanged&&maxForecastDelta<=1e-9&&confidenceDelta<=1e-9};
}
async function parse(root,c){
  const file=path.join(root,c.transcript),stream=()=>fs.existsSync(file)?fs.createReadStream(file):fs.createReadStream(file+".gz").pipe(createGunzip());
  const digest=crypto.createHash("sha256");let bytes=0;for await(const chunk of stream()){digest.update(chunk);bytes+=chunk.length;}assert.equal(digest.digest("hex"),c.sha256);assert.equal(bytes,c.bytes);
  let header=null,terminal=null;const rows=[];
  for await(const line of readline.createInterface({input:stream(),crlfDelay:Infinity})){
    for(const[prefix,target]of[["SHARD_QUALITY_HEADER=","header"],["SHARD_QUALITY_TERMINAL=","terminal"],["SHARD_QUALITY_ROW=","row"]]){
      const at=line.indexOf(prefix);if(at<0)continue;const value=JSON.parse(line.slice(at+prefix.length));
      if(target==="header"){assert.equal(header,null);header=value;}else if(target==="terminal"){assert.equal(terminal,null);terminal=value;}else rows.push(summarize(value));
    }
  }
  assert.equal(header.questions,96);assert.equal(header.past_count,4000);assert.equal(header.nomination_multiplier,3);assert.equal(header.cold_law_only,true);
  assert.equal(header.shard_routes.length,4);assert.equal(header.shard_routes.reduce((s,n)=>s+n,0),4000);assert.ok(header.shard_routes.every(n=>n>600));
  assert.equal(terminal.functional,true);assert.equal(terminal.rows,768);assert.equal(terminal.future_count,256);assert.equal(rows.length,768);
  const byKey=new Map(rows.map(x=>[`${x.phase}/${x.k}/${x.query}`,x]));assert.equal(byKey.size,768);
  const cells=[],comparisons=[];
  for(const k of[50,200]){
    for(const phase of["past","unchanged_repeat","future","future_repeat"]){
      const rr=Array.from({length:96},(_,i)=>byKey.get(`${phase}/${k}/${i}`));assert.ok(rr.every(Boolean));
      const metric={arm:c.arm,pair:c.pair,k,phase,queries:96,meanTieRecall:mean(rr.map(x=>x.tieRecall)),meanStrictIDRecall:mean(rr.map(x=>x.strictRecall)),meanExactScoreRegret:mean(rr.map(x=>x.regret)),meanPacketDeficit:mean(rr.map(x=>x.packetDeficit)),meanPacketIDOverlap:mean(rr.map(x=>x.packetIDOverlap)),packetCountMismatch:rr.filter(x=>x.packedCount!==x.referencePackedCount).length,commonReferenceLaws:rr.reduce((s,x)=>s+x.commonReferenceLaws,0),maxReferenceLawDelta:Math.max(...rr.map(x=>x.maxReferenceLawDelta)),maxBinaryBrierDifferenceBound:Math.max(...rr.map(x=>x.maxBinaryBrierDifferenceBound)),recallP99MS:q(rr.map(x=>x.recallNS),.99)/1e6};
      metric.annPass=metric.meanTieRecall>=.95&&metric.meanExactScoreRegret<=.005;metric.packetPass=metric.meanPacketDeficit<=.01;cells.push(metric);
    }
    for(const[from,to,label]of[["past","unchanged_repeat","past_repeat"],["future","future_repeat","future_repeat"],["unchanged_repeat","future","future_counterfactual"]])for(const key of["actual","reference"]){
      const values=Array.from({length:96},(_,i)=>compare(byKey.get(`${from}/${k}/${i}`),byKey.get(`${to}/${k}/${i}`),key));
      comparisons.push({arm:c.arm,pair:c.pair,k,label,key,queries:96,changedPackets:values.filter(x=>x.packetChanged).length,changedFrontiers:values.filter(x=>x.frontierChanged).length,strictEqualQueries:values.filter(x=>x.strictEqual).length,
        changedNominations:values.filter(x=>x.nominationChanged).length,maxNativeScoreDelta:Math.max(...values.map(x=>x.maxNativeScoreDelta)),commonLaws:values.reduce((s,x)=>s+x.commonLaws,0),maxForecastDelta:Math.max(...values.map(x=>x.maxForecastDelta)),maxLawDelta:Math.max(...values.map(x=>x.maxLawDelta)),maxConfidenceDelta:Math.max(...values.map(x=>x.confidenceDelta)),pass:values.every(x=>x.strictEqual)});
    }
  }
  return {arm:c.arm,pair:c.pair,header,terminal,cells,comparisons,exact:rows.map(x=>[`${x.phase}/${x.k}/${x.query}`,x.exactDigest])};
}
export async function evaluate(root){
  const run=JSON.parse(fs.readFileSync(path.join(root,"run-results.json")));assert.equal(run.completed,true);assert.equal(run.preflightPassed,true);assert.equal(run.functionalPass,true);
  const commands=run.commands.filter(c=>c.full);assert.equal(commands.length,4);assert.ok(commands.every(c=>c.status===0&&!c.dataRaceReported));
  const arms=[];for(const c of commands)arms.push(await parse(root,c));
  const cells=arms.flatMap(a=>a.cells),comparisons=arms.flatMap(a=>a.comparisons),paired=[];
  for(const pair of[0,1]){
    const control=arms.find(a=>a.pair===pair&&a.arm==="control"),candidate=arms.find(a=>a.pair===pair&&a.arm==="candidate");assert.deepEqual(control.exact,candidate.exact);
    for(const k of[50,200])for(const phase of["past","unchanged_repeat","future","future_repeat"]){
      const c=cells.find(x=>x.pair===pair&&x.arm==="control"&&x.k===k&&x.phase===phase),n=cells.find(x=>x.pair===pair&&x.arm==="candidate"&&x.k===k&&x.phase===phase);
      paired.push({pair,k,phase,tieRecallDifference:n.meanTieRecall-c.meanTieRecall,pass:n.meanTieRecall-c.meanTieRecall>=-.02});
    }
  }
  return {verified:true,functionalPass:true,cells,comparisons,paired,annAbsolutePass:cells.every(c=>c.annPass),packetQualityPass:cells.every(c=>c.packetPass),pairedQualityPass:paired.every(p=>p.pass),
    repeatEqualityPass:comparisons.filter(x=>x.key==="actual"&&x.label!=="future_counterfactual").every(x=>x.pass),counterfactualEqualityPass:comparisons.filter(x=>x.key==="actual"&&x.label==="future_counterfactual").every(x=>x.pass),exactReferenceStabilityPass:comparisons.filter(x=>x.key==="reference").every(x=>x.pass),
    finiteDesignOnly:true,coldLawOnly:true,semanticTaskAccuracyTested:false,populationTailGuarantee:false,wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){
  const root=path.dirname(fileURLToPath(import.meta.url)),result=await evaluate(root);
  fs.writeFileSync(path.join(root,"evaluation.json"),JSON.stringify(result,null,2)+"\n");
  console.log(JSON.stringify(Object.fromEntries(Object.entries(result).filter(([k])=>!["cells","comparisons","paired"].includes(k)))));
}
