import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {acquisitionBelief,acquisitionChoice,TYPES} from './interval-acquisition.mjs';
import {lookaheadChoice} from './interval-lookahead.mjs';
function rng(key){
  let a=createHash('sha256').update(key).digest().readUInt32LE(0);
  return ()=>{a=(a+0x6D2B79F5)>>>0;let t=a;t=Math.imul(t^(t>>>15),t|1);t^=t+Math.imul(t^(t>>>7),t|61);return ((t^(t>>>14))>>>0)/4294967296;};
}
const brier=(p,y)=>p.reduce((s,v,i)=>s+(v-Number(i===y))**2,0);
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const intervals=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,v)=>s+(v-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.3*se,upper:m+3.3*se};};
const records=[],summaries=[],gates=[],policies=['lookahead','target','random','entropy','fixed'];
for(let split=0;split<2;split++)for(const noise of [.1,.2,.3])for(const mask of [0,5,10,15])for(let index=0;index<64;index++){
  const seed='2026091402:'+split+':'+noise+':'+mask+':'+index;
  const truth=Math.floor(rng(seed+':truth')()*16),draw=rng(seed+':reports');
  const roots=TYPES.map(()=>draw()),fresh=Array.from({length:6},()=>TYPES.map(()=>draw()));
  const probability=t=>{const bit=t===7?((truth&1)^((truth>>2)&1)):((truth>>t)&1),error=t===2?.01:noise;return bit?1-error:error;};
  const initial=TYPES.map((t,i)=>({action:[0,t,0],outcome:roots[i]<probability(t)}));
  const arms={},cache=new Map();
  for(const policy of policies){
    const history=initial.slice(),trace=[],random=rng(seed+':actions:'+policy);let area=0;
    for(let step=0;step<6;step++){
      const forecast=acquisitionBelief(history).forecast;
      const action=policy==='lookahead'?lookaheadChoice(history,cache):acquisitionChoice(history,policy,random,cache),[,t,slot]=action,j=TYPES.indexOf(t);
      const outcome=(mask&(1<<j)?roots[j]:fresh[slot][j])<probability(t);
      area+=brier(forecast,truth%4)/6;
      trace.push({credit:4+step*2,cost:2,forecast,action,outcome});history.push({action,outcome});
    }
    const final=acquisitionBelief(history).forecast,chosen=final.indexOf(Math.max(...final));
    assert.equal(new Set(history.map(s=>JSON.stringify(s.action))).size,10);
    arms[policy]={trace,final,final_brier:brier(final,truth%4),area_brier:area,correct:chosen===truth%4,confident_wrong:Math.max(...final)>=.9&&chosen!==truth%4,credits:16};
  }
  records.push({split,noise,mask,index,seed,truth,initial,arms});
}
for(let split=0;split<2;split++)for(const noise of [.1,.2,.3])for(const mask of [0,5,10,15]){
  const rs=records.filter(r=>r.split===split&&r.noise===noise&&r.mask===mask);assert.equal(rs.length,64);
  for(const policy of policies)summaries.push({split,noise,mask,policy,...Object.fromEntries(['final_brier','area_brier','correct','confident_wrong'].map(k=>[k,mean(rs.map(r=>Number(r.arms[policy][k])))]))});
  for(const control of ['random','entropy','target']){
    const final=intervals(rs.map(r=>r.arms[control].final_brier-r.arms.lookahead.final_brier));
    gates.push({split,noise,mask,control,kind:'final_nonharm',...final,passed:final.lower>=-.01});
    if(mask!==15&&control!=='target'){const area=intervals(rs.map(r=>r.arms[control].area_brier-r.arms.lookahead.area_brier));gates.push({split,noise,mask,control,kind:'area_gain',...area,passed:area.lower>0});}
  }
}
assert.equal(records.length,1536);assert.equal(gates.length,108);
const paths=['interval-lookahead.mjs','test-interval-lookahead.mjs','interval-acquisition.mjs','test-interval-acquisition.mjs','interval-lookahead-experiment.mjs','noise-envelope.mjs','INTERVAL_LOOKAHEAD_PROTOCOL.md'].map(p=>'research/hypothesis-observation/'+p);
console.log(JSON.stringify({scope:'Fresh finite two-query interval-aware acquisition pilot; equal16 credits; descriptive paired intervals, no simultaneous coverage',seedBase:2026091402,hashes:Object.fromEntries(paths.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),records,summaries,gates,screen_passed:gates.every(g=>g.passed)}));


