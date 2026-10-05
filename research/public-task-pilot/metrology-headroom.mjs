// POST-HOC feasibility only. No serving, fitting, resampling or gate changes.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import{fileURLToPath}from'node:url';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
export function ceiling(per){
 assert.equal(per.length,36);assert.equal(new Set(per.map(p=>p.id)).size,36);
 const groups=new Map();for(const p of per){assert([0,1].includes(p.top1));assert(['literal','paraphrase'].includes(p.wording));assert.equal(p.nominated,1);assert.equal(p.retained,1);assert.equal(p.survival,1);if(!groups.has(p.cluster))groups.set(p.cluster,[]);groups.get(p.cluster).push(p)}
 assert.equal(groups.size,6);assert([...groups.values()].every(ps=>ps.length===6&&ps.filter(p=>p.wording==='literal').length===3));
 const bands=[...groups].map(([cluster,ps])=>({cluster,min:-ps.filter(p=>p.wording==='paraphrase'&&p.top1===1).length,max:ps.filter(p=>p.top1===0).length}));
 let evaluated=0,eligible=0,best=null;const visit=(at,counts)=>{
  if(at<6){const b=bands[at];for(let d=b.min;d<=b.max;d++)visit(at+1,[...counts,d]);return}
  evaluated++;const net=counts.reduce((s,x)=>s+x,0);if(net<2)return;eligible++;
  const vector=counts.map(x=>x/6),mean=vector.reduce((s,x)=>s+x,0)/6,se=Math.sqrt(vector.reduce((s,x)=>s+(x-mean)**2,0)/30),lower=mean-2.015048373*se;
  if(best===null||lower>best.lower)best={net,vector,mean,se,lower};
 };visit(0,[]);
 const missing=per.filter(p=>p.top1===0).length;
 return{baselineTop1:36-missing,maximumNetGains:missing,bands,enumeratedVectors:evaluated,eligibleVectors:eligible,best,oracleGatePossible:best!==null&&best.lower>0,optimisticOracleNotImplementable:true};
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 assert.equal(process.argv.length,4,'explicit consumed results and NEW output');const input=process.argv[2],out=process.argv[3],raw=fs.readFileSync(input),r=JSON.parse(raw);
 assert.equal(r.technical_pass,true);assert.equal(Object.keys(r.cells).length,4);
 const cells=Object.fromEntries(Object.entries(r.cells).map(([key,v])=>[key,ceiling(v.baseline.per)]));
 const report={time:new Date().toISOString(),input,inputSHA256:hash(raw),sourceSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),postHoc:true,noNewModelRuns:true,noThresholdChanges:true,noResampling:true,cells,allFourOracleGatesPossible:Object.values(cells).every(c=>c.oracleGatePossible),allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
 fs.writeFileSync(out,JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});assert.equal(hash(fs.readFileSync(input)),report.inputSHA256);console.log(JSON.stringify(report,null,2));
}
