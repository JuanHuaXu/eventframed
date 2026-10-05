// Independent outcome audit. Runtime collector never imports this module.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/public-task-pilot/ecmascript-v1/';
const read=p=>JSON.parse(fs.readFileSync(root+p)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const stop=new Set('a an the is was of to at in on and what which when according retained records date s'.split(' '));
const words=s=>new Set(s.toLowerCase().split(/[^a-z0-9]+/).filter(s=>s&&!stop.has(s)));
function coverage(q, fields){const query=words(q),union=words(fields.join(' '));assert(query.size>0);return [...query].filter(w=>union.has(w)).length/query.size}
export function audit(raw,phase,weight=0){
 const queries=read('prepared/queries.json').filter(q=>(phase==='fit')===(q.split==='fit'));
 const corpus=read('prepared/corpus.json'),expectedFrames=Object.fromEntries(corpus.map(f=>[f.id,['who','what','where','when','why','how'].map(k=>f.frame[k])]));
 const rows=raw.filter(r=>r.Type==='case'),header=raw[0],footer=raw.at(-1);
 assert.equal(header.type,'header');assert.equal(header.phase,phase);assert.equal(footer.type,'footer');assert.equal(footer.complete,216);assert.equal(rows.length,216);assert.equal(footer.hashes_verified,true);
 for(const [p,h]of Object.entries(header.files))assert.equal(hash(fs.readFileSync(p)),h,'changed source '+p);
 const arms=phase==='fit'?['baseline','incumbent','w025','w050','w075','w100']:['baseline','incumbent','selected'];
 const weights={baseline:0,incumbent:0,w025:.25,w050:.5,w075:.75,w100:1,selected:weight};
 const keyed=new Map();let frameRecords=0;
 for(const r of rows){assert(queries.some(q=>q.id===r.Case&&q.split===r.Split),'foreign query');assert(arms.includes(r.Arm));const k=r.Case+'/'+r.Arm;assert(!keyed.has(k),'duplicate case');keyed.set(k,r);assert.equal(r.JournalStored,true);assert.equal(r.Weight,weights[r.Arm]);assert(r.RecallNS>0&&r.ImportNS>0);assert.equal(r.MemoBefore.misses,r.MemoAfter.misses,'uncached timed Recall');assert.equal(r.Order.length,54);assert(r.Packed.length>0&&r.Packed.length<=10);assert.equal(new Set(r.Order.map(v=>v.id)).size,54);assert.equal(new Set(r.Packed.map(v=>v.id)).size,r.Packed.length);
  for(const v of [...r.Order,...r.Packed]){assert(v.id in expectedFrames,'foreign candidate');assert(Number.isFinite(v.score)&&v.score>=0&&v.score<=1);assert(v.law&&Object.values(v.law).every(Number.isFinite));}
  const byID=new Map(r.Order.map(v=>[v.id,v]));for(const v of r.Packed){assert(byID.has(v.id),'packed outside frontier');assert.deepEqual(v.law,byID.get(v.id).law);assert.equal(v.score,byID.get(v.id).score)}
  if(r.Frames){assert.deepEqual(r.Frames,expectedFrames,'structured import changed');frameRecords++}
 }
 assert.equal(frameRecords,1);
 for(const q of queries){const baseline=keyed.get(q.id+'/baseline');assert(baseline,'missing baseline');const idx=baseline.Order.map(v=>v.id);const byID=new Map(baseline.Order.map(v=>[v.id,v]));
  for(const a of arms){const r=keyed.get(q.id+'/'+a);assert(r,'missing arm');assert.deepEqual([...r.Order.map(v=>v.id)].sort(),[...idx].sort(),'nomination changed');for(const v of r.Order)assert.deepEqual(v.law,byID.get(v.id).law,'law changed');
   if(a==='baseline'){assert.equal(r.HookCalls,0);assert.equal(r.Lexical,null);continue}
   assert.equal(r.HookCalls,1);assert.equal(r.Lexical.length,54);assert.equal(r.HookScores.length,54);
   const utility=idx.map((id,i)=>{const x=coverage(q.text,expectedFrames[id]);assert(Math.abs(x-r.Lexical[i])<1e-12,'lexical/metadata/order mismatch');const s=(1-weights[a])/(i+1)+weights[a]*x;assert(Math.abs(s-r.HookScores[i])<1e-12,'formula mismatch');return {id,score:s,i}}).sort((a,b)=>b.score-a.score||a.i-b.i);
   assert.deepEqual(r.Order.map(v=>v.id),utility.map(v=>v.id),'stable sort mismatch');for(let i=0;i<54;i++)assert(Math.abs(r.Order[i].score-utility[i].score)<1e-12,'journal score');
   if(a==='incumbent'){assert.deepEqual(r.Packed.map(v=>v.id),baseline.Packed.map(v=>v.id),'passthrough packet changed')}
  }
 }
 return {keyed,queries,rows,arms,source_files:Object.keys(header.files).length};
}
export function metrics(check, labels, split, arm){
 const per=[],cluster=new Map();for(const q of check.queries.filter(q=>q.split===split)){
  const label=labels.find(l=>l.question_id===q.id);assert(label,'missing label');const r=check.keyed.get(q.id+'/'+arm),ids=r.Packed.map(v=>v.id),rank=ids.indexOf(label.target_id)+1;
  const v={id:q.id,cluster:label.cluster,wording:q.wording,top1:Number(ids[0]===label.target_id),survival:Number(rank>0),rr:rank>0?1/rank:0,nominated:Number(r.Order.some(v=>v.id===label.target_id))};per.push(v);if(!cluster.has(v.cluster))cluster.set(v.cluster,[]);cluster.get(v.cluster).push(v);
 }
 assert.equal(per.length,36);assert.equal(cluster.size,6);
 return {top1:per.reduce((a,v)=>a+v.top1,0),survival:per.reduce((a,v)=>a+v.survival,0),nomination:per.reduce((a,v)=>a+v.nominated,0),mean_reciprocal_rank:per.reduce((a,v)=>a+v.rr,0)/36,per,clusters:Object.fromEntries([...cluster].map(([k,v])=>[k,v.reduce((a,r)=>a+r.top1,0)/v.length]))};
}
export function comparison(base,selected){const vector=Object.keys(base.clusters).map(k=>selected.clusters[k]-base.clusters[k]),mean=vector.reduce((a,v)=>a+v,0)/6,se=Math.sqrt(vector.reduce((a,v)=>a+(v-mean)**2,0)/5/6),pairs=base.per.map((v,i)=>[v,selected.per[i]]);const gains=pairs.filter(([a,b])=>b.top1>a.top1).length,losses=pairs.filter(([a,b])=>b.top1<a.top1).length,literalLosses=pairs.filter(([a,b])=>a.wording==='literal'&&b.top1<a.top1).length,survivalLosses=pairs.filter(([a,b])=>b.survival<a.survival).length;const positive=vector.filter(v=>v>0).length,negative=vector.filter(v=>v<0).length,n=positive+negative;let sign=0;for(let k=positive;k<=n;k++){let c=1;for(let j=1;j<=k;j++)c=c*(n-j+1)/j;sign+=c/2**n}const lower=mean-2.015048373*se;return {gains,losses,literal_losses:literalLosses,survival_losses:survivalLosses,cluster_vector:vector,cluster_mean:mean,one_sided95_lower:lower,two_sided95:[mean-2.570581836*se,mean+2.570581836*se],exact_cluster_sign_p:sign,pass:gains-losses>=2&&literalLosses===0&&survivalLosses===0&&lower>0}}
function timing(check){const out={};for(const arm of check.arms){const rows=check.rows.filter(r=>r.Arm===arm),quant=(values,p)=>values.sort((a,b)=>a-b)[Math.ceil(p*values.length)-1];out[arm]={cases:rows.length,recall_median_ms:quant(rows.map(r=>r.RecallNS/1e6),.5),recall_p99_ms:quant(rows.map(r=>r.RecallNS/1e6),.99),hook_p99_ms:quant(rows.map(r=>r.HookNS/1e6),.99)}}return out}
if(process.argv[1]?.endsWith('/magnitude-supplement.mjs')){
 const phase=process.argv[2],rawPath=process.argv[3],output=process.argv[4];const raw=fs.readFileSync(rawPath,'utf8').trim().split('\n').map(l=>JSON.parse(l));
 const weight=phase==='heldout'?read('magnitude-model.json').weight:0,check=audit(raw,phase,weight);let report;
 if(phase==='fit'){const labels=read('magnitude-fit-labels.json');assert.equal(labels.length,36);assert(labels.every(l=>l.split==='fit'));const scores=check.arms.filter(a=>a!=='baseline').map(a=>({arm:a,weight:{incumbent:0,w025:.25,w050:.5,w075:.75,w100:1}[a],...metrics(check,labels,'fit',a)})).sort((a,b)=>b.top1-a.top1||b.survival-a.survival||a.weight-b.weight);report={time:new Date().toISOString(),fit_only:true,weight:scores[0].weight,chosen:scores[0].arm,scores,raw_sha256:hash(fs.readFileSync(rawPath)),technical_pass:true,timing:timing(check),whole_goal_complete:false}}
 else{const labels=read('prepared/oracle.json'),splits={};for(const split of ['design','confirmation']){const b=metrics(check,labels,split,'baseline'),s=metrics(check,labels,split,'selected');splits[split]={baseline:b,selected:s,comparison:comparison(b,s)}}report={time:new Date().toISOString(),weight,raw_sha256:hash(fs.readFileSync(rawPath)),model_sha256:hash(fs.readFileSync(root+'magnitude-model.json')),technical_pass:true,splits,adoption_pass:Object.values(splits).every(s=>s.comparison.pass),timing:timing(check),whole_goal_complete:false}}
 fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({phase,technical_pass:report.technical_pass,weight,chosen:report.chosen,adoption_pass:report.adoption_pass},null,2));
}

