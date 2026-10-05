// Preparation-only evidence-family split. Gold relations never enter serving.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import{fileURLToPath}from'node:url';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
export function group(train,test){
 const parent=new Map();function root(x){if(!parent.has(x))parent.set(x,x);let r=x;while(parent.get(r)!==r)r=parent.get(r);while(x!==r){const next=parent.get(x);parent.set(x,r);x=next}return r}
 const querySplit=new Map(),testQueries=new Set();
 for(const[rows,split]of [[train,'train'],[test,'test']])for(const r of rows){assert(r.query&&r.document&&r.score===1);assert(!querySplit.has(r.query)||querySplit.get(r.query)===split,'official query split overlap');querySplit.set(r.query,split);if(split==='test')testQueries.add(r.query);const a=root('q:'+r.query),b=root('d:'+r.document);parent.set(a,b)}
 const families=new Map();for(const k of parent.keys()){const r=root(k);if(!families.has(r))families.set(r,[]);families.get(r).push(k)}
 const components=[...families.values()].map(nodes=>{const queries=nodes.filter(n=>n.startsWith('q:')).map(n=>n.slice(2)).sort(),documents=nodes.filter(n=>n.startsWith('d:')).map(n=>n.slice(2)).sort(),id=hash(JSON.stringify({queries,documents})),confirmation=queries.some(q=>testQueries.has(q));return{id,queries,documents,containsOfficialTest:confirmation}}).sort((a,b)=>a.id.localeCompare(b.id));
 const splits={fit:[],calibration:[],confirmation:[],excludedOfficialTrain:[]},units={fit:[],calibration:[],confirmation:[]};
 for(const c of components){if(c.containsOfficialTest){splits.confirmation.push(...c.queries.filter(q=>testQueries.has(q)));splits.excludedOfficialTrain.push(...c.queries.filter(q=>!testQueries.has(q)));units.confirmation.push(c.id)}else{const split=parseInt(c.id.slice(0,8),16)%3===0?'calibration':'fit';splits[split].push(...c.queries);units[split].push(c.id)}}
 for(const a of Object.values(splits))a.sort();assert.equal(splits.confirmation.length,testQueries.size);assert.equal(Object.values(splits).reduce((s,a)=>s+a.length,0),querySplit.size);
 return{components,splits,units};
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 const dir='research/public-task-pilot/scifact-v1',manifest=JSON.parse(fs.readFileSync(dir+'/source.json'));
 for(const[p,h]of Object.entries(manifest.artifacts))assert.equal(hash(fs.readFileSync(dir+'/'+p)),h);
 const r=group(JSON.parse(fs.readFileSync(dir+'/labels-train.json')),JSON.parse(fs.readFileSync(dir+'/labels-test.json'))),owners=new Map();
 for(const[split,ids]of Object.entries(r.units))for(const id of ids){const c=r.components.find(c=>c.id===id);for(const d of c.documents){assert(!owners.has(d)||owners.get(d)===split,'evidence family crosses split');owners.set(d,split)}}
 const report={time:new Date().toISOString(),sourceManifestSHA256:hash(fs.readFileSync(dir+'/source.json')),scriptSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),...r,counts:{components:r.components.length,queries:Object.fromEntries(Object.entries(r.splits).map(([k,v])=>[k,v.length])),units:Object.fromEntries(Object.entries(r.units).map(([k,v])=>[k,v.length]))},allOfficialTestQueriesRetained:true,exactEvidenceOverlapPrevented:true,topicIndependenceUnproven:true,partitionUsesIdentitiesNotModelPerformance:true,noModelRuns:true,noFitting:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
 fs.writeFileSync(dir+'/split-plan.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({counts:report.counts,allOfficialTestQueriesRetained:true,exactEvidenceOverlapPrevented:true,topicIndependenceUnproven:true,noModelRuns:true},null,2));
}
