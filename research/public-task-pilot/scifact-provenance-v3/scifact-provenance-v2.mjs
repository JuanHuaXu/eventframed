// Public source-only audit. No model execution, fitting, or production access.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
export function checkClaimsV2(queries,rels,original){
 const q=new Map(queries.map(x=>[x.id,x.text])),byID=new Map();
 for(const c of original){assert(!byID.has(String(c.id)),'duplicate source claim');byID.set(String(c.id),c)}
 const ids=new Set(rels.map(x=>x.query));assert.equal(ids.size,original.length,'partition coverage');
 for(const id of ids){const c=byID.get(id);assert(c,'unknown original claim');assert.equal(q.get(id),c.claim,'source claim text');const actual=rels.filter(x=>x.query===id).map(x=>x.document).sort();assert.equal(new Set(actual).size,actual.length,'duplicate relation');assert(rels.filter(x=>x.query===id).every(x=>x.score===1),'relevance domain');const expected=Object.keys(c.evidence).sort();assert.deepEqual(actual,expected,'evidence relation identity')}
 return {claims:ids.size,relations:rels.length,textAndEvidenceExact:true};
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 const root='research/public-task-pilot/scifact-provenance-v2',old='research/public-task-pilot/scifact-v1';assert(!fs.existsSync(root),'exclusive provenance root');fs.mkdirSync(root,{mode:0o700});
 function save(name,b){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
 const started=new Date().toISOString(),t=performance.now(),sources={};
 async function acquire(name,url,limit){const r=await fetch(url,{signal:AbortSignal.timeout(120000)});assert(r.ok,'public source acquisition '+name);let size=0;const chunks=[];for await(const c of r.body){size+=c.length;assert(size<=limit,'source size bound');chunks.push(c)}const b=Buffer.concat(chunks);save(name,b);sources[name]={url,sha256:hash(b),bytes:b.length};return b}
 const license=await acquire('LICENSE-upstream.md','https://raw.githubusercontent.com/allenai/scifact/master/LICENSE.md',65536);
 assert(license.toString().includes('CC BY 4.0')&&license.toString().includes('ODC-By 1.0'),'declared license categories');
 await acquire('download-upstream.sh','https://raw.githubusercontent.com/allenai/scifact/master/script/download-data.sh',65536);
 const archive=await acquire('data.tar.gz','https://scifact.s3-us-west-2.amazonaws.com/release/latest/data.tar.gz',32<<20);
 const names=execFileSync('tar',['-tzf',root+'/data.tar.gz'],{encoding:'utf8',maxBuffer:2<<20}).trim().split('\n');
 const members={};const data={};
 for(const name of ['data/claims_train.jsonl','data/claims_dev.jsonl','data/claims_test.jsonl']){assert.equal(names.filter(x=>x===name).length,1,'explicit archive member');const b=execFileSync('tar',['-xOzf',root+'/data.tar.gz',name],{maxBuffer:32<<20});members[name]=hash(b);data[name]=b.toString('utf8').trim().split('\n').map(x=>JSON.parse(x))}
 const oldReport=JSON.parse(fs.readFileSync(old+'/source.json')),prepared={};
 for(const name of ['queries.json','labels-train.json','labels-test.json']){const b=fs.readFileSync(old+'/'+name);assert.equal(hash(b),oldReport.artifacts[name],'unchanged BEIR preparation');prepared[name]={sha256:hash(b),bytes:b.length}}
 const read=n=>JSON.parse(fs.readFileSync(old+'/'+n));
 const train=checkClaimsV2(read('queries.json'),read('labels-train.json'),data['data/claims_train.jsonl']);
 const dev=checkClaimsV2(read('queries.json'),read('labels-test.json'),data['data/claims_dev.jsonl']);
 assert.equal(train.claims,809);assert.equal(dev.claims,300);
 assert(data['data/claims_test.jsonl'].every(x=>!Object.hasOwn(x,'evidence')),'upstream test is unlabeled');
 assert.equal(data['data/claims_test.jsonl'].length,300);
 const report={start:started,end:new Date().toISOString(),wallMS:performance.now()-t,sources,members,prepared,sourceArchiveSHA256:hash(archive),scriptSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),mapping:{BEIRTrain:train,BEIRTestIsUpstreamDev:dev,upstreamUnlabeledTestClaims:300},licenseMetadata:{claimsAndAnnotations:'CC BY 4.0',abstracts:'ODC-By 1.0 (S2ORC)',code:'Apache 2.0',datasetNotRelicensedByEventframed:true,redistributionNotPerformed:true},noModelRuns:true,noFitting:true,confirmationNotConsumedByPrediction:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
 save('report.json',Buffer.from(JSON.stringify(report,null,2)+'\n'));console.log(JSON.stringify(report,null,2));
}
