import fs from 'node:fs';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/retention-v64-diagnostic',hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
async function load(p){const out=[];for await(const line of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))out.push(JSON.parse(line));return out}
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json')),complete=JSON.parse(fs.readFileSync(root+'/completed.json'));
assert(complete.checks.every(x=>x.exitCode===0));for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles}))assert.equal(await fileHash(p),h,p);
for(const[p,h]of Object.entries(freeze.files))assert.equal(await fileHash(root+'/source/'+p),h,p);
for(const[p,h]of Object.entries(freeze.data))assert.equal(await fileHash(p),h,p);
for(const[p,h]of Object.entries(complete.artifacts))assert.equal(await fileHash(root+'/'+p),h,p);
const rows=await load(root+'/diagnostic.jsonl'),old=await load('research/tree-v60-diagnostic/diagnostic.jsonl');assert.equal(rows.length,41);assert.equal(old.length,41);
assert.equal(rows[0].SeedBase,2026105407);assert.deepEqual(rows[0].Sources,freeze.files);
const modes=['full','adaptive','no_pair','random','uncertainty'],totals={},cells=[],stationary={},byRegime={},control=[];
const mean=x=>x.reduce((a,b)=>a+b,0)/x.length,near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<2e-10,`${a} ${b}`);
for(const m of modes){totals[m]={cells:0,risk:0,priority:0,terminalRisk:0,utility:0,maxMS:0,totalMS:0,proposalMS:0,observedW1Brier:0,observedW2Brier:0,requests:0,observed:0,missing:0,shortestExpired:0};stationary[m]=[];byRegime[m]={}}
const strip=a=>{const b=structuredClone(a);delete b.Costs;delete b.Breakdown;return b};
for(let w=1;w<rows.length;w++) {
 const p=rows[w].Population,world=p.World;assert.deepEqual(p,old[w].Population);assert.equal(rows[w].Arms.length,15);
 for(let j=0;j<15;j++) {
  const a=rows[w].Arms[j],m=a.Mode;assert.equal(m,modes[j%5]);assert.equal(a.Schedule,old[w].Arms[Math.floor(j/5)*8].Schedule);
  if(j%5<2){assert.deepEqual(strip(a),strip(old[w].Arms[Math.floor(j/5)*8+j%5]));control.push([w,j])}
  assert.equal(a.Issued.length,2400);assert.equal(a.Receipts.length,2400);let risk=0,priority=0;
  for(let k=0;k<2400;k++){const q=a.Issued[k],r=world.Rates[Math.floor(k/150)][k%150],loss=q*q-2*q*r+r;assert(q>=0&&q<=1);risk+=loss/2400;priority+=loss*(k%150<10?3:1)/(16*170)}near(risk,a.IssuedBrier);near(priority,a.IssuedPriority);
  for(const s of a.Snapshots) {const rates=world.Rates[Math.min(15,Math.floor(s.Tick/150))];let brier=0,pri=0,use=0,bias=0;
   for(let i=0;i<150;i++){const q=s.Forecast[i],r=rates[i],loss=q*q-2*q*r+r;brier+=loss/150;pri+=loss*(i<10?3:1)/170}
   const order=s.Forecast.map((q,i)=>[q,i]).sort((a,b)=>b[0]-a[0]||a[1]-b[1]);for(const[q,i]of order.slice(0,10)){use+=rates[i]/10;bias+=(q-rates[i])/10}
   near(brier,s.Brier);near(pri,s.PriorityBrier);near(use,s.PacketUsefulness);near(bias,s.PacketBias);
  }
  let recovery=0;for(let j=0;j<(world.Changes??[]).length;j++){const start=world.Changes[j],end=world.Changes[j+1]??16;let delay=end-start+1,n=0;for(const s of a.Snapshots){if(s.Tick>=2400)continue;const r=Math.floor(s.Tick/150);if(r<start||r>=end)continue;n=s.Brier<=.2&&s.PacketUsefulness>=.75?n+1:0;if(n===2){delay=r-start+1;break}}recovery+=delay/world.Changes.length}near(recovery,a.Recovery);
  const audits=a.AuditOutcomes??[],decisions=a.Decisions??[],paired=j%5>=3;assert.equal(audits.length,paired?400:0);assert.equal(decisions.length,paired?16:0);
  const c=a.Costs;assert(c.ElapsedNS>=c.AccountedNS);if(j%5>=2){const b=a.Breakdown;assert.equal(b.AccountedNS,b.ScheduleNS+b.SetupNS+b.IssueNS+b.FirstResolveNS+b.ProposalNS+b.RequestNS+b.SecondResolveNS+b.SnapshotNS);assert.equal(b.ElapsedNS,c.ElapsedNS)}
  const t=totals[m],last=a.Snapshots.at(-1);t.cells++;t.risk+=risk/120;t.priority+=priority/120;t.terminalRisk+=last.Brier/120;t.utility+=last.PacketUsefulness/120;t.maxMS=Math.max(t.maxMS,c.ElapsedNS/1e6);t.totalMS+=c.ElapsedNS/1e6;t.proposalMS+=(a.Breakdown.ProposalNS??0)/1e6;t.requests+=audits.length;t.observed+=audits.filter(x=>x.Available).length;t.missing+=audits.filter(x=>!x.Available).length;t.shortestExpired+=audits.filter(x=>x.Expired).length;
  let w1=0,w2=0,n2=0;
  if(j%5>=2){assert.equal(a.ObservedIssued.length,2400);for(let k=0;k<2400;k++){const q=a.ObservedIssued[k],y=+world.Outcomes[Math.floor(k/150)][k%150];w1+=(q-y)**2/2400;near(a.Receipts.find(x=>x.IssuedAt===k).Forecast,q)}
   for(const o of audits){assert.equal(o.Available,p.Available[o.Trial]);assert.equal(o.Value,o.Available?p.Second[o.Trial]:false);assert.equal(o.Expired,o.Trial<Math.min(o.ArrivedAt+1,2400)-600);if(o.Available){w2+=(o.Forecast-(+o.Value))**2;n2++}}
   t.observedW1Brier+=w1/120;t.observedW2Brier+=n2?w2/n2/120:0;
  }
  const full=rows[w].Arms[Math.floor(j/5)*5],adaptive=rows[w].Arms[Math.floor(j/5)*5+1],single=rows[w].Arms[Math.floor(j/5)*5+2],older=old[w].Arms[Math.floor(j/5)*8+j%5];
  const cell={geometry:world.Geometry,regime:world.Regime,schedule:a.Schedule,mode:m,risk,fullGain:full.IssuedBrier-risk,adaptiveHarm:risk-adaptive.IssuedBrier,singleGain:single.IssuedBrier-risk,v60Gain:older.IssuedBrier-risk,hasChanges:(world.Changes??[]).length>0,recoveryGainAdaptive:adaptive.Recovery-a.Recovery,recoveryGainSingle:single.Recovery-a.Recovery,elapsedMS:c.ElapsedNS/1e6};cells.push(cell);
  if(!cell.hasChanges)stationary[m].push(cell);(byRegime[m][world.Regime]??=[]).push(cell);
 }
}
const summaries={};for(const m of modes.slice(2)){const c=cells.filter(x=>x.mode===m),shift=c.filter(x=>x.hasChanges),stable=stationary[m];summaries[m]={fullGain:mean(c.map(x=>x.fullGain)),adaptiveHarm:mean(c.map(x=>x.adaptiveHarm)),harmOver01:c.filter(x=>x.adaptiveHarm>.01).length,singleGain:mean(c.map(x=>x.singleGain)),v60Gain:mean(c.map(x=>x.v60Gain)),v60Wins:c.filter(x=>x.v60Gain>1e-12).length,v60Losses:c.filter(x=>x.v60Gain< -1e-12).length,stationaryCells:stable.length,stationaryHarm:mean(stable.map(x=>x.adaptiveHarm)),stationaryMaxHarm:Math.max(...stable.map(x=>x.adaptiveHarm)),shiftCells:shift.length,recoveryGainAdaptive:mean(shift.map(x=>x.recoveryGainAdaptive)),recoveryGainSingle:mean(shift.map(x=>x.recoveryGainSingle)),over400MS:c.filter(x=>x.elapsedMS>400).length,gates:{gain01:mean(c.map(x=>x.fullGain))>=.01,noAdaptiveHarm01:c.every(x=>x.adaptiveHarm<=.01),recovery:mean(shift.map(x=>x.recoveryGainAdaptive))>0,completeLoop400:c.every(x=>x.elapsedMS<=400)},byRegime:Object.fromEntries(Object.entries(byRegime[m]).map(([r,x])=>[r,{risk:mean(x.map(y=>y.risk)),v60Gain:mean(x.map(y=>y.v60Gain)),adaptiveHarm:mean(x.map(y=>y.adaptiveHarm))}]))}}
const result={study:'retention-v64',stage:'complete 5-arm consumed-cohort diagnostic, n1percell, no adoption',worlds:40,arms:600,distinctY:96000,bankArms:360,independentIssueComparisons:864000,controlsBitwiseEqual:control.length,populationsIdentical:40,totals,summaries,cells,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
assert.equal(control.length,240);fs.writeFileSync(root+'/readback.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({totals,summaries},null,2));
