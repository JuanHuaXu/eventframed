import fs from 'node:fs';
import crypto from 'node:crypto';
const raw=fs.readFileSync('docs/experiments/mmm-member-integration-v80.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v80'||rows.length!==640)throw Error('incomplete integration');
for(const[p,h]of Object.entries(header.Hashes))if(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')!==h)throw Error(`changed source ${p}`);
const scenarios=['stable','member_shift','common_shift','recurring','null'];
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const interval=xs=>{const m=mean(xs),r=3.5*Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/63/64);return {mean:m,lower:m-r,upper:m+r}};
const wilson=k=>{const n=64,p=k/n,z=1.959963984540054;return (p+z*z/(2*n)+z*Math.sqrt(p*(1-p)/n+z*z/(4*n*n)))/(1+z*z/n)};
let pilotPass=true;const summaries=[],comparisons=[];
for(const split of ['design','confirmation'])for(const scenario of scenarios){
  const rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario);if(rs.length!==64)throw Error('cell count');
  const at=scenario==='recurring'?128:256;
  for(let i=0;i<64;i++){
    const r=rs[i];if(r.Index!==i||r.Arms.length!==5||r.GateFirst.length!==2||!/^[a-f0-9]{64}$/.test(r.Tape))throw Error('identity/shape');
    for(const a of r.Arms){if(!Number.isInteger(a.SplitAt)||a.SplitAt < -1||a.SplitAt>=512)throw Error('split');for(const [key,n]of [['Full',512],['Post',512-at]]){const m=a[key];if(m.N!==n||!Number.isFinite(m.Brier)||m.Brier<0||m.Brier>n||!Number.isFinite(m.LogLoss)||m.LogLoss<0||!Number.isInteger(m.Cost)||m.Cost>6*n||m.Cost<0||m.Correct<0||m.Correct>n)throw Error('metrics');}}
  }
  const delay=x=>x<at?512-at:x-at;
  for(let arm=0;arm<5;arm++){
    const splits=rs.filter(r=>r.Arms[arm].SplitAt>=0).length,early=rs.filter(r=>r.Arms[arm].SplitAt>=0&&r.Arms[arm].SplitAt<at).length;
    summaries.push({split,scenario,arm:rs[0].Arms[arm].Arm,fullBrier:mean(rs.map(r=>r.Arms[arm].Full.Brier/512)),postBrier:mean(rs.map(r=>r.Arms[arm].Post.Brier/(512-at))),postAccuracy:mean(rs.map(r=>r.Arms[arm].Post.Correct/(512-at))),foreground:mean(rs.map(r=>r.Arms[arm].Full.Cost/512)),splits,early,restrictedDelay:mean(rs.map(r=>delay(r.Arms[arm].SplitAt))),splitWilsonUpper:wilson(splits)});
  }
  const oldDelay=mean(rs.map(r=>delay(r.Arms[2].SplitAt))),newDelay=mean(rs.map(r=>delay(r.Arms[3].SplitAt)));
  const early=a=>rs.filter(r=>r.Arms[a].SplitAt>=0&&r.Arms[a].SplitAt<at).length;
  const full=interval(rs.map(r=>(r.Arms[2].Full.Brier-r.Arms[3].Full.Brier)/512));
  const post=interval(rs.map(r=>(r.Arms[2].Post.Brier-r.Arms[3].Post.Brier)/(512-at)));
  const ok=scenario==='member_shift'?oldDelay-newDelay>=.1*oldDelay&&early(3)<=early(2)&&post.mean>=.005&&post.lower>0:full.lower>=-.01&&post.lower>=-.01;
  pilotPass&&=ok;
  comparisons.push({split,scenario,oldDelay,newDelay,delayGainPercent:100*(oldDelay-newDelay)/oldDelay,fullBrierGain:full,postBrierGain:post,pilotCellPass:ok,monitorCost:mean(rs.map(r=>r.MonitorCost/512)),auditCost:mean(rs.map(r=>r.AuditCost/512)),fits:mean(rs.map(r=>r.Fits)),gateRestrictedDelay:[0,1].map(a=>mean(rs.map(r=>delay(r.GateFirst[a]))))});
}
console.log(JSON.stringify({sha256:crypto.createHash('sha256').update(raw).digest('hex'),sourceCount:Object.keys(header.Hashes).length,trajectories:rows.length,pilotPass,overall:pilotPass?'INCONCLUSIVE: 2% false-revocation requirement unestablished':'FAIL pilot improvement; 2% false-revocation requirement also unestablished',summaries,comparisons},null,2));
