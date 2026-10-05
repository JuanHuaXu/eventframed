import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';
import {criticFeatures} from './query-critic.mjs';
import {fitQueryFactorial,chooseQueryFactorial as chooseCritic} from './query-error-factorial.mjs';

const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const experiment=JSON.parse(fs.readFileSync(process.argv[2],'utf8')),target=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
const projection=JSON.parse(fs.readFileSync(process.argv[5],'utf8'));assert.equal(sha(process.argv[5]),experiment.projectionSHA256);
assert.equal(sha(process.argv[3]),experiment.targetSHA256);assert.equal(sha(process.argv[4]),experiment.rawSHA256);
const reader=createInterface({input:fs.createReadStream(process.argv[4]),crlfDelay:Infinity});let header=true,index=0;const train=[],evaluation=[];
for await(const line of reader){if(header){header=false;continue;}const r=JSON.parse(line),t=target.records[index++];assert.equal(r.Original.Index,t.index);
  const o=r.Original.Decision,d=r.Candidate.Decision;
  const savedIndex=index-1;
  const features=()=> (o.Pool??[]).map(origin=>({origin,x:experiment.augmented?criticFeatures(o,d,origin).concat(projection.records[savedIndex].features.find(c=>c.origin===origin).values):criticFeatures(o,d,origin)}));
  if(t.phase===0&&t.schedule===1)train.push(features().map(c=>({...c,y:t.branches[0].value-t.branches.find(b=>b.origin===c.origin).value,weight:1/o.Pool.length})));
  if(t.phase===1&&t.schedule===1)evaluation.push({features,expected:experiment.records[index-1].decision});
}
assert.equal(evaluation.length,672);assert.equal(index,2688);
const fits=[];for(let i=0;i<3;i++){const start=performance.now(),m=fitQueryFactorial(train,experiment.quadratic);fits.push(performance.now()-start);assert.deepEqual(m,experiment.model);}
const model=experiment.model;
const once=r=>{const t=performance.now(),got=chooseCritic(model,r.features());const ms=performance.now()-t;assert.deepEqual(got,r.expected);return ms;};
const firstMS=once(evaluation[0]);for(let i=0;i<100;i++)once(evaluation[i%evaluation.length]);
const samples=[];for(let repeat=0;repeat<3;repeat++)for(const r of evaluation)samples.push(once(r));samples.sort((a,b)=>a-b);
const q=p=>samples[Math.ceil(p*samples.length)-1];
console.log(JSON.stringify({scope:'Component timing only, excludes parsing, prequential projection creation, Bayesian bundles, posterior fitting, retrieval, I/O, queues and contention',
  node:process.version,platform:process.platform,arch:process.arch,pools:train.length,rows:train.flat().length,fitMS:fits,firstSelectionMS:firstMS,warmupCalls:100,calls:samples.length,
  selectionMS:{median:q(.5),p95:q(.95),p99:q(.99),max:samples.at(-1)},experimentSHA256:sha(process.argv[2]),scriptSHA256:sha(import.meta.filename),criticSHA256:sha(new URL('./query-error-factorial.mjs',import.meta.url))}));
