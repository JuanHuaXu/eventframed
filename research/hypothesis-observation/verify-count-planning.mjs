import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'count-planning-exact.json'),d=JSON.parse(raw);
for(const[p,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const replay=execFileSync(process.execPath,[dir+'count-planning-exact.mjs'],{maxBuffer:64*1024*1024});assert(raw.equals(replay));
const risk=s=>1-s.forecast.reduce((a,p)=>a+p*p,0),zero=Array(8).fill(0).join(',');
let valueChecks=0,normalizationChecks=0,maxError=0,bellmanChecks=0;
for(const root of d.roots){
  const states=new Map(root.states.map(s=>[s.counts.join(','),s]));
  for(const s of root.states){
    const remaining=6-s.counts.reduce((a,n)=>a+n,0);
    if(!remaining){assert.equal(s.fullCost,0);continue;}
    const costs=Array.from({length:4},(_,a)=>[0,1].reduce((sum,y)=>{
      const counts=s.counts.slice();counts[2*a+y]++;const child=states.get(counts.join(','));
      return sum+child.mass/s.mass*(risk(child)+child.fullCost);
    },0));
    assert(Math.abs(s.fullCost-Math.min(...costs))<1e-12);bellmanChecks++;
  }
  for(const policy of Object.keys(d.totals)){
    let occupancy=new Map([[zero,1]]),preSum=0,postSum=0;
    for(let step=0;step<6;step++){
      const next=new Map();assert(Math.abs([...occupancy.values()].reduce((s,p)=>s+p,0)-1)<1e-12);normalizationChecks++;
      for(const [key,p]of occupancy){
        const s=states.get(key);preSum+=p*risk(s);
        for(let a=0;a<4;a++){
          const actionWeight=policy==='random'?.25:Number(s.actions[policy]===a);if(!actionWeight)continue;
          for(let y=0;y<2;y++){
            const counts=s.counts.slice();counts[2*a+y]++;const k=counts.join(','),child=states.get(k);
            const w=p*actionWeight*child.mass/s.mass;next.set(k,(next.get(k)||0)+w);postSum+=w*risk(child);
          }
        }
      }
      occupancy=next;
    }
    const finalMass=[...occupancy.values()].reduce((s,p)=>s+p,0);assert(Math.abs(finalMass-1)<1e-12);normalizationChecks++;
    const finalRisk=[...occupancy].reduce((s,[key,p])=>s+p*risk(states.get(key)),0);
    for(const[k,value]of Object.entries({preSum,postSum,finalRisk})){
      const error=Math.abs(value-root.values[policy][k]);assert(error<1e-11);maxError=Math.max(maxError,error);valueChecks++;
    }
  }
}
const componentTests=JSON.parse(execFileSync(process.execPath,[dir+'test-count-planning.mjs']));
const out={scope:'Byte-exact replay; separate forward occupancy and Bellman verification on saved beliefs, not independent inference',inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-count-planning.mjs')).digest('hex'),replay:true,valueChecks,normalizationChecks,bellmanChecks,maxError,componentTests};
writeFileSync(dir+'count-planning-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));

