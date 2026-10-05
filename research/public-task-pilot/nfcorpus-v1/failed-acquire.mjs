// Public research-only acquisition. Partition membership is not a relevance fit.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
const root='research/public-task-pilot/nfcorpus-v1',hash=(b,alg='sha256')=>crypto.createHash(alg).update(b).digest('hex');
assert(!fs.existsSync(root));fs.mkdirSync(root,{mode:0o700});
const url='https://public.ukp.informatik.tu-darmstadt.de/thakur/BEIR/datasets/nfcorpus.zip';
const began=new Date().toISOString();
execFileSync('/usr/bin/curl',['--fail','--location','--max-time','120','--proto','=https','--proto-redir','=https','--output',path.resolve(root+'/source.zip'),url],{stdio:['ignore','pipe','pipe']});
const zip=fs.readFileSync(root+'/source.zip');assert.equal(hash(zip,'md5'),'a89dba18a62ef92f7d323ec890a0d38d');
const names=execFileSync('/usr/bin/unzip',['-Z1',root+'/source.zip'],{encoding:'utf8'}).trim().split('\n');
assert(names.every(n=>n.startsWith('nfcorpus/')&&!n.includes('..')&&!n.startsWith('/')&&!n.includes('\\')));
const get=n=>execFileSync('/usr/bin/unzip',['-p',root+'/source.zip','nfcorpus/'+n],{maxBuffer:32<<20});
const corpusRaw=get('corpus.jsonl'),queriesRaw=get('queries.jsonl'),testRaw=get('qrels/test.tsv');
const decode=b=>b.toString('utf8').trim().split('\n').map(JSON.parse);
const corpus=decode(corpusRaw),queries=decode(queriesRaw);assert.equal(corpus.length,3633);
const seen=new Set(),records=corpus.map(d=>{assert.deepEqual(Object.keys(d).sort(),['_id','metadata','text','title']);assert.equal(typeof d._id,'string');assert(!seen.has(d._id));seen.add(d._id);assert.equal(typeof d.title,'string');assert.equal(typeof d.text,'string');assert(d.title&&d.text);return{id:d._id,title:d.title,text:d.text};});
const qmap=new Map();for(const q of queries){assert.deepEqual(Object.keys(q).sort(),['_id','metadata','text']);assert.equal(typeof q._id,'string');assert(!qmap.has(q._id));assert.equal(typeof q.text,'string');assert(q.text);qmap.set(q._id,{id:q._id,text:q.text});}
const lines=testRaw.toString('utf8').trim().split('\n');assert.equal(lines[0],'query-id\tcorpus-id\tscore');
// Selection uses ONLY official partition membership (column zero). Relevance
// values/doc identities are not provided to any predictor or model selection.
const ids=[...new Set(lines.slice(1).map(l=>l.split('\t')[0]))].sort();assert.equal(ids.length,323);
const projection={partition:'nfcorpus-official-test',queries:ids.map(id=>{assert(qmap.has(id));return qmap.get(id);})};
const write=(n,b)=>fs.writeFileSync(root+'/'+n,b,{flag:'wx',mode:0o600});
write('corpus-source.jsonl',corpusRaw);write('queries-source.jsonl',queriesRaw);write('test-qrels-source.tsv',testRaw);
write('corpus.json',JSON.stringify(records)+'\n');write('test-only.json',JSON.stringify(projection)+'\n');
const scientificModel='research/public-task-pilot/scifact-idf-pairrank-v1/predictions.json',model=JSON.parse(fs.readFileSync(scientificModel));assert.equal(model.models[0].fold,-1);assert.equal(model.calibrationPredictions,0);assert.equal(model.confirmationPredictions,0);
write('frozen-idf-model.json',JSON.stringify(model.models[0].model)+'\n');
const artifacts={};for(const n of fs.readdirSync(root)){const b=fs.readFileSync(root+'/'+n);artifacts[n]={sha256:hash(b),bytes:b.length};}
write('source.json',JSON.stringify({time:new Date().toISOString(),began,url,md5:hash(zip,'md5'),expectedMD5:'a89dba18a62ef92f7d323ec890a0d38d',archiveEntries:names,artifacts,documents:3633,officialTestQueries:323,allQueryRows:queries.length,membershipColumnOnly:true,relevanceValuesUsed:false,predictions:0,outcomesEvaluated:0,modelFittedOnNF:false,modelOrigin:{path:scientificModel,sha256:hash(fs.readFileSync(scientificModel)),fold:-1,fitQueries:351},primarySource:'https://www.cl.uni-heidelberg.de/statnlpgroup/nfcorpus/',terms:'Owner grants academic use; other NutritionFacts.org uses require consulting owner. Local research only; no automatic redistribution/commercial authority. HF card labels cc-by-sa-4.0, not treated as superseding owner terms.',relevanceNature:'Automatically constructed direct/indirect link and topic-tag judgments, not verified factual truth or agent usefulness.',nativeDatabaseUsed:false,noProductionAccess:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',scriptSHA256:hash(fs.readFileSync('research/public-task-pilot/nfcorpus-v1-acquire.mjs'))},null,2)+'\n');
console.log(JSON.stringify({root,documents:3633,testQueries:323,predictions:0,outcomesEvaluated:0,modelFrozenFromFIT:true}));
