// Independent relation-first reconstruction from the original public archive.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';import{execFileSync}from'node:child_process';
const root='research/public-task-pilot/scifact-provenance-v3',old='research/public-task-pilot/scifact-v1',raw='research/public-task-pilot/scifact-provenance-v2/data.tar.gz';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),r=JSON.parse(fs.readFileSync(root+'/report.json'));
for(const[p,x]of Object.entries(r.sources))assert.equal(hash(fs.readFileSync(p)),x.sha256,'source '+p);
for(const[p,x]of Object.entries(r.prepared))assert.equal(hash(fs.readFileSync(old+'/'+p)),x.sha256,'prepared '+p);
const query=new Map(JSON.parse(fs.readFileSync(old+'/queries.json')).map(x=>[x.id,x.text]));
const checks={};
for(const[part,label,key]of [['train','train','BEIRTrain'],['dev','test','BEIRTestIsUpstreamLabeledDev']]){
 const name='data/claims_'+part+'.jsonl',b=execFileSync('tar',['-xOzf',raw,name],{maxBuffer:32<<20});assert.equal(hash(b),r.members[name]);const claims=b.toString().trim().split('\n').map(JSON.parse),byID=new Map(claims.map(c=>[String(c.id),c]));assert.equal(byID.size,claims.length);
 const rels=JSON.parse(fs.readFileSync(old+'/labels-'+label+'.json')),seen=new Set(),used=new Set();let withEvidence=0,withoutEvidence=0;
 for(const x of [...rels].reverse()){
  const c=byID.get(x.query);assert(c);assert.equal(query.get(x.query),c.claim);assert.equal(x.score,1);const pair=x.query+'\0'+x.document;assert(!seen.has(pair));seen.add(pair);assert(c.cited_doc_ids.some(n=>String(n)===x.document));used.add(x.query);if(Object.hasOwn(c.evidence,x.document))withEvidence++;else withoutEvidence++;
 }
 for(const c of claims)for(const n of c.cited_doc_ids)assert(seen.has(String(c.id)+'\0'+String(n)),'missing citation relation');
 assert.equal(used.size,claims.length);
 const value={claims:claims.length,relations:rels.length,textAndCitationsExact:true,emptyEvidenceClaims:claims.filter(c=>Object.keys(c.evidence).length===0).length,citationRelationsWithEvidence:withEvidence,citationRelationsWithoutEvidence:withoutEvidence};assert.deepEqual(value,r.mapping[key]);checks[part]=value;
}
const test=execFileSync('tar',['-xOzf',raw,'data/claims_test.jsonl'],{maxBuffer:32<<20});assert.equal(hash(test),r.members['data/claims_test.jsonl']);const unlabeled=test.toString().trim().split('\n').map(JSON.parse);assert.equal(unlabeled.length,r.mapping.upstreamUnlabeledTestClaims);assert(unlabeled.every(c=>!Object.hasOwn(c,'evidence')));
const out={time:new Date().toISOString(),reportSHA256:hash(fs.readFileSync(root+'/report.json')),auditorSHA256:hash(fs.readFileSync('research/public-task-pilot/scifact-provenance-v3-audit.mjs')),independentRelationFirstReconstruction:true,checks,upstreamUnlabeledTestClaims:unlabeled.length,preparedDataUnchanged:true,noPredictionOrFitting:true,allSevenWholeGoals:'OPEN'};
const fd=fs.openSync(root+'/audit.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(out,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}console.log(JSON.stringify(out,null,2));
