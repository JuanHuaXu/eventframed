import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

// With D boundaries between the first and last selected labels, the event of
// no boundary has probability q=(1-h)^D and evidence Z0. Hence Zseg>=q*Z0,
// and the equal-prior segment weight is >=q/(1+q). This audits averaged bounds
// against archived averaged weights, not individual unrecorded mixture weights.
const raw=process.argv[2],diagnostic=process.argv[3];
const data=readFileSync(diagnostic),s=JSON.parse(data),hash=createHash('sha256');
const stream=createReadStream(raw);stream.on('data',b=>hash.update(b));
const sums=new Map();let records=0,header;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line);records++;
  for(let w=0;w<2;w++){
    const key=[r.Phase,r.Case,r.Schedule,w].join('/');
    if(!sums.has(key))sums.set(key,{count:0,bounds:Array(8).fill(0)});
    const a=sums.get(key);a.count++;
    for(let f=0;f<8;f++){
      const os=r.Fits[f].Origins[w];assert(os.length>0);
      const span=os.at(-1)-os[0];assert(span>=0);
      const q=.99**span;a.bounds[f]+=q/(1+q);
    }
  }
}
assert.equal(records,2688);assert.equal(hash.digest('hex'),s.artifactSHA256);
const rows=[];let checks=0,minSlack=Infinity;
for(const g of s.groups.filter(g=>g.segment===0)){
  const key=[g.phase,g.case,g.schedule,g.window].join('/'),a=sums.get(key);assert.equal(a.count,32);
  const lower=a.bounds.map(x=>x/32),slack=lower.map((v,f)=>g.segmentWeightByFit[f]-v);
  for(const v of slack){assert(v>=-1e-12);minSlack=Math.min(minSlack,v);checks++;}
  rows.push({phase:g.phase,case:g.case,schedule:g.schedule,window:g.window,meanLowerBounds:lower,meanWeights:g.segmentWeightByFit});
}
assert.equal(checks,1344);
console.log(JSON.stringify({scope:'Analytic averaged static-atom weight lower bound; not a forecast-error bound',artifactSHA256:s.artifactSHA256,diagnosticSHA256:createHash('sha256').update(data).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),checks,minSlack,rows},null,2));
