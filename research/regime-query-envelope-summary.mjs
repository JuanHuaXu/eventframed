import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditEnvelope} from './regime-query-envelope-audit.mjs';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';

const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function input(file){const stream=fs.createReadStream(file),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
// Attach every iterator before awaiting: each hashing listener starts flowing.
const streams=process.argv.slice(2).map(input);assert.equal(streams.length,4);
const headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
assert.equal(headers[0].Version,'regime-query-envelope-v1');assert.equal(headers[1].Version,'soft-learners-v120');
assert.equal(headers[2].Version,'regime-query-outcome-v1');assert.equal(headers[3].Version,'regime-query-disjoint-v1');
assert.equal(headers[0].Records,2688);assert.equal(headers[0].Workers,4);
for(const [file,hash]of Object.entries(headers[0].Hashes)){const p=path.resolve('internal/observationlearners',file);assert(p.startsWith(process.cwd()+path.sep));assert.equal(sha(p),hash);}
const records=[],ids=new Set();let checks=0,fits=0,branchCount=0;
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
for(;;){
  const lines=[];for(const s of streams)lines.push(await s.it.next());
  if(lines.some(x=>x.done)){assert(lines.every(x=>x.done));break;}
  const [r,source,old,disjoint]=lines.map(x=>JSON.parse(x.value));
  const id=`${source.Phase}:${source.Case}:${source.Index}:${source.Schedule}`;assert(!ids.has(id));ids.add(id);
  assert(!disjoint.Error);assert.deepEqual(disjoint.Original,old);
  const oa=auditRegimeOutcome(old,source),da=auditRegimeOutcome(disjoint.Candidate,source,137);
  const audited=auditEnvelope(r,source);checks+=audited.checks;fits+=audited.fits;branchCount+=audited.branches.length;
  const branches=audited.branches.map(b=>({...b,loss:mean(b.pred.map((p,i)=>{const q=source.Steps[161+i].Q;assert(Number.isFinite(q)&&q>=0&&q<=1);return (p-q)**2+q*(1-q);} ))}));
  for(const bundle of [old,disjoint.Candidate])for(let a=0;a<4;a++){
    const branch=branches.find(b=>b.selected===bundle.Decision.Selected[a]);assert(branch);
    assert.deepEqual(branch.pred,bundle.Predictions[a]);assert.deepEqual(branch.at,bundle.AtPublication[a]);
    assert.deepEqual(branch.support,bundle.Origins[a]);assert.equal(branch.logEvidence,bundle.LogEvidence[a]);
  }
  const base=branches[0].loss,paid=branches.slice(1),losses=paid.map(b=>b.loss);
  const best=paid.length?Math.min(...losses):base,worst=paid.length?Math.max(...losses):base;
  const average=paid.length?mean(losses):base;
  records.push({phase:source.Phase,case:source.Case,index:source.Index,schedule:source.Schedule,
    losses:[...oa.brier,da.brier[3],average,best,worst,Math.min(base,best)],
    branches:branches.map(b=>({origin:b.selected,loss:b.loss,cost:b.cost,redundant:b.redundant})),
    anyGain:Number(best<base-1e-12),allHarm:Number(paid.length>0&&best>base+1e-12)});
}
assert.equal(records.length,2688);
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[1],'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
for(const h of [headers[0],headers[2],headers[3]])assert.equal(h.InputSHA256,hashes[1]);assert.equal(headers[3].BaselineSHA256,hashes[2]);
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  groups.push({phase,case:c,schedule,meanLoss:Array.from({length:9},(_,a)=>mean(rs.map(r=>r.losses[a]))),
    bestPaidGains:Array.from({length:6},(_,a)=>ci(rs.map(r=>r.losses[a]-r.losses[6]))),
    anyGain:rs.reduce((s,r)=>s+r.anyGain,0),allHarm:rs.reduce((s,r)=>s+r.allHarm,0)});
}
const delayed=records.filter(r=>r.schedule===1);
console.log(JSON.stringify({scope:'All-candidate hindsight opportunity, not a deployable selection policy or fresh confirmation',
  arms:['noquery','random','entropy','joint','disjoint','uniform-pool','best-paid-hindsight','worst-paid-hindsight','best-with-abstention-hindsight'],
  artifactHashes:hashes,scriptSHA256:sha(import.meta.filename),auditSHA256:sha(new URL('./regime-query-envelope-audit.mjs',import.meta.url)),
  oldAuditSHA256:sha(new URL('./regime-query-outcome-audit.mjs',import.meta.url)),checks,fits,branchCount,
  delayed:{records:delayed.length,meanLoss:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.losses[a]))),
    anyGain:delayed.reduce((s,r)=>s+r.anyGain,0),allHarm:delayed.reduce((s,r)=>s+r.allHarm,0)},groups,records}));
