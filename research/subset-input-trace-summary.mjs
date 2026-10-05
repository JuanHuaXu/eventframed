import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(hash(fs.readFileSync('docs/experiments/mmm-subset-input-integration-v1.jsonl')),header.ArchiveHash);
for(const [name,h]of Object.entries(header.Hashes)){assert.equal(hash(header.Sources[name]),h);assert.equal(hash(fs.readFileSync('internal/observationgate/'+name)),h);}
assert.equal(rows.length,32);
const complete=p=>(p.mask&30)===30;
const floor=p=>complete(p)?.0475:.25;
const expected=p=>{let q=.5;if(complete(p)){let n=0;for(let b=1;b<=4;b++)n+=(p.values>>b)&1;q=n%2?.95:.05;}return q*(1-p.p)**2+(1-q)*p.p**2;};
const cells=[];
for(const phase of ['design','confirmation']){
 const rs=rows.filter(r=>r.Phase===phase);assert.equal(rs.length,16);
 assert.deepEqual(rs.map(r=>r.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
 const a={phase,n:0,controlComplete:0,coupledComplete:0,lostComplete:0,gainedComplete:0,subsetGuided:0,confidenceIncomplete:0,subsetConfidenceIncomplete:0,subsetBudgetIncomplete:0,controlFloor:0,coupledFloor:0,controlExpected:0,coupledExpected:0,stops:{},costs:{}};
 for(const r of rs){assert.equal(r.Steps.length,256);for(let i=0;i<256;i++){
  const t=r.Steps[i];assert.equal(t.Step,256+i);a.n++;
  for(const p of [t.Control,t.Coupled]) {
   assert(Number.isInteger(p.mask)&&p.mask>=0&&p.mask<512);
   assert(Number.isInteger(p.values)&&p.values>=0&&(p.values&~p.mask)===0);
   assert(Number.isInteger(p.cost)&&p.cost>=1&&p.cost<=6);
   assert(Number.isFinite(p.p)&&p.p>=0&&p.p<=1);
  }
  const c=complete(t.Control),b=complete(t.Coupled);
  a.controlComplete+=Number(c);a.coupledComplete+=Number(b);a.lostComplete+=Number(c&&!b);a.gainedComplete+=Number(!c&&b);
  a.subsetGuided+=Number(t.CoupledSubset);
  a.confidenceIncomplete+=Number(t.Stop==='confidence'&&!b);
  a.subsetConfidenceIncomplete+=Number(t.CoupledSubset&&t.Stop==='confidence'&&!b);
  a.subsetBudgetIncomplete+=Number(t.CoupledSubset&&t.Stop==='budget'&&!b);
  a.controlFloor+=floor(t.Control);a.coupledFloor+=floor(t.Coupled);
  a.controlExpected+=expected(t.Control);a.coupledExpected+=expected(t.Coupled);
  a.stops[t.Stop]=(a.stops[t.Stop]??0)+1;
  a.costs[t.Coupled.cost]=(a.costs[t.Coupled.cost]??0)+1;
 }}
 for(const k of ['controlFloor','coupledFloor','controlExpected','coupledExpected'])a[k]/=a.n;
 cells.push(a);
}
const result={sourceHash:hash(bytes),archiveHash:header.ArchiveHash,cells,limits:'Consumed parity4 trajectories only; known independent-input noisy-parity oracle is evaluator-only, not available to the observer. Aggregate mechanism diagnostic, not a new rescue.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(result,null,2));
