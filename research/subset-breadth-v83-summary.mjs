import fs from 'node:fs';
import crypto from 'node:crypto';
const raw=fs.readFileSync('docs/experiments/mmm-subset-breadth-v83.jsonl');
const [header,...records]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v83'||records.length!==1280)throw Error('incomplete artifact');
for(const[p,h]of Object.entries(header.Hashes))if(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')!==h)throw Error(`source changed ${p}`);
const cases=[['stable_uniform','stable','bit',false],['bit_uniform','member_shift','bit',false],['xor2_uniform','member_shift','xor2',false],['majority3_uniform','member_shift','majority3',false],['mux_uniform','member_shift','mux',false],['parity4_uniform','member_shift','parity4',false],['bit_dependent','member_shift','bit',true],['xor2_dependent','member_shift','xor2',true],['stable_dependent','stable','bit',true],['null_uniform','null','bit',false]];
for(let k=0;k<records.length;k++){const r=records[k],phase=Math.floor(k/640),j=Math.floor((k%640)/64),c=cases[j];if(r.TrainSeed!==2026118300*1e6+j||r.StreamBase!==2026118301+10*j+phase||JSON.stringify([r.Config.Name,r.Config.Mode,r.Config.Target,r.Config.Dependent])!==JSON.stringify(c)||r.Result.Original.Split!==['design','confirmation'][phase]||r.Result.Original.Scenario!==c[0]||r.Result.Original.Index!==k%64)throw Error('configuration/seed/order');}
const rows=records.map(r=>r.Result);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const interval=xs=>{const m=mean(xs),r=3.5*Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/63/64);return {mean:m,lower:m-r,upper:m+r}};
let pass=true;const summaries=[],comparisons=[];
for(const split of ['design','confirmation'])for(const scenario of cases.map(c=>c[0])){
  const rs=rows.filter(r=>r.Original.Split===split&&r.Original.Scenario===scenario);if(rs.length!==64)throw Error('cell count');
  const at=256;
  for(let i=0;i<64;i++){
    const r=rs[i];if(r.Original.Index!==i||r.Original.Arms.length!==5||!/^[a-f0-9]{64}$/.test(r.CandidateTape)||r.SubsetFits*3!==r.Original.Fits||r.Candidate.SplitAt!==r.Original.Arms[3].SplitAt||!Number.isInteger(r.SubsetGuides)||r.SubsetGuides<0||r.SubsetGuides>512)throw Error('identity/accounting');
    for(const a of [...r.Original.Arms,r.Candidate])for(const [key,n]of [['Full',512],['Post',512-at]]){const m=a[key];if(m.N!==n||!Number.isFinite(m.Brier)||m.Brier<0||m.Brier>n||!Number.isFinite(m.LogLoss)||m.LogLoss<0||!Number.isInteger(m.Cost)||m.Cost<0||m.Cost>6*n||m.Correct<0||m.Correct>n)throw Error('metrics');}
  }
  for(let arm=0;arm<6;arm++){
    const get=r=>arm===5?r.Candidate:r.Original.Arms[arm];
    summaries.push({split,scenario,arm:get(rs[0]).Arm,fullBrier:mean(rs.map(r=>get(r).Full.Brier/512)),postBrier:mean(rs.map(r=>get(r).Post.Brier/(512-at))),postAccuracy:mean(rs.map(r=>get(r).Post.Correct/(512-at))),postLogLoss:mean(rs.map(r=>get(r).Post.LogLoss/(512-at))),foreground:mean(rs.map(r=>get(r).Full.Cost/512))});
  }
  const full=interval(rs.map(r=>(r.Original.Arms[3].Full.Brier-r.Candidate.Full.Brier)/512));
  const post=interval(rs.map(r=>(r.Original.Arms[3].Post.Brier-r.Candidate.Post.Brier)/(512-at)));
  const ok=(!scenario.startsWith('stable')&&scenario!=='null_uniform')?post.mean>=.005&&post.lower>0:full.lower>=-.01&&post.lower>=-.01;pass&&=ok;
  comparisons.push({split,scenario,fullGain:full,postGain:post,cellPass:ok,subsetGuideFraction:mean(rs.map(r=>r.SubsetGuides/512)),subsetFits:mean(rs.map(r=>r.SubsetFits)),auditCost:mean(rs.map(r=>r.Original.AuditCost/512)),monitorCost:mean(rs.map(r=>r.Original.MonitorCost/512))});
}
console.log(JSON.stringify({sha256:crypto.createHash('sha256').update(raw).digest('hex'),sourceCount:Object.keys(header.Hashes).length,trajectories:rows.length,overall:pass?'PASS finite breadth screen conditional on fitted bases':'FAIL',summaries,comparisons},null,2));

