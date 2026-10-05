import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditEnvelope} from './regime-query-envelope-audit.mjs';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';
import {expectedQueryValue} from './regime-query-counterfactual-value.mjs';

const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function input(file){const s=fs.createReadStream(file),hash=crypto.createHash('sha256');s.on('data',b=>hash.update(b));return {it:createInterface({input:s,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=process.argv.slice(2).map(input);assert.equal(streams.length,5);
const headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
['regime-query-counterfactual-v1','soft-learners-v120','regime-query-envelope-v1','regime-query-outcome-v1','regime-query-disjoint-v1'].forEach((v,i)=>assert.equal(headers[i].Version,v));
assert.equal(headers[0].Records,2688);assert.equal(headers[0].Workers,4);
for(const [file,hash]of Object.entries(headers[0].Hashes)){const p=path.resolve('internal/observationlearners',file);assert(p.startsWith(process.cwd()+path.sep));assert.equal(sha(p),hash);}
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const records=[],ids=new Set();let checks=0,fits=0,nonredundant=0;
for(;;){
  const lines=[];for(const s of streams)lines.push(await s.it.next());if(lines.some(x=>x.done)){assert(lines.every(x=>x.done));break;}
  const [r,source,baseline,old,disjoint]=lines.map(x=>JSON.parse(x.value));assert(!r.Error);assert.deepEqual(r.Original,baseline);
  const id=`${source.Phase}:${source.Case}:${source.Index}:${source.Schedule}`;assert(!ids.has(id));ids.add(id);
  assert.deepEqual(disjoint.Original,old);auditRegimeOutcome(old,source);auditRegimeOutcome(disjoint.Candidate,source,137);
  const flipped=structuredClone(source);for(let j=152;j<160;j++)if(flipped.Steps[j].Missing||j+flipped.Steps[j].Delay>161)flipped.Steps[j].Y=!flipped.Steps[j].Y;
  const a=auditEnvelope(r.Original,source),b=auditEnvelope(r.Flipped,flipped);checks+=a.checks+b.checks;fits+=a.fits+b.fits;
  assert.deepEqual(r.Original.Choices,r.Flipped.Choices);
  const loss=branch=>mean(branch.pred.map((p,i)=>{const q=source.Steps[161+i].Q;assert(q>=0&&q<=1);return (p-q)**2+q*(1-q);}));
  const branches=a.branches.map((x,i)=>{
    const y=b.branches[i];assert.equal(x.selected,y.selected);assert.deepEqual(x.support,y.support);assert.equal(x.redundant,y.redundant);
    const actual=loss(x),counter=loss(y);
    if(x.selected<0||x.redundant){assert.deepEqual(x.pred,y.pred);assert.equal(x.logEvidence,y.logEvidence);return {origin:x.selected,value:actual,actual,counter,redundant:x.redundant};}
    nonredundant++;const s=source.Steps[x.selected],q=s.Q;assert(Number.isFinite(q)&&q>=0&&q<=1);
    return {origin:x.selected,value:expectedQueryValue(s.Y,q,actual,counter),actual,counter,q,y:s.Y,redundant:false};
  });
  for(const bundle of [old,disjoint.Candidate])for(let arm=0;arm<4;arm++){const i=r.Original.Choices.indexOf(bundle.Decision.Selected[arm]);assert(i>=0);assert.deepEqual(a.branches[i].pred,bundle.Predictions[arm]);}
  const base=branches[0].value,paid=branches.slice(1),values=paid.map(x=>x.value),v=j=>branches.find(x=>x.origin===j).value;
  const best=paid.length?Math.min(...values):base;
  records.push({phase:source.Phase,case:source.Case,index:source.Index,schedule:source.Schedule,
    values:[base,...old.Decision.Selected.slice(1).map(v),v(disjoint.Candidate.Decision.Selected[3]),paid.length?mean(values):base,best,Math.min(base,best),paid.length?Math.min(...paid.map(x=>x.actual)):base],
    anyExpectedGain:Number(best<base-1e-12),allExpectedHarm:Number(paid.length>0&&best>base+1e-12),branches});
}
assert.equal(records.length,2688);const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[1],'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
for(const i of [0,2,3,4])assert.equal(headers[i].InputSHA256,hashes[1]);assert.equal(headers[4].BaselineSHA256,hashes[3]);
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  groups.push({phase,case:c,schedule,meanValues:Array.from({length:9},(_,a)=>mean(rs.map(r=>r.values[a]))),bestExpectedGains:Array.from({length:6},(_,a)=>ci(rs.map(r=>r.values[a]-r.values[6]))),anyExpectedGain:rs.reduce((s,r)=>s+r.anyExpectedGain,0),allExpectedHarm:rs.reduce((s,r)=>s+r.allExpectedHarm,0)});
}
const delayed=records.filter(r=>r.schedule===1);
console.log(JSON.stringify({scope:'Generator-informed action values conditional on natural evidence at publication; not an online policy or fresh data',
  arms:['noquery','random-expected','entropy-expected','joint-expected','disjoint-expected','uniform-expected','best-expected-paid-oracle','best-expected-with-abstention-oracle','old-realized-answer-best'],
  artifactHashes:hashes,scriptSHA256:sha(import.meta.filename),valueSHA256:sha(new URL('./regime-query-counterfactual-value.mjs',import.meta.url)),envelopeAuditSHA256:sha(new URL('./regime-query-envelope-audit.mjs',import.meta.url)),outcomeAuditSHA256:sha(new URL('./regime-query-outcome-audit.mjs',import.meta.url)),
  checks,fits,nonredundant,delayed:{records:delayed.length,meanValues:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.values[a]))),anyExpectedGain:delayed.reduce((s,r)=>s+r.anyExpectedGain,0),allExpectedHarm:delayed.reduce((s,r)=>s+r.allExpectedHarm,0)},groups,records}));
