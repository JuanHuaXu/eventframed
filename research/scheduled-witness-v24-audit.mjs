import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import vm from 'node:vm';
import {execFileSync} from 'node:child_process';

const root=process.cwd(),dir=path.resolve('research/scheduled-witness-v24');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(ok,why)=>{if(!ok)throw Error(why);};
const ns=x=>{const m=x.match(/^(.*?)(?:\.(\d+))?Z$/);need(m,'UTC timestamp');return BigInt(Date.parse(m[1]+'Z'))*1000000n+BigInt((m[2]??'').padEnd(9,'0'));};
const same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
// Reuse the sealed independent source/nomination/law/timing checker unchanged.
// This cohort records UTC views and genuine wire proofs, so no post-hoc repair.
const source=fs.readFileSync('research/durable-witness-v23-audit.mjs','utf8');
const core=source.slice(source.indexOf('const root ='),source.indexOf('function filesBelow'));
const {audit}=vm.runInNewContext(core+'\n({audit})',{fs,path,crypto,process},{timeout:1000});

function auditScheduled(row) {
  const summary=audit(row);
  need(row.Enabled===true && (row.Errors??[]).length===0,'conditional witness/control regime');
  const proofs=new Map(row.JournalProofs.map(p=>[p.Index,p]));
  need(proofs.size===128,'wire proof cardinality');
  for(const r of row.Reads) {
    const p=proofs.get(r.Index);need(p && same(p.Snapshot,r.Snapshot) && p.SHA256===hash(JSON.stringify(r.Decisions)),'durable journal wire proof');
    const b=row.FinalWitness.Journals[r.Journal];need(b && same(b.Snapshot,r.Snapshot),'durable query binding');
  }
  const stats=row.Admission;
  need(!stats.Closed && !stats.ActiveWriter && stats.ActiveReaders===0 && stats.Queued.every(n=>n===0),'admission terminal state');
  const counts=[0,0,0],waits=[[],[],[]];
  const boundaries=[];
  for(const r of row.Trace) {
    need([0,1,2].includes(r.Kind) && r.Error==='' && ns(r.Begin)<=ns(r.Acquired) && ns(r.Acquired)<=ns(r.Released),'lease trace/timing');
    counts[r.Kind]++;waits[r.Kind].push(Number(ns(r.Acquired)-ns(r.Begin)));
    boundaries.push({time:ns(r.Acquired),start:true,kind:r.Kind},{time:ns(r.Released),start:false,kind:r.Kind});
  }
  need(counts[0]===16 && counts[1]===129 && counts[2]===row.Commits.length-145,'complete lease attempts');
  if(row.Scheduled) {
    need(same(stats.Granted,counts) && stats.Cancelled.every(n=>n===0) && stats.PeakQueued<=512 && stats.MaxReaders<=8,'grant/cap conservation');
    boundaries.sort((a,b)=>a.time<b.time?-1:a.time>b.time?1:Number(a.start)-Number(b.start));
    let readers=0,writers=0,peak=0;
    for(const e of boundaries) {
      const change=e.start?1:-1;
      if(e.kind===1)readers+=change;else writers+=change;
      need(readers>=0 && writers>=0 && readers<=8 && writers<=1 && (!writers || !readers),'shared/exclusive interval violation');
      peak=Math.max(peak,readers);
    }
    need(readers===0 && writers===0 && peak<=stats.MaxReaders,'lease trace drain');
  } else need(stats.Granted.every(n=>n===0) && stats.MaxReaders===0,'control secretly scheduled');
  return {...summary,scheduled:row.Scheduled,admission:stats,lease_counts:counts,queue_wait_max_ns:waits.map(a=>Math.max(...a)),journal_wire_proofs:proofs.size};
}

function filesBelow(dir) {return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?filesBelow(path.join(dir,e.name)):e.name.endsWith('.go')?[path.join(dir,e.name)]:[]);}
function verify() {const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const [p,h]of Object.entries(f.files))need(hash(fs.readFileSync(p))===h,'changed source '+p);return f;}
if(process.argv.includes('--freeze')) {
  fs.mkdirSync(dir,{recursive:true});
  const files=[...filesBelow('internal'),'go.mod','go.sum','research/durable-witness-v23-audit.mjs','research/scheduled-witness-v24-audit.mjs','research/scheduled-witness-v24-run.mjs','docs/experiments/mmm-scheduled-witness-v24-protocol.md'].sort();
  const f={time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]))};
  fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify(f,null,2)+'\n',{flag:'wx'});console.log('frozen',files.length,'files');
} else {
  const frozen=verify(),raw=path.join(dir,'raw.ndjson'),rows=[];let header=false,footer=false,first;
  for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})) {
    if(!line)continue;const r=JSON.parse(line);
    if(r.Type==='header'){need(!header && r.Trials===8 && r.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header/freeze');header=true;continue;}
    if(r.Type==='footer'){need(!footer && rows.length===8 && r.Trials===8,'footer');footer=true;continue;}
    need(header && !footer,'trial envelope');rows.push(auditScheduled(r));if(r.Scheduled && !r.Visible && !first)first=r;
  }
  need(header && footer && rows.length===8 && new Set(rows.map(r=>[r.trial,r.scheduled,r.visible].join('/'))).size===8,'factorial completeness');
  const controls=[];
  for(const [name,mutate]of [
    ['missing-write',r=>r.Writes.pop()],['future-nominee',r=>r.Reads[0].Decisions[0].event_id='future125'],
    ['pin-mismatch',r=>r.Reads[0].PinSnapshot.evidence_epoch++],['future-label',r=>r.Outcomes[0].Request.available_at='2099-01-01T00:00:00Z'],
    ['chain-prior',r=>r.Commits[0].Prior='corrupt'],['retagged-source',r=>r.DurableSources[0].evidence_epoch++],
    ['wire-proof',r=>r.JournalProofs[0].SHA256='corrupt'],['grant-count',r=>r.Admission.Granted[0]++],
    ['overlap',r=>{const w=r.Trace.find(t=>t.Kind===2),read=r.Trace.find(t=>t.Kind===1 && ns(t.Acquired)>ns(w.Begin));w.Begin=read.Begin;w.Acquired=read.Acquired;w.Released=read.Released;}],
    ['timing-fabrication',r=>r.Metrics.offer_p99_ns=0],['gate-fabrication',r=>r.Pass=!r.Pass],
  ]) {const copy=structuredClone(first);mutate(copy);let rejected=false;try{auditScheduled(copy);}catch{rejected=true;};need(rejected,'corrupt tape accepted '+name);controls.push(name);}
  verify();const report={type:'prospectively-frozen-audit',time:new Date().toISOString(),raw_sha256:hash(fs.readFileSync(raw)),source_files:Object.keys(frozen.files).length,controls,trials:rows,all_seven_goals:'OPEN'};
  const output=process.argv.find(x=>x.startsWith('--output='))?.slice(9);if(output)fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
  console.log(JSON.stringify({...report,trials:rows.map(r=>({...r,freshness:{used:r.freshness.filter(x=>x.used).length,censored:r.freshness.filter(x=>!x.used).length,max_ns:Math.max(0,...r.freshness.filter(x=>x.used).map(x=>x.first_use_ns))}}))},null,2));
}
