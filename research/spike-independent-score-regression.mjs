import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {independentArms} from './spike-independent-score.mjs';

// Consumed data is used only to verify implementation equivalence, not tune.
const dir='docs/experiments/';
const compact=new Map(fs.readFileSync(dir+'mmm-spike-continuous-v1.jsonl','utf8').trim().split('\n').map(JSON.parse).map(r=>[[r.Phase,r.Case,r.Index,r.Schedule].join(':'),r]));
const reference=new Map(JSON.parse(fs.readFileSync(dir+'mmm-spike-fixed-share-local-v1-results.json','utf8')).results.map(r=>[r.key,r]));
let header=true,count=0;
for await(const line of readline.createInterface({input:fs.createReadStream(dir+'mmm-soft-learners-v120.jsonl'),crlfDelay:Infinity})){
  const src=JSON.parse(line);if(header){header=false;continue;}
  const key=[src.Phase,src.Case,src.Index,src.Schedule].join(':'),r=compact.get(key);if(!r)continue;
  const rows=r.P.map((c,i)=>{const s=src.Steps[i];return{b:s.P[12],c,y:s.Y,delay:s.Delay,missing:s.Missing};});
  const arms=independentArms(rows),old=reference.get(key);assert(old);
  assert.deepEqual(arms[0],old.forecasts);
  for(let a=0;a<3;a++){
    let expected=0,realized=0;
    for(let i=0;i<256;i++){
      const p=arms[a][i].p,q=src.Steps[i].Q,y=Number(rows[i].y);
      expected+=((p-q)**2+q*(1-q))/256;realized+=(p-y)**2/256;
    }
    assert.equal(expected,old.expected[a]);assert.equal(realized,old.realized[a]);
  }
  compact.delete(key);reference.delete(key);count++;
}
assert.equal(count,672);assert.equal(compact.size,0);assert.equal(reference.size,0);
console.log('PASS: 172032 guarded forecasts and all 672 paired scores exactly reproduce consumed-data controls');
