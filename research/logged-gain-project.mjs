import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';

const [sourcePath,popPath,policyPath]=process.argv.slice(2);
function stream(p){const file=fs.createReadStream(p),hash=crypto.createHash('sha256');file.on('data',b=>hash.update(b));return {iterator:createInterface({input:file,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
// Attach both iterators before consuming either stream.
const [source,pop]=[sourcePath,popPath].map(stream),policies=JSON.parse(fs.readFileSync(policyPath));
const sourceHeader=JSON.parse((await source.iterator.next()).value),popHeader=JSON.parse((await pop.iterator.next()).value);
assert.equal(sourceHeader.Version,'soft-learners-v120');assert.equal(popHeader.Version,'population-query-v1');
assert.deepEqual(policies.arms.slice(1,4),['random','entropy','joint8']);
const records=[];let checked=0,identities=0;
for(;;){
  const a=await source.iterator.next(),b=await pop.iterator.next();
  if(a.done||b.done){assert(a.done&&b.done);break;}
  const s=JSON.parse(a.value),p=JSON.parse(b.value),policy=policies.records[records.length];assert(policy&&!p.Error);
  for(const [up,low]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(s[up],p[up]);assert.equal(s[up],policy[low]);}
  const origins=policy.selected.slice(1,4),outcomes=s.Steps.slice(161,192).map(x=>x.Y);assert.equal(outcomes.length,31);assert(outcomes.every(x=>typeof x==='boolean'));
  const losses=origins.map(origin=>{
    const branch=p.Branches.find(b=>b.Origin===origin);assert(branch);
    const answer=origin<0?0:Number(s.Steps[origin].Y);if(origin>=0)assert.equal(Boolean(answer),branch.ActualY);
    const predictions=branch.Predictions[answer];assert.equal(predictions.length,31);
    let direct=0,alternate=0;
    for(let j=0;j<31;j++){
      const v=predictions[j],y=Number(outcomes[j]);assert(Number.isFinite(v)&&v>=0&&v<=1);
      direct+=(v-y)**2/31;alternate+=(outcomes[j]?(1-v)**2:v*v)/31;checked++;
    }
    assert(Math.abs(direct-alternate)<1e-13);return direct;
  });
  for(let i=0;i<3;i++)for(let j=0;j<i;j++)if(origins[i]===origins[j]){assert.equal(losses[i],losses[j]);identities++;}
  records.push({phase:s.Phase,case:s.Case,index:s.Index,schedule:s.Schedule,origins,losses});
}
assert.equal(records.length,2688);
const inputHash=source.hash.digest('hex');assert.equal(inputHash,popHeader.InputSHA256);assert.equal(inputHash,'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
console.log(JSON.stringify({scope:'Full-information simulator audit table; realized future outcomes, not teacher risk; not supplied wholesale to logged estimator',hashes:{[sourcePath]:inputHash,[popPath]:pop.hash.digest('hex'),[policyPath]:sha(policyPath),[import.meta.filename]:sha(import.meta.filename)},checkedForecastOutcomes:checked,identityChecks:identities,records}));
