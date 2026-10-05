import fs from 'node:fs';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/joint-rate-v67-diagnostic';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
async function* load(p){for await(const s of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(s)}
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles,...freeze.data}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(freeze.files))assert.equal(await hash(root+'/source/'+p),h,p);
for(const n of ['race','vet','experiment','audit'])assert.equal(JSON.parse(fs.readFileSync(root+'/'+n+'-command.json')).exitCode,0,n);
const rows=load(root+'/diagnostic.jsonl'),old=load('research/joint-v66-diagnostic/diagnostic.jsonl');
const m=(await rows.next()).value;assert.equal(m.Worlds,40);assert.deepEqual(m.Sources,freeze.files);assert.equal((await old.next()).value.Worlds,40);
const names=['static','nominal','memoryless'],policies=['no_pair','uncertainty'],totals={},cells=[];let nominalEqual=0,worlds=0;
const mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
const strip=a=>{const b=structuredClone(a);delete b.Costs;delete b.Breakdown;return b};
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<2e-10,`${a} != ${b}`);
for(const name of names)for(const mode of policies)totals[name+'/'+mode]={cells:0,risk:0,terminal:0,utility:0,maxMS:0,totalMS:0,requests:0,missing:0};
for(let n=0;n<40;n++){
 const w=(await rows.next()).value,v=(await old.next()).value;assert(w&&v);assert.deepEqual(w.Population,v.Population);assert.equal(w.Arms.length,18);worlds++;
 for(let j=0;j<18;j++){
  const a=w.Arms[j],name=names[Math.floor((j%6)/2)],mode=policies[j%2],schedule=Math.floor(j/6),older=v.Arms[schedule*8+(mode==='no_pair'?2:4)],full=v.Arms[schedule*8],adaptive=v.Arms[schedule*8+1];assert.equal(a.Mode,mode);assert.equal(a.Schedule,older.Schedule);
  if(name==='nominal'){assert.deepEqual(strip(a),strip(older));nominalEqual++}
  assert.equal(a.Issued.length,2400);let risk=0;
  for(let k=0;k<2400;k++){const q=a.Issued[k],r=w.Population.World.Rates[Math.floor(k/150)][k%150];assert(q>=0&&q<=1);risk+=(q*q-2*q*r+r)/2400;if(name==='memoryless')near(q,w.Population.World.Base[k%150])}near(risk,a.IssuedBrier);
  const t=totals[name+'/'+mode],audits=a.AuditOutcomes??[],last=a.Snapshots.at(-1),c=a.Breakdown;
  assert.equal(c.AccountedNS,c.ScheduleNS+c.SetupNS+c.IssueNS+c.FirstResolveNS+c.ProposalNS+c.RequestNS+c.SecondResolveNS+c.SnapshotNS);assert(c.AccountedNS<=c.ElapsedNS);assert.equal(c.ElapsedNS,a.Costs.ElapsedNS);
  assert.equal(audits.length,mode==='no_pair'?0:400);assert(audits.every(x=>!x.Expired));t.cells++;t.risk+=risk/120;t.terminal+=last.Brier/120;t.utility+=last.PacketUsefulness/120;t.maxMS=Math.max(t.maxMS,c.ElapsedNS/1e6);t.totalMS+=c.ElapsedNS/1e6;t.requests+=audits.length;t.missing+=audits.filter(x=>!x.Available).length;
  cells.push({name,mode,geometry:w.Population.World.Geometry,regime:w.Population.World.Regime,schedule:a.Schedule,risk,nominalGain:older.IssuedBrier-risk,fullGain:full.IssuedBrier-risk,adaptiveHarm:risk-adaptive.IssuedBrier,hasChanges:(w.Population.World.Changes??[]).length>0,recoveryGain:adaptive.Recovery-a.Recovery,elapsedMS:c.ElapsedNS/1e6});
 }
}
assert((await rows.next()).done);assert((await old.next()).done);assert.equal(nominalEqual,240);
const summaries={};for(const name of names)for(const mode of policies){const c=cells.filter(x=>x.name===name&&x.mode===mode),shift=c.filter(x=>x.hasChanges),stable=c.filter(x=>!x.hasChanges);summaries[name+'/'+mode]={nominalGain:mean(c.map(x=>x.nominalGain)),nominalWins:c.filter(x=>x.nominalGain>1e-12).length,nominalLosses:c.filter(x=>x.nominalGain< -1e-12).length,fullGain:mean(c.map(x=>x.fullGain)),harmOver01:c.filter(x=>x.adaptiveHarm>.01).length,stationaryHarm:mean(stable.map(x=>x.adaptiveHarm)),recoveryGainAdaptive:mean(shift.map(x=>x.recoveryGain)),gates:{gain01:mean(c.map(x=>x.fullGain))>=.01,noAdaptiveHarm01:c.every(x=>x.adaptiveHarm<=.01),recovery:mean(shift.map(x=>x.recoveryGain))>0,completeLoop400:c.every(x=>x.elapsedMS<=400)},byRegime:Object.fromEntries([...new Set(c.map(x=>x.regime))].map(r=>{const z=c.filter(x=>x.regime===r);return[r,{risk:mean(z.map(x=>x.risk)),nominalGain:mean(z.map(x=>x.nominalGain))}]}))}}
const result={study:'joint-rate-v67',stage:'post-V66 consumed-cohort boundary ablation, NOT confirmation or adoption',worlds,arms:720,distinctY:96000,independentIssuedPackets:1728000,scalarIssuedComparisons:3456000,nominalBitwiseEqual:nominalEqual,populationsIdentical:40,totals,summaries,cells,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/readback.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({totals,summaries:Object.fromEntries(Object.entries(summaries).map(([m,x])=>[m,{...x,byRegime:undefined}]))},null,2));
