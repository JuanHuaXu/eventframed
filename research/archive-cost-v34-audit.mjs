import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import readline from 'node:readline';import vm from 'node:vm';import{execFileSync}from'node:child_process';
const dir=path.resolve('research/archive-cost-v34'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(b,x)=>{if(!b)throw Error(x)};
const sealed=fs.readFileSync('research/batch-archive-v33-audit.mjs','utf8');
const core=sealed.slice(sealed.indexOf('const dir='),sealed.indexOf('function below'));
const {check:original}=vm.runInNewContext(core+'\n({check})',{fs,path,crypto,vm,process},{timeout:1000});
function check(row){
 const r=original(row);
 need([-1,0,1,2,3].includes(row.Mode)&&row.Archived===(row.Mode!==-1),'factor mode');
 need(row.SharedAncestry===(row.Mode>=0&&!!(row.Mode&1))&&row.DeferredProofs===(row.Mode>=0&&!!(row.Mode&2)),'factor flags');
 if(row.Archived){
  need(row.AncestryScans===(row.SharedAncestry?row.Batches.length:129),'full ancestry scans');
  need(row.ProofBeforeAck===(row.DeferredProofs?0:129),'proof formatting timing');
  need(Number.isSafeInteger(row.PostDrainProofNS)&&row.PostDrainProofNS>0,'post-drain cost');
 }else need(row.AncestryScans===0&&row.ProofBeforeAck===0&&row.PostDrainProofNS===0,'control hidden overhead');
 return {...r,mode:row.Mode,shared_ancestry:row.SharedAncestry,deferred_proofs:row.DeferredProofs,ancestry_scans:row.AncestryScans,proof_before_ack:row.ProofBeforeAck,post_drain_proof_ns:row.PostDrainProofNS};
}
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[])}
function verify(){const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(f.files))need(hash(fs.readFileSync(p))===h,'changed source '+p);return f}
if(process.argv.includes('--freeze')){
 fs.mkdirSync(dir,{recursive:true});
 const checks=JSON.parse(fs.readFileSync(path.join(dir,'technical-prefreeze/checks.json')));
 need(!checks.changed.length&&checks.checks.every(x=>x.code===0),'technical preflight');
 const files=[...below('internal'),'go.mod','go.sum','cmd/research-archive-cost-gen/main.go','cmd/research-archive-gen/main.go','research/durable-witness-v23-audit.mjs','research/batch-archive-v33-audit.mjs','research/archive-cost-v34-audit.mjs','research/archive-cost-v34-run.mjs','docs/experiments/mmm-archive-cost-v34-protocol.md'].sort();
 const f={time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]))};
 fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify(f,null,2)+'\n',{flag:'wx'});console.log('frozen',files.length);
}else{
 const f=verify(),raw=path.join(dir,'raw.ndjson'),summaries=[];
 let header=false,footer=false,first;
 for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})){
  if(!line)continue;const row=JSON.parse(line);
  if(row.Type==='header'){need(!header&&row.Trials===20&&row.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header');header=true;continue;}
  if(row.Type==='footer'){need(!footer&&summaries.length===20&&row.Trials===20,'footer');footer=true;continue;}
  need(header&&!footer,'envelope');summaries.push(check(row));if(row.Mode===3&&!row.Visible&&!first)first=row;
 }
 need(header&&footer&&summaries.length===20&&new Set(summaries.map(r=>[r.trial,r.mode,r.visible].join('/'))).size===20,'complete factorial');
 const run=JSON.parse(fs.readFileSync(path.join(dir,'run.json')));need(run.code===0&&run.raw_sha256===hash(fs.readFileSync(raw)),'raw/run');
 const controls=[];
 for(const[name,mutate]of[
  ['grant-count',r=>r.Admission.Granted[0]++],['missing-write',r=>r.Writes.pop()],['future-nominee',r=>r.Reads[0].Decisions[0].event_id='future125'],['pin-mismatch',r=>r.Reads[0].PinSnapshot.evidence_epoch++],['future-label',r=>r.Outcomes[0].Request.available_at='2099-01-01T00:00:00Z'],['chain-prior',r=>r.Commits[0].Prior='corrupt'],['retagged-source',r=>r.DurableSources[0].evidence_epoch++],['wire-proof',r=>r.JournalProofs[0].SHA256='bad'],['queue-hidden-work',r=>r.ArchiveQueue.Finished--],['duplicate-batch',r=>r.Batches[1].IDs[0]=r.Batches[0].IDs[0]],['delivered-premature-ack',r=>{r.Batches.find(b=>r.Reads.some(x=>b.IDs.includes(x.Journal))).End='2099-01-01T00:00:00Z'}],['capsule-wire',r=>r.Captures[0].Encoded+=' '],['capsule-root',r=>r.Captures[0].Root='unknown'],['late-capture',r=>r.Captures[0].CapturedAt='2099-01-01T00:00:00Z'],['timing-fabrication',r=>r.Metrics.offer_p99_ns=0],['gate-fabrication',r=>r.Pass=!r.Pass],['scan-suppression',r=>r.AncestryScans--],['proof-timing-fabrication',r=>r.ProofBeforeAck++],['factor-relabelling',r=>r.Mode=0],['hidden-post-drain-cost',r=>r.PostDrainProofNS=0]
 ]){
  const row=structuredClone(first);mutate(row);let rejected=false;try{check(row)}catch{rejected=true};need(rejected,'corruption escaped '+name);controls.push(name);
 }
 verify();const report={type:'prospectively-frozen-factorial-audit',time:new Date().toISOString(),raw_sha256:run.raw_sha256,source_files:Object.keys(f.files).length,controls,scope_limit:'setup-prime ack lacks an independently bounded timestamp; all128loaded acknowledgments bounded',trials:summaries,all_seven_goals:'OPEN'};
 fs.writeFileSync(path.join(dir,'audit.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});
 console.log(JSON.stringify({...report,trials:summaries.map(r=>({...r,freshness:{used:r.freshness.filter(x=>x.used).length,censored:r.freshness.filter(x=>!x.used).length,max_ns:Math.max(0,...r.freshness.filter(x=>x.used).map(x=>x.first_use_ns))}}))},null,2));
}
