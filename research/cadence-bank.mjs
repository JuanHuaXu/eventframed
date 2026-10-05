import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {switchBudgetCosts} from './switch-budget-oracle.mjs';
const summary=JSON.parse(fs.readFileSync(process.argv[4]));
const streams=[process.argv[2],process.argv[3]].map(p=>fs.createReadStream(p));
const hashes=streams.map(()=>crypto.createHash('sha256'));
streams.forEach((s,i)=>s.on('data',b=>hashes[i].update(b)));
const its=streams.map(s=>createInterface({input:s,crlfDelay:Infinity})[Symbol.asyncIterator]());
const header=JSON.parse((await its[0].next()).value);assert.equal(header.Version,'refit-cadence-v1');assert.equal(JSON.parse((await its[1].next()).value).Version,'soft-learners-v120');
const records=[];let identities=0;
for await(const line of its[0]) {
 const r=JSON.parse(line),next=await its[1].next();assert(!next.done);const raw=JSON.parse(next.value),prior=summary.records[records.length];assert(prior&&!r.Error);
 assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[raw.Phase,raw.Case,raw.Index,raw.Schedule]);assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[prior.phase,prior.case,prior.index,prior.schedule]);
 const metrics=[];
 for(let cadence=0;cadence<2;cadence++) {
  const ps=cadence===0?raw.Steps.map(s=>[0,1,2,3,12].map(a=>s.P[a])):r.Result.Predictions;
  const windows=[];
  for(const start of [0,192]) {
   let floor=0,hull=0,served=0,framewise=0;const costs=[];
   for(let t=start;t<256;t++) {
    const q=raw.Steps[t].Q,base=q*(1-q),experts=ps[t].slice(0,4);assert(experts.length===4&&experts.every(p=>Number.isFinite(p)&&p>0&&p<1));
    const lo=Math.min(...experts),hi=Math.max(...experts),projection=Math.max(lo,Math.min(hi,q));assert(ps[t][4]>=lo-1e-10&&ps[t][4]<=hi+1e-10);
    const row=experts.map(p=>base+(p-q)**2);costs.push(row);floor+=base;hull+=base+(projection-q)**2;served+=base+(ps[t][4]-q)**2;framewise+=Math.min(...row);
   }
   const n=256-start,dp=switchBudgetCosts(costs,2);floor/=n;hull/=n;served/=n;framewise/=n;
   const bestFixed=dp[0]/n,oneSwitch=dp[1]/n,twoSwitch=dp[2]/n,bankGap=hull-floor,selectionGap=served-hull;
   assert(floor<=hull+1e-12&&hull<=framewise+1e-12&&framewise<=twoSwitch+1e-12&&twoSwitch<=oneSwitch+1e-12&&oneSwitch<=bestFixed+1e-12);
   assert(bankGap>=-1e-12&&selectionGap>=-1e-12);assert(Math.abs(served-floor-bankGap-selectionGap)<1e-12);identities++;
   assert(Math.abs(served-prior.brier[cadence][4][start===0?0:1])<1e-12);
   windows.push({floor,hull,served,framewise,bestFixed,oneSwitch,twoSwitch,bankGap,selectionGap});
  }
  metrics.push(windows);
 }
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,metrics});
}
assert((await its[1].next()).done);assert.equal(records.length,2688);assert.equal(hashes[0].digest('hex'),summary.artifactSHA256);assert.equal(hashes[1].digest('hex'),summary.inputSHA256);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((a,x)=>a+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++) {
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
 const keys=Object.keys(rs[0].metrics[0][segment]);
 groups.push({phase,case:c,schedule,segment,banks:[0,1].map(a=>Object.fromEntries(keys.map(k=>[k,mean(rs.map(r=>r.metrics[a][segment][k]))]))),bankGapReduction:ci(rs.map(r=>r.metrics[0][segment].bankGap-r.metrics[1][segment].bankGap)),selectionGapReduction:ci(rs.map(r=>r.metrics[0][segment].selectionGap-r.metrics[1][segment].selectionGap))});
}
console.log(JSON.stringify({scope:'Consumed hindsight bank diagnostic, not deployable or learned',sourceSHA256:summary.inputSHA256,cadenceArtifactSHA256:summary.artifactSHA256,scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),identities,records,groups}));
