import fs from 'node:fs';
import crypto from 'node:crypto';
const sha=data=>crypto.createHash('sha256').update(data).digest('hex');
const raw=fs.readFileSync('docs/experiments/mmm-age-breadth-v88.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v88'||rows.length!==2560)throw Error('incomplete artifact');
for(const[p,h]of Object.entries(header.Hashes))if(sha(fs.readFileSync(p))!==h)throw Error(`source changed ${p}`);
const configs=[['stable_uniform','stable','bit',false],['bit_uniform','member_shift','bit',false],['xor2_uniform','member_shift','xor2',false],['majority3_uniform','member_shift','majority3',false],['mux_uniform','member_shift','mux',false],['parity4_uniform','member_shift','parity4',false],['bit_dependent','member_shift','bit',true],['xor2_dependent','member_shift','xor2',true],['stable_dependent','stable','bit',true],['null_uniform','null','bit',false]];
const phases=['design','confirmation'];
const schedules=[{Name:'immediate',Delay:0,Jitter:false,Missing:0},{Name:'jitter31_missing20',Delay:0,Jitter:true,Missing:.2}];
const integer=(x,a,b)=>Number.isInteger(x)&&x>=a&&x<=b;
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const interval=xs=>{if(xs.length!==64)throw Error('cell size');const m=mean(xs),d=3.5*Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/63/64);return{mean:m,lower:m-d,upper:m+d}};
const latent=new Map(),trainSeeds=new Set();
for(let k=0;k<rows.length;k++){
 const r=rows[k],phase=Math.floor(k/1280),c=Math.floor(k%1280/128),s=Math.floor(k%128/64);
 if(r.Split!==phases[phase]||r.Scenario!==configs[c][0]||r.Index!==k%64||r.StreamBase!==2026118801+10*c+phase||r.TrainSeed!==2026118800*1e6+c*10000+phase*1000+r.Index||JSON.stringify([r.Config.Name,r.Config.Mode,r.Config.Target,r.Config.Dependent])!==JSON.stringify(configs[c])||JSON.stringify(r.Schedule)!==JSON.stringify(schedules[s]))throw Error('configuration/seed/order');
 const key=`${r.Split}/${r.Scenario}/${r.Index}`;
 if(latent.has(key)&&latent.get(key)!==r.LatentTape)throw Error('latent mismatch');
 latent.set(key,r.LatentTape);trainSeeds.add(r.TrainSeed);
 if(![r.Tape,r.LatentTape].every(h=>/^[a-f0-9]{64}$/.test(h)))throw Error('tape');
 if(!integer(r.AgeFits,0,r.SubsetFits)||!integer(r.AgeSamples,16*r.AgeFits,64*r.AgeFits))throw Error('age accounting');
 for(const name of ['Received','ReceivedAudits','Missing','PendingPackets','MaxPending','CountFits','SubsetFits','MonitorCost','AuditCost'])if(!integer(r[name],0,name.endsWith('Cost')?18*512:512))throw Error('counter');
 if(r.Received+r.Missing+r.PendingPackets!==512||r.ReceivedAudits>r.Received||r.MaxPending>64||r.CountFits!==3*r.SubsetFits||r.SubsetFits!==(r.ReceivedAudits<32?0:1+Math.floor((r.ReceivedAudits-32)/16))||r.AuditCost%18!==0||r.ReceivedAudits>r.AuditCost/18)throw Error('fit/feedback accounting');
 if(r.Arms.length!==2||r.Stats.length!==2||JSON.stringify(r.Stats[0])!==JSON.stringify(r.Stats[1])||r.Arms[0].SplitAt!==r.Arms[1].SplitAt)throw Error('unequal policies');
 const st=r.Stats[0];if(!Object.values(st).every(v=>integer(v,0,512))||st.Pending!==0||st.Applied+st.Stale!==r.Received||st.Censored!==r.Missing+r.PendingPackets)throw Error('journal');
 if(s===0&&(st.Applied!==512||st.Stale!==0||st.Censored!==0))throw Error('immediate feedback');
 for(let a=0;a<2;a++){
  const arm=r.Arms[a];if(arm.Arm!==['retained_control','age_challenger'][a]||!integer(arm.SplitAt,-1,511))throw Error('arm');
  for(const[part,n]of [['Full',512],['Post',256]]){const m=arm[part];if(m.N!==n||!Number.isFinite(m.Brier)||m.Brier<0||m.Brier>n||!Number.isFinite(m.LogLoss)||m.LogLoss<0||!integer(m.Correct,0,n)||!integer(m.Cost,0,6*n))throw Error('metric')}
 }
}
if(latent.size!==1280||trainSeeds.size!==1280)throw Error('independent fits');
let pass=true;const summaries=[],comparisons=[],versusImmediate=[];
for(const split of phases)for(const cfg of configs)for(const schedule of schedules){
 const scenario=cfg[0],rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario&&r.Schedule.Name===schedule.Name);
 const immediate=rows.filter(r=>r.Split===split&&r.Scenario===scenario&&r.Schedule.Name==='immediate');
 if(rs.length!==64||immediate.length!==64)throw Error('cell count');
 for(let a=0;a<2;a++){
  summaries.push({split,scenario,schedule:schedule.Name,arm:rs[0].Arms[a].Arm,fullBrier:mean(rs.map(r=>r.Arms[a].Full.Brier/512)),postBrier:mean(rs.map(r=>r.Arms[a].Post.Brier/256)),postAccuracy:mean(rs.map(r=>r.Arms[a].Post.Correct/256)),postLogLoss:mean(rs.map(r=>r.Arms[a].Post.LogLoss/256)),foreground:mean(rs.map(r=>r.Arms[a].Full.Cost/512))});
  versusImmediate.push({split,scenario,schedule:schedule.Name,arm:rs[0].Arms[a].Arm,postBrierIncrease:interval(rs.map((r,i)=>(r.Arms[a].Post.Brier-immediate[i].Arms[a].Post.Brier)/256))});
 }
 const fullGain=interval(rs.map(r=>(r.Arms[0].Full.Brier-r.Arms[1].Full.Brier)/512)),postGain=interval(rs.map(r=>(r.Arms[0].Post.Brier-r.Arms[1].Post.Brier)/256));
 const primary=cfg[1]==='member_shift'&&schedule.Name==='jitter31_missing20';
 const cellPass=primary?postGain.mean>=.005&&postGain.lower>0:fullGain.lower>=-.01&&postGain.lower>=-.01;
 pass&&=cellPass;
 comparisons.push({split,scenario,schedule:schedule.Name,primary,cellPass,fullGain,postGain,ageFits:mean(rs.map(r=>r.AgeFits)),ageSamples:mean(rs.map(r=>r.AgeSamples)),applied:mean(rs.map(r=>r.Stats[0].Applied)),stale:mean(rs.map(r=>r.Stats[0].Stale)),censored:mean(rs.map(r=>r.Stats[0].Censored)),maxPending:Math.max(...rs.map(r=>r.MaxPending)),auditCost:mean(rs.map(r=>r.AuditCost/512)),monitorCost:mean(rs.map(r=>r.MonitorCost/512))});
}
console.log(JSON.stringify({sha256:sha(raw),sourceCount:Object.keys(header.Hashes).length,underlyingTrajectories:latent.size,independentBaseFits:trainSeeds.size,scheduleRuns:rows.length,overall:pass?'PASS finite age breadth screen':'FAIL',summaries,comparisons,versusImmediate},null,2));
