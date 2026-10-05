// Postcollection diagnosis. Preserve strict native trace failures; do not repair them.
import fs from 'node:fs';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/mean-anchor-v75-diagnostic',source='research/mean-anchor-v75-nomination-audit.mjs';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
async function* load(p){for await(const s of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(s)}
const done=JSON.parse(fs.readFileSync(root+'/completed.json')),oldDone=JSON.parse(fs.readFileSync('research/mean-joint-v74-diagnostic/completed.json')),read=JSON.parse(fs.readFileSync(root+'/readback.json'));assert(done.allJobsTerminal&&done.checks.every(x=>x.exitCode===0));
assert.equal(await hash(root+'/diagnostic.jsonl'),done.artifacts['diagnostic.jsonl']);assert.equal(await hash('research/mean-joint-v74-diagnostic/diagnostic.jsonl'),oldDone.artifacts['diagnostic.jsonl']);
const streams=[load(root+'/diagnostic.jsonl'),load('research/mean-joint-v74-diagnostic/diagnostic.jsonl')];for(const s of streams)assert.equal((await s.next()).value.Worlds,40);
const result={arms:0,issuedCleanNoisyScalarChecks:0,maxIssuedLawDefect:0,maxFinalRiskDefect:0,maxSnapshotLawDefect:0,changedOrderArms:0,changedSetArms:0,changedOrderRounds:0,changedSetRounds:0,orderOnlyRounds:0,maxQueryScoreDefectOnChangedRounds:0,maxScoreSpanAmongChangedMembers:0,commonAuditForecastChecks:0,maxCommonAuditForecastDefect:0,unmatchedAuditTrials:0,examples:[],fullStrictEquivalence:false};
const ids=a=>[...a].sort((x,y)=>x-y),same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
for(let w=0;w<40;w++){
 const x=(await streams[0].next()).value,y=(await streams[1].next()).value;assert.deepEqual(x.Population,y.Population);assert.equal(x.Arms.length,90);assert.equal(y.Arms.length,90);
 for(let j=0;j<90;j++){
  const a=x.Arms[j],b=y.Arms[j];if(!a.Mode.startsWith('mean'))continue;result.arms++;assert.equal(a.Mode,b.Mode);assert.equal(a.Schedule,b.Schedule);
  for(const field of ['Issued','ObservedIssued']){assert.equal(a[field].length,2400);for(let k=0;k<2400;k++){result.issuedCleanNoisyScalarChecks++;result.maxIssuedLawDefect=Math.max(result.maxIssuedLawDefect,Math.abs(a[field][k]-b[field][k]))}}
  result.maxFinalRiskDefect=Math.max(result.maxFinalRiskDefect,Math.abs(a.IssuedBrier-b.IssuedBrier));assert.equal(a.Snapshots.length,b.Snapshots.length);for(let s=0;s<a.Snapshots.length;s++)for(let k=0;k<150;k++)result.maxSnapshotLawDefect=Math.max(result.maxSnapshotLawDefect,Math.abs(a.Snapshots[s].Forecast[k]-b.Snapshots[s].Forecast[k]));
  assert.equal((a.Decisions??[]).length,(b.Decisions??[]).length);let orderArm=false,setArm=false;
  for(let d=0;d<(a.Decisions??[]).length;d++){
   const u=a.Decisions[d],v=b.Decisions[d];assert.equal(u.Round,v.Round);assert.equal(u.At,v.At);if(same(u.Members,v.Members))continue;
   orderArm=true;result.changedOrderRounds++;const changedSet=!same(ids(u.Members),ids(v.Members));if(changedSet){setArm=true;result.changedSetRounds++}else result.orderOnlyRounds++;
   const policy=a.Mode.endsWith('uncertainty')?'Uncertainty':'EdgeCut';assert(a.Mode.endsWith('uncertainty')||a.Mode.endsWith('falsification'));
   let defect=0;for(let k=0;k<150;k++)defect=Math.max(defect,Math.abs(u.Options[k][policy]-v.Options[k][policy]));result.maxQueryScoreDefectOnChangedRounds=Math.max(result.maxQueryScoreDefectOnChangedRounds,defect);
   const us=new Set(u.Members),vs=new Set(v.Members),swapped=[...new Set([...us,...vs])].filter(k=>us.has(k)!==vs.has(k));const scores=swapped.flatMap(k=>[u.Options[k][policy],v.Options[k][policy]]),span=scores.length?Math.max(...scores)-Math.min(...scores):0;result.maxScoreSpanAmongChangedMembers=Math.max(result.maxScoreSpanAmongChangedMembers,span);
   if(result.examples.length<12)result.examples.push({geometry:x.Population.World.Geometry,regime:x.Population.World.Regime,schedule:a.Schedule,mode:a.Mode,round:u.Round,orderOnly:!changedSet,queryScoreDefect:defect,swappedMembers:swapped,swappedScoreSpan:span,selected:u.Members,priorSelected:v.Members});
  }
  result.changedOrderArms+=+orderArm;result.changedSetArms+=+setArm;
  const oldAudits=new Map((b.AuditOutcomes??[]).map(z=>[z.Trial,z]));assert.equal(oldAudits.size,(b.AuditOutcomes??[]).length);for(const z of a.AuditOutcomes??[]){const o=oldAudits.get(z.Trial);if(!o){result.unmatchedAuditTrials++;continue}assert.equal(z.RequestedAt,o.RequestedAt);assert.equal(z.ArrivedAt,o.ArrivedAt);assert.equal(z.Available,o.Available);assert.equal(z.Value,o.Value);result.commonAuditForecastChecks++;result.maxCommonAuditForecastDefect=Math.max(result.maxCommonAuditForecastDefect,Math.abs(z.Forecast-o.Forecast))}
 }
}
for(const s of streams)assert((await s.next()).done);assert.equal(result.arms,1440);assert.equal(result.changedOrderArms,read.equivalence.selectionChangedArms);assert.equal(result.changedOrderRounds,read.equivalence.selectionChangedRounds);
result.source=source;result.sourceSHA256=await hash(source);result.dataSHA256=done.artifacts['diagnostic.jsonl'];result.priorDataSHA256=oldDone.artifacts['diagnostic.jsonl'];result.scope='postcollection diagnostic of native floating-point nominations; joined event metrics do NOT replace or erase strict recorded trace failures, not future confirmation or equalcost success';result.goals=Array(7).fill('OPEN');
fs.writeFileSync(root+'/nomination-audit.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));
