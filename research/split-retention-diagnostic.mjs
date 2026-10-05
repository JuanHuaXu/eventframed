import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const mean=a=>a.length?a.reduce((s,x)=>s+x,0)/a.length:null;

// A hypothetical strict epoch-origin policy, NOT a change to live training.
function retained(fit, split, checkpoint){
  const active=split>=0&&split<checkpoint;
  if(!fit)return [];
  if(!active)return fit.Origins.slice(-64);
  if(fit.Clock<split)return [];
  return fit.Origins.filter(i=>i>split).slice(-64);
}
assert.deepEqual(retained({Clock:10,Origins:[1,4,8]},-1,12),[1,4,8]);
assert.deepEqual(retained({Clock:10,Origins:[1,4,8]},11,12),[]);
assert.deepEqual(retained({Clock:10,Origins:[1,4,8]},10,10),[1,4,8]);
assert.deepEqual(retained({Clock:10,Origins:[1,4,8]},4,12),[8]);
assert.deepEqual(retained(undefined,4,12),[]);

const [parentPath,cleanPath,output]=process.argv.slice(2);
const raw=fs.readFileSync(parentPath),cr=fs.readFileSync(cleanPath);
assert.equal(hash(raw),'4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655');
const clean=JSON.parse(cr);assert.equal(clean.rawSHA256,'ba5e0fcdd9e28a0cca2e027c9fdc69bd2945eed24a0a8906683c61fa2ce1d090');
const [, ...parents]=raw.toString().trim().split('\n').map(JSON.parse);
const map=new Map(parents.map(p=>[[p.Phase,p.Case,p.Index].join('/'),p])),records=[];
assert.equal(parents.length,128);assert.equal(clean.records.length,1024);
for(const r of clean.records){
 const p=map.get([r.phase,r.case,r.index].join('/'));assert.ok(p);
 const tape=p[r.schedule],split=tape.Arms[2].SplitAt;
 assert.ok(Number.isInteger(split)&&split>=-1&&split<544);
 const fit=tape.Fits.filter(f=>f.Clock<r.checkpoint).at(-1);
 assert.equal(fit?.Clock??-1,r.fitClock);
 const ids=retained(fit,split,r.checkpoint),active=split>=0&&split<r.checkpoint;
 const actual=fit?.Origins.slice(-64)??[];
 assert.equal(actual.length,r.counts[0]);
 for(const i of ids){const f=tape.Frames[i];assert.ok(f.Audit&&!f.Missing&&f.Arrival<=fit.Clock);if(active)assert.ok(i>split);}
 const changed=r.case.includes('_to_')&&r.checkpoint>=256;
 const oracle=changed?(fit?.Origins.filter(i=>i>=256)??[]).slice(-64):actual;
 assert.equal(oracle.length,r.counts[2]);
 const cleanKept=changed?ids.filter(i=>i>=256).length:null;
 records.push({phase:r.phase,case:r.case,index:r.index,schedule:r.schedule,checkpoint:r.checkpoint,fitClock:r.fitClock,split,active,
  actualCount:actual.length,retainedCount:ids.length,oracleCount:oracle.length,cleanKept,
  cleanDiscarded:changed?oracle.length-cleanKept:null,staleKept:changed?ids.length-cleanKept:null,
  invalidatedWithoutRefit:active&&(!fit||fit.Clock<split),ids});
}
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed'])for(const checkpoint of [128,256,384,480]){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule&&r.checkpoint===checkpoint);assert.equal(a.length,16);
 const changed=name.includes('_to_')&&checkpoint>=256;
 cells.push({phase,case:name,schedule,checkpoint,active:a.filter(r=>r.active).length,empty:a.filter(r=>r.retainedCount===0).length,
  invalidatedWithoutRefit:a.filter(r=>r.invalidatedWithoutRefit).length,
  meanRetained:mean(a.map(r=>r.retainedCount)),meanOracle:mean(a.map(r=>r.oracleCount)),
  meanCleanKept:changed?mean(a.map(r=>r.cleanKept)):null,meanStaleKept:changed?mean(a.map(r=>r.staleKept)):null,
  meanCleanDiscarded:changed?mean(a.map(r=>r.cleanDiscarded)):null,
  activeMeanRetained:mean(a.filter(r=>r.active).map(r=>r.retainedCount)),activeMeanOracle:mean(a.filter(r=>r.active).map(r=>r.oracleCount))});
}
const result={parentSHA256:hash(raw),cleanSummarySHA256:hash(cr),scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),records,cells,
 limits:'Metadata-only hypothetical strict post-detection origin filter. Does not refit models, change predictions, estimate onset, or validate a rescue. True boundary labels only evaluator retention counts. No outcome labels consulted; missing splits preserve existing mixed training.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.case.includes('_to_')&&c.checkpoint>=384)));
