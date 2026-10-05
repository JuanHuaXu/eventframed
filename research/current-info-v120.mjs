import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {switchWeights} from './switch-distribution.mjs';
import {currentStateInformation} from './current-state-information.mjs';
const arms=[0,1,2,3,10,11],prior=[.95,.01,.01,.01,.01,.01];
function run(steps,stop=steps.length){
 const paid=new Map(),queries=[],ps=[];
 for(let t=0;t<stop;t++){
  const rows=steps.slice(0,t).map((s,j)=>({p:arms.map(i=>s.P[i]),y:(!s.Missing&&j+s.Delay<=t)||(paid.has(j)&&paid.get(j)<=t)?s.Y:null}));
  const base=switchWeights(rows,prior);ps.push(base.weights.reduce((v,w,i)=>v+w*steps[t].P[arms[i]],0));
  if(t===0||t%8!==0)continue;
  let selected,best=-Infinity;
  for(let j=t-8;j<t;j++)if(rows[j].y===null){const score=currentStateInformation(rows,prior,j,base).information;if(score>best){best=score;selected=j;}}
  if(selected===undefined)continue;
  assert(!paid.has(selected));paid.set(selected,t+1);queries.push({clock:t,origin:selected,reveal:t+1,information:best});
 }
 return{ps,queries};
}
const referencePath='docs/experiments/mmm-expedite-v120.json',reference=JSON.parse(fs.readFileSync(referencePath));
const input=process.argv[2],stream=fs.createReadStream(input),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));
const records=[];let header,asOf=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
 if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
 const r=JSON.parse(line),result=run(r.Steps),scores=[0,0];assert.equal(r.Steps.length,256);
 result.ps.forEach((p,t)=>{assert(Number.isFinite(p)&&p>=0&&p<=1);const q=r.Steps[t].Q,l=(p-q)**2+q*(1-q);scores[0]+=l/256;if(t>=192)scores[1]+=l/64;});
 const control=reference.records[records.length];assert.deepEqual([control.phase,control.case,control.index,control.schedule],[r.Phase,r.Case,r.Index,r.Schedule]);
 for(const a of [1,2,3])assert.equal(result.queries.length,control.queries[a].length);
 if(r.Schedule===0){assert.equal(result.queries.length,0);scores.forEach((v,i)=>assert.equal(v,control.scores[0][i]));}
 if(r.Index===0)for(const clock of [32,160,255]){
  const paid=new Set(result.queries.filter(q=>q.reveal<=clock).map(q=>q.origin));
  const poisoned=r.Steps.map((s,j)=>({...s,Q:1-s.Q,Y:j>=clock||((s.Missing||j+s.Delay>clock)&&!paid.has(j))?!s.Y:s.Y}));
  const check=run(poisoned,clock+1);assert.deepEqual(check.ps,result.ps.slice(0,clock+1));assert.deepEqual(check.queries,result.queries.filter(q=>q.clock<=clock));asOf++;
 }
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores,queries:result.queries,controls:control.scores});
}
assert.equal(records.length,2688);assert.equal(asOf,252);assert.equal(hash.copy().digest('hex'),reference.artifactSHA256);
const mean=a=>a.reduce((a,b)=>a+b,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const meta={phase,case:c,schedule,segment};groups.push({...meta,brier:mean(rs.map(r=>r.scores[segment])),controls:[0,1,2,3].map(a=>mean(rs.map(r=>r.controls[a][segment]))),meanQueries:mean(rs.map(r=>r.queries.length))});
 for(const control of [0,1,2,3]){const gain=interval(rs.map(r=>r.controls[control][segment]-r.scores[segment]));gates.push({...meta,control,type:'nonharm',...gain,pass:gain.lower>=-.01});
  if(schedule===1&&segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,control,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});}
}
const paths=[import.meta.filename,'research/switch-distribution.mjs','research/current-state-information.mjs','research/current-state-information-test.mjs','docs/experiments/mmm-current-info-v120-protocol.md',referencePath];
console.log(JSON.stringify({scope:'Consumed fixed-forecast current-state information acquisition; perfect label service assumed',artifactSHA256:hash.digest('hex'),hashes:Object.fromEntries(paths.map(p=>[p,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')])),asOf,summary:{nonharm:gates.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:672,gains:gates.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:64,status:gates.every(g=>g.pass)?'PASS':'FAIL'},records,groups,gates}));
