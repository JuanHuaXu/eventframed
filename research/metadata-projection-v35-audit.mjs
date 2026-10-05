import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import readline from 'node:readline';import vm from 'node:vm';import{execFileSync}from'node:child_process';
const dir=path.resolve('research/metadata-projection-v35'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(b,x)=>{if(!b)throw Error(x)},same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
const ns=x=>{const m=x.match(/^(.*?)(?:\.(\d+))?Z$/);need(m,'UTC');return BigInt(Date.parse(m[1]+'Z'))*1000000n+BigInt((m[2]??'').padEnd(9,'0'))};
const sealed=fs.readFileSync('research/archive-cost-v34-audit.mjs','utf8');
const start=sealed.indexOf('const dir='),end=sealed.indexOf('\nfunction below(');
need(start>=0&&end>start,'sealed checker declaration boundary');
const core=sealed.slice(start,end);
const {check:base}=vm.runInNewContext(core+'\n({check})',{fs,path,crypto,vm,process},{timeout:1000});
const sorted=m=>Object.fromEntries(Object.keys(m).sort().map(k=>[k,m[k]]));
function check(row){
 const r=base(row);need([-1,3].includes(row.Mode)&&typeof row.Projected==='boolean','projection factorial');
 const batchByDigest=new Map(row.Batches.map(b=>[b.After,b]));
 let state={Head:row.FinalWitness.Base,Hash:'',Log:null,Sources:{},Journals:{},Certificate:null,Base:row.FinalWitness.Base};
 state.Hash=hash(JSON.stringify(state));
 const roots=new Map([[state.Hash,{state,seq:0}]]);
 for(const c of row.Commits){
  const tr=JSON.parse(c.Payload),sources={...state.Sources},journals={...state.Journals};
  for(const m of tr.Mutations??[])if(m.Kind===2){need(row.FinalWitness.Sources[m.AffectedKeys[0]],'unknown outcome source');sources[m.AffectedKeys[0]]=row.FinalWitness.Sources[m.AffectedKeys[0]];}
  let certificate=state.Certificate;
  const batch=batchByDigest.get(c.Digest);
  if(batch){for(const id of batch.IDs)journals[id]=row.FinalWitness.Journals[id];if(certificate===null)certificate=row.FinalWitness.Certificate;}
  const added=tr.Mutations??[];
  state={Head:tr.After,Hash:c.Digest,Log:added.length?[...(state.Log??[]),...added]:state.Log,Sources:sorted(sources),Journals:sorted(journals),Certificate:certificate,Base:state.Base};
  need(hash(JSON.stringify({...state,Hash:''}))===tr.StateHash,'independent prefix state checksum');
  roots.set(c.Digest,{state,seq:c.Seq});
 }
 need(same(state,row.FinalWitness),'complete prefix replay');
 const reads=new Map(row.Reads.map(x=>[x.Journal,x]));
 const traces=row.MetadataReads;
 need(traces.length===129&&new Set(traces.map(x=>x.Journal)).size===129&&traces.every(x=>row.FinalWitness.Journals[x.Journal]),'metadata trace completeness');
 let calls=0,success=0;
 for(const t of traces){
  const read=reads.get(t.Journal),head=read?.Snapshot??row.InitialSnapshot;
  need(t.Getters.length>=2&&t.Getters[0].Method==='GetSelectionCertificate'&&t.Getters[1].Method==='GetOmittedInfluenceCertificate','getter order');
  const actualKeys=[];
  for(const [i,g]of t.Getters.entries()){
   need(typeof g.NativeOK==='boolean','native result trace');calls++;success+=Number(g.NativeOK);
   if(i<2)need(g.Key===''&&g.NativeOK,'certificate native presence');
   else{
    need(g.Method==='GetBayesianPosterior'&&typeof g.Key==='string','getter method/key');actualKeys.push(g.Key);
    const src=row.FinalWitness.Sources[g.Key];
    need(g.NativeOK===!!(src&&src.Record.Base.runtime_version<=head.runtime_version),'native posterior presence');
   }
  }
  if(read){const expected=read.Selection&&read.Omitted?read.Decisions.filter(d=>d.activated).map(d=>d.posterior_key):[];need(same(actualKeys.sort(),expected.sort()),'actual selection-conditioned getter keys');}
  else need(actualKeys.length<=150&&t.Getters.slice(2).every(x=>!x.NativeOK),'prime source/nomination scope');
 }
 const stats=row.Metadata;
 if(row.Projected){
  need(stats.Captured===129&&stats.ProjectedGets===calls&&stats.NativeSuccess===success&&stats.LegacyOwnerResults===0,'projected counter recomputation');
  need(row.MetadataProofs.length===129&&new Set(row.MetadataProofs.map(p=>p.Journal)).size===129,'metadata projection count');
  let bytes=0;
  for(const p of row.MetadataProofs){
   bytes+=Buffer.byteLength(p.Encoded);const copy=JSON.parse(p.Encoded),root=roots.get(p.Root),read=reads.get(p.Journal);
   need(hash(p.Encoded)===p.SHA256&&copy.Hash===p.Root&&root&&same(copy.Head,p.Snapshot),'projection digest/root/head');
   const expected={Head:root.state.Head,Hash:root.state.Hash,Log:root.state.Log,Sources:root.state.Sources,Certificate:root.state.Certificate};
   need(same(copy,expected),'owned projection differs from full committed prefix');
   const b=row.Batches.find(b=>b.IDs.includes(p.Journal)),commit=row.Commits.find(c=>c.Digest===b.After);
   need(root.seq<commit.Seq&&same(row.FinalWitness.Journals[p.Journal].Snapshot,p.Snapshot),'projection ancestry/original binding');
   if(read)need(same(read.Snapshot,p.Snapshot)&&ns(read.Start)<=ns(p.CapturedAt)&&ns(p.CapturedAt)<=ns(b.Begin),'projection before handoff');
  }
  need(stats.CopiesBytes===bytes,'owned bytes recomputation');
 }else need(row.MetadataProofs===null&&stats.Captured===0&&stats.ProjectedGets===0&&stats.NativeSuccess===0&&stats.LegacyOwnerResults===success&&stats.CopiesBytes===0,'control hidden projection/counter');
 return {...r,projected:row.Projected,metadata:stats,metadata_calls:calls,metadata_native_success:success};
}
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[])}
function verify(){const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(f.files))need(hash(fs.readFileSync(p))===h,'changed source '+p);return f}
if(process.argv.includes('--freeze')){
 fs.mkdirSync(dir,{recursive:true});const checks=JSON.parse(fs.readFileSync(path.join(dir,'technical-prefreeze/checks.json')));need(!checks.changed.length&&checks.checks.every(x=>x.code===0),'technical preflight');
 const files=[...below('internal'),'go.mod','go.sum','cmd/research-metadata-projection-gen/main.go','research/durable-witness-v23-audit.mjs','research/batch-archive-v33-audit.mjs','research/archive-cost-v34-audit.mjs','research/metadata-projection-v35-audit.mjs','research/metadata-projection-v35-run.mjs','docs/experiments/mmm-metadata-projection-v35-protocol.md'].sort();
 fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]))},null,2)+'\n',{flag:'wx'});console.log('frozen',files.length);
}else{
 const f=verify(),raw=path.join(dir,'raw.ndjson'),summaries=[];let header=false,footer=false,first;
 for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})){
  if(!line)continue;const row=JSON.parse(line);
  if(row.Type==='header'){need(!header&&row.Trials===16&&row.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header');header=true;continue;}
  if(row.Type==='footer'){need(!footer&&summaries.length===16&&row.Trials===16,'footer');footer=true;continue;}
  need(header&&!footer,'envelope');summaries.push(check(row));if(row.Projected&&row.Archived&&!row.Visible&&!first)first=row;
 }
 need(header&&footer&&summaries.length===16&&new Set(summaries.map(r=>[r.trial,r.mode,r.projected,r.visible].join('/'))).size===16,'complete factorial');
 const run=JSON.parse(fs.readFileSync(path.join(dir,'run.json')));need(run.code===0&&run.raw_sha256===hash(fs.readFileSync(raw)),'raw/run');
 const recode=(p,change)=>{const s=JSON.parse(p.Encoded);change(s);p.Encoded=JSON.stringify(s);p.SHA256=hash(p.Encoded)};
 const controls=[];
 for(const[name,mutate]of[
  ['missing-write',r=>r.Writes.pop()],['future-nominee',r=>r.Reads[0].Decisions[0].event_id='future125'],['future-label',r=>r.Outcomes[0].Request.available_at='2099-01-01T00:00:00Z'],['chain-prior',r=>r.Commits[0].Prior='corrupt'],['retagged-source',r=>r.DurableSources[0].evidence_epoch++],['wire-proof',r=>r.JournalProofs[0].SHA256='bad'],['hidden-queue',r=>r.ArchiveQueue.Finished--],['premature-ack',r=>{r.Batches.find(b=>r.Reads.some(x=>b.IDs.includes(x.Journal))).End='2099-01-01T00:00:00Z'}],['capsule-wire',r=>r.Captures[0].Encoded+=' '],['timing-fabrication',r=>r.Metrics.offer_p99_ns=0],['gate-fabrication',r=>r.Pass=!r.Pass],['projection-root',r=>r.MetadataProofs[0].Root='unknown'],['projection-digest',r=>r.MetadataProofs[0].Encoded+=' '],['projection-head',r=>recode(r.MetadataProofs[0],s=>s.Head.evidence_epoch++)],['projection-source',r=>recode(r.MetadataProofs.find(p=>Object.keys(JSON.parse(p.Encoded).Sources).length),s=>{s.Sources[Object.keys(s.Sources)[0]].Posterior.alpha++})],['projection-future-log',r=>recode(r.MetadataProofs[0],s=>{s.Log=[...(s.Log??[]),r.FinalWitness.Log.at(-1)]})],['projection-late',r=>r.MetadataProofs[1].CapturedAt='2099-01-01T00:00:00Z'],['copy-byte-count',r=>r.Metadata.CopiesBytes++],['getter-counter',r=>r.Metadata.NativeSuccess++],['trace-missing',r=>r.MetadataReads.pop()],['trace-native-result',r=>r.MetadataReads[0].Getters[0].NativeOK=false],['trace-key',r=>r.MetadataReads.find(t=>t.Getters.length>2).Getters[2].Key='unknown'],['factor-relabelling',r=>r.Projected=false]
 ]){const row=structuredClone(first);mutate(row);let rejected=false;try{check(row)}catch{rejected=true};need(rejected,'corruption escaped '+name);controls.push(name)}
 verify();const report={type:'prospectively-frozen-projection-audit',time:new Date().toISOString(),raw_sha256:run.raw_sha256,source_files:Object.keys(f.files).length,controls,scope_limit:'setup-prime ack independently unbounded; its getter keys checked only for initial no-posterior/bounded count, not a saved full nomination oracle',trials:summaries,all_seven_goals:'OPEN'};
 fs.writeFileSync(path.join(dir,'audit.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({...report,trials:summaries.map(r=>({...r,freshness:{used:r.freshness.filter(x=>x.used).length,censored:r.freshness.filter(x=>!x.used).length,max_ns:Math.max(0,...r.freshness.filter(x=>x.used).map(x=>x.first_use_ns))}}))},null,2));
}
