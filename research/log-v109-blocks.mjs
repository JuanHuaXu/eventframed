import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const bytes=fs.readFileSync('docs/experiments/mmm-log-v109.json');
const a=JSON.parse(bytes),hash=crypto.createHash('sha256').update(bytes).digest('hex');
assert.equal(hash,'6173c8d4c0232608876af25aaeba5e1ffbfddcf541c7c0870d51720ea03ca1ad');
const cells=[];
for(const phase of ['design','confirmation'])for(const name of ['majority_to_parity','parity_to_majority'])for(const schedule of [0,1]){
 const rs=a.Records.filter(r=>r.Phase===phase&&r.Case===name&&r.Schedule===schedule);assert.equal(rs.length,32);
 for(let block=0;block<8;block++){
  const metrics=Array.from({length:7},(_,arm)=>{
   let brier=0,logLoss=0,accuracy=0,cost=0;
   for(const r of rs)for(const s of r.Steps.slice(block*32,block*32+32)){
    const p=s.P[arm],q=s.Q;brier+=((p-q)**2+q*(1-q))/1024;
    logLoss+=(-q*Math.log(p)-(1-q)*Math.log1p(-p))/1024;
    accuracy+=(p>=.5?q:1-q)/1024;cost+=s.Cost[arm]/1024;
   }
   return{brier,logLoss,accuracy,cost};
  });
  cells.push({phase,case:name,schedule,block,start:block*32,metrics});
 }
 for(let arm=0;arm<7;arm++)for(let seg=0;seg<2;seg++){
  const z=cells.filter(c=>c.phase===phase&&c.case===name&&c.schedule===schedule&&c.block>=(seg?4:0));
  const got=z.reduce((s,c)=>s+c.metrics[arm].brier,0)/z.length;
  const want=rs.reduce((s,r)=>s+r.Metrics[arm][seg].Brier,0)/32;
  assert.ok(Math.abs(got-want)<1e-12);
 }
}
assert.equal(cells.length,64);
console.log(JSON.stringify({status:'POSTHOC_CONSUMED_DIAGNOSTIC',parentSHA256:hash,scope:'Both phases, both switches, both schedules; no filtered trajectories. Blocks are descriptive, not replacement gates.',cells},null,2));
