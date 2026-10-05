import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const pop = x => { let n = 0; for (; x; x &= x - 1) n++; return n; };
const mean = a => a.reduce((s,x)=>s+x,0)/a.length;
const [input,parentPath,output] = process.argv.slice(2);
const raw=fs.readFileSync(input), parentRaw=fs.readFileSync(parentPath);
const parse=b=>b.toString().trim().split('\n').map(JSON.parse);
const [header,...rows]=parse(raw),[, ...parents]=parse(parentRaw);
assert.equal(hash(parentRaw),header.ParentSHA256);
for(const [p,h] of Object.entries(header.Hashes)) assert.equal(hash(fs.readFileSync(p)),h,p);
assert.equal(rows.length,1024); assert.equal(parents.length,128);
const map=new Map(parents.map(p=>[[p.Phase,p.Case,p.Index].join('/'),p]));
const names=['stable_majority3','stable_parity4','majority_to_parity','parity_to_majority'];
const seen=new Set(),records=[];let checks=0,noFit=0;
for(const r of rows){
 const key=[r.Phase,r.Case,r.Index].join('/'),id=[key,r.Schedule,r.Checkpoint].join('/');
 assert.ok(!seen.has(id));seen.add(id);assert.ok([128,256,384,480].includes(r.Checkpoint));
 const p=map.get(key);assert.ok(p);const tape=p[r.Schedule];assert.ok(tape);
 const fit=tape.Fits.filter(f=>f.Clock<r.Checkpoint).at(-1);
 assert.equal(r.FitClock,fit?.Clock??-1);if(!fit)noFit++;
 const scenario=names.indexOf(r.Case),changed=scenario>=2&&r.Checkpoint>=256;
 assert.equal(r.Boundary,changed?256:0);
 const ids=r.Origins.map(x=>x??[]),actual=fit?.Origins.slice(-64)??[],clean=(fit?.Origins.filter(i=>i>=r.Boundary)??[]).slice(-64);
 assert.deepEqual(ids[0],actual);assert.deepEqual(ids[2],clean);assert.equal(ids[1].length,clean.length);
 if(clean.length===actual.length)assert.deepEqual(ids[1],actual);
 for(const group of ids)for(let j=0;j<group.length;j++){
  const i=group[j],f=tape.Frames[i];assert.ok(f&&f.Audit&&!f.Missing&&f.Arrival<=r.FitClock);
  assert.ok(j===0||group[j-1]<i);assert.ok(fit.Origins.includes(i));
 }
 for(const i of ids[1])assert.ok(actual.includes(i));
 const target=p.Masks[Number(changed)],majority=(scenario===0||scenario===2)!==changed;
 const risk=ids.map(()=>[0,0]);
 for(let arm=0;arm<3;arm++){
  assert.equal(r.Predictions[arm].length,512);
  const n=Array(512).fill(0),yes=Array(512).fill(0);
  for(const i of ids[arm]){const f=tape.Frames[i];n[f.X]++;yes[f.X]+=Number(f.Y);}
  for(let x=0;x<512;x++){
   const truth=majority?pop(x&target)>=2:pop(x&target)%2===1,q=.05+.9*Number(truth);
   const predictions=r.Predictions[arm][x];assert.equal(predictions.length,2);
   assert.ok(Math.abs(predictions[0]-(yes[x]+1)/(n[x]+2))<1e-12);checks++;
   for(let m=0;m<2;m++){
    const v=predictions[m];assert.ok(Number.isFinite(v)&&v>=0&&v<=1);
    if(!ids[arm].length)assert.equal(v,.5);
    risk[arm][m]+=(.0475+(v-q)**2)/512;
   }
  }
 }
 if(!changed)assert.deepEqual(r.Predictions[0],r.Predictions[2]);
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,checkpoint:r.Checkpoint,fitClock:r.FitClock,counts:ids.map(a=>a.length),risk});
}
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of names)for(const schedule of ['Immediate','Delayed'])for(const checkpoint of [128,256,384,480]){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule&&r.checkpoint===checkpoint);assert.equal(a.length,16);
 cells.push({phase,case:name,schedule,checkpoint,counts:[0,1,2].map(j=>mean(a.map(r=>r.counts[j]))),risk:[0,1,2].map(j=>[0,1].map(m=>mean(a.map(r=>r.risk[j][m]))))});
}
const result={rawSHA256:hash(raw),scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),sources:Object.keys(header.Hashes).length,checks,noFit,cells,records,limits:'Consumed hindsight full-input expected risk; not served quality, a deployable reset, or fresh confirmation. Matched controls use one seeded sample each. No uncertainty guarantee is inferred from cell means.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({checks,noFit,cells:cells.filter(c=>c.case.includes('_to_')&&c.checkpoint>=384)}));
