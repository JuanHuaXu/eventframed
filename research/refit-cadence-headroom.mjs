import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const summary=JSON.parse(fs.readFileSync(process.argv[2]));
const stream=fs.createReadStream(process.argv[3]),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));
const reader=createInterface({input:stream,crlfDelay:Infinity});const rows=[];let count=0;
for await(const line of reader){const raw=JSON.parse(line);if(!raw.Steps)continue;
 const r=summary.records[count++];assert(r);assert.deepEqual([r.phase,r.case,r.index,r.schedule],[raw.Phase,raw.Case,raw.Index,raw.Schedule]);
 const floor=raw.Steps.slice(192).reduce((sum,s)=>sum+s.Q*(1-s.Q)/64,0);
 rows.push({...r,floor});
}
assert.equal(count,2688);assert.equal(hash.digest('hex'),summary.inputSHA256);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((sum,x)=>sum+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(const c of [1,2,4,5,7,8,19,20]){
 const rs=rows.filter(r=>r.phase===phase&&r.case===c&&r.schedule===1);assert.equal(rs.length,32);
 const headroom=mean(rs.map(r=>r.brier[0][4][1]-r.floor));
 groups.push({phase,case:c,terminalFloor:mean(rs.map(r=>r.floor)),baselineHeadroom:headroom,notRuledOutByBayesFloor:headroom>=.005,first192Gain:ci(rs.map(r=>(256*(r.brier[0][4][0]-r.brier[1][4][0])-64*(r.brier[0][4][1]-r.brier[1][4][1]))/192))});
}
console.log(JSON.stringify({scope:'Post-result analytic Brier floor and time-window diagnostic; no gate changes or policy selection',inputSHA256:summary.inputSHA256,groups}));
