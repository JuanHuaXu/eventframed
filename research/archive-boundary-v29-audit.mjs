import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import {isDeepStrictEqual as eq} from 'node:util';
const sha=x=>crypto.createHash('sha256').update(x).digest('hex');
const dir=path.resolve(process.argv[2]??'research/archive-boundary-v29');
const fail=x=>{throw new Error(x);};
function audit(raw){
  if(raw.Type!=='historical-owned-archive-technical'||raw.LoadOrQualityProof!==false)fail('wrong scope');
  const kinds=['future','visible','outcome','mixed','cancel'];
  if(!eq(raw.Cases.map(x=>x.Kind),kinds))fail('case set');
  for(const r of raw.Cases){
    if(sha(r.JournalWire)!==r.WireSHA256||!eq(JSON.parse(r.JournalWire),r.Journal))fail('wire');
    if(!eq(r.Captured,r.Journal.snapshot)||eq(r.Captured,r.Current))fail('historical snapshot');
    if(r.CurrentGateAccepts!==(r.Kind==='future'))fail('ordinary validity control');
    if(!r.FullStreamAccepted||r.FullStreamResponse.duplicate||r.FullStreamResponse.snapshot.runtime_version<=r.Current.runtime_version)fail('historical feedback');
    if(r.SelectedRejected!==['visible','mixed'].includes(r.Kind))fail('selection guard');
    if((r.PostCalls??[]).length)fail('late reads');
    if(!/^[a-f0-9]{64}$/.test(r.CaptureRoot))fail('capture provenance');
    const times=['Handoff','MutationBegin','MutationEnd','Ack'].map(k=>Date.parse(r[k]));
    if(times.some(x=>!Number.isFinite(x))||times.some((x,i)=>i&&x<times[i-1])||times[3]-times[2]<4)fail('barrier order');
    const decisions=r.Journal.report.decisions;
    if(decisions.length!==150||new Set(decisions.map(d=>d.event_id)).size!==150)fail('frontier');
    if(r.Kind==='cancel'){
      if(r.Packet.bayesian_shadow?.journal_id)fail('canceled packet');
    }else{
      if(!eq(r.Packet.snapshot,r.Captured)||r.Packet.candidates.length!==10)fail('packet capture');
      if(!eq(r.Packet.bayesian_shadow,r.Journal.report))fail('packet report');
      const map=new Map(decisions.map(d=>[d.event_id,d]));
      for(const c of r.Packet.candidates)if(!eq(c.forecast,map.get(c.event.id)?.forecast))fail('packed law');
    }
  }
}
const freeze=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
for(const[p,h]of Object.entries(freeze.files))if(sha(fs.readFileSync(p))!==h)fail('changed source '+p);
const run=JSON.parse(fs.readFileSync(path.join(dir,'run.json')));
if(run.code!==0||run.changed.length)fail('run failed');
const bytes=fs.readFileSync(path.join(dir,'raw.json'));
if(sha(bytes)!==run.raw_sha256)fail('raw digest');
const raw=JSON.parse(bytes);audit(raw);
const controls={
  wire:r=>{r.Cases[0].WireSHA256='0'.repeat(64)},
  snapshot:r=>{r.Cases[0].Captured.runtime_version++},
  timing:r=>{r.Cases[0].Ack=r.Cases[0].Handoff},
  validity:r=>{r.Cases[1].CurrentGateAccepts=true},
  selection:r=>{r.Cases[1].SelectedRejected=false},
  lateRead:r=>{r.Cases[0].PostCalls=['Snapshot']},
  law:r=>{r.Cases[0].Packet.candidates[0].forecast.rank_score+=1},
};
for(const[name,mutate]of Object.entries(controls)){
  const clone=structuredClone(raw);mutate(clone);let rejected=false;
  try{audit(clone)}catch{rejected=true}if(!rejected)fail('control escaped '+name);
}
const report={scope:'technical-only',pass:true,cases:raw.Cases.length,source_files:Object.keys(freeze.files).length,raw_sha256:sha(bytes),corruption_controls:Object.keys(controls),generated_persistence:'Go inverse-AST test',authority_controls:'six live Go controls',adoption:false};
fs.writeFileSync(path.join(dir,'audit.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(report,null,2));
