// Stream large raw artifacts; never overwrite original trial/audit evidence.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import{execFileSync}from'node:child_process';
const hash=x=>crypto.createHash('sha256').update(x).digest('hex'),need=(ok,why)=>{if(!ok)throw Error(why)};
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const names=fs.readdirSync('research').filter(d=>/(witness-v2[3-6]|boundary-v2[7-9]|archive-(lifecycle|requests)-v3[0-2]|batch-archive-v33|archive-cost-v34|metadata-projection-v35|metadata-core-v36)$/.test(d));
need(names.length===14,'missing frozen lineage');const checked=[];
for(const name of names){const dir=path.join('research',name),f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
 for(const[p,h]of Object.entries(f.files))need(await fileHash(p)===h,'source changed '+p);
 const runPath=path.join(dir,'run.json');let run=null;if(fs.existsSync(runPath)){run=JSON.parse(fs.readFileSync(runPath));if(run.raw_sha256){const candidates=['raw.json','raw.ndjson'].map(p=>path.join(dir,p)).filter(p=>fs.existsSync(p));need(candidates.length===1,'raw ambiguity '+name);need(await fileHash(candidates[0])===run.raw_sha256,'raw changed '+name)}}
 checked.push({name,source_files:Object.keys(f.files).length,run_code:run?.code??null,raw_sha256:run?.raw_sha256??null});
}
const v35='research/metadata-projection-v35',supplement=JSON.parse(fs.readFileSync(path.join(v35,'supplement.json')));
need(supplement.original_audit_exit===1&&supplement.controls.length===23&&supplement.trials.length===16&&supplement.trials.every(r=>!r.pass),'V35 evidence');
need(await fileHash(path.join(v35,'audit-attempt.log'))===supplement.original_audit_log_sha256,'V35 failure changed');
need(await fileHash('research/metadata-projection-v35-supplement.mjs')===supplement.supplement_source_sha256,'supplement changed');
const dir='research/metadata-core-v36',audit=JSON.parse(fs.readFileSync(path.join(dir,'audit.json')));
need(audit.trials.length===16&&audit.controls.length===31&&audit.trials.every(r=>!r.pass),'V36 evidence');
need(audit.raw_sha256===await fileHash(path.join(dir,'raw.ndjson')),'V36 audit/raw');
const pre=JSON.parse(fs.readFileSync(path.join(dir,'technical-prefreeze/checks.json')));need(pre.changed.length===0&&pre.checks.every(x=>x.code===0),'preflight');
for(let i=0;i<pre.checks.length;i++)need(await fileHash(path.join(dir,`technical-prefreeze/check-${i}.log`))===pre.checks[i].log_sha256,'preflight log');
const profile=JSON.parse(fs.readFileSync(path.join(dir,'profile.json')));need(profile.code===0,'profile nonterminal/failed');
for(const v of Object.values(profile.files))need(await fileHash(path.join(dir,v.path))===v.sha256,'profile file changed');
need(await fileHash('research/metadata-core-v36-profile.mjs')===profile.script_sha256,'profile script changed');
for(const kind of profile.summary_files)need(fs.statSync(path.join(dir,`profile-${kind}.txt`)).size>0,'profile summary missing');
const diff=execFileSync('git',['diff','--numstat'],{encoding:'utf8'}).trim();
const expected='48\t0\t.learnings/ERRORS.md\n1\t0\tinternal/model/api.go\n21\t4\tinternal/productioneval/codex.go\n105\t0\tinternal/productioneval/codex_test.go\n25\t0\tinternal/service/service.go\n5\t1\tinternal/store/libravdbstore/store.go';need(diff===expected,'preexisting tracked diff changed');
const report={time:new Date().toISOString(),type:'checkpoint-readback',checked,source_and_raw_unchanged:true,v35_preserved_auditor_exit:1,v35_posthoc_controls:23,v36_prospective_controls:31,v36_all16adoption_fail:true,v36_profile_code:0,tracked_diff:diff,script_sha256:hash(fs.readFileSync('research/metadata-core-v36-checkpoint-verify.mjs')),all_seven_goals:'OPEN',goal:'ACTIVE',weekly_usage_percent:9};
fs.writeFileSync(path.join(dir,'checkpoint-verification.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
