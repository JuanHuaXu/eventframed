// Independent arithmetic/readback. No candidate fitting or cohort resampling.
import fs from 'node:fs';import crypto from 'node:crypto';import readline from 'node:readline';import assert from 'node:assert/strict';
const stage=process.argv[2];assert(['diagnostic','normal'].includes(stage));
const root='research/hybrid-v48-study-'+stage,done=JSON.parse(fs.readFileSync(root+'/completed.json')),freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));
assert(done.sourceUnchanged&&done.stage===stage);const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
for(const[p,h]of Object.entries(freeze.files))assert.equal(hash(fs.readFileSync(p)),h,p);
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80,'fresh weekly usage required');
const modes=['static','slow','round'],schedules=['immediate','fixed150','uniform299'];
const stationary=new Set(['aligned','independent','curved','baseline_matched','stationary_noise10']);
const cost=JSON.parse(fs.readFileSync(root+'/cost-report.json'));
function mean(x){return x.reduce((s,v)=>s+v/x.length,0)}
function estimate(x){const m=mean(x),se=x.length>1?Math.sqrt(x.reduce((s,v)=>s+(v-m)**2,0)/(x.length-1)/x.length):null;return{n:x.length,mean:m,se,lower:se===null?null:m-3.5*se,upper:se===null?null:m+3.5*se}}
function near(a,b){assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=3e-12,'metric arithmetic mismatch')}
function metrics(p,q){assert.equal(p.length,150);assert.equal(q.length,150);assert(q.every(v=>Number.isFinite(v)&&v>0&&v<1));const loss=q.map((v,i)=>v*v-2*v*p[i]+p[i]);const top=q.map((v,i)=>({v,i})).sort((a,b)=>b.v-a.v||a.i-b.i).slice(0,10);return{brier:mean(loss),priority:loss.reduce((s,v,i)=>s+v*(i<10?3:1)/170,0),usefulness:mean(top.map(x=>p[x.i]))}}
function derive(w,a){assert.equal(a.Issued.length,2400);const snapshots=a.Snapshots.map(s=>({...s,...metrics(w.Rates[Math.min(15,Math.floor(s.Tick/150))],s.Forecast)}));const loss=a.Issued.map((q,j)=>q*q-2*q*w.Rates[Math.floor(j/150)][j%150]+w.Rates[Math.floor(j/150)][j%150]);let recovery=0;for(let k=0;k<(w.Changes??[]).length;k++){const start=w.Changes[k],end=w.Changes[k+1]??16;let consecutive=0,delay=end-start+1;for(const s of snapshots){if(s.Tick>=2400)continue;const round=Math.floor(s.Tick/150);if(round<start||round>=end)continue;consecutive=s.brier<=.20&&s.usefulness>=.75?consecutive+1:0;if(consecutive===2){delay=round-start+1;break}}recovery+=delay/w.Changes.length}
const last=snapshots.at(-1);return{IssuedBrier:mean(loss),IssuedPriority:loss.reduce((s,v,j)=>s+v*(j%150<10?3:1)/(16*170),0),Recovery:recovery,FinalBrier:last.brier,FinalPriority:last.priority,FinalUsefulness:last.usefulness,elapsed:a.Costs?.ElapsedNS??null}}
async function* lines(file){const digest=crypto.createHash('sha256'),input=fs.createReadStream(file);input.on('data',b=>digest.update(b));for await(const line of readline.createInterface({input,crlfDelay:Infinity}))yield JSON.parse(line);assert.equal(digest.digest('hex'),done.artifacts[file.slice(root.length+1)],'raw artifact hash')}
const summaries={},seen=new Set();
for(const split of stage==='diagnostic'?['diagnostic']:['design','confirmation']){
 const fi=lines(root+'/'+split+'-fixture.jsonl')[Symbol.asyncIterator](),ri=lines(root+'/'+split+'.jsonl')[Symbol.asyncIterator]();const fm=(await fi.next()).value,rm=(await ri.next()).value;assert.equal(fm.Kind,'fixture_manifest');assert.equal(rm.Kind,'hybrid_study_manifest');assert.equal(fm.Split,split);assert.equal(rm.Split,split);assert.equal(fm.Worlds,stage==='diagnostic'?28:448);assert.equal(rm.Worlds,fm.Worlds);assert.equal(fm.SeedBase,freeze.seedBases[split]);assert.equal(rm.SeedBase,fm.SeedBase);
 const groups=new Map();let count=0,labels=0,snaps=0,maxElapsed=0;
 for(;;){const fx=await fi.next(),rx=await ri.next();assert.equal(fx.done,rx.done,'paired tape coverage');if(fx.done)break;const f=fx.value,r=rx.value,w=f.World.Population;assert.equal(r.Seed,w.Seed);assert(!seen.has(w.Seed),'actual seed collision');seen.add(w.Seed);assert.equal(f.World.Arms.length,9);assert.equal(r.Arms.length,9);count++;labels+=2400;
  for(let s=0;s<3;s++){
   const key=w.Geometry+'/'+w.Regime+'/'+schedules[s];if(!groups.has(key))groups.set(key,[]);const controls=f.World.Arms.slice(s*3,s*3+2).map(a=>{const v=derive(w,a);near(v.IssuedBrier,a.IssuedBrier);near(v.IssuedPriority,a.IssuedPriority);near(v.Recovery,a.Recovery);return v});
   const candidates=r.Arms.slice(s*3,s*3+3).map((a,j)=>{assert.equal(a.Mode,modes[j]);assert.equal(a.Schedule,schedules[s]);assert.equal(a.Advice.length,2400);assert.equal(a.Receipts.length,2400);snaps+=a.Snapshots.length;const v=derive(w,a);assert(Number.isInteger(v.elapsed)&&v.elapsed>0);maxElapsed=Math.max(maxElapsed,v.elapsed);return v});groups.get(key).push({seed:w.Seed,controls,candidates,changes:(w.Changes??[]).length});
  }
 }
 assert.equal(count,fm.Worlds);assert.equal(groups.size,84);const cells={},failed=Object.fromEntries(modes.map(m=>[m,0])),diagnostic={};
 for(const[key,rows]of groups){assert.equal(rows.length,stage==='diagnostic'?1:16);cells[key]={};diagnostic[key]={};const isStationary=stationary.has(key.split('/')[1]);const maximum=Math.max(...rows.flatMap(r=>[...r.controls,...r.candidates].map(v=>v.elapsed)));
  for(let j=0;j<3;j++){
   const checks={Work:maximum<=400000000,Allocation:cost.constructorMaxBytes<=8<<20};let pass=checks.Work&&checks.Allocation;diagnostic[key][modes[j]]={};
   for(let control=0;control<2;control++)for(const field of ['IssuedBrier','IssuedPriority','FinalBrier','FinalPriority','FinalUsefulness']){
    const gain=estimate(rows.map(r=>(r.controls[control][field]-r.candidates[j][field])*(field==='FinalUsefulness'?-1:1)));diagnostic[key][modes[j]][['full','adaptive'][control]+'/'+field]=gain;
    if(stage==='normal'){const improve=control===0&&!isStationary&&(field==='IssuedBrier'||field==='IssuedPriority');const ok=improve?gain.mean>=.01&&gain.lower>0:gain.lower>=-.01;checks[['full','adaptive'][control]+'/'+field]={gain,improve,pass:ok};pass&&=ok}
   }
   if(stage==='normal'&&rows[0].changes){const gain=estimate(rows.map(r=>r.controls[0].Recovery-r.candidates[j].Recovery)),baseline=mean(rows.map(r=>r.controls[0].Recovery)),ok=gain.lower>0&&gain.mean>=.1*baseline;checks.Recovery={gain,baseline,pass:ok};pass&&=ok}
   cells[key][modes[j]]={checks,pass:stage==='normal'?pass:null};if(stage==='normal'&&!pass)failed[modes[j]]++;
  }
 }
 summaries[split]={worlds:count,distinctLabels:labels,candidateArms:count*9,candidateSnapshots:snaps,maximumElapsedNS:maxElapsed,cells,pairedContrasts:diagnostic,failedCells:stage==='normal'?failed:null,qualityAdoptionEvaluated:stage==='normal'};
}
const report={time:new Date().toISOString(),stage,summaries,weeklyUsage:usage,sourceAndRawHashes:true,independentMetricArithmetic:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false,qualityAdoption:stage==='normal'&&modes.some(m=>Object.values(summaries).every(s=>s.failedCells[m]===0))};
const fd=fs.openSync(root+'/readback.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(report,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify({stage,summaries:Object.fromEntries(Object.entries(summaries).map(([s,r])=>[s,{worlds:r.worlds,distinctLabels:r.distinctLabels,candidateArms:r.candidateArms,candidateSnapshots:r.candidateSnapshots,maximumElapsedNS:r.maximumElapsedNS,failedCells:r.failedCells}])),qualityAdoption:report.qualityAdoption},null,2));


