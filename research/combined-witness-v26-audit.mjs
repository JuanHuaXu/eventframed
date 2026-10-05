import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import vm from 'node:vm';
import {execFileSync} from 'node:child_process';

const dir=path.resolve('research/combined-witness-v26'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(ok,why)=>{if(!ok)throw Error(why);},same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
const ns=x=>{const m=x.match(/^(.*?)(?:\.(\d+))?Z$/);need(m,'UTC timestamp');return BigInt(Date.parse(m[1]+'Z'))*1000000n+BigInt((m[2]??'').padEnd(9,'0'));};
const source=fs.readFileSync('research/durable-witness-v23-audit.mjs','utf8');
const core=source.slice(source.indexOf('const root ='),source.indexOf('function filesBelow'));
const {audit}=vm.runInNewContext(core+'\n({audit})',{fs,path,crypto,process},{timeout:1000});
function auditCombined(row){
 const result=audit(row);need(row.Enabled && row.Joined && typeof row.Scheduled==='boolean' && (row.Errors??[]).length===0,'wrong regime');
 const proofs=new Map(row.JournalProofs.map(p=>[p.Index,p]));need(proofs.size===128,'wire proof count');
 const journals=row.FinalWitness.Journals;need(Object.keys(journals).length===129,'journal binding cardinality');
 for(const r of row.Reads){const p=proofs.get(r.Index);need(p && same(p.Snapshot,r.Snapshot) && p.SHA256===hash(JSON.stringify(r.Decisions)),'wire readback');need(journals[r.Journal] && same(journals[r.Journal].Snapshot,r.Snapshot),'missing query binding');}
 const insertBatches=row.Commits.filter(c=>(JSON.parse(c.Payload).Mutations??[]).some(m=>m.Kind===1)).length;
 const counts=[0,0,0];for(const r of row.Trace){need([0,1,2].includes(r.Kind) && r.Error==='' && ns(r.Begin)<=ns(r.Acquired) && ns(r.Acquired)<=ns(r.Released),'trace');counts[r.Kind]++;}
 need(same(counts,[16,129,insertBatches]),'lease conservation');
 const stats=row.Admission;need(!stats.Closed && !stats.ActiveWriter && stats.ActiveReaders===0 && stats.Queued.every(n=>n===0) ,'undrained scheduler');
 if(row.Scheduled){
  need(same(stats.Granted,counts) && stats.Cancelled.every(n=>n===0) && stats.MaxReaders<=8 && stats.PeakQueued<=512,'grant/cap conservation');
  const boundaries=row.Trace.flatMap(r=>[{at:ns(r.Acquired),start:true,kind:r.Kind},{at:ns(r.Released),start:false,kind:r.Kind}]);
  boundaries.sort((a,b)=>a.at<b.at?-1:a.at>b.at?1:Number(a.start)-Number(b.start));
  let readers=0,writers=0,peak=0;for(const b of boundaries){const step=b.start?1:-1;if(b.kind===1)readers+=step;else writers+=step;need(readers>=0 && writers>=0 && readers<=8 && writers<=1 && (!writers || !readers),'lease overlap');peak=Math.max(peak,readers);}
  need(readers===0 && writers===0 && peak<=stats.MaxReaders,'lease drain');
 }else need(stats.Granted.every(n=>n===0) && stats.MaxReaders===0,'hidden scheduler');
 let entries=0;const seen=new Set();const batchSizes=[];
 if(row.Joined){
  need(row.Batches.length>=Math.ceil(129/4) && row.Batches.length<=129,'batch count');
  for(const b of row.Batches){
   need(b.Error==='' && b.IDs.length>=1 && b.IDs.length<=4 && ns(b.Begin)<=ns(b.End),'batch bounds/timing');batchSizes.push(b.IDs.length);
   const commit=row.Commits.find(c=>c.Digest===b.After);need(commit && commit.Prior===b.Before && (JSON.parse(commit.Payload).Mutations??[]).length===0,'batch chain binding');
   for(const id of b.IDs){need(!seen.has(id) && journals[id],'duplicate/missing batch binding');seen.add(id);entries++;const r=row.Reads.find(r=>r.Journal===id);if(r)need(ns(r.Start)<=ns(b.Begin) && ns(b.End)<=ns(r.End),'premature acknowledgment');}
  }
  need(entries===129 && row.Commits.length===row.Batches.length+16+insertBatches,'joint publication conservation');
 }else need(row.Batches===null && row.Commits.length===129+16+insertBatches,'two-stage control');
 return {...result,joined:row.Joined,scheduled:row.Scheduled,admission:stats,journal_wire_proofs:proofs.size,insert_batches:insertBatches,journal_batches:row.Joined?row.Batches.length:null,batch_histogram:[1,2,3,4].map(k=>batchSizes.filter(n=>n===k).length),lease_counts:counts};
}
function filesBelow(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?filesBelow(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[]);}
function verify(){const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(f.files))need(hash(fs.readFileSync(p))===h,'changed source '+p);return f;}
if(process.argv.includes('--freeze')){
 fs.mkdirSync(dir,{recursive:true});const files=[...filesBelow('internal'),'go.mod','go.sum','research/durable-witness-v23-audit.mjs','research/combined-witness-v26-audit.mjs','research/combined-witness-v26-run.mjs','docs/experiments/mmm-combined-witness-v26-protocol.md'].sort();
 const f={time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]))};fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify(f,null,2)+'\n',{flag:'wx'});console.log('frozen',files.length,'files');
}else{
 const f=verify(),raw=path.join(dir,'raw.ndjson'),rows=[];let header=false,footer=false,first;
 for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})){if(!line)continue;const r=JSON.parse(line);if(r.Type==='header'){need(!header && r.Trials===8 && r.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header');header=true;continue;}if(r.Type==='footer'){need(!footer && rows.length===8 && r.Trials===8,'footer');footer=true;continue;}need(header&&!footer,'envelope');rows.push(auditCombined(r));if(r.Scheduled&&!r.Visible&&!first)first=r;}
 need(header&&footer&&rows.length===8&&new Set(rows.map(r=>[r.trial,r.scheduled,r.visible].join('/'))).size===8,'factorial completeness');
 const controls=[];
 for(const[name,mutate]of[
  ['grant-count',r=>r.Admission.Granted[0]++],['overlap',r=>{const w=r.Trace.find(t=>t.Kind===2),read=r.Trace.find(t=>t.Kind===1 && ns(t.Acquired)>ns(w.Begin));w.Begin=read.Begin;w.Acquired=read.Acquired;w.Released=read.Released;}],['missing-write',r=>r.Writes.pop()],['future-nominee',r=>r.Reads[0].Decisions[0].event_id='future125'],['pin-mismatch',r=>r.Reads[0].PinSnapshot.evidence_epoch++],['future-label',r=>r.Outcomes[0].Request.available_at='2099-01-01T00:00:00Z'],['chain-prior',r=>r.Commits[0].Prior='corrupt'],['retagged-source',r=>r.DurableSources[0].evidence_epoch++],['wire-proof',r=>r.JournalProofs[0].SHA256='bad'],['missing-binding',r=>delete r.FinalWitness.Journals[r.Reads[0].Journal]],['duplicate-batch',r=>r.Batches[1].IDs[0]=r.Batches[0].IDs[0]],['premature-ack',r=>{const b=r.Batches.find(b=>b.IDs.includes(r.Reads[0].Journal));b.End='2099-01-01T00:00:00Z';}],['timing-fabrication',r=>r.Metrics.offer_p99_ns=0],['gate-fabrication',r=>r.Pass=!r.Pass],
 ]){const copy=structuredClone(first);mutate(copy);let rejected=false;try{auditCombined(copy);}catch{rejected=true;}need(rejected,'corruption accepted '+name);controls.push(name);}
 verify();const report={type:'prospectively-frozen-audit',time:new Date().toISOString(),raw_sha256:hash(fs.readFileSync(raw)),source_files:Object.keys(f.files).length,controls,trials:rows,all_seven_goals:'OPEN'};
 const output=process.argv.find(x=>x.startsWith('--output='))?.slice(9);if(output)fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
 console.log(JSON.stringify({...report,trials:rows.map(r=>({...r,freshness:{used:r.freshness.filter(x=>x.used).length,censored:r.freshness.filter(x=>!x.used).length,max_ns:Math.max(0,...r.freshness.filter(x=>x.used).map(x=>x.first_use_ns))}}))},null,2));
}
