// Post-run supplementary identity audit; not a predeclared adoption criterion.
import fs from 'node:fs';import crypto from 'node:crypto';import path from 'node:path';import{isDeepStrictEqual as eq}from'node:util';
const dir=path.resolve('research/archive-requests-v32');const sha=x=>crypto.createHash('sha256').update(x).digest('hex');const fail=x=>{throw Error(x)};
function audit(raw){
 if(raw.Type!=='request-identity-diagnostic'||raw.LoadedLatencyProof!==false||!eq(raw.Trials.map(t=>t.Mode),['native-clock','distinct-asof','identical-retry']))fail('scope');
 const results=[];
 for(const t of raw.Trials){
  if(t.Captures.length!==16||t.Responses.length!==16||new Set(t.Responses.map(r=>r.Index)).size!==16)fail('count');
  const map=new Map();
  for(const c of t.Captures){
   if(!/^bj_[0-9a-f]{32}$/.test(c.id)||c.report.decisions.length!==150||new Set(c.report.decisions.map(d=>d.event_id)).size!==150)fail('capture');
   if(map.has(c.id)&&!eq(map.get(c.id),c))fail('same id/different wire');
   map.set(c.id,c);
  }
  const used=new Map();
  for(const r of t.Responses){
   if(r.Error!=='')fail('request failed');
   const p=r.Packet,c=map.get(p.bayesian_shadow.journal_id);if(!c)fail('unbound');
   if(c.as_of!==r.Request.as_of||c.tenant_id!==r.Request.tenant_id||c.session_id!==r.Request.session_id||!eq(p.snapshot,c.snapshot)||!eq(p.bayesian_shadow,c.report))fail('request/capture');
   if(p.candidates.length!==10)fail('packing');
   const decisions=new Map(c.report.decisions.map(d=>[d.event_id,d]));for(const x of p.candidates)if(!eq(x.forecast,decisions.get(x.event.id)?.forecast))fail('law');
   used.set(c.id,(used.get(c.id)??0)+1);
  }
  if(used.size!==map.size||t.Unique!==map.size||t.Duplicates!==16-map.size)fail('multiplicity');
  if(t.Mode==='distinct-asof'&&map.size!==16)fail('distinct');if(t.Mode==='identical-retry'&&map.size!==1)fail('retry');
  // Full strings, not Date.parse milliseconds, preserve nanosecond distinctions.
  const timestamps=new Set(t.Responses.map(r=>r.Request.as_of));
  if(t.Mode==='native-clock'&&timestamps.size!==map.size)fail('clock/identity explanation');
  results.push({mode:t.Mode,requests:16,unique:map.size,duplicates:16-map.size,distinct_timestamps:timestamps.size,diagnostic_elapsed_ms:t.ElapsedNS/1e6});
 }
 return results;
}
const freeze=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(freeze.files))if(sha(fs.readFileSync(p))!==h)fail('source changed '+p);
const run=JSON.parse(fs.readFileSync(path.join(dir,'run.json')));const bytes=fs.readFileSync(path.join(dir,'raw.json'));if(run.code!==0||run.changed.length||sha(bytes)!==run.raw_sha256)fail('run/hash');
const raw=JSON.parse(bytes),results=audit(raw);const controls={
 loss:r=>{r.Trials[0].Responses.pop()},
 error:r=>{r.Trials[0].Responses[0].Error='injected'},
 asof:r=>{r.Trials[0].Responses[0].Request.as_of='2000-01-01T00:00:00Z'},
 law:r=>{r.Trials[0].Responses[0].Packet.candidates[0].forecast.rank_score+=1},
 multiplicity:r=>{r.Trials[0].Duplicates++},
 retag:r=>{r.Trials[0].Responses[0].Packet.snapshot.evidence_epoch++},
 duplicateWire:r=>{r.Trials[2].Captures[1].report.decisions[0].forecast.rank_score+=1},
};for(const[name,mutate]of Object.entries(controls)){let rejected=false;const r=structuredClone(raw);mutate(r);try{audit(r)}catch{rejected=true}if(!rejected)fail('negative control escaped '+name)}
const report={type:'post-hoc supplemental identity audit',pass:true,raw_sha256:sha(bytes),audit_source_sha256:sha(fs.readFileSync('research/archive-requests-v32-audit.mjs')),source_files:Object.keys(freeze.files).length,results,corruption_controls:Object.keys(controls),adoption:false};fs.writeFileSync(path.join(dir,'supplement.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
