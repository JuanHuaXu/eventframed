import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [resultPath,coveragePath,sourcePath]=process.argv.slice(2),result=JSON.parse(fs.readFileSync(resultPath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for(const [p,h]of Object.entries(result.hashes))assert.equal(sha(p),h);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[coveragePath,sourcePath].map(input);for(const s of streams)await s.it.next();
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10);
const fields=['jointLog','conditionalLog','brier','targetExpectedLog','targetExpectedBrier','teacherJointLog'];
let recordCount=0,pairCount=0,lossChecks=0,trainingPairs=0,d0=0,d1=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [r,source]=next.map(x=>JSON.parse(x.value)),saved=result.records[recordCount++];assert(saved);
  for(const [up,lo]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[up],source[up]);assert.equal(r[up],saved[lo]);}
  const bs=r.Branches??[],metrics=Object.fromEntries(fields.map(f=>[f,[0,0,0]]));let pairs=0;
  for(const branch of bs){
    const query=source.Steps[branch.Origin],p=branch.Mass[1],qy=Number(query.Y),w=1/(31*bs.length);
    for(let age=0;age<31;age++){
      const target=source.Steps[161+age],ty=Number(target.Y),decay=.99**age;
      const f=branch.Conditional.map(v=>.5+decay*(v[target.X]-.5)),base=.5+decay*(r.Base[target.X]-.5);
      const covariance=p*(1-p)*(f[1]-f[0]),independent=[(1-p)*(1-base),(1-p)*base,p*(1-base),p*base];
      const actual=2*qy+ty,queryMass=qy?p:1-p;
      if(r.Phase===0){
        const delta=f[qy]-base,sgn=ty?1:-1;
        d0-=w*sgn*delta/(ty?base:1-base);d1-=w*sgn*delta/(ty?f[qy]:1-f[qy]);trainingPairs++;
      }
      for(const [a,lambda]of result.lambdas.entries()){
        const cells=independent.map((v,i)=>v+lambda*covariance*([1,-1,-1,1][i]));assert(cells.every(v=>v>0));near(cells.reduce((a,b)=>a+b,0),1);
        const conditional=cells[2*qy+1]/queryMass,logJoint=-Math.log(cells[actual]),logConditional=logJoint+Math.log(queryMass);
        const truth=[(1-query.Q)*(1-target.Q),(1-query.Q)*target.Q,query.Q*(1-target.Q),query.Q*target.Q];
        const values={jointLog:logJoint,conditionalLog:logConditional,brier:(conditional-ty)**2,targetExpectedLog:-(1-target.Q)*Math.log(cells[2*qy]/queryMass)-target.Q*Math.log(conditional),targetExpectedBrier:conditional*conditional-2*conditional*target.Q+target.Q,teacherJointLog:-truth.reduce((sum,v,i)=>sum+v*Math.log(cells[i]),0)};
        for(const f of fields)metrics[f][a]+=w*values[f];
      }
      pairs++;
    }
  }
  assert.equal(pairs,saved.pairs);pairCount+=pairs;
  for(const f of fields){if(!pairs)assert.equal(saved.metrics[f],null);else for(let a=0;a<3;a++){near(metrics[f][a],saved.metrics[f][a]);lossChecks++;}}
}
assert.equal(recordCount,2688);assert.equal(trainingPairs,143840);assert.equal(pairCount,287587);
near(d0,result.fit.d0);near(d1,result.fit.d1);
// The objective is a sum of negative logs of positive affine probabilities.
// Convexity plus a negative right-end derivative proves this endpoint optimum.
assert(d1<0);assert.equal(result.fit.lambda,1);
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],result.hashes[coveragePath]);assert.equal(hashes[1],result.hashes[sourcePath]);
const mean=xs=>xs.reduce((sum,x)=>sum+x/xs.length,0);let means=0,bounds=0;
for(const g of result.groups){
  const rows=result.records.filter(r=>r.phase===g.phase&&r.case===g.case&&r.schedule===g.schedule);assert.equal(rows.length,32);
  if(!g.schedule){assert.equal(g.active,0);continue;}
  for(const f of fields){
    for(let a=0;a<3;a++){near(mean(rows.map(r=>r.metrics[f][a])),g.metrics[f][a]);means++;}
    for(let ctrl=0;ctrl<2;ctrl++){
      const xs=rows.map(r=>r.metrics[f][ctrl]-r.metrics[f][2]),m=mean(xs);let pairVariance=0;
      for(let i=0;i<32;i++)for(let j=i+1;j<32;j++)pairVariance+=(xs[i]-xs[j])**2;
      const se=Math.sqrt(pairVariance/(32*32*31)),v=g.gains[f][ctrl];near(v.mean,m);near(v.lower,m-3.5*se);near(v.upper,m+3.5*se);bounds++;
    }
  }
}
for(const f of ['jointLog','targetExpectedBrier']){
  const groups=result.groups.filter(g=>g.phase===1&&g.schedule===1);
  const nonharm=!groups.some(g=>g.gains[f].some(v=>v.lower<-.001)),transition=!groups.filter(g=>g.case>=19).some(g=>g.gains[f].some(v=>v.lower<=0));
  assert.deepEqual(result.screens[f],{nonharm,transition,pass:nonharm&&transition});
}
console.log(JSON.stringify({status:'PASS',recordCount,pairCount,trainingPairs,lossChecks,means,bounds,d0,d1,endpointOptimum:true,hashes:{[resultPath]:sha(resultPath),[import.meta.filename]:sha(import.meta.filename)},scope:'Independent covariance-cell reconstruction, phase0-only derivative/convex optimum proof, source rescoring, pairwise-variance aggregation and screens.'}));
