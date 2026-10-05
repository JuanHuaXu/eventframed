// Independent streaming exact readback and paired cost calculation.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import readline from 'node:readline';
const root='research/memo-v49-ablation';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const f=JSON.parse(fs.readFileSync(root+'/freeze.json')),d=JSON.parse(fs.readFileSync(root+'/completed.json'));
assert(d.sourceUnchanged&&d.referenceUnchanged&&d.checks.length===5&&d.checks.every(c=>c.code===0));
for(const[p,h]of Object.entries(f.files)){assert.equal(hash(fs.readFileSync(p)),h,p);assert.equal(hash(fs.readFileSync(root+'/source/'+p)),h,'frozen '+p)}
for(const[p,h]of Object.entries(d.artifacts))assert.equal(await digest(root+'/'+p),h,'artifact '+p);
for(const[p,h]of Object.entries(f.reference))assert.equal(await digest(f.referenceRoot+'/'+p),h,'reference '+p);
for(const c of d.checks){assert.deepEqual(JSON.parse(fs.readFileSync(root+'/'+c.name+'-command.json')),c);assert.equal(await digest(root+'/'+c.name+'.log'),c.logSHA256)}
function cost(c){for(const x of Object.values(c))assert(Number.isSafeInteger(x)&&x>0);assert.equal(c.AccountedNS,c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS);assert(c.ElapsedNS>=c.AccountedNS)}
function* fields(a){for(const[k,v]of Object.entries(a))if(k!=='Costs')yield[k,v]}
const report=JSON.parse(fs.readFileSync(root+'/cost-report.json'));
assert(report.exactPairEquality&&!report.adoption);
assert.equal(report.loops.length,9);const screenKeys=new Set();
const screen={};for(const name of ['original','memo']){
 const bytes=report.constructorMaxBytes[name];assert(Number.isSafeInteger(bytes)&&bytes>0);
 screen[name]={constructorMaxBytes:bytes,allocationPass:bytes<=8<<20,maximumLoopNS:0,loopPass:true};
}
for(const[j,r]of report.loops.entries()){
 assert.equal(r.mode,['static','slow','round'][j%3]);assert.equal(r.schedule,['immediate','fixed150','uniform299'][Math.floor(j/3)]);assert.equal(r.memoFirst,j%2!==0);
 assert(!screenKeys.has(r.mode+'/'+r.schedule));screenKeys.add(r.mode+'/'+r.schedule);
 for(const name of ['original','memo']){cost(r[name]);screen[name].maximumLoopNS=Math.max(screen[name].maximumLoopNS,r[name].ElapsedNS);screen[name].loopPass&&=r[name].ElapsedNS<=400000000}
}
function lines(p){return readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity})[Symbol.asyncIterator]()}
const streams=[lines(root+'/original.jsonl'),lines(root+'/memo.jsonl'),lines(f.referenceRoot+'/diagnostic.jsonl'),lines(f.referenceRoot+'/diagnostic-fixture.jsonl')];
const heads=await Promise.all(streams.map(s=>s.next()));assert(heads.every(x=>!x.done));const manifests=heads.map(x=>JSON.parse(x.value));assert.deepEqual(manifests[0],manifests[1]);assert.deepEqual(manifests[0],manifests[2]);manifests[3].Kind='hybrid_study_manifest';assert.deepEqual(manifests[0],manifests[3]);
const paired=[],seeds=new Set(),cells=new Set();let worlds=0,snapshots=0,labels=0;
while(true){const rows=await Promise.all(streams.map(s=>s.next()));if(rows.every(x=>x.done))break;assert(rows.every(x=>!x.done),'unequal tapes');const[a,b,old,fixture]=rows.map(x=>JSON.parse(x.value));assert.equal(a.Seed,b.Seed);assert.equal(a.Seed,old.Seed);assert.equal(a.Seed,fixture.World.Population.Seed);assert(!seeds.has(a.Seed));seeds.add(a.Seed);assert.equal(a.Arms.length,9);assert.equal(b.Arms.length,9);assert.equal(old.Arms.length,9);
 labels+=fixture.World.Population.Outcomes.flat().length;
 for(let j=0;j<9;j++){
  const x=a.Arms[j],y=b.Arms[j],z=old.Arms[j];assert.deepEqual(Object.fromEntries(fields(x)),Object.fromEntries(fields(y)),'candidate differs');assert.deepEqual(Object.fromEntries(fields(x)),Object.fromEntries(fields(z)),'original differs');cost(x.Costs);cost(y.Costs);
  assert.equal(x.Mode,['static','slow','round'][j%3]);assert.equal(x.Schedule,['immediate','fixed150','uniform299'][Math.floor(j/3)]);assert.equal(x.Issued.length,2400);assert.equal(x.Receipts.length,2400);
  if(j%3===0)cells.add(fixture.World.Population.Geometry+'/'+fixture.World.Population.Regime+'/'+x.Schedule);
  snapshots+=x.Snapshots.length;
  paired.push({geometry:fixture.World.Population.Geometry,regime:fixture.World.Population.Regime,mode:x.Mode,schedule:x.Schedule,memoFirst:(worlds*9+j)%2!==0,original:x.Costs,memo:y.Costs});
 }
 worlds++;
}
assert.equal(worlds,28);assert.equal(worlds,manifests[0].Worlds);assert.equal(cells.size,84);assert.equal(snapshots,4200);assert.equal(labels,67200);assert.equal(paired.length,252);
const total={};for(const name of ['original','memo']){const elapsed=paired.map(r=>r[name].ElapsedNS),setup=paired.map(r=>r[name].SetupNS);total[name]={maximumLoopNS:Math.max(...elapsed),loopPass:elapsed.every(x=>x<=400000000),meanLoopNS:elapsed.reduce((a,b)=>a+b,0)/elapsed.length,meanSetupNS:setup.reduce((a,b)=>a+b,0)/setup.length}}
const out={time:new Date().toISOString(),worlds,cells:cells.size,pairedArms:paired.length,snapshotsPerVariant:snapshots,distinctLabels:labels,consumedData:true,exactAllFieldsExceptCost:true,sourceCopies:Object.keys(f.files).length,commands:d.checks.length,screen,total,paired,qualityConfirmation:false,adoption:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
const fd=fs.openSync(root+'/readback.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(out,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify({...out,paired:undefined},null,2));
