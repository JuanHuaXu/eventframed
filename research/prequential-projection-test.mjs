import assert from 'node:assert/strict';
import {projectPrequential,prequentialFeatures} from './prequential-projection.mjs';
const s={Steps:Array.from({length:256},()=>({X:0,P:Array(15).fill(.25),Y:false,Missing:true,Delay:0}))};
assert.deepEqual(prequentialFeatures(projectPrequential(s),152),[0,.5,.25,.5,0,.25,0,.25,.5,1]);
s.Steps[128]={...s.Steps[128],Missing:false,Delay:32,Y:true};
s.Steps[159]={...s.Steps[159],Missing:false,Delay:1,Y:false};
const v=projectPrequential(s),f=prequentialFeatures(v,152);
assert.equal(v.observed.length,2);assert.deepEqual(f,[2/32,.5,.3125,.625,1/8,.0625,2/32,.3125,.625,33/64]);
const before=JSON.stringify(s),poison=structuredClone(s);
for(let i=0;i<256;i++){const r=poison.Steps[i];r.Q=NaN;if(i!==128&&i!==159)r.Y=!r.Y;if(i<128||i>=160){r.P=null;r.X=-1;}}
assert.deepEqual(projectPrequential(poison),v);assert.equal(JSON.stringify(s),before);
const delayed=structuredClone(s);delayed.Steps[128].Delay=33;assert.equal(projectPrequential(delayed).observed.length,1);
const hidden=structuredClone(s);for(let j=129;j<159;j++){hidden.Steps[j].Missing=false;hidden.Steps[j].Delay=200-j;Object.defineProperty(hidden.Steps[j],'Y',{get(){throw Error('unarrived label');}});}
assert.deepEqual(projectPrequential(hidden),v);
for(const key of ['Q','Phase','Case','Seeds','Teacher','Initial','Fits','Metrics'])Object.defineProperty(s,key,{get(){throw Error('forbidden '+key);}});
for(let j=128;j<160;j++)Object.defineProperty(s.Steps[j],'Q',{get(){throw Error('oracle Q');}});
assert.deepEqual(projectPrequential(s),v);
const changed=structuredClone(v);changed.observed[0].y=false;assert.notDeepEqual(prequentialFeatures(changed,152),f);
assert.throws(()=>prequentialFeatures(v,160));const bad=structuredClone(v);bad.observed[0].arrived=161;assert.throws(()=>prequentialFeatures(bad,152));
const dup=structuredClone(v);dup.observed.push(dup.observed[0]);assert.throws(()=>prequentialFeatures(dup,152));
console.log('PASS: hand calculations, empty/boundary observations, future-label and Q traps, delivery equivalence, ownership, sensitivity and invalid views');
