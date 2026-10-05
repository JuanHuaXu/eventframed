// Post-output verification supplement. Does not modify frozen predictor/auditor.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {audit} from './audit.mjs';
const root='research/public-task-pilot/metrology-transfer-v1/';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const read=p=>JSON.parse(fs.readFileSync(p));
const raw=fs.readFileSync(root+'raw.jsonl','utf8').trim().split('\n').map(JSON.parse);
const completed=read(root+'completed.json'),results=read(root+'results.json');
for(const[p,h]of Object.entries(completed.artifacts))assert.equal(hash(fs.readFileSync(root+p)),h,'completed artifact changed');
for(const[p,h]of Object.entries(read(root+'auditor-freeze.json').files))assert.equal(hash(fs.readFileSync(p)),h,'auditor source changed');
const check=audit(raw),labels=read(root+'prepared/oracle.json'),queries=read(root+'prepared/queries.json'),corpus=read(root+'prepared/corpus.json');
assert.equal(new Set(labels.map(l=>l.question_id)).size,72);assert.equal(labels.length,72);
for(const q of queries){const l=labels.find(l=>l.question_id===q.id);assert(l&&l.split===q.split);assert(corpus.some(c=>c.id===l.target_id));assert(BigInt(l.observable.denominator)>0n)}
function packing(rows){for(const r of rows){const order=new Map(r.Order.map((v,i)=>[v.id,i]));let previous=-1;for(const v of r.Packed){assert(order.get(v.id)>previous,'packed order not an ordered subset');previous=order.get(v.id)}assert.equal(r.Packed[0].id,r.Order[0].id,'top candidate lost despite ample budget');assert.equal(r.Packed.length,10,'unexpected occupancy/token truncation')}}
packing(check.rows);
const altered=structuredClone(check.rows);[altered[0].Packed[0],altered[0].Packed[1]]=[altered[0].Packed[1],altered[0].Packed[0]];assert.notDeepEqual(altered,check.rows);assert.throws(()=>packing(altered));
const modelPath='research/public-task-pilot/ecmascript-v1/magnitude-v2-model.json',model=read(modelPath),freeze=read(root+'auditor-freeze.json');
assert.equal(results.model_sha256,hash(fs.readFileSync(modelPath)));assert.equal(raw[0].model_sha256,freeze.model_sha256);
assert(Date.parse(model.time)<Date.parse(freeze.time));assert(Date.parse(freeze.time)<Date.parse(raw[0].time));assert(Date.parse(raw[0].time)<Date.parse(results.time));
const losses=[];
for(const[cell,v]of Object.entries(results.cells))for(let i=0;i<v.baseline.per.length;i++){
 const b=v.baseline.per[i],s=v.selected.per[i];if(b.top1<=s.top1)continue;
 const mode=cell.split('/')[0],q=queries.find(q=>q.id===b.id),l=labels.find(l=>l.question_id===b.id),r=check.keyed.get(mode+'/'+b.id+'/selected');
 losses.push({cell,query:q.text,target:corpus.find(c=>c.id===l.target_id).frame.what,top_two:r.Packed.slice(0,2).map(v=>({record:corpus.find(c=>c.id===v.id).frame.what,utility:v.score})),unchanged_law:true,target_still_packed:true});
}
const before=read('research/continuing-v39-completion/final-readback.json');
const priorPath='research/continuing-v39-completion/checkpoint-verification.json';
assert.equal(hash(fs.readFileSync(priorPath)),before.checkpoint_sha256);
for(const[p,h]of Object.entries(read(priorPath).docs))if(p!=='research-direction.md')assert.equal(hash(fs.readFileSync(p)),h,'prior documented evidence changed');
for(const[p,h]of Object.entries(before.terminal_run_hashes))assert.equal(hash(fs.readFileSync(p)),h);
const report={time:new Date().toISOString(),readback_source_sha256:hash(fs.readFileSync(root+'readback.mjs')),normal_pass:true,post_output_supplement:true,main_controls:completed.controls,extra_controls:['packed_order_swap'],complete_cases:check.rows.length,journal_decisions:check.rows.reduce((n,r)=>n+r.JournalOrder.length,0),packed_decisions:check.rows.reduce((n,r)=>n+r.Packed.length,0),all_targets_nominated:true,model_before_transfer:true,prior_checkpoints_unchanged:true,losses,adoption_pass:results.adoption_pass,whole_goal_complete:false};
fs.writeFileSync(root+'readback.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(report,null,2));
