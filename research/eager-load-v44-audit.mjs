// Full prefix/wire authority is inherited; publication work is separately audited.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import vm from 'node:vm';import readline from 'node:readline';
import {readEnvelope} from './eager-load-v44-stream.mjs';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
const need=(x,why)=>{if(!x)throw Error(why)};
const ns=x=>{const m=x.match(/^(.*?)(?:\.(\d+))?Z$/);need(m,'UTC timestamp');return BigInt(Date.parse(m[1]+'Z'))*1000000n+BigInt((m[2]??'').padEnd(9,'0'))};
export function checkOwnerTime(p){
 if(p.Owner)need(typeof p.OwnerAt==='string'&&ns(p.Begin)<=ns(p.OwnerAt)&&ns(p.OwnerAt)<=ns(p.End),'owner acquisition outside preparation');
 else need(p.OwnerAt==='0001-01-01T00:00:00Z','hidden owner acquisition');
}
export function checkReuseTime(p,at){checkOwnerTime(p);need(p.Owner&&ns(at)<=ns(p.OwnerAt),'reused core not yet available at owner acquisition')}
const src=fs.readFileSync('research/metadata-projection-v35-audit.mjs','utf8');
const begin=src.indexOf('const dir='),end=src.indexOf('\nfunction below(');need(begin>=0&&end>begin,'sealed checker boundary');
const ret='return {...r,projected:row.Projected,metadata:stats,metadata_calls:calls,metadata_native_success:success};';
const section=src.slice(begin,end);need(section.split(ret).length===2,'unique root exposure boundary');
const {check:sealed}=vm.runInNewContext(section.replace(ret,'return {...r,projected:row.Projected,metadata:stats,metadata_calls:calls,metadata_native_success:success,readRoots:roots};')+'\n({check})',{fs,path,crypto,vm,process,Buffer},{timeout:1000});

// Binding is per operation, not a counter delta inferred during concurrent calls.
export function checkBindings(row,traces){
 const reads=new Map(row.Reads.map(r=>[r.Journal,r])),outcomes=new Map(row.Outcomes.map(o=>[o.Request.idempotency_key,o])),writes=new Map(row.Writes.map(w=>['write-'+String(w.Index).padStart(3,'0'),w]));
 const seen={journal:new Set(),outcome:new Set(),write:new Set()};
 for(const p of traces){
  need(Object.hasOwn(seen,p.Kind)&&Array.isArray(p.OperationIDs)&&p.OperationIDs.length>0,'preparation operation identity');
  need(p.Kind==='write'||p.OperationIDs.length===1,'nonbatch operation identity');
  let batchAck;
  for(const id of p.OperationIDs){
   need(typeof id==='string'&&id!==''&&!seen[p.Kind].has(id),'duplicate preparation operation');seen[p.Kind].add(id);
   if(p.Kind==='journal'){
    const b=row.Batches.find(b=>b.IDs.includes(id)),r=reads.get(id);
    need(b&&b.Error===''&&ns(b.End)<=ns(p.Begin),'preparation before journal durability');
    if(r)need(r.Error===''&&ns(r.Start)<=ns(p.Begin)&&ns(p.End)<=ns(r.End),'journal preparation outside returned ack');
    else need(row.FinalWitness.Journals[id]&&same(row.FinalWitness.Journals[id].Snapshot,row.InitialSnapshot),'unknown prime journal');
   }else if(p.Kind==='outcome'){
    const o=outcomes.get(id);need(o&&o.Error===''&&ns(o.Offer)<=ns(p.Begin)&&ns(p.End)<=ns(o.Published),'outcome preparation outside returned ack');
   }else{
    const w=writes.get(id);need(w&&w.Error===''&&ns(w.Offer)<=ns(p.Begin)&&ns(p.End)<=ns(w.Ack),'write preparation outside returned ack');
    const key=w.Ack+'/'+JSON.stringify(w.Snapshot);if(batchAck===undefined)batchAck=key;else need(batchAck===key,'preparation spans distinct write batches');
   }
  }
 }
 if(row.EagerEnabled){
  need(seen.journal.size===129&&row.Reads.every(r=>seen.journal.has(r.Journal)),'journal preparation coverage');
  need(seen.outcome.size===16&&row.Outcomes.every(o=>seen.outcome.has(o.Request.idempotency_key)),'outcome preparation coverage');
  need(seen.write.size===128&&[...writes.keys()].every(id=>seen.write.has(id)),'write preparation coverage');
 }
 return seen;
}

export function check(row){
 need(row.Projected===true&&row.CoreEnabled===true&&row.WarmEnabled===true&&typeof row.EagerEnabled==='boolean','declared V44 factors');
 const logical=structuredClone(row);logical.Metadata.CopiesBytes=row.MetadataProofs.reduce((s,p)=>s+Buffer.byteLength(p.Encoded),0);
 const result=sealed(logical),roots=result.readRoots;delete result.readRoots;
 const byHead=new Map();let certificate;
 for(const{state}of roots.values()){
  const key=JSON.stringify(state.Head),fields={Head:state.Head,Log:state.Log,Sources:state.Sources};
  if(byHead.has(key))need(same(byHead.get(key),fields),'read fields changed without head motion');else byHead.set(key,fields);
  if(state.Certificate!==null){if(certificate)need(same(certificate,state.Certificate),'certificate mutated after binding');else certificate=state.Certificate}else need(!certificate,'certificate reverted to nil');
 }
 const publisher=new Map(),traces=row.Preparations??[];let owner=0,preparationNS=0n;const kinds={journal:0,outcome:0,write:0};
 checkBindings(row,traces);
 for(const p of traces){
  need(Object.hasOwn(kinds,p.Kind)&&typeof p.Owner==='boolean'&&typeof p.Built==='boolean'&&typeof p.Reused==='boolean'&&typeof p.InitialNil==='boolean','preparation domain');kinds[p.Kind]++;
  checkOwnerTime(p);
  const duration=ns(p.End)-ns(p.Begin);need(duration>=0n,'preparation reversed');preparationNS+=duration;owner+=Number(p.Owner);
  need(Number(p.Built)+Number(p.Reused)+Number(p.InitialNil)<=1,'preparation states overlap');
  if(p.Built){
   const state=JSON.parse(p.Encoded),prefix=roots.get(p.Root);need(p.Owner&&p.Error===''&&prefix&&state.Hash===p.Root&&same(state.Head,p.Head)&&state.Certificate!==null,'publisher head/root authority');
   const expected={Head:prefix.state.Head,Hash:prefix.state.Hash,Log:prefix.state.Log,Sources:prefix.state.Sources,Certificate:prefix.state.Certificate};need(same(state,expected),'publisher differs from full committed prefix');
   need(ns(p.OwnerAt)<=ns(p.BuiltAt)&&ns(p.BuiltAt)<=ns(p.End),'publisher build time');need(!publisher.has(p.Root),'duplicate publisher root');publisher.set(p.Root,p);
  }else{
   need(p.Encoded===''&&p.BuiltAt==='0001-01-01T00:00:00Z','hidden publisher bytes');
   if(p.Reused)need(p.Owner&&p.Error===''&&roots.has(p.Root)&&same(roots.get(p.Root).state.Head,p.Head)&&roots.get(p.Root).state.Certificate!==null,'invalid reuse');
   if(p.InitialNil)need(p.Owner&&p.Error===''&&p.Root==='','invalid initial-nil preparation');
   if(p.Error==='')need(p.Reused||p.InitialNil,'successful preparation without disposition');
  }
 }
 if(!row.EagerEnabled)need(traces.length===0,'disabled hidden preparation');
 else{
  need(kinds.journal===129&&kinds.outcome===16,'complete delegated preparation kinds');
  const successfulBatches=new Set(row.Writes.filter(w=>w.Error==='').map(w=>w.Ack+'/'+JSON.stringify(w.Snapshot)));need(kinds.write===successfulBatches.size,'successful write preparation count');
 }
 const stats=row.CoreStats,builds=new Map();let physical=0;
 need(stats.Builds===row.CoreBuilds.length&&stats.Builds-publisher.size>=0&&stats.Builds-publisher.size+stats.Hits===129,'read and publisher build conservation');
 need(stats.OwnerAcquisitions>=stats.Builds-publisher.size&&stats.OwnerAcquisitions<=129,'read-side owner accounting');
 for(const b of row.CoreBuilds){
  const state=JSON.parse(b.Encoded),prefix=roots.get(b.Root);need(prefix&&state.Hash===b.Root&&!builds.has(b.Root),'core committed root');
  const expected={Head:prefix.state.Head,Hash:prefix.state.Hash,Log:prefix.state.Log,Sources:prefix.state.Sources,Certificate:prefix.state.Certificate};need(same(state,expected),'core full prefix');builds.set(b.Root,b);physical+=Buffer.byteLength(b.Encoded);
  const ps=row.MetadataProofs.filter(p=>p.Root===b.Root),pub=publisher.get(b.Root);
  need(ps.length>0||pub,'unused nonpublisher build');
  need(ps.every(p=>p.Encoded===b.Encoded&&ns(b.At)<=ns(p.CapturedAt)),'core before requests');
  if(pub)need(pub.Encoded===b.Encoded&&pub.BuiltAt===b.At,'publisher physical proof mismatch');
  if(state.Certificate===null)need(!pub&&ps.length===1,'nil certificate reused/published');
 }
 need([...publisher.keys()].every(k=>builds.has(k)),'publisher omitted physical build');
 for(const p of traces.filter(p=>p.Reused)){need(builds.has(p.Root),'reused core not physically built');checkReuseTime(p,builds.get(p.Root).At)}
 need(row.MetadataProofs.every(p=>builds.has(p.Root)),'request without physical core');
 need(stats.PhysicalBytes===physical&&row.Metadata.CopiesBytes===physical,'ALL physical bytes including unused');
 const proofs=new Map(row.MetadataProofs.map(p=>[p.Journal,p])),reads=new Map(row.Reads.map(r=>[r.Journal,r]));let fast=0;
 need(row.WarmUses.length===129&&new Set(row.WarmUses.map(u=>u.Journal)).size===129,'warm trace conservation');
 for(const u of row.WarmUses){const p=proofs.get(u.Journal),b=builds.get(u.Root),r=reads.get(u.Journal);need(p&&b&&u.Root===p.Root&&typeof u.Fast==='boolean','warm root binding');need(ns(u.CheckedAt)<=ns(p.CapturedAt)&&(!r||ns(r.Start)<=ns(u.CheckedAt)),'warm check time');if(u.Fast){need(ns(b.At)<=ns(u.CheckedAt)&&JSON.parse(b.Encoded).Certificate!==null,'warm core unavailable');fast++}}
 need(row.WarmStats.Fast===fast&&row.WarmStats.Cold===129-fast&&stats.Hits>=fast,'warm counters');
 return{...result,metadata:row.Metadata,eager:row.EagerEnabled,warm:row.WarmStats,core:stats,publication:{attempts:traces.length,owners:owner,built:publisher.size,unused:[...publisher.keys()].filter(k=>!row.MetadataProofs.some(p=>p.Root===k)).length,aggregateNS:String(preparationNS),kinds}};
}

const coreSource=fs.readFileSync('research/metadata-core-v36-audit.mjs','utf8');
const cs=coreSource.indexOf('const cs=source.indexOf('),ce=coreSource.indexOf('\nfunction corruptions(',cs);need(cs>=0&&ce>cs,'sealed mutation boundary');
const{mutations:old}=vm.runInNewContext(coreSource.slice(cs,ce)+'\n({mutations})',{source:src,need,vm,crypto},{timeout:1000});
export const mutations=[...old,
 ['missing-preparation',r=>r.Preparations.pop()],['publisher-bytes',r=>{r.Preparations.find(p=>p.Built).Encoded+=' '}],['publisher-root',r=>{r.Preparations.find(p=>p.Built).Root='unknown'}],['publisher-time',r=>{r.Preparations.find(p=>p.Built).BuiltAt='2099-01-01T00:00:00Z'}],['owner-domain',r=>{r.Preparations.find(p=>p.Built).Owner=false}],['overlapping-dispositions',r=>{r.Preparations.find(p=>p.Built).Reused=true}],['publisher-factor',r=>r.EagerEnabled=false],
 ['operation-id',r=>r.Preparations[0].OperationIDs=['unknown']],['operation-id-duplicate',r=>r.Preparations.push(structuredClone(r.Preparations[0]))],['preparation-after-ack',r=>{const p=r.Preparations.find(p=>p.Kind==='outcome');p.Begin=p.End='2099-01-01T00:00:00Z'}],['preparation-before-durability',r=>{const p=r.Preparations.find(p=>p.Kind==='journal');p.Begin=p.End='2000-01-01T00:00:00Z'}],['missing-write-id',r=>r.Preparations.find(p=>p.Kind==='write').OperationIDs.pop()],
 ['missing-warm-use',r=>r.WarmUses.pop()],['warm-root',r=>r.WarmUses[0].Root='unknown'],['warm-counter',r=>r.WarmStats.Fast++],['warm-late-check',r=>r.WarmUses[0].CheckedAt='2099-01-01T00:00:00Z'],['warm-first-nil',r=>{r.WarmUses[0].Fast=true;r.WarmStats.Fast++;r.WarmStats.Cold--}],['warm-factor',r=>r.WarmEnabled=false],
 ['owner-time-missing',r=>{delete r.Preparations.find(p=>p.Owner).OwnerAt}],['owner-time-before-call',r=>{r.Preparations.find(p=>p.Owner).OwnerAt='2000-01-01T00:00:00Z'}],['owner-time-after-return',r=>{r.Preparations.find(p=>p.Owner).OwnerAt='2099-01-01T00:00:00Z'}]
];
export function corruptions(first){
 const controls=[];for(const[name,mutate]of mutations){const r=structuredClone(first);mutate(r);let rejected=false;try{check(r)}catch{rejected=true};need(rejected,'corruption escaped '+name);controls.push(name)}return controls;
}

// This fixture tests auditor logic, not performance or new experimental results.
export function preflightFixture(consumed){
 const r=structuredClone(consumed);r.EagerEnabled=true;r.Preparations=[];
 const trace=(Kind,OperationIDs,at)=>({Kind,OperationIDs,Begin:at,End:at,OwnerAt:'0001-01-01T00:00:00Z',BuiltAt:'0001-01-01T00:00:00Z',Owner:false,Built:false,Reused:false,InitialNil:false,Head:{},Root:'',Encoded:'',Error:'synthetic auditor-preflight optional failure'});
 const reads=new Map(r.Reads.map(x=>[x.Journal,x]));
 for(const b of r.Batches)for(const id of b.IDs)r.Preparations.push(trace('journal',[id],reads.get(id)?.End??b.End));
 for(const o of r.Outcomes)r.Preparations.push(trace('outcome',[o.Request.idempotency_key],o.Published));
 const batches=new Map();for(const w of r.Writes){const key=w.Ack+'/'+JSON.stringify(w.Snapshot);if(!batches.has(key))batches.set(key,trace('write',[],w.Ack));batches.get(key).OperationIDs.push('write-'+String(w.Index).padStart(3,'0'))}r.Preparations.push(...batches.values());
 const p=r.Preparations.reduce((a,b)=>ns(a.End)>ns(b.End)?a:b),s=r.FinalWitness;
 need(!r.CoreBuilds.some(b=>b.Root===s.Hash),'synthetic unused publisher root already built');
 const Encoded=JSON.stringify({Head:s.Head,Hash:s.Hash,Log:s.Log,Sources:s.Sources,Certificate:s.Certificate});
 Object.assign(p,{Owner:true,OwnerAt:p.End,Built:true,Head:s.Head,Root:s.Hash,Encoded,BuiltAt:p.End,Error:''});
 r.CoreBuilds.push({Root:s.Hash,Encoded,At:p.End});r.CoreStats.Builds++;r.CoreStats.PhysicalBytes+=Buffer.byteLength(Encoded);r.Metadata.CopiesBytes+=Buffer.byteLength(Encoded);
 return r;
}

// Run as a library from the frozen runner. No new outputs are overwritten here.
if(process.argv[1]&&path.resolve(process.argv[1])===path.resolve('research/eager-load-v44-audit.mjs')){
 if(process.argv[2]==='--self-test'){
  let consumed;for await(const line of readline.createInterface({input:fs.createReadStream('research/warm-search-v37/raw.ndjson'),crlfDelay:Infinity})){const r=JSON.parse(line);if(r.WarmEnabled&&r.Archived&&!r.Visible){consumed=r;break}}
  need(consumed,'consumed V37 fixture missing');const off=structuredClone(consumed);off.EagerEnabled=false;off.Preparations=[];check(off);
  const fixture=preflightFixture(consumed),positive=check(fixture),controls=corruptions(fixture);
  console.log(JSON.stringify({time:new Date().toISOString(),type:'auditor-logic-only-before-V44-normal',fixture:'consumed V37 with explicitly synthetic preparation traces and one unused publisher; NOT a new load result',controls,unusedPublisherValidated:positive.publication.unused,wholeGoalPassUnproven:true},null,2));
 }else{
 const input=process.argv[2];need(input,'explicit isolated raw file');
 const summaries=[],factors=new Set();let controls;
 const envelope=await readEnvelope(input,row=>{
  summaries.push(check(row));const factor=[row.Trial,row.Visible,row.Archived,row.EagerEnabled].join('/');need(!factors.has(factor),'duplicate factorial trial');factors.add(factor);
  if(!controls&&row.EagerEnabled&&row.Archived&&!row.Visible)controls=corruptions(row);
 });
 need(factors.size===16&&controls,'complete factorial and positive experimental arm');
 console.log(JSON.stringify({time:new Date().toISOString(),rawSHA256:envelope.rawSHA256,controls,trials:summaries,allSevenWholeGoals:'OPEN'},null,2));
 }
}
