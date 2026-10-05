import fs from 'node:fs';
import readline from 'node:readline';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/shared-v68-diagnostic';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
async function* load(p){for await(const s of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(s)}
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles,...freeze.data}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(freeze.files))assert.equal(await hash(root+'/source/'+p),h,p);
for(const[p,h]of Object.entries(freeze.generatedCompilerCopies))assert.equal(await hash(root+'/'+p),h,p);
for(const n of ['race-model','race-fixture','vet','allocation','benchmark','experiment','audit'])assert.equal(JSON.parse(fs.readFileSync(root+'/'+n+'-command.json')).exitCode,0,n);
const streams=[load(root+'/diagnostic.jsonl'),load('research/joint-v66-diagnostic/diagnostic.jsonl'),load('research/tree-v60-diagnostic/diagnostic.jsonl')];
const m=(await streams[0].next()).value;assert.equal(m.Worlds,40);assert.equal(m.SeedBase,2026105407);assert.deepEqual(m.Sources,freeze.files);
for(const s of streams.slice(1))assert.equal((await s.next()).value.Worlds,40);
const modes=['full','adaptive','hybrid_no_pair','hybrid_random','hybrid_uncertainty','hybrid_information','hybrid_falsification','hybrid_predictive','local_no_pair','local_uncertainty','noise_no_pair','noise_uncertainty','shared_no_pair','shared_uncertainty'];
const totals={},cells=[];let controls=0;
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<2e-10,`${a} != ${b}`),mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
const strip=a=>{const b=structuredClone(a);delete b.Costs;delete b.Breakdown;return b};
for(const mode of modes)totals[mode]={cells:0,risk:0,priority:0,terminalRisk:0,utility:0,maxMS:0,totalMS:0,proposalMS:0,requests:0,missing:0,observedW1Brier:0,observedW2Brier:0};
for(let n=0;n<40;n++){
 const w=(await streams[0].next()).value,old=(await streams[1].next()).value,v60=(await streams[2].next()).value;
 assert(w&&old&&v60);assert.deepEqual(w.Population,old.Population);assert.deepEqual(w.Population,v60.Population);assert.equal(w.Arms.length,42);
 const world=w.Population.World;
 for(let j=0;j<42;j++){
  const a=w.Arms[j],i=j%14,schedule=Math.floor(j/14),mode=modes[i],full=w.Arms[schedule*14],adaptive=w.Arms[schedule*14+1],single=w.Arms[schedule*14+2],oldLocal=old.Arms[schedule*8+(mode.endsWith('no_pair')?2:4)];
  assert.equal(a.Mode,mode);assert.equal(a.Schedule,old.Arms[schedule*8].Schedule);
  if(i<2){assert.deepEqual(strip(a),strip(old.Arms[schedule*8+i]));assert.deepEqual(strip(a),strip(v60.Arms[schedule*8+i]));controls++}
  assert.equal(a.Issued.length,2400);assert.equal(a.Receipts.length,2400);
  let risk=0,priority=0;
  for(let k=0;k<2400;k++){const q=a.Issued[k],r=world.Rates[Math.floor(k/150)][k%150],loss=q*q-2*q*r+r;assert(q>=0&&q<=1);risk+=loss/2400;priority+=loss*(k%150<10?3:1)/(16*170)}near(risk,a.IssuedBrier);near(priority,a.IssuedPriority);
  for(const s of a.Snapshots){assert.equal(s.Forecast.length,150);const rates=world.Rates[Math.min(15,Math.floor(s.Tick/150))];let b=0,p=0,use=0,bias=0;
   for(let k=0;k<150;k++){const q=s.Forecast[k],r=rates[k],loss=q*q-2*q*r+r;assert(q>=0&&q<=1);b+=loss/150;p+=loss*(k<10?3:1)/170}
   const order=s.Forecast.map((q,k)=>[q,k]).sort((a,b)=>b[0]-a[0]||a[1]-b[1]);for(const[q,k]of order.slice(0,10)){use+=rates[k]/10;bias+=(q-rates[k])/10}
   near(b,s.Brier);near(p,s.PriorityBrier);near(use,s.PacketUsefulness);near(bias,s.PacketBias);
  }
  let recovery=0;
  for(let z=0;z<(world.Changes??[]).length;z++){const start=world.Changes[z],end=world.Changes[z+1]??16;let delay=end-start+1,streak=0;for(const s of a.Snapshots){if(s.Tick>=2400)continue;const r=Math.floor(s.Tick/150);if(r<start||r>=end)continue;streak=s.Brier<=.2&&s.PacketUsefulness>=.75?streak+1:0;if(streak===2){delay=r-start+1;break}}recovery+=delay/world.Changes.length}near(recovery,a.Recovery);
  const audits=a.AuditOutcomes??[],paired=i>=2&&!mode.endsWith('no_pair');assert.equal(audits.length,paired?400:0);assert.equal((a.Decisions??[]).length,paired?16:0);
  const c=a.Costs;assert(c.AccountedNS<=c.ElapsedNS);
  if(i>=2){const b=a.Breakdown;assert.equal(b.AccountedNS,b.ScheduleNS+b.SetupNS+b.IssueNS+b.FirstResolveNS+b.ProposalNS+b.RequestNS+b.SecondResolveNS+b.SnapshotNS);assert.equal(b.ElapsedNS,c.ElapsedNS)}
  const t=totals[mode],last=a.Snapshots.at(-1);t.cells++;t.risk+=risk/120;t.priority+=priority/120;t.terminalRisk+=last.Brier/120;t.utility+=last.PacketUsefulness/120;t.maxMS=Math.max(t.maxMS,c.ElapsedNS/1e6);t.totalMS+=c.ElapsedNS/1e6;t.proposalMS+=(a.Breakdown.ProposalNS??0)/1e6;t.requests+=audits.length;t.missing+=audits.filter(x=>!x.Available).length;
  if(i>=2){assert.equal(a.ObservedIssued.length,2400);const receipts=new Map(a.Receipts.map(x=>[x.IssuedAt,x]));let w1=0,w2=0,count=0;
   for(let k=0;k<2400;k++){const q=a.ObservedIssued[k],y=+world.Outcomes[Math.floor(k/150)][k%150];w1+=(q-y)**2/2400;near(receipts.get(k).Forecast,q)}
   for(const z of audits){assert(!z.Expired);assert.equal(z.Available,w.Population.Available[z.Trial]);assert.equal(z.Value,z.Available?w.Population.Second[z.Trial]:false);if(z.Available){w2+=(z.Forecast-(+z.Value))**2;count++}}
   t.observedW1Brier+=w1/120;t.observedW2Brier+=count?w2/count/120:0;
  }
  cells.push({mode,geometry:world.Geometry,regime:world.Regime,schedule:a.Schedule,risk,fullGain:full.IssuedBrier-risk,adaptiveHarm:risk-adaptive.IssuedBrier,singleGain:single.IssuedBrier-risk,v66LocalGain:i>=2?oldLocal.IssuedBrier-risk:null,hasChanges:(world.Changes??[]).length>0,recoveryGainAdaptive:adaptive.Recovery-a.Recovery,elapsedMS:c.ElapsedNS/1e6});
 }
}
for(const s of streams)assert((await s.next()).done);assert.equal(controls,240);
const summaries={};for(const mode of modes.slice(2)){const c=cells.filter(x=>x.mode===mode),shift=c.filter(x=>x.hasChanges),stable=c.filter(x=>!x.hasChanges);summaries[mode]={fullGain:mean(c.map(x=>x.fullGain)),adaptiveHarm:mean(c.map(x=>x.adaptiveHarm)),harmOver01:c.filter(x=>x.adaptiveHarm>.01).length,singleGain:mean(c.map(x=>x.singleGain)),v66LocalGain:mean(c.map(x=>x.v66LocalGain)),stationaryHarm:mean(stable.map(x=>x.adaptiveHarm)),stationaryMaxHarm:Math.max(...stable.map(x=>x.adaptiveHarm)),recoveryGainAdaptive:mean(shift.map(x=>x.recoveryGainAdaptive)),over400MS:c.filter(x=>x.elapsedMS>400).length,gates:{gain01:mean(c.map(x=>x.fullGain))>=.01,noAdaptiveHarm01:c.every(x=>x.adaptiveHarm<=.01),recovery:mean(shift.map(x=>x.recoveryGainAdaptive))>0,completeLoop400:c.every(x=>x.elapsedMS<=400)},byRegime:Object.fromEntries([...new Set(c.map(x=>x.regime))].map(r=>{const z=c.filter(x=>x.regime===r);return[r,{risk:mean(z.map(x=>x.risk)),adaptiveHarm:mean(z.map(x=>x.adaptiveHarm))}]}))}}
const allocation=JSON.parse(fs.readFileSync(root+'/allocation.json')),costComparison={};
for(const against of ['hybrid_random','hybrid_uncertainty'])costComparison[against]={falsificationRiskGain:totals[against].risk-totals.hybrid_falsification.risk,falsificationCoreCostRatio:totals.hybrid_falsification.totalMS/totals[against].totalMS,predictiveRiskGain:totals[against].risk-totals.hybrid_predictive.risk,predictiveCoreCostRatio:totals.hybrid_predictive.totalMS/totals[against].totalMS};
const result={study:'shared-v68',stage:'1680-arm consumed-cohort diagnostic, n1 per cell, NOT confirmation',worlds:40,arms:1680,modelArms:1440,distinctY:96000,independentIssuedPackets:3456000,scalarIssuedComparisons:6912000,controlsBitwiseEqual:controls,populationsIdentical:40,allocation,allocationGate:Object.values(allocation.AllocatedBytes).every(x=>x<=allocation.CapBytes),totals,summaries,costComparison,cells,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/readback.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({allocation,totals,summaries:Object.fromEntries(Object.entries(summaries).map(([m,s])=>[m,{...s,byRegime:undefined}])),costComparison},null,2));
