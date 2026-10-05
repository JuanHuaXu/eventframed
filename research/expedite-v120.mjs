import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {switchWeights} from './switch-distribution.mjs';
const arms=[0,1,2,3,10,11],prior=[.95,.01,.01,.01,.01,.01];
const entropy=p=>{p=Math.max(1e-12,Math.min(1-1e-12,p));return-p*Math.log(p)-(1-p)*Math.log1p(-p);};
function run(steps,key,policy,stop=steps.length){
 const paid=new Map(),queries=[],ps=[];
 for(let t=0;t<stop;t++){
  const rows=steps.slice(0,t).map((s,j)=>({p:arms.map(i=>s.P[i]),y:(!s.Missing&&j+s.Delay<=t)||(paid.has(j)&&paid.get(j)<=t)?s.Y:null}));
  const{weights}=switchWeights(rows,prior);assert(Math.abs(weights.reduce((a,b)=>a+b,0)-1)<1e-11);
  ps.push(weights.reduce((v,w,i)=>v+w*steps[t].P[arms[i]],0));
  if(policy===0||t===0||t%8!==0)continue;
  const pool=[];for(let j=t-8;j<t;j++)if(rows[j].y===null)pool.push(j);
  if(!pool.length)continue;
  let selected;
  if(policy===1){const value=crypto.createHash('sha256').update(key+':'+t).digest().readUInt32BE(0)/2**32;selected=pool[Math.floor(value*pool.length)];}
  else{
   let best=-Infinity;
   for(const j of pool){const p=arms.map(i=>steps[j].P[i]),m=weights.reduce((v,w,i)=>v+w*p[i],0);
    const score=entropy(m)-(policy===3?weights.reduce((v,w,i)=>v+w*entropy(p[i]),0):0);
    if(score>best){selected=j;best=score;}
   }
  }
  assert(!paid.has(selected));paid.set(selected,t+1);queries.push({clock:t,origin:selected,reveal:t+1});
 }
 return{ps,queries};
}
const referencePath='docs/experiments/mmm-switch-distribution-v120.json',reference=JSON.parse(fs.readFileSync(referencePath));
const input=process.argv[2],stream=fs.createReadStream(input),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));
const records=[];let header,asOf=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
 if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
 const r=JSON.parse(line),key=[r.Phase,r.Case,r.Index,r.Schedule].join(':'),scores=[],queries=[];
 assert.equal(r.Steps.length,256);
 for(let policy=0;policy<4;policy++){
  const result=run(r.Steps,key,policy),bs=[0,0];queries.push(result.queries);
  result.ps.forEach((p,t)=>{const q=r.Steps[t].Q,l=(p-q)**2+q*(1-q);bs[0]+=l/256;if(t>=192)bs[1]+=l/64;});scores.push(bs);
  if(r.Schedule===0)assert.equal(result.queries.length,0);
  if(r.Index===0)for(const clock of [32,160,255]){
   const seenPaid=new Map(result.queries.filter(q=>q.reveal<=clock).map(q=>[q.origin,q.reveal]));
   const poisoned=r.Steps.map((s,j)=>({...s,Q:1-s.Q,Y:j>=clock||((s.Missing||j+s.Delay>clock)&&!seenPaid.has(j))?!s.Y:s.Y}));
   const rerun=run(poisoned,key,policy,clock+1);
   assert.deepEqual(rerun.ps,result.ps.slice(0,clock+1));assert.deepEqual(rerun.queries,result.queries.filter(q=>q.clock<=clock));asOf++;
  }
 }
 assert.equal(queries[1].length,queries[2].length);assert.equal(queries[2].length,queries[3].length);
 const control=reference.records[records.length];assert.deepEqual([control.phase,control.case,control.index,control.schedule],[r.Phase,r.Case,r.Index,r.Schedule]);
 scores[0].forEach((v,i)=>assert(Math.abs(v-control.scores[2][i])<1e-12));
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores,queries});
}
assert.equal(records.length,2688);assert.equal(asOf,1008);assert.equal(hash.copy().digest('hex'),reference.artifactSHA256);
const mean=a=>a.reduce((a,b)=>a+b,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const meta={phase,case:c,schedule,segment};groups.push({...meta,brier:[0,1,2,3].map(a=>mean(rs.map(r=>r.scores[a][segment]))),meanQueries:[0,1,2,3].map(a=>mean(rs.map(r=>r.queries[a].length)))});
 for(const control of [0,1,2]){const gain=interval(rs.map(r=>r.scores[control][segment]-r.scores[3][segment]));gates.push({...meta,control,type:'nonharm',...gain,pass:gain.lower>=-.01});
  if(schedule===1&&segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,control,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});}
}
const paths=[import.meta.filename,'research/switch-distribution.mjs','docs/experiments/mmm-expedite-v120-protocol.md',referencePath];
console.log(JSON.stringify({scope:'Consumed fixed-forecast equal-query-count expedite diagnostic; oracle label service assumed',artifactSHA256:hash.digest('hex'),hashes:Object.fromEntries(paths.map(p=>[p,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')])),asOf,summary:{nonharm:gates.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:504,gains:gates.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:48,status:gates.every(g=>g.pass)?'PASS':'FAIL'},records,groups,gates}));
