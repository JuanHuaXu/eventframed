import fs from 'node:fs';
import readline from 'node:readline';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [source,output]=process.argv.slice(2);assert(source&&output);
let algebraChecks=0;
for(let bi=0;bi<=10;bi++)for(let qi=0;qi<=10;qi++)for(let pi=0;pi<=10;pi++) {
  const b=bi/10,q=qi/10,p=pi/10,noise=q*(1-q);
  const gain=((b-q)**2+noise)-((p-q)**2+noise);
  assert(gain<=(b-q)**2+1e-12);algebraChecks++;
}
const digest=crypto.createHash('sha256'),stream=fs.createReadStream(source);
stream.on('data',b=>digest.update(b));
const groups=new Map(),identities=new Set();let header=true;
for await(const line of readline.createInterface({input:stream,crlfDelay:Infinity})) {
  const r=JSON.parse(line);
  if(header) {assert.equal(r.Version,'soft-learners-v120');assert.equal(r.TransferBase,2200112200);assert.equal(r.BooleanBase,2204112300);header=false;continue;}
  const id=[r.Phase,r.Case,r.Index,r.Schedule].join(':');assert(!identities.has(id));identities.add(id);
  assert.equal(r.Steps.length,256);
  const changing=r.Case<9?r.Case%3!==0:r.Case>=19;if(!changing)continue;
  const key=[r.Phase,r.Case,r.Schedule].join(':');if(!groups.has(key))groups.set(key,[]);
  const gains=[0,2,12,13,14].map(a=>r.Steps.slice(192).reduce((v,s)=>{
    assert(Number.isFinite(s.Q)&&s.Q>=0&&s.Q<=1);assert(s.P[a]>=0&&s.P[a]<=1);
    return v+(s.P[a]-s.Q)**2/64;
  },0));
  groups.get(key).push({index:r.Index,gains});
}
assert.equal(identities.size,2688);assert.equal(groups.size,32);
const artifactSHA256=digest.digest('hex');assert.equal(artifactSHA256,'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
const cells=[];
for(const [key,rows] of groups) {
  assert.deepEqual(rows.map(r=>r.index).sort((a,b)=>a-b),Array.from({length:32},(_,i)=>i));
  for(const cap of [64,32])for(const control of cap===64?[0,12,13]:[2,12,14]) {
    const a=[0,2,12,13,14].indexOf(control),upperGain=rows.reduce((v,r)=>v+r.gains[a]/32,0);
    cells.push({key,cap,control,oracleMaximumMeanGain:upperGain,belowRequiredGain:upperGain<.005-1e-12});
  }
}
assert.equal(cells.length,192);
const out={artifactSHA256,algebraChecks,records:identities.size,cells,
  impossibleMeanGates:cells.filter(c=>c.belowRequiredGain).length,
  limitations:'Exact conditional-risk ceiling on archived v120 records, Q evaluation only. No refitting or fresh confirmation. Passing this necessary mean condition does not prove learnability, confidence-gate feasibility, or success.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({...out,cells:cells.filter(c=>c.belowRequiredGain)},null,2));
