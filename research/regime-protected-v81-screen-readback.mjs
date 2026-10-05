// Independent arithmetic/journal audit, not a second inference oracle.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/regime-protected-v81-screen-initial',json=p=>JSON.parse(fs.readFileSync(p)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const done=json(root+'/completed.json');assert(done.allJobsTerminal&&done.allChecksPass&&done.sourceUnchanged);
for(const c of done.checks){assert.equal(c.exitCode,0);assert.equal(hash(fs.readFileSync(root+'/'+c.name+'.log')),c.logSHA256)}
for(const[p,h]of Object.entries(done.dataHashes))assert.equal(hash(fs.readFileSync(root+'/data/'+p)),h,p);
const summary=json(root+'/data/summary.json');assert.equal(summary.length,144);
const arms=['v80_top36','v81_protected36','v81_reset9'],aggregate={},paired=[],seen=new Set();let maxLossError=0,maxMetadataError=0,forecasts=0,attempts=0;
for(let id=0;id<48;id++){
 const x=json(root+'/data/'+`case-${String(id).padStart(3,'0')}.json`),f=x.Fixture;
 assert.equal(f.Members,50);assert.equal(f.Rounds,16);assert.equal(f.Events.length,800);
 assert([2026105501,2026105503].includes(f.Seed));assert(['stationary','abrupt','gradual','recurring'].includes(f.Generator));assert([0,.1,.2].includes(f.Noise));
 const key=JSON.stringify([f.Seed,f.Generator,f.Noise,f.Delayed]);assert(!seen.has(key));seen.add(key);
 assert.deepEqual(f.Base,Array.from({length:50},(_,i)=>[.30,.45,.65,.85][i%4]));
 const flat=[];let baseline=0;
 for(let i=0;i<800;i++){
  const e=f.Events[i],round=Math.floor(i/50);let mix=0;
  if(f.Generator==='abrupt')mix=round>=8?1:0;
  if(f.Generator==='gradual')mix=Math.max(0,Math.min(1,(round-4)/8));
  if(f.Generator==='recurring')mix=Math.floor(round/4)%2;
  const p=(1-mix)*f.Base[i%50]+mix*(1-f.Base[i%50]);
  assert.equal(e.Ordinal,i);assert.equal(e.Member,i%50);
  for(const[a,b]of [[e.Truth,p],[e.Observed,f.Noise+(1-2*f.Noise)*p]]){maxMetadataError=Math.max(maxMetadataError,Math.abs(a-b));assert(Math.abs(a-b)<2e-15)}
  assert([0,1].includes(e.First)&&[0,1].includes(e.Second));assert.equal(e.HasSecond,i%50<25);
  assert.equal(e.FirstAt,i+(f.Delayed?50:0));assert.equal(e.SecondAt,i+(f.Delayed?100:1));
  baseline+=p*(1-p)+(f.Base[e.Member]-p)**2;
 }
 for(let at=0;at<f.Packets.length;at++)for(const p of f.Packets[at]??[]){
  const e=f.Events[p.Ordinal];assert(p.Ordinal<=at);assert.equal(p.At,at);
  assert(p.Which===1||p.Which===2);assert.equal(p.Value,p.Which===1?e.First:e.Second);assert.equal(at,p.Which===1?e.FirstAt:e.SecondAt);if(p.Which===2)assert(e.HasSecond);
  flat.push(p);
 }
 assert.equal(flat.length,1200);assert.equal(flat.filter(p=>p.Which===1).length,800);assert.equal(flat.filter(p=>p.Which===2).length,400);
 const journalKeys=new Set(flat.map(p=>`${p.Ordinal}/${p.Which}`));assert.equal(journalKeys.size,1200);
 assert(Math.abs(baseline/800-x.StaticCleanBrier)<1e-14);
 assert.deepEqual(x.Results.map(r=>r.Arm).sort(),arms.slice().sort());const by={};
 for(const r of x.Results){
  by[r.Arm]=r;assert.equal(r.Issued.length,800);assert.equal(r.Operations.length,1200);assert(r.CoreMS>0&&r.AllocatedBytes>0);
  let b=0,w=0,real=0,a1=0,a2=0,j1=0,j2=0,qa=0;
  for(let i=0;i<800;i++){
   const z=r.Issued[i],e=f.Events[i];assert.equal(z.Ordinal,i);assert(z.Clean>=0&&z.Clean<=1&&z.First>=0&&z.First<=1&&/^[0-9a-f]{64}$/.test(z.LawID));
   b+=e.Truth*(1-e.Truth)+(z.Clean-e.Truth)**2;w+=e.Observed*(1-e.Observed)+(z.First-e.Observed)**2;real+=(z.First-e.First)**2;
  }
  for(let i=0;i<1200;i++){
   const op=r.Operations[i];for(const k of ['Ordinal','Which','Value','At'])assert.equal(op[k],flat[i][k]);
   assert.equal(typeof op.RevealError,'string');assert.equal(typeof op.QueryError,'string');
   if(op.Which===1){assert.equal(op.QueryError,'');op.RevealError?j1++:a1++}else{op.RevealError?j2++:a2++;if(op.QueryError)qa++}
  }
  for(const[v,stored]of [[b/800,r.CleanBrier],[w/800,r.FirstBrier],[real/800,r.RealizedFirstBrier]]){maxLossError=Math.max(maxLossError,Math.abs(v-stored));assert(Math.abs(v-stored)<2e-14)}
  assert.equal(a1,r.AcceptedFirst);assert.equal(a2,r.AcceptedSecond);assert.equal(j1,r.RejectedFirst);assert.equal(j2,r.RejectedSecond);assert.equal(qa,r.QueryAbstentions);
  const sr=summary.find(s=>s.Case===id&&s.Arm===r.Arm);assert(sr);for(const k of ['CleanBrier','FirstBrier','CoreMS','AllocatedBytes','RejectedFirst','RejectedSecond','QueryAbstentions'])assert.equal(sr[k],r[k]);
  const agg=aggregate[r.Arm]??={cases:0,clean:0,first:0,coreMS:0,maxCoreMS:0,totalAllocatedBytes:0,rejectedFirst:0,rejectedSecond:0,queryAbstentions:0};
  agg.cases++;agg.clean+=b/800;agg.first+=w/800;agg.coreMS+=r.CoreMS;agg.maxCoreMS=Math.max(agg.maxCoreMS,r.CoreMS);agg.totalAllocatedBytes+=r.AllocatedBytes;agg.rejectedFirst+=j1;agg.rejectedSecond+=j2;agg.queryAbstentions+=qa;
  forecasts+=800;attempts+=1200;
 }
 const old=by.v80_top36,protect=by.v81_protected36,reset=by.v81_reset9;
 paired.push({case:id,seed:f.Seed,generator:f.Generator,noise:f.Noise,delayed:f.Delayed,cleanGain:old.CleanBrier-protect.CleanBrier,firstGain:old.FirstBrier-protect.FirstBrier,coreRatio:protect.CoreMS/old.CoreMS,resetCleanGain:old.CleanBrier-reset.CleanBrier,staticCleanBrier:x.StaticCleanBrier});
}
assert.equal(seen.size,48);assert.equal(forecasts,115200);assert.equal(attempts,172800);
for(const a of Object.values(aggregate)){assert.equal(a.cases,48);a.meanClean=a.clean/a.cases;a.meanFirst=a.first/a.cases;a.meanCoreMS=a.coreMS/a.cases;delete a.clean;delete a.first}
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length,stationary=paired.filter(p=>p.generator==='stationary'),shifted=paired.filter(p=>p.generator!=='stationary');
const screen={maxStationaryHarm:Math.max(...stationary.map(p=>-p.cleanGain)),stationaryHarmCasesOver001:stationary.filter(p=>-p.cleanGain>.01).length,meanShiftedCleanGain:mean(shifted.map(p=>p.cleanGain)),shiftedPositiveCases:shifted.filter(p=>p.cleanGain>0).length,shiftedCases:shifted.length};
screen.rejectQualityRescueInterpretation=screen.stationaryHarmCasesOver001>0||screen.meanShiftedCleanGain<=0;
const out={time:new Date().toISOString(),sourceSHA256:hash(fs.readFileSync('research/regime-protected-v81-screen-readback.mjs')),completedSHA256:hash(fs.readFileSync(root+'/completed.json')),maxLossError,maxMetadataError,forecasts,attempts,aggregate,paired,screen,
 proofScope:'all issued losses, actual packet identities/counts/arrival metadata independently recomputed; small future-suffix fork tests; not independent replay of all inference states or broad future-leak proof',
 limitations:['development rather than untouched confirmation','small fixed generators/members, not original native Full/Adaptive gates','not equal total-cost acquisition','serial whole-arm cost, not loaded latency or peak RSS','generator true rates retained by scorer, not passed to learner'],
 goals:Array(7).fill('OPEN'),goal:'ACTIVE',scientificAdoption:false};
fs.writeFileSync(root+'/readback.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({forecasts,attempts,maxLossError,maxMetadataError,aggregate,screen},null,2));
