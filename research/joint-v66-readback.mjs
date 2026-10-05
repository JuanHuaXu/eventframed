import fs from 'node:fs';
import readline from 'node:readline';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/joint-v66-diagnostic';
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
async function* load(p){for await(const s of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(s)}
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles,...freeze.data}))assert.equal(await digest(p),h,p);
for(const[p,h]of Object.entries(freeze.files))assert.equal(await digest(root+'/source/'+p),h,p);
for(const[p,h]of Object.entries(freeze.generatedCompilerCopies))assert.equal(await digest(root+'/'+p),h,p);
for(const n of ['race-model','race-fixture','vet','allocation','benchmark','experiment','audit'])assert.equal(JSON.parse(fs.readFileSync(root+'/'+n+'-command.json')).exitCode,0,n);
const current=load(root+'/diagnostic.jsonl'),old=load('research/tree-v60-diagnostic/diagnostic.jsonl'),v64=load('research/retention-v64b-diagnostic/diagnostic.jsonl');
const manifest=(await current.next()).value;
assert.equal(manifest.Worlds,40);assert.equal(manifest.SeedBase,2026105407);assert.deepEqual(manifest.Sources,freeze.files);
for(const r of [old,v64])assert.equal((await r.next()).value.Worlds,40);
const modes=['full','adaptive','no_pair','random','uncertainty','information','falsification','predictive'];
const totals={},cells=[];let controls=0,worlds=0;
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<2e-10,`${a} != ${b}`);
const mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
const strip=a=>{const b=structuredClone(a);delete b.Costs;delete b.Breakdown;return b};
for(const m of modes)totals[m]={cells:0,risk:0,priority:0,terminalRisk:0,utility:0,maxMS:0,totalMS:0,proposalMS:0,requests:0,observed:0,missing:0,observedW1Brier:0,observedW2Brier:0};
for(let w=0;w<40;w++){
 const x=(await current.next()).value,o=(await old.next()).value,b=(await v64.next()).value;
 assert(x&&o&&b);assert.deepEqual(x.Population,o.Population);assert.deepEqual(x.Population,b.Population);assert.equal(x.Arms.length,24);worlds++;
 const p=x.Population,world=p.World;
 for(let j=0;j<24;j++){
  const a=x.Arms[j],mode=modes[j%8],full=x.Arms[Math.floor(j/8)*8],adaptive=x.Arms[Math.floor(j/8)*8+1],single=x.Arms[Math.floor(j/8)*8+2];
  assert.equal(a.Mode,mode);assert.equal(a.Schedule,o.Arms[j].Schedule);
  if(j%8<2){assert.deepEqual(strip(a),strip(o.Arms[j]));assert.deepEqual(strip(a),strip(b.Arms[Math.floor(j/8)*5+j%8]));controls++}
  assert.equal(a.Issued.length,2400);assert.equal(a.Receipts.length,2400);
  let risk=0,priority=0;
  for(let k=0;k<2400;k++){const q=a.Issued[k],r=world.Rates[Math.floor(k/150)][k%150],loss=q*q-2*q*r+r;assert(q>=0&&q<=1);risk+=loss/2400;priority+=loss*(k%150<10?3:1)/(16*170)}near(risk,a.IssuedBrier);near(priority,a.IssuedPriority);
  for(const s of a.Snapshots){assert.equal(s.Forecast.length,150);const r=world.Rates[Math.min(15,Math.floor(s.Tick/150))];let loss=0,pri=0,use=0,bias=0;
   for(let i=0;i<150;i++){const q=s.Forecast[i],v=q*q-2*q*r[i]+r[i];assert(q>=0&&q<=1);loss+=v/150;pri+=v*(i<10?3:1)/170}
   const order=s.Forecast.map((q,i)=>[q,i]).sort((a,b)=>b[0]-a[0]||a[1]-b[1]);for(const[q,i]of order.slice(0,10)){use+=r[i]/10;bias+=(q-r[i])/10}
   near(loss,s.Brier);near(pri,s.PriorityBrier);near(use,s.PacketUsefulness);near(bias,s.PacketBias);
  }
  let recovery=0;
  for(let z=0;z<(world.Changes??[]).length;z++){const start=world.Changes[z],end=world.Changes[z+1]??16;let delay=end-start+1,n=0;for(const s of a.Snapshots){if(s.Tick>=2400)continue;const r=Math.floor(s.Tick/150);if(r<start||r>=end)continue;n=s.Brier<=.2&&s.PacketUsefulness>=.75?n+1:0;if(n===2){delay=r-start+1;break}}recovery+=delay/world.Changes.length}near(recovery,a.Recovery);
  const audits=a.AuditOutcomes??[],decisions=a.Decisions??[],paired=j%8>=3;
  assert.equal(audits.length,paired?400:0);assert.equal(decisions.length,paired?16:0);
  const cost=a.Costs;assert(cost.AccountedNS<=cost.ElapsedNS);
  if(j%8>=2){const c=a.Breakdown;assert.equal(c.AccountedNS,c.ScheduleNS+c.SetupNS+c.IssueNS+c.FirstResolveNS+c.ProposalNS+c.RequestNS+c.SecondResolveNS+c.SnapshotNS);assert.equal(c.ElapsedNS,cost.ElapsedNS)}
  const t=totals[mode],last=a.Snapshots.at(-1);t.cells++;t.risk+=risk/120;t.priority+=priority/120;t.terminalRisk+=last.Brier/120;t.utility+=last.PacketUsefulness/120;t.maxMS=Math.max(t.maxMS,cost.ElapsedNS/1e6);t.totalMS+=cost.ElapsedNS/1e6;t.proposalMS+=(a.Breakdown.ProposalNS??0)/1e6;t.requests+=audits.length;t.observed+=audits.filter(x=>x.Available).length;t.missing+=audits.filter(x=>!x.Available).length;
  if(j%8>=2){assert.equal(a.ObservedIssued.length,2400);const receipts=new Map(a.Receipts.map(x=>[x.IssuedAt,x]));let w1=0,w2=0,n2=0;
   for(let k=0;k<2400;k++){const q=a.ObservedIssued[k],y=+world.Outcomes[Math.floor(k/150)][k%150];w1+=(q-y)**2/2400;near(receipts.get(k).Forecast,q)}
   for(const z of audits){assert.equal(z.Expired,false);assert.equal(z.Available,p.Available[z.Trial]);assert.equal(z.Value,z.Available?p.Second[z.Trial]:false);if(z.Available){w2+=(z.Forecast-(+z.Value))**2;n2++}}
   t.observedW1Brier+=w1/120;t.observedW2Brier+=n2?w2/n2/120:0;
  }
  const prev64=j%8<5?b.Arms[Math.floor(j/8)*5+j%8]:null;
  cells.push({geometry:world.Geometry,regime:world.Regime,schedule:a.Schedule,mode,risk,fullGain:full.IssuedBrier-risk,adaptiveHarm:risk-adaptive.IssuedBrier,singleGain:single.IssuedBrier-risk,v60Gain:o.Arms[j].IssuedBrier-risk,v64Gain:prev64?prev64.IssuedBrier-risk:null,hasChanges:(world.Changes??[]).length>0,recoveryGainAdaptive:adaptive.Recovery-a.Recovery,recoveryGainSingle:single.Recovery-a.Recovery,elapsedMS:cost.ElapsedNS/1e6});
 }
}
for(const r of [current,old,v64])assert((await r.next()).done);
assert.equal(controls,240);assert.equal(worlds,40);
const summaries={};
for(const m of modes.slice(2)){const c=cells.filter(x=>x.mode===m),shift=c.filter(x=>x.hasChanges),stable=c.filter(x=>!x.hasChanges),prior=c.filter(x=>x.v64Gain!==null);summaries[m]={fullGain:mean(c.map(x=>x.fullGain)),adaptiveHarm:mean(c.map(x=>x.adaptiveHarm)),harmOver01:c.filter(x=>x.adaptiveHarm>.01).length,singleGain:mean(c.map(x=>x.singleGain)),v60Gain:mean(c.map(x=>x.v60Gain)),v64Gain:prior.length?mean(prior.map(x=>x.v64Gain)):null,v60Wins:c.filter(x=>x.v60Gain>1e-12).length,v60Losses:c.filter(x=>x.v60Gain< -1e-12).length,stationaryCells:stable.length,stationaryHarm:mean(stable.map(x=>x.adaptiveHarm)),stationaryMaxHarm:Math.max(...stable.map(x=>x.adaptiveHarm)),recoveryGainAdaptive:mean(shift.map(x=>x.recoveryGainAdaptive)),recoveryGainSingle:mean(shift.map(x=>x.recoveryGainSingle)),over400MS:c.filter(x=>x.elapsedMS>400).length,gates:{gain01:mean(c.map(x=>x.fullGain))>=.01,noAdaptiveHarm01:c.every(x=>x.adaptiveHarm<=.01),recovery:mean(shift.map(x=>x.recoveryGainAdaptive))>0,completeLoop400:c.every(x=>x.elapsedMS<=400)},byRegime:Object.fromEntries([...new Set(c.map(x=>x.regime))].map(r=>{const z=c.filter(x=>x.regime===r);return[r,{risk:mean(z.map(x=>x.risk)),adaptiveHarm:mean(z.map(x=>x.adaptiveHarm)),v60Gain:mean(z.map(x=>x.v60Gain))}]}))}}
const allocation=JSON.parse(fs.readFileSync(root+'/allocation.json'));
const result={study:'joint-v66',stage:'full eight-arm consumed-cohort diagnostic, n1 per cell, NOT confirmation',worlds,arms:960,distinctY:96000,modelArms:720,independentIssuedPackets:1728000,scalarIssuedComparisons:3456000,controlsBitwiseEqual:controls,populationsIdentical:40,allocation,allocationGate:Object.values(allocation.AllocatedBytes).every(x=>x<=allocation.CapBytes),totals,summaries,cells,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/readback.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({allocation,totals,summaries},null,2));
