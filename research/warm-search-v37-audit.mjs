import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import readline from 'node:readline';import vm from 'node:vm';import{execFileSync}from'node:child_process';
const dir=path.resolve('research/warm-search-v37'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(ok,x)=>{if(!ok)throw Error(x)},ns=x=>{const m=x.match(/^(.*?)(?:\.(\d+))?Z$/);need(m,'UTC');return BigInt(Date.parse(m[1]+'Z'))*1000000n+BigInt((m[2]??'').padEnd(9,'0'))};
const source=fs.readFileSync('research/metadata-core-v36-audit.mjs','utf8');
const start=source.indexOf('const dir='),end=source.indexOf('\nfunction below(');need(start>=0&&end>start,'V36 checker boundary');
const{check:prior}=vm.runInNewContext(source.slice(start,end)+'\n({check})',{fs,path,crypto,vm,process,Buffer,structuredClone},{timeout:1000});
function check(row){
 const r=prior(row);need(row.CoreEnabled===true&&typeof row.WarmEnabled==='boolean','warm factor');
 need(row.WarmUses.length===129&&new Set(row.WarmUses.map(x=>x.Journal)).size===129,'warm uses complete');
 const builds=new Map(row.CoreBuilds.map(b=>[b.Root,b])),proofs=new Map(row.MetadataProofs.map(p=>[p.Journal,p])),reads=new Map(row.Reads.map(x=>[x.Journal,x]));
 let fast=0;
 for(const u of row.WarmUses){
  const p=proofs.get(u.Journal),b=builds.get(u.Root),read=reads.get(u.Journal);
  need(p&&b&&p.Root===u.Root&&typeof u.Fast==='boolean','warm request/root binding');
  need(ns(u.CheckedAt)<=ns(p.CapturedAt)&&(!read||ns(read.Start)<=ns(u.CheckedAt)),'warm check time');
  if(u.Fast){need(row.WarmEnabled&&ns(b.At)<=ns(u.CheckedAt)&&JSON.parse(b.Encoded).Certificate!==null,'warm immutable core existed');fast++;}
 }
 need(row.WarmStats.Fast===fast&&row.WarmStats.Cold===129-fast&&row.CoreStats.Hits>=fast,'warm counter conservation');
 if(!row.WarmEnabled)need(fast===0,'control hidden warm use');
 return{...r,warm_enabled:row.WarmEnabled,warm_stats:row.WarmStats};
}
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[])}
function verify(){const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(f.files))need(hash(fs.readFileSync(p))===h,'changed source '+p);return f}
const ms=source.indexOf('const cs=source.indexOf('),me=source.indexOf('\nfunction corruptions(',ms);need(ms>=0&&me>ms,'sealed mutations boundary');
const{mutations:old}=vm.runInNewContext(source.slice(ms,me)+'\n({mutations})',{source:fs.readFileSync('research/metadata-projection-v35-audit.mjs','utf8'),need,vm,crypto},{timeout:1000});
const mutations=[...old,
 ['missing-warm-use',r=>r.WarmUses.pop()],['warm-root',r=>r.WarmUses[0].Root='unknown'],['warm-counter',r=>r.WarmStats.Fast++],['warm-late-check',r=>r.WarmUses[0].CheckedAt='2099-01-01T00:00:00Z'],['warm-first-nil',r=>{r.WarmUses[0].Fast=true;r.WarmStats.Fast++;r.WarmStats.Cold--}],['warm-factor',r=>r.WarmEnabled=false]
];
function corruptions(first){const out=[];for(const[name,mutate]of mutations){const r=structuredClone(first);mutate(r);let rejected=false;try{check(r)}catch{rejected=true}need(rejected,'corruption escaped '+name);out.push(name)}return out}
if(process.argv.includes('--self-test')){
 let first;for await(const line of readline.createInterface({input:fs.createReadStream('research/metadata-core-v36/raw.ndjson'),crlfDelay:Infinity})){const r=JSON.parse(line);if(r.CoreEnabled&&r.Archived&&!r.Visible){first=r;break}}
 need(first,'fixture missing');const seen=new Set();first.WarmEnabled=true;
 first.WarmUses=first.MetadataProofs.map(p=>{const fast=seen.has(p.Root)&&JSON.parse(p.Encoded).Certificate!==null;seen.add(p.Root);return{Journal:p.Journal,Root:p.Root,Fast:fast,CheckedAt:p.CapturedAt}});
 const fast=first.WarmUses.filter(u=>u.Fast).length;first.WarmStats={Fast:fast,Cold:129-fast};
 check(first);console.log(JSON.stringify({type:'pre-normal-auditor-self-test',fixture:'consumed V36 only; never V37 results',controls:corruptions(first)},null,2));
}else if(process.argv.includes('--freeze')){
 fs.mkdirSync(dir,{recursive:true});const c=JSON.parse(fs.readFileSync(path.join(dir,'technical-prefreeze/checks.json')));need(c.changed.length===0&&c.checks.every(x=>x.code===0),'preflight');
 const files=[...below('internal'),'go.mod','go.sum','cmd/research-warm-search-gen/main.go','research/durable-witness-v23-audit.mjs','research/batch-archive-v33-audit.mjs','research/archive-cost-v34-audit.mjs','research/metadata-projection-v35-audit.mjs','research/metadata-core-v36-audit.mjs','research/warm-search-v37-audit.mjs','research/warm-search-v37-run.mjs','docs/experiments/mmm-warm-search-v37-protocol.md'].sort();
 fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]))},null,2)+'\n',{flag:'wx'});console.log('frozen',files.length);
}else{
 const f=verify(),raw=path.join(dir,'raw.ndjson'),summaries=[];let header=false,footer=false,first;
 for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})){if(!line)continue;const row=JSON.parse(line);
  if(row.Type==='header'){need(!header&&row.Trials===16&&row.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header');header=true;continue}
  if(row.Type==='footer'){need(!footer&&summaries.length===16&&row.Trials===16,'footer');footer=true;continue}
  need(header&&!footer,'envelope');summaries.push(check(row));if(row.WarmEnabled&&row.Archived&&!row.Visible&&!first)first=row;
 }
 need(header&&footer&&summaries.length===16&&new Set(summaries.map(r=>[r.trial,r.mode,r.warm_enabled,r.visible].join('/'))).size===16,'complete factorial');
 const run=JSON.parse(fs.readFileSync(path.join(dir,'run.json')));need(run.code===0&&run.raw_sha256===hash(fs.readFileSync(raw)),'raw/run');
 const controls=corruptions(first);verify();const report={type:'prospectively-frozen-warm-search-audit',time:new Date().toISOString(),source_files:Object.keys(f.files).length,raw_sha256:run.raw_sha256,controls,scope_limit:'prime full nomination/independent ack not saved; all128delivered requests checked',trials:summaries,all_seven_goals:'OPEN'};
 fs.writeFileSync(path.join(dir,'audit.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({...report,trials:summaries.map(r=>({...r,freshness:{used:r.freshness.filter(x=>x.used).length,censored:r.freshness.filter(x=>!x.used).length,max_ns:Math.max(0,...r.freshness.filter(x=>x.used).map(x=>x.first_use_ns))}}))},null,2));
}
