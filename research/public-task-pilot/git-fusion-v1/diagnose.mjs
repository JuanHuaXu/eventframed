// Explicitly post-hoc localization, not another confirmation or adoption gate.
import fs from 'node:fs';
import crypto from 'node:crypto';
const root='research/public-task-pilot/git-fusion-v1/';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const rawBytes=fs.readFileSync(root+'raw.jsonl');
const raw=rawBytes.toString('utf8').trim().split('\n').map(JSON.parse).filter(r=>r.type==='case');
const scored=JSON.parse(fs.readFileSync(root+'scored.json','utf8'));
const queries=JSON.parse(fs.readFileSync(root+'queries.json','utf8'));
const oracle=JSON.parse(fs.readFileSync(root+'oracle.json','utf8'));
const out=[];
for(const o of oracle.filter(o=>o.target)){
  const rows=Object.fromEntries(raw.filter(r=>r.case===o.case_id).map(r=>[r.arm,r]));
  const top=r=>r.packed[0]?.id===o.target,surv=r=>r.packed.some(v=>v.id===o.target);
  if(top(rows.baseline)&&top(rows.rrf)&&top(rows.lexical)&&surv(rows.baseline))continue;
  const incumbent=rows.incumbent.order.map(v=>v.id),i=incumbent.indexOf(o.target);
  const ranks=rows.rrf.lexical.map((v,i)=>({v,i})).sort((a,b)=>b.v-a.v||a.i-b.i);
  out.push({case:o.case_id,question:queries.find(q=>q.case_id===o.case_id).question,target:o.target,answer:o.answer,split:o.split,family:o.family,
    target_incumbent_rank:i+1,target_lexical_rank:ranks.findIndex(v=>v.i===i)+1,
    target_coverage:rows.rrf.lexical[i],baseline_first_coverage:rows.rrf.lexical[0],
    baseline_top1:top(rows.baseline),lexical_top1:top(rows.lexical),rrf_top1:top(rows.rrf),
    baseline_survival:surv(rows.baseline),rrf_survival:surv(rows.rrf),protected_survival:surv(rows.protected),
    rrf_first:rows.rrf.packed[0]?.id,rrf_target_rank:rows.rrf.packed.findIndex(v=>v.id===o.target)+1,
    suppressed:rows.rrf.correlated_suppressed});
}
const report={status:'post-hoc descriptive only; no fit, rerun, threshold or gate change',raw_hash:hash(rawBytes),scored_hash:hash(fs.readFileSync(root+'scored.json')),script_hash:hash(fs.readFileSync(root+'diagnose.mjs')),technical_audit:scored.technical,cases:out};
fs.writeFileSync(root+'diagnostic.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({losses:out.filter(v=>v.baseline_top1&&!v.rrf_top1),lexical_repaired:out.filter(v=>!v.baseline_top1&&v.lexical_top1),packet_repaired:out.filter(v=>!v.baseline_survival&&v.rrf_survival)},null,2));
