import assert from 'node:assert/strict';
import {checkClaimsV2} from './scifact-provenance-v2.mjs';
const queries=[{id:'1',text:'A claim'},{id:'2',text:'Other claim'}],rels=[{query:'1',document:'9',score:1},{query:'1',document:'8',score:1}],claims=[{id:1,claim:'A claim',evidence:{8:[{label:'SUPPORT'}],9:[{label:'CONTRADICT'}]}}];
assert.equal(checkClaimsV2(queries,rels,claims).claims,1);
assert.deepEqual(checkClaimsV2(queries,[...rels].reverse(),claims),checkClaimsV2(queries,rels,claims));
const cases=[
 (q,r,c)=>q[0].text='different',
 (q,r,c)=>r[0].document='7',
 (q,r,c)=>r.push({...r[0]}),
 (q,r,c)=>r[0].score=0,
 (q,r,c)=>r[0].query='2',
 (q,r,c)=>c.push({...c[0]}),
 (q,r,c)=>c[0].id=3,
 (q,r,c)=>delete c[0].evidence['8'],
];
for(const mutate of cases){const[q,r,c]=structuredClone([queries,rels,claims]);mutate(q,r,c);assert.throws(()=>checkClaimsV2(q,r,c))}
console.log('SciFact provenance: exact positive/order controls and eight corruption rejections PASS');
