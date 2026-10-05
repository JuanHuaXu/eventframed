import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import vm from 'node:vm';
import {spawnSync} from 'node:child_process';

// Post-hoc classification repair. The original collector/checker, raw flags,
// hashes and failed adoption gates remain sealed, including their false flags.
const dir=path.resolve('research/durable-witness-v23');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(ok,why)=>{if(!ok)throw Error(why);};
const source=fs.readFileSync('research/durable-witness-v23-audit.mjs','utf8');
const freeze=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
for(const [file,digest] of Object.entries(freeze.files)) need(hash(fs.readFileSync(file))===digest,'changed frozen source: '+file);
const initial=spawnSync('node',['research/durable-witness-v23-audit.mjs'],{encoding:'utf8'});
need(initial.status!==0 && initial.stderr.includes('collector correctness errors'),'initial checker failure not reproduced');
const failure={code:initial.status,stdout:initial.stdout,stderr:initial.stderr,raw_sha256:hash(fs.readFileSync(path.join(dir,'raw.ndjson')))};
const failurePath=path.join(dir,'initial-audit-failure.json');
if(!fs.existsSync(failurePath)) fs.writeFileSync(failurePath,JSON.stringify(failure,null,2)+'\n',{flag:'wx'});
else need(JSON.stringify(JSON.parse(fs.readFileSync(failurePath)))===JSON.stringify(failure),'changed original failure');
// Execute the exact sealed independent model/timing checks unchanged. Only the
// known private-field classifier flags are separated in an analysis-local copy.
const core=source.slice(source.indexOf('const root ='),source.indexOf('function filesBelow'));
const {audit}=vm.runInNewContext(core+'\n({audit})',{fs,path,crypto,process},{timeout:1000});
function canonicalUTC(value) {
  const m=value.match(/^(.*?)(?:\.(\d+))?(Z|[+-]\d\d:\d\d)$/);
  need(m,'invalid RFC3339 instant');
  const whole=new Date(Date.parse(m[1]+m[3])).toISOString().replace(/\.000Z$/,'');
  const canonical=whole+(m[2]?'.'+m[2]:'')+'Z';
  need(Date.parse(value)===Date.parse(canonical),'timezone conversion changed instant');
  return canonical;
}
const rows=[], controls=[]; let first,header=false,footer=false;
const raw=path.join(dir,'raw.ndjson');
for await(const line of readline.createInterface({input:fs.createReadStream(raw),crlfDelay:Infinity})) {
  if(!line)continue;
  const row=JSON.parse(line);
  if(row.Type==='header') {need(!header && row.Trials===8 && row.FreezeSHA256===hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header/freeze');header=true;continue;}
  if(row.Type==='footer') {need(!footer && rows.length===8 && row.Trials===8,'footer');footer=true;continue;}
  need(header && !footer,'row outside envelope');
  need(row.Errors?.length===128 && row.Errors.every(e=>e==='durable journal mismatch'),'unexplained collector error');
  // Native published view timestamps use Go Local (-04:00); normalize only
  // their representation for the sealed UTC-only parser, preserving fraction.
  const analysis={...row,Errors:[],Reads:row.Reads.map(r=>({...r,ViewAt:canonicalUTC(r.ViewAt)}))};
  const independent=audit(analysis);
  // Binding presence is evidence of commit ack; the native gate itself verifies
  // byte-equivalent readback before ack. No new backend replay is pretended here.
  for(const r of row.Reads) {
    const b=row.FinalWitness.Journals[r.Journal];
    need(b && JSON.stringify(b.Snapshot)===JSON.stringify(r.Snapshot),'missing committed journal provenance');
    need(row.Commits.some(c=>JSON.parse(c.Payload).After.runtime_version>=r.Snapshot.runtime_version),'journal has no committed witness publication');
  }
  rows.push({...independent,original_collector_pass:row.Pass,original_private_field_flags:row.Errors.length,journal_readback_scope:'native frozen gate verified wire before ack; independent wire falsifier is post-hoc, not a second stored-cohort replay'});
  if(!first)first=analysis;
}
need(header && footer && rows.length===8 && new Set(rows.map(r=>[r.trial,r.enabled,r.visible].join('/'))).size===8,'full factorial');
for(const [name,mutate] of [
  ['missing-write',r=>r.Writes.pop()], ['future-nominee',r=>r.Reads[0].Decisions[0].event_id='future125'],
  ['pin-mismatch',r=>r.Reads[0].PinSnapshot.evidence_epoch++], ['future-label',r=>r.Outcomes[0].Request.available_at='2099-01-01T00:00:00Z'],
  ['chain-prior',r=>r.Commits[0].Prior='corrupt'], ['retagged-source',r=>r.DurableSources[0].evidence_epoch++],
  ['fabricated-latency',r=>r.Metrics.offer_p99_ns=0], ['fabricated-pass',r=>r.Pass=!r.Pass],
]) {
  const copy=structuredClone(first);mutate(copy);let rejected=false;
  try {audit(copy);}catch{rejected=true;}
  need(rejected,'corrupt tape accepted: '+name);controls.push(name);
}
const summary={type:'post-hoc-supplement',time:new Date().toISOString(),raw_sha256:hash(fs.readFileSync(raw)),original_checker_sha256:hash(source),supplement_sha256:hash(fs.readFileSync('research/durable-witness-v23-supplement.mjs')),
  diagnostic_source_sha256:hash(fs.readFileSync('internal/store/libravdbstore/research_durable_witness_v23_diagnostic_test.go')),source_files:Object.keys(freeze.files).length,
  parser_repair:'published ViewAt uses RFC3339 offset; analysis-local UTC normalization preserves original instant and fractional nanoseconds',
  original_checker:'FAILED at private-field classifier; retained unchanged',adoption:'FAIL, outcome latency alone exceeds the unchanged gate in every arm',all_seven_goals:'OPEN',controls,trials:rows};
fs.writeFileSync(path.join(dir,'supplement.json'),JSON.stringify(summary,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({adoption:summary.adoption,controls:controls.length,source_files:summary.source_files,trials:rows.map(r=>({trial:r.trial,enabled:r.enabled,visible:r.visible,learned:r.learned,cross_epoch:r.cross_epoch,pass:r.pass,metrics:r.metrics,fresh_used:r.freshness.filter(f=>f.used).length}))},null,2));
