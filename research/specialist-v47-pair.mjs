// Same-world diagnostic attribution; never used to refit a consumed cohort.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import readline from 'node:readline';import {execFileSync} from 'node:child_process';
const base='research/specialist-v47-study-diagnostic',root='research/specialist-v47-pair';
assert(!fs.existsSync(root),'exclusive paired root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const files={...JSON.parse(fs.readFileSync(base+'/freeze.json')).files};
for(const p of ['internal/researchswitch/local_pair_v47_test.go','research/specialist-v47-pair.mjs'])files[p]=hash(fs.readFileSync(p));
const unchanged=()=>{for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)};
unchanged();assert(JSON.parse(fs.readFileSync(base+'/artifact-audit.json')).sourceAndArtifactHashes);
fs.mkdirSync(root,{mode:0o700});
function save(p,x){const fd=fs.openSync(root+'/'+p,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
for(const[p,h]of Object.entries(files)){const b=fs.readFileSync(p);assert.equal(hash(b),h);const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,b,{flag:'wx',mode:0o600})}
save('freeze.json',{files,purpose:'n1 same-world alpha0 scope attribution; no policy fitting or adoption',worlds:28,cells:84});
let code=0,log='';const start=performance.now();
try{log=execFileSync('go',['test','./internal/researchswitch','-run','^TestLocalV47GlobalStaticPair$','-v','-count=1','-timeout=20m'],{encoding:'utf8',timeout:1300000,maxBuffer:16<<20,env:{...process.env,EVENTFRAME_SPECIALIST_V47_FIXTURE:path.resolve(base+'/diagnostic-fixture.jsonl'),EVENTFRAME_SPECIALIST_V47_PAIR:path.resolve(root+'/global-static.jsonl')}})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
fs.writeFileSync(root+'/pair.log',log,{flag:'wx',mode:0o600});
save('command.json',{code,wallMS:performance.now()-start,logSHA256:hash(log)});console.log(log);unchanged();assert.equal(code,0,'paired collection/reference');
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
async function* rows(p){for await(const s of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(s)}
const done=JSON.parse(fs.readFileSync(base+'/completed.json'));
for(const p of ['diagnostic-fixture.jsonl','diagnostic.jsonl'])assert.equal(await digest(base+'/'+p),done.artifacts[p]);
const fi=rows(base+'/diagnostic-fixture.jsonl')[Symbol.asyncIterator](),li=rows(base+'/diagnostic.jsonl')[Symbol.asyncIterator](),gi=rows(root+'/global-static.jsonl')[Symbol.asyncIterator]();
const fm=(await fi.next()).value,lm=(await li.next()).value,gm=(await gi.next()).value;
assert.equal(fm.Worlds,28);assert.equal(lm.Worlds,28);assert.equal(gm.Worlds,28);assert.equal(gm.Kind,'global_static_pair_manifest');
for(const m of [lm,gm]){assert.equal(m.SeedBase,fm.SeedBase);assert.equal(m.Split,fm.Split)}
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
function loss(w,a){assert.equal(a.Issued.length,2400);return mean(a.Issued.map((q,j)=>{assert(Number.isFinite(q)&&q>0&&q<1);const p=w.Rates[Math.floor(j/150)][j%150];return p*(1-q)**2+(1-p)*q*q}))}
const cells=[];let worlds=0;
for(;;){const f=await fi.next(),l=await li.next(),g=await gi.next();assert.equal(f.done,l.done);assert.equal(f.done,g.done);if(f.done)break;
 const w=f.value.World.Population;assert.equal(l.value.Seed,w.Seed);assert.equal(g.value.Seed,w.Seed);assert.equal(l.value.Arms.length,9);assert.equal(g.value.Arms.length,3);worlds++;
 for(let s=0;s<3;s++){
  const la=l.value.Arms[s*3],ga=g.value.Arms[s];assert.equal(la.Mode,'static');assert.equal(ga.Mode,'static');assert.equal(ga.Schedule,la.Schedule);
  assert.deepEqual(ga.Advice,la.Advice,'exact same original expert advice');assert.equal(ga.Snapshots.length,la.Snapshots.length);
  for(let j=0;j<ga.Snapshots.length;j++){assert.equal(ga.Snapshots[j].Tick,la.Snapshots[j].Tick);assert.deepEqual(ga.Snapshots[j].Advice,la.Snapshots[j].Advice,'exact same current expert advice')}
  const gl=loss(w,ga),ll=loss(w,la);cells.push({geometry:w.Geometry,regime:w.Regime,schedule:la.Schedule,seed:w.Seed,globalIssuedBrier:gl,localIssuedBrier:ll,localGain:gl-ll,globalElapsedNS:ga.Costs.ElapsedNS,localElapsedNS:la.Costs.ElapsedNS});
 }
}
assert.equal(worlds,28);assert.equal(cells.length,84);
const summary={worlds,cells:84,localPositiveCells:cells.filter(c=>c.localGain>0).length,macroLocalGain:mean(cells.map(c=>c.localGain)),nPerCell:1,intervals:null,adoption:false};
save('completed.json',{files,globalRawSHA256:await digest(root+'/global-static.jsonl'),fixtureRawSHA256:done.artifacts['diagnostic-fixture.jsonl'],localRawSHA256:done.artifacts['diagnostic.jsonl'],sourceUnchanged:true,independentGlobalReference:true,exactSharedExpertAdvice:true,summary,cells,allSevenWholeGoals:'OPEN'});
console.log(JSON.stringify(summary,null,2));
