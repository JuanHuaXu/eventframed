import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {binomialUpper} from './paired-risk-v73.mjs';
const [file]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const [header,...rows]=fs.readFileSync(file,'utf8').trim().split('\n').map(JSON.parse);
assert.equal(header.Version,'member-coverage-v1');assert.equal(header.PerCell,512);assert.deepEqual(header.Seeds,[2026091507,2026091508]);assert.equal(rows.length,5120);
for(const [p,h]of Object.entries(header.Hashes))assert.equal(sha(p),h,p);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/(xs.length-1)/xs.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const summaries=[],comparisons=[],falseRevocations=[];let pilotPass=true;
for(const split of ['design','confirmation'])for(const scenario of ['stable','member_shift','common_shift','recurring','null']){
  const rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario);assert.equal(rs.length,512);assert.equal(new Set(rs.map(r=>r.Index)).size,512);
  const at=scenario==='recurring'?128:256,delay=t=>t<at?512-at:t-at;
  for(let i=0;i<512;i++){
    const r=rs[i];assert.equal(r.Index,i);assert.equal(r.Arms.length,5);assert.equal(r.GateFirst.length,2);assert(/^[a-f0-9]{64}$/.test(r.Tape));
    for(const t of r.GateFirst)assert(Number.isInteger(t)&&t>=-1&&t<512);
    for(const [a,arm]of r.Arms.entries()){
      assert(Number.isInteger(arm.SplitAt)&&arm.SplitAt>=-1&&arm.SplitAt<512);
      if((a===2||a===3)&&arm.SplitAt>=0){assert(r.GateFirst[a-2]>=0);assert(arm.SplitAt>=r.GateFirst[a-2]);}
      for(const [field,n]of [['Full',512],['Post',512-at]]){
        const m=arm[field];assert.equal(m.N,n);assert(Number.isFinite(m.Brier)&&m.Brier>=0&&m.Brier<=n);assert(Number.isFinite(m.LogLoss)&&m.LogLoss>=0);assert(Number.isInteger(m.Correct)&&m.Correct>=0&&m.Correct<=n);assert(Number.isInteger(m.Cost)&&m.Cost>=0&&m.Cost<=6*n);
      }
    }
    for(const f of ['MonitorCost','AuditCost','Fits'])assert(Number.isInteger(r[f])&&r[f]>=0);
  }
  for(let a=0;a<5;a++){
    const ts=rs.map(r=>r.Arms[a].SplitAt),valid=ts.filter(t=>t>=at),early=ts.filter(t=>t>=0&&t<at).length,absent=ts.filter(t=>t<0).length;
    summaries.push({split,scenario,arm:rs[0].Arms[a].Arm,n:512,splits:512-absent,early,absent,validDetections:valid.length,detectedOnlyDelay:valid.length?mean(valid.map(t=>t-at)):null,restrictedDelay:mean(ts.map(delay)),fullBrier:mean(rs.map(r=>r.Arms[a].Full.Brier/512)),postBrier:mean(rs.map(r=>r.Arms[a].Post.Brier/(512-at))),postAccuracy:mean(rs.map(r=>r.Arms[a].Post.Correct/(512-at))),foreground:mean(rs.map(r=>r.Arms[a].Full.Cost/512))});
    if([2,3].includes(a)&&['stable','common_shift','null'].includes(scenario)){
      const k=512-absent,upper=binomialUpper(k,512,.05/12);falseRevocations.push({split,scenario,arm:rs[0].Arms[a].Arm,k,n:512,alpha:.05/12,upper,pass:upper<=.02});
    }
  }
  const oldDelay=mean(rs.map(r=>delay(r.Arms[2].SplitAt))),newDelay=mean(rs.map(r=>delay(r.Arms[3].SplitAt))),early=a=>rs.filter(r=>r.Arms[a].SplitAt>=0&&r.Arms[a].SplitAt<at).length;
  const full=ci(rs.map(r=>(r.Arms[2].Full.Brier-r.Arms[3].Full.Brier)/512)),post=ci(rs.map(r=>(r.Arms[2].Post.Brier-r.Arms[3].Post.Brier)/(512-at)));
  const pass=scenario==='member_shift'?oldDelay-newDelay>=.1*oldDelay&&early(3)<=early(2)&&post.mean>=.005&&post.lower>0:full.lower>=-.01&&post.lower>=-.01;pilotPass&&=pass;
  comparisons.push({split,scenario,oldDelay,newDelay,delayGainPercent:100*(oldDelay-newDelay)/oldDelay,fullBrierGain:full,postBrierGain:post,pilotCellPass:pass,monitorCost:mean(rs.map(r=>r.MonitorCost/512)),auditCost:mean(rs.map(r=>r.AuditCost/512)),fits:mean(rs.map(r=>r.Fits))});
}
assert.equal(falseRevocations.length,12);const coveragePass=falseRevocations.every(x=>x.pass);
console.log(JSON.stringify({scope:'Fresh fixed-model finite-horizon member integration; twelve simultaneous false-revocation bounds, not real-world/anytime target-law certification',hashes:Object.fromEntries([file,import.meta.filename,'research/paired-risk-v73.mjs','docs/experiments/mmm-member-coverage-v1-protocol.md'].map(p=>[p,sha(p)])),sourceCount:Object.keys(header.Hashes).length,trajectories:rows.length,frames:rows.length*512,pilotPass,coveragePass,overall:pilotPass&&coveragePass?'PASS finite integration screen only':'FAIL full pilot criteria',falseRevocations,summaries,comparisons}));
