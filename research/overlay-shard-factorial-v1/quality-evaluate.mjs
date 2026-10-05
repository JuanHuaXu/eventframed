import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {parse} from './quality-core.mjs';
export async function evaluate(root){
 const run=JSON.parse(fs.readFileSync(path.join(root,'run-results.json')));
 assert.equal(run.completed,true);assert.equal(run.preflightPassed,true);
 const commands=run.commands.filter(c=>c.kind==='quality');assert.equal(commands.length,8);
 const orders=[['rebuild','overlay','shard','both'],['both','shard','overlay','rebuild']];
 assert.deepEqual(commands.map(c=>[c.pair,c.arm]),orders.flatMap((a,p)=>a.map(x=>[p,x])));
 const arms=[];for(const c of commands){assert.equal(c.status,0);assert.equal(c.functional,true);assert.equal(c.dataRaceReported,false);arms.push(await parse(root,c));}
 const cells=arms.flatMap(a=>a.cells),comparisons=arms.flatMap(a=>a.comparisons),paired=[];
 for(const pair of[0,1])for(const[reference,candidate]of[['rebuild','overlay'],['rebuild','shard'],['rebuild','both'],['shard','both'],['overlay','both']]){
  const a=arms.find(x=>x.pair===pair&&x.arm===reference),b=arms.find(x=>x.pair===pair&&x.arm===candidate);assert.deepEqual(a.exact,b.exact);
  for(const k of[50,200])for(const phase of['past','unchanged_repeat','future','future_repeat']){
   const c=a.cells.find(x=>x.k===k&&x.phase===phase),n=b.cells.find(x=>x.k===k&&x.phase===phase),difference=n.meanTieRecall-c.meanTieRecall;
   paired.push({pair,reference,candidate,k,phase,tieRecallDifference:difference,pass:difference>=-.02});
  }
 }
 const armVerdicts=Object.fromEntries(['rebuild','overlay','shard','both'].map(arm=>{
  const cc=cells.filter(c=>c.arm===arm),dd=comparisons.filter(c=>c.arm===arm);
  return[arm,{ann:cc.every(x=>x.annPass),packet:cc.every(x=>x.packetPass),repeat:dd.filter(x=>x.key==='actual'&&x.label!=='future_counterfactual').every(x=>x.pass),future:dd.filter(x=>x.key==='actual'&&x.label==='future_counterfactual').every(x=>x.pass),exactReference:dd.filter(x=>x.key==='reference').every(x=>x.pass)}];
 }));
 return{verified:true,cells,comparisons,paired,armVerdicts,pairedQualityPass:paired.every(x=>x.pass),finiteDesignOnly:true,coldLawOnly:true,semanticTaskAccuracyTested:false,populationTailGuarantee:false,wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){const root=path.dirname(fileURLToPath(import.meta.url)),x=await evaluate(root);fs.writeFileSync(path.join(root,'quality-evaluation.json'),JSON.stringify(x,null,2)+'\n');console.log(JSON.stringify({verified:x.verified,arms:x.armVerdicts,pairedQualityPass:x.pairedQualityPass}));}
