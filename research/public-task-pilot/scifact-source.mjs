// Public benchmark acquisition only; no private data, fitting, model or installs.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import{execFileSync}from'node:child_process';import{fileURLToPath}from'node:url';
const URL='https://public.ukp.informatik.tu-darmstadt.de/thakur/BEIR/datasets/scifact.zip',MD5='5f7d1de60b170fc8027bb7898e2efca1';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex');
export function jsonLines(bytes){return bytes.toString('utf8').trim().split('\n').map(line=>JSON.parse(line))}
export function qrels(bytes){
 const lines=bytes.toString('utf8').trim().split(/\r?\n/);assert.equal(lines.shift(),'query-id\tcorpus-id\tscore');const seen=new Set();
 return lines.map(line=>{const parts=line.split('\t');assert.equal(parts.length,3);const[query,document,s]=parts;assert(query&&document);assert(/^[0-9]+$/.test(s));const key=query+'\0'+document;assert(!seen.has(key));seen.add(key);return{query,document,score:Number(s)}});
}
export function validate(corpus,queries,train,test){
 const docs=new Set(),qs=new Set();for(const d of corpus){assert(typeof d._id==='string'&&d._id&&typeof d.title==='string'&&typeof d.text==='string');assert(d.title.trim()&&d.text.trim());assert(!docs.has(d._id));docs.add(d._id)}
 for(const q of queries){assert(typeof q._id==='string'&&q._id&&typeof q.text==='string'&&q.text.trim());assert(!qs.has(q._id));qs.add(q._id)}
 const a=new Set(),b=new Set();for(const[r,out]of [[train,a],[test,b]])for(const x of r){assert(qs.has(x.query)&&docs.has(x.document));assert.equal(x.score,1);out.add(x.query)}
 assert([...a].every(k=>!b.has(k)),'official train/test query overlap');
 const trainDocs=new Set(train.map(x=>x.document)),testDocs=new Set(test.map(x=>x.document));
 return{documents:docs.size,queries:qs.size,trainQueries:a.size,testQueries:b.size,trainQrels:train.length,testQrels:test.length,sharedEvidenceDocuments:[...trainDocs].filter(k=>testDocs.has(k)).length,unlabeledQueries:[...qs].filter(k=>!a.has(k)&&!b.has(k)).length};
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 const root='research/public-task-pilot/scifact-v1';assert(!fs.existsSync(root),'exclusive public source root');fs.mkdirSync(root,{mode:0o700});
 const start=new Date().toISOString(),begin=performance.now();const response=await fetch(URL,{signal:AbortSignal.timeout(120000)});assert(response.ok,'public benchmark download');const chunks=[];let size=0;
 for await(const chunk of response.body){size+=chunk.length;assert(size<=32<<20,'archive size bound');chunks.push(chunk)}const archive=Buffer.concat(chunks);
 assert.equal(crypto.createHash('md5').update(archive).digest('hex'),MD5,'official BEIR checksum');fs.writeFileSync(root+'/source.zip',archive,{flag:'wx',mode:0o600});
 const names=execFileSync('unzip',['-Z','-1',root+'/source.zip'],{encoding:'utf8',maxBuffer:1<<20}).trim().split('\n');
 const allowed=['scifact/corpus.jsonl','scifact/queries.jsonl','scifact/qrels/train.tsv','scifact/qrels/test.tsv'];assert(allowed.every(n=>names.filter(p=>p===n).length===1));
 // Extract only explicit members to stdout. No archive-controlled path reaches
 // a filesystem destination; ZIP links/traversal cannot write outside this root.
 const data=Object.fromEntries(allowed.map(n=>[n,execFileSync('unzip',['-p',root+'/source.zip',n],{maxBuffer:32<<20})]));
 const corpus=jsonLines(data[allowed[0]]),queries=jsonLines(data[allowed[1]]),train=qrels(data[allowed[2]]),test=qrels(data[allowed[3]]),inventory=validate(corpus,queries,train,test);
 assert.equal(inventory.documents,5183);assert.equal(inventory.testQueries,300);
 const prepared={
  'corpus.json':corpus.map(d=>({id:d._id,title:d.title,text:d.text})),
  'queries.json':queries.map(q=>({id:q._id,text:q.text})),
  'labels-train.json':train,'labels-test.json':test,
 };
 const artifacts={};for(const[n,value]of Object.entries(prepared)){const bytes=Buffer.from(JSON.stringify(value,null,2)+'\n');fs.writeFileSync(root+'/'+n,bytes,{flag:'wx',mode:0o600});artifacts[n]=sha(bytes)}
 const report={start,end:new Date().toISOString(),wallMS:performance.now()-begin,url:URL,officialMD5:MD5,sourceSHA256:sha(archive),sourceBytes:size,scriptSHA256:sha(fs.readFileSync(fileURLToPath(import.meta.url))),inventory,members:Object.fromEntries(allowed.map(n=>[n,sha(data[n])])),artifacts,metadataAuthorsAndFieldsExcluded:true,queriesContainNoLabels:true,noModelRuns:true,noFitting:true,noPretrainingNoveltyClaim:true,noPublicationOrMedicalTruthClaim:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
 fs.writeFileSync(root+'/source.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(report,null,2));
}
