// Independent per-question enumeration, not the aggregate-band search.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
assert.equal(process.argv.length,4);const file=process.argv[2],output=process.argv[3],bytes=fs.readFileSync(file),r=JSON.parse(bytes),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(hash(fs.readFileSync(r.input)),r.inputSHA256);assert.equal(hash(fs.readFileSync('research/public-task-pilot/metrology-headroom.mjs')),r.sourceSHA256);
const result=JSON.parse(fs.readFileSync(r.input)),checked={};
for(const[key,c]of Object.entries(r.cells)){
 const per=result.cells[key].baseline.per,groups=[...new Set(per.map(p=>p.cluster))];assert.equal(groups.length,6);assert(per.filter(p=>p.wording==='literal').every(p=>p.top1===1));
 const variable=per.filter(p=>p.wording==='paraphrase').map(p=>({key:groups.indexOf(p.cluster),baseline:p.top1}));assert.equal(variable.length,18);
 let maximum=-Infinity,eligible=0;const counts=new Int32Array(6);
 for(let mask=0;mask<2**18;mask++){
  counts.fill(0);let net=0;
  for(let j=0;j<18;j++){const d=((mask>>>j)&1)-variable[j].baseline;counts[variable[j].key]+=d;net+=d}
  if(net<2)continue;eligible++;
  // Algebraically separate count-scale calculation: mean=net/36,
  // SE^2=(sum count^2 - net^2/6)/(5*6*36).
  const squares=counts.reduce((s,x)=>s+x*x,0),se=Math.sqrt((squares-net*net/6)/1080),lower=net/36-2.015048373*se;
  maximum=Math.max(maximum,lower);
 }
 assert.equal(maximum>0,c.oracleGatePossible);assert(Math.abs(maximum-c.best.lower)<2e-14);
 checked[key]={perQuestionAssignments:2**18,eligibleAssignments:eligible,maximumLower:maximum,oracleGatePossible:maximum>0};
}
const report={time:new Date().toISOString(),headroomSHA256:hash(bytes),inputSHA256:r.inputSHA256,sourceSHA256:hash(fs.readFileSync('research/public-task-pilot/metrology-headroom-verify.mjs')),independentExhaustion:true,checked,modelOrProtocolChanged:false,allSevenWholeGoals:'OPEN'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(report,null,2));
