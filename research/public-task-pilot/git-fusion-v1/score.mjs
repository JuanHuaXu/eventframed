// Post-collection auditor: no candidate fitting or source edits.
import fs from 'node:fs';
import crypto from 'node:crypto';
const root='research/public-task-pilot/git-fusion-v1/';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const load=p=>JSON.parse(fs.readFileSync(p,'utf8'));
const mean=x=>x.reduce((s,v)=>s+v,0)/x.length;
const equal=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
const assert=(v,m)=>{if(!v)throw Error(m);};
const arms=['baseline','incumbent','lexical','rrf','protected'];
const words=s=>new Set(s.toLowerCase().split(/[^a-z0-9]+/).filter(w=>w&&!new Set('a an the is was of to at in on and what which when according retained records date s'.split(' ')).has(w)));
function interval(x){
  const m=mean(x),variance=x.reduce((s,v)=>s+(v-m)**2,0)/(x.length-1),se=Math.sqrt(variance/x.length);
  return {mean:m,lower95:m-1.894578605*se,two_sided95:[m-2.364624252*se,m+2.364624252*se],clusters:x};
}
function sign(x){const pos=x.filter(v=>v>0).length,neg=x.filter(v=>v<0).length,n=pos+neg;let probability=0,choose=1;for(let k=0;k<=n;k++){if(k>=pos)probability+=choose/2**n;choose=choose*(n-k)/(k+1)};return {pos,neg,ties:x.length-n,p_one_sided:probability};}
function expected(lexical,prefix){
  const order=lexical.map((_,i)=>i).sort((a,b)=>lexical[b]-lexical[a]||a-b),ranks=[];order.forEach((i,j)=>ranks[i]=j+1);
  return lexical.map((_,i)=>{const v=30.5*(1/(61+i)+1/(60+ranks[i]));return prefix? v/2+(i<prefix?.5:0):v;});
}
function audit(raw){
  assert(raw.length===502,'incomplete raw');const header=raw[0],footer=raw.at(-1),rows=raw.slice(1,-1);
  assert(header.type==='header'&&footer.type==='footer'&&footer.complete===500&&footer.hashes_verified&&header.cases===100&&equal(header.arms,arms)&&header.digest==='0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f','header/footer');
  const freeze=load(root+'freeze.json'),af=load(root+'audit-freeze.json');
  assert(equal(header.hashes,freeze),'header freeze mismatch');
  for(const [p,h]of Object.entries({...freeze,...af}))assert(hash(fs.readFileSync(p))===h,`source freeze ${p}`);
  const corpus=load(root+'corpus.json'),queries=load(root+'queries.json'),oracle=load(root+'oracle.json'),facts=load(root+'facts.json');
  assert(corpus.length===48&&queries.length===100&&oracle.length===100&&facts.length===48,'fixture shape');
  const ids=new Set(corpus.map(x=>x.fixture_id)),qs=new Map(queries.map(x=>[x.case_id,x.question]));
  assert(ids.size===48&&qs.size===100,'duplicate input');
  const labels=new Map(oracle.map(x=>[x.case_id,x]));assert(labels.size===100,'duplicate oracle');
  for(const o of oracle){assert(qs.has(o.case_id),'unknown label');if(o.target){const f=corpus.find(f=>f.fixture_id===o.target);assert(f&&f.text.includes(o.answer)&&!qs.get(o.case_id).includes(o.answer),'grounding/answer leak');}}
  const cases=new Map();for(const r of rows){assert(r.type==='case'&&qs.has(r.case)&&arms.includes(r.arm),'unknown case/arm');const key=r.case+'|'+r.arm;assert(!cases.has(key),'duplicate case');cases.set(key,r);}
  const frames=rows.find(r=>r.frames)?.frames;assert(frames&&Object.keys(frames).length===48,'frame coverage');
  const gains=[];let packetSetChanges=0,prefixChecks=0;
  for(const [id,question]of qs){
    const control=cases.get(id+'|incumbent'),base=cases.get(id+'|baseline');assert(control&&base,'missing control');
    const incumbent=control.order.map(v=>v.id);
    assert(equal(base.order.map(v=>v.id),incumbent),'hook incumbent order mismatch');
    assert(equal(base.packed.map(v=>v.id),control.packed.map(v=>v.id)),'incumbent packet mismatch');
    const laws=new Map(base.order.map(v=>[v.id,v.law]));
    for(const arm of arms){
      const r=cases.get(id+'|'+arm);assert(r,'missing arm');
      assert(r.journal_stored&&r.memo_after.misses===r.memo_before.misses&&r.recall_ns>0,'journal/cold Recall');
      for(const set of [r.order,r.packed]){assert(new Set(set.map(v=>v.id)).size===set.length,'duplicate output');for(const v of set){assert(ids.has(v.id)&&Number.isFinite(v.score)&&v.score>=0&&v.score<=1&&equal(v.law,laws.get(v.id)),'ID/score/law invariant');}}
      assert(equal([...r.order.map(v=>v.id)].sort(),[...incumbent].sort()),'nomination drift');
      assert(r.packed.length<=10&&r.packed.every(v=>r.order.some(w=>w.id===v.id)),'packet cap/subset');
      if(arm==='baseline'){assert(r.hook_calls===0&&r.lexical===null&&r.hook_scores===null,'baseline hook');continue;}
      assert(r.hook_calls===1&&r.lexical.length===incumbent.length&&r.hook_scores.length===incumbent.length,'hook shape');
      const terms=words(question),lex=incumbent.map(target=>{const union=new Set(frames[target].flatMap(s=>[...words(s)]));return [...terms].filter(w=>union.has(w)).length/terms.size;});
      assert(r.lexical.every((v,i)=>Math.abs(v-lex[i])<1e-14),'5W1H hook input mismatch');
      const scores=arm==='incumbent'?incumbent.map((_,i)=>1/(i+1)):arm==='lexical'?lex:expected(lex,arm==='protected'?10:0);
      assert(r.hook_scores.every((v,i)=>Math.abs(v-scores[i])<1e-14),'independent score mismatch');
      const ordering=incumbent.map((id,i)=>({id,i})).sort((a,b)=>scores[b.i]-scores[a.i]||a.i-b.i);
      assert(equal(r.order.map(v=>v.id),ordering.map(v=>v.id)),'applied rank mismatch');
      assert(r.order.every(v=>Math.abs(v.score-scores[incumbent.indexOf(v.id)])<1e-14),'journal rank mismatch');
      assert(r.packed.every(v=>Math.abs(v.score-scores[incumbent.indexOf(v.id)])<1e-14),'served rank mismatch');
      if(arm==='protected'){prefixChecks++;assert(equal(r.order.slice(0,10).map(v=>v.id).sort(),incumbent.slice(0,10).sort()),'prefix set');if(!equal(r.packed.map(v=>v.id).sort(),base.packed.map(v=>v.id).sort()))packetSetChanges++;}
    }
    for(const arm of arms){const r=cases.get(id+'|'+arm),o=labels.get(id),rank=o.target?r.packed.findIndex(v=>v.id===o.target):-1;gains.push({...o,arm,top1:rank===0?1:0,survival:rank>=0?1:0,rr:rank>=0?1/(rank+1):0,nominated:o.target&&r.order.some(v=>v.id===o.target)?1:0,packet_count:r.packed.length});}
  }
  const summary={};
  for(const split of ['design','confirmation']){
    summary[split]={};const clusterNames=[...new Set(oracle.filter(o=>o.split===split&&o.target).map(o=>o.family))];assert(clusterNames.length===8,'cluster split');
    for(const arm of arms){
      const data=gains.filter(r=>r.split===split&&r.arm===arm&&r.target),baseline=gains.filter(r=>r.split===split&&r.arm==='baseline'&&r.target),bm=new Map(baseline.map(r=>[r.case_id,r]));
      const paired=data.map(r=>({...r,delta:r.top1-bm.get(r.case_id).top1,survival_delta:r.survival-bm.get(r.case_id).survival}));
      const cluster=clusterNames.map(f=>mean(paired.filter(r=>r.family===f).map(r=>r.delta)));
      const bound=interval(cluster),net=paired.reduce((s,r)=>s+r.delta,0),literalLosses=paired.filter(r=>r.wording==='literal'&&r.delta<0).length,survivalLosses=paired.filter(r=>r.survival_delta<0).length;
      const cells={};for(const wording of ['literal','paraphrase']){const a=data.filter(r=>r.wording===wording);cells[wording]={n:a.length,top1:a.reduce((s,r)=>s+r.top1,0),survival:a.reduce((s,r)=>s+r.survival,0),nominated:a.reduce((s,r)=>s+r.nominated,0),mrr:mean(a.map(r=>r.rr))};}
      summary[split][arm]={cells,net_top1:net,gains:paired.filter(r=>r.delta>0).map(r=>r.case_id),losses:paired.filter(r=>r.delta<0).map(r=>r.case_id),literal_losses:literalLosses,survival_losses:survivalLosses,cluster_bound:bound,sign_test:sign(cluster),screen:net>=2&&literalLosses===0&&survivalLosses===0&&bound.lower95>0?'PASS':'FAIL'};
    }
  }
  const perf={};for(const arm of arms){const r=rows.filter(r=>r.arm===arm),times=r.map(r=>r.recall_ns/1e6).sort((a,b)=>a-b),q=p=>times[Math.ceil(p*times.length)-1];perf[arm]={n:r.length,median_ms:q(.5),p95_ms:q(.95),p99_ms:q(.99),max_ms:q(1),capture_total_ms:r.reduce((s,r)=>s+r.capture_ns/1e6,0),callback_total_ms:r.reduce((s,r)=>s+r.hook_ns/1e6,0)};}
  return {technical:'PASS',quality:summary,performance:perf,memo:footer.memo,protected_prefix_checks:prefixChecks,protected_packet_set_changes:packetSetChanges,absent:gains.filter(r=>!r.target),rows:gains,oracle_hash:hash(fs.readFileSync(root+'oracle.json'))};
}
if(process.argv[2]==='--self-test'){
  assert(equal(expected([0,.8,1,.2],0),[0,.8,1,.2].map((_,i)=>30.5*(1/(61+i)+1/(60+[4,2,1,3][i])))),'rank oracle');
  assert(sign([1,1,1,1,1,1,1,1]).p_one_sided===1/256,'sign test');
  assert(interval([0,0,0,0,0,0,0,0]).lower95===0,'zero interval');console.log('scorer arithmetic PASS');
}else{
  assert(process.argv.length===4,'usage: score.mjs RAW.jsonl NEW.json (output exclusive)');
  const bytes=fs.readFileSync(process.argv[2]),raw=bytes.toString('utf8').trim().split('\n').map(s=>JSON.parse(s)),out=audit(raw);out.raw_hash=hash(bytes);
  const failures=[];
  for(const [name,mutate]of [
    ['score',r=>{r[2].hook_scores[0]=.001;}],
    ['law',r=>{r[2].order[0].law.useful=.001;}],
    ['duplicate',r=>{r[2].order[1]=r[2].order[0];}],
    ['missing',r=>r.splice(10,1)],
    ['warm',r=>{r[2].memo_after.misses++;}],
    ['freeze',r=>{r[0].hashes['internal/researchfusion/fusion.go']='wrong';}],
    ['journal',r=>{r[2].journal_stored=false;}],
    ['lexical',r=>{r[2].lexical[0]=.111111111;}],
    ['order',r=>{r[2].order.reverse();}]
  ]){const r=structuredClone(raw);mutate(r);let rejected=false;try{audit(r)}catch{rejected=true}assert(rejected,`negative control accepted ${name}`);failures.push(name);}
  out.rejected_negative_controls=failures;
  fs.writeFileSync(process.argv[3],JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});
  console.log(JSON.stringify({technical:out.technical,quality:out.quality,performance:out.performance,memo:out.memo,protected_packet_set_changes:out.protected_packet_set_changes,rejected_negative_controls:failures},null,2));
}
