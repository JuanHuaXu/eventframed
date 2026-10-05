// Independent final readback; never overwrites trial, freeze, or audit evidence.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import{execFileSync}from'node:child_process';
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const need=(ok,x)=>{if(!ok)throw Error(x)};
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const names=fs.readdirSync('research').filter(d=>/(witness-v2[3-6]|boundary-v2[7-9]|archive-(lifecycle|requests)-v3[0-2]|batch-archive-v33|archive-cost-v34)$/.test(d));
need(names.length===12,'missing frozen archive lineage');
const checked=[];
for(const name of names){
 const dir=path.join('research',name),f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
 for(const[p,h]of Object.entries(f.files))need(await fileHash(p)===h,'source changed '+p);
 const runPath=path.join(dir,'run.json');let run=null;
 if(fs.existsSync(runPath)){
  run=JSON.parse(fs.readFileSync(runPath));
  if(run.raw_sha256){
   const candidates=['raw.json','raw.ndjson'].map(p=>path.join(dir,p)).filter(p=>fs.existsSync(p));
   need(candidates.length===1,'ambiguous/missing raw artifact '+name);
   need(await fileHash(candidates[0])===run.raw_sha256,'raw changed '+name);
  }
 }
 checked.push({name,source_files:Object.keys(f.files).length,run_code:run?.code??null,raw_sha256:run?.raw_sha256??null});
}
const dir='research/archive-cost-v34',audit=JSON.parse(fs.readFileSync(path.join(dir,'audit.json')));
need(audit.controls.length===20&&audit.trials.length===20&&audit.trials.every(r=>!r.pass),'V34 negative evidence');
need(audit.raw_sha256===await fileHash(path.join(dir,'raw.ndjson')),'audit/raw');
const profileDir='research/batch-archive-v33',detail=JSON.parse(fs.readFileSync(path.join(profileDir,'profile-detail.json')));
for(const r of detail.reports)need(await fileHash(path.join(profileDir,`profile-${r.name}.txt`))===r.sha256,'profile report '+r.name);
need(detail.generation.code===0&&await fileHash(path.join(profileDir,'independent-generation.log'))===detail.generation.sha256,'generation terminal');
need(await fileHash('research/batch-archive-v33-profile-detail.mjs')===detail.script_sha256,'profile detail script');
const supplement=JSON.parse(fs.readFileSync(path.join(profileDir,'supplement.json')));
need(supplement.original_audit_exit===1&&await fileHash(path.join(profileDir,'audit-attempt.log'))===supplement.original_audit_log_sha256,'preserved original auditor failure');
need(await fileHash('research/batch-archive-v33-supplement.mjs')===supplement.supplement_source_sha256,'supplement source');
const diff=execFileSync('git',['diff','--numstat'],{encoding:'utf8'}).trim();
const expected='48\t0\t.learnings/ERRORS.md\n1\t0\tinternal/model/api.go\n21\t4\tinternal/productioneval/codex.go\n105\t0\tinternal/productioneval/codex_test.go\n25\t0\tinternal/service/service.go\n5\t1\tinternal/store/libravdbstore/store.go';
need(diff===expected,'preexisting tracked edits changed');
const report={time:new Date().toISOString(),type:'checkpoint-readback',local_verifier_repair:'first attempt exit1: assumed raw.ndjson for V29; actual recorded artifact is raw.json. Repaired artifact discovery, no trial/runtime/audit changes. Corrected preexisting per-file numstat assumption against authoritative git output; aggregate205/5 unchanged.',checked,source_and_raw_unchanged:true,preserved_v33_auditor_exit:1,v33_profile_generation_code:0,v34_all20adoption_fail:true,v34_corruptions_rejected:20,tracked_diff:diff,script_sha256:hash(fs.readFileSync('research/archive-cost-v34-checkpoint-verify.mjs')),all_seven_goals:'OPEN',goal:'ACTIVE',weekly_usage_percent:8};
fs.writeFileSync(path.join(dir,'checkpoint-verification.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
