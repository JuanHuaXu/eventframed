import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import vm from 'node:vm';
import {execFileSync} from 'node:child_process';
const dir=path.resolve('research/metadata-core-v36');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(ok,why)=>{if(!ok)throw Error(why)},same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
const ns=x=>{const m=x.match(/^(.*?)(?:\.(\d+))?Z$/);need(m,'UTC');return BigInt(Date.parse(m[1]+'Z'))*1000000n+BigInt((m[2]??'').padEnd(9,'0'))};
const source=fs.readFileSync('research/metadata-projection-v35-audit.mjs','utf8');
const start=source.indexOf('const dir='),end=source.indexOf('\nfunction below(');
need(start>=0&&end>start,'V35 checker boundary');
const ret='return {...r,projected:row.Projected,metadata:stats,metadata_calls:calls,metadata_native_success:success};';
let core=source.slice(start,end);need(core.split(ret).length===2,'root exposure boundary');
core=core.replace(ret,'return {...r,projected:row.Projected,metadata:stats,metadata_calls:calls,metadata_native_success:success,readRoots:roots};');
const {check:prior}=vm.runInNewContext(core+'\n({check})',{fs,path,crypto,vm,process,Buffer},{timeout:1000});
function check(row){
  need(row.Projected===true&&typeof row.CoreEnabled==='boolean','core factor');
  const copy=structuredClone(row);
  // Legacy checker independently verifies all logical request projections. Its
  // old per-request byte equality is replaced only in this view, then actual
  // physical bytes are independently checked against the build trace below.
  if(row.CoreEnabled)copy.Metadata.CopiesBytes=row.MetadataProofs.reduce((n,p)=>n+Buffer.byteLength(p.Encoded),0);
  const r=prior(copy),roots=r.readRoots;delete r.readRoots;
  const byHead=new Map();let certificate;
  for(const {state}of roots.values()){
    const key=JSON.stringify(state.Head), fields={Head:state.Head,Log:state.Log,Sources:state.Sources};
    if(byHead.has(key))need(same(byHead.get(key),fields),'read fields changed without head motion');else byHead.set(key,fields);
    if(state.Certificate!==null){if(certificate)need(same(certificate,state.Certificate),'certificate mutated after first binding');else certificate=state.Certificate;}
    else need(!certificate,'certificate reverted to nil');
  }
  const s=row.CoreStats;
  if(!row.CoreEnabled){
    need(row.CoreBuilds===null&&Object.values(s).every(x=>x===0),'control hidden shared work');
  }else{
    need(s.Builds===row.CoreBuilds.length&&s.Builds+s.Hits===129&&s.OwnerAcquisitions>=s.Builds&&s.OwnerAcquisitions<=129,'build/hit/owner conservation');
    const builds=new Map();let physical=0;
    for(const b of row.CoreBuilds){
      const state=JSON.parse(b.Encoded),prefix=roots.get(b.Root);
      need(prefix&&state.Hash===b.Root,'build root');
      const expected={Head:prefix.state.Head,Hash:prefix.state.Hash,Log:prefix.state.Log,Sources:prefix.state.Sources,Certificate:prefix.state.Certificate};
      need(same(state,expected),'build fields not committed prefix');
      need(!builds.has(b.Root),'duplicate core root');builds.set(b.Root,b);physical+=Buffer.byteLength(b.Encoded);
      const projections=row.MetadataProofs.filter(p=>p.Root===b.Root);
      need(projections.length>0&&projections.every(p=>p.Encoded===b.Encoded&&ns(b.At)<=ns(p.CapturedAt)),'build precedes bound requests');
      if(state.Certificate===null)need(projections.length===1,'nil certificate reused');
    }
    need(row.MetadataProofs.every(p=>builds.has(p.Root)),'projection without build');
    need(s.PhysicalBytes===physical&&row.Metadata.CopiesBytes===physical,'physical copy bytes');
    // A cached core cannot cross runtime-head motion; each projection's own
    // scoped snapshot was already independently validated by the sealed checker.
    for(const p of row.MetadataProofs){const b=builds.get(p.Root);need(same(JSON.parse(b.Encoded).Head,p.Snapshot),'reuse head mismatch');}
  }
  return {...r,metadata:row.Metadata,core_enabled:row.CoreEnabled,core_stats:s};
}
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[])}
function verify(){const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(f.files))need(hash(fs.readFileSync(p))===h,'changed source '+p);return f}
const cs=source.indexOf('for(const[name,mutate]of['),ce=source.indexOf(']){const row=',cs);need(cs>=0&&ce>cs,'controls boundary');
let list=source.slice(cs+'for(const[name,mutate]of'.length,ce+1);
const old="r.MetadataReads.find(t=>t.Getters.length>2).Getters[2].Key='unknown'",fixed="r.MetadataReads.find(t=>r.Reads.some(x=>x.Journal===t.Journal)&&t.Getters.length>2).Getters[2].Key='unknown'";
need(list.split(old).length===2,'unique delivered target');list=list.replace(old,fixed);
const controlsVM=vm.runInNewContext('const hash=x=>crypto.createHash("sha256").update(x).digest("hex");const recode=(p,change)=>{const s=JSON.parse(p.Encoded);change(s);p.Encoded=JSON.stringify(s);p.SHA256=hash(p.Encoded)};'+list,{crypto},{timeout:1000});
const mutations=[...controlsVM,
 ['missing-core-build',r=>r.CoreBuilds.pop()],['core-root',r=>r.CoreBuilds[0].Root='unknown'],['core-fields',r=>{const s=JSON.parse(r.CoreBuilds[0].Encoded);s.Head.evidence_epoch++;r.CoreBuilds[0].Encoded=JSON.stringify(s)}],['core-late',r=>r.CoreBuilds[0].At='2099-01-01T00:00:00Z'],['hidden-copy',r=>r.CoreStats.PhysicalBytes--],['hit-count',r=>r.CoreStats.Hits++],['owner-count',r=>r.CoreStats.OwnerAcquisitions=0],['core-factor',r=>r.CoreEnabled=false]
];
function corruptions(first){const controls=[];for(const[name,mutate]of mutations){const r=structuredClone(first);mutate(r);let bad=false;try{check(r)}catch{bad=true}need(bad,'corruption escaped '+name);controls.push(name)}return controls}
if(process.argv.includes('--self-test')){
  // Fixture only: V35 is already consumed and is never a V36 normal result.
  let fixture;for await(const line of readline.createInterface({input:fs.createReadStream('research/metadata-projection-v35/raw.ndjson'),crlfDelay:Infinity})){const r=JSON.parse(line);if(r.Projected&&r.Archived&&!r.Visible){fixture=r;break}}
  need(fixture,'fixture missing');fixture.CoreEnabled=true;
  const built=new Map();for(const p of fixture.MetadataProofs){const prev=built.get(p.Root);if(!prev||ns(p.CapturedAt)<ns(prev.At))built.set(p.Root,{Root:p.Root,Encoded:p.Encoded,At:p.CapturedAt});}
  fixture.CoreBuilds=[...built.values()];
  const bytes=fixture.CoreBuilds.reduce((n,p)=>n+Buffer.byteLength(p.Encoded),0);
  fixture.Metadata.CopiesBytes=bytes;
  fixture.CoreStats={Builds:built.size,Hits:129-built.size,OwnerAcquisitions:built.size,PhysicalBytes:bytes};
  check(fixture);const controls=corruptions(fixture);
  console.log(JSON.stringify({type:'pre-normal-auditor-self-test',controls,fixture:'consumed V35, not V36 results'},null,2));
}else if(process.argv.includes('--freeze')){
  fs.mkdirSync(dir,{recursive:true});const c=JSON.parse(fs.readFileSync(path.join(dir,'technical-prefreeze/checks.json')));
  need(c.changed.length===0&&c.checks.every(x=>x.code===0),'technical preflight');
  const files=[...below('internal'),'go.mod','go.sum','cmd/research-metadata-core-gen/main.go','research/durable-witness-v23-audit.mjs','research/batch-archive-v33-audit.mjs','research/archive-cost-v34-audit.mjs','research/metadata-projection-v35-audit.mjs','research/metadata-core-v36-audit.mjs','research/metadata-core-v36-run.mjs','docs/experiments/mmm-metadata-core-v36-protocol.md'].sort();
  fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]))},null,2)+'\n',{flag:'wx'});console.log('frozen',files.length);
}else{
  const f=verify(),raw=path.join(dir,'raw.ndjson'),summaries=[];let header=false,footer=false,first;
  for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})){if(!line)continue;const row=JSON.parse(line);
    if(row.Type==='header'){need(!header&&row.Trials===16&&row.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header');header=true;continue}
    if(row.Type==='footer'){need(!footer&&summaries.length===16&&row.Trials===16,'footer');footer=true;continue}
    need(header&&!footer,'envelope');summaries.push(check(row));if(row.CoreEnabled&&row.Archived&&!row.Visible&&!first)first=row;
  }
  need(header&&footer&&summaries.length===16&&new Set(summaries.map(r=>[r.trial,r.mode,r.core_enabled,r.visible].join('/'))).size===16,'complete factorial');
  const run=JSON.parse(fs.readFileSync(path.join(dir,'run.json')));need(run.code===0&&run.raw_sha256===hash(fs.readFileSync(raw)),'raw/run');
  const controls=corruptions(first);verify();
  const report={type:'prospectively-frozen-core-audit',time:new Date().toISOString(),raw_sha256:run.raw_sha256,source_files:Object.keys(f.files).length,controls,scope_limit:'prime full nomination/independent ack not recorded; all 128 delivered requests per trial checked',trials:summaries,all_seven_goals:'OPEN'};
  fs.writeFileSync(path.join(dir,'audit.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});
  console.log(JSON.stringify({...report,trials:summaries.map(r=>({...r,freshness:{used:r.freshness.filter(x=>x.used).length,censored:r.freshness.filter(x=>!x.used).length,max_ns:Math.max(0,...r.freshness.filter(x=>x.used).map(x=>x.first_use_ns))}}))},null,2));
}
