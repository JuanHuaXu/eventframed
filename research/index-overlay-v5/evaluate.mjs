import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
export const mean=a=>a.reduce((x,y)=>x+y,0)/a.length;
const quantile=(a,p)=>[...a].sort((x,y)=>x-y)[Math.ceil(p*a.length)-1];
export function quality(row){
 const a=row.actual,e=row.exact_top;assert.equal(a.length,50);assert.equal(e.length,50);
 for(const values of[a,e]){assert.equal(new Set(values.map(x=>x.id)).size,50);for(const x of values){assert.match(x.id,/^e-\d{5}$/);assert.ok(Number(x.id.slice(2))<row.records);assert.ok(Number.isFinite(x.exact)&&Math.abs(x.exact)<=1+1e-6);}}
 for(const x of a)assert.ok(Number.isFinite(x.native));
 for(let i=1;i<e.length;i++){assert.ok(e[i-1].exact>=e[i].exact);if(e[i-1].exact===e[i].exact)assert.ok(e[i-1].id<e[i].id);}
 const cutoff=e.at(-1).exact,required=new Set(e.filter(x=>x.exact>cutoff+1e-6).map(x=>x.id)),exactIDs=new Set(e.map(x=>x.id));
 const hits=a.filter(x=>required.has(x.id)).length,ties=a.filter(x=>Math.abs(x.exact-cutoff)<=1e-6).length;
 return {tieRecall:(hits+Math.min(50-required.size,ties))/50,strictRecall:a.filter(x=>exactIDs.has(x.id)).length/50,regret:Math.max(0,mean(e.map(x=>x.exact))-mean(a.map(x=>x.exact))),nativeError:Math.max(...a.map(x=>Math.abs(x.native-x.exact)))};
}
export function evaluate(root){
 const run=JSON.parse(fs.readFileSync(path.join(root,'cost-results.json'))),pins=fs.readFileSync(path.join(root,'SOURCE_PINS.json'));
 assert.equal(hash(pins),run.sourcePinsSHA256);assert.equal(hash(fs.readFileSync(path.join(root,'PROTOCOL.md'))),run.protocolSHA256);
 const cells=[],failed=[];const keyed=new Map();
 for(const c of run.results){
  assert.ok(['control','candidate'].includes(c.arm));assert.ok([0,1,2].includes(c.repeat));assert.ok([32,128].includes(c.dimension));assert.ok([256,1024].includes(c.initial));
  const key=[c.repeat,c.dimension,c.initial,c.arm].join('/');assert.ok(!keyed.has(key));
  const data=fs.readFileSync(path.join(root,c.label+'.txt'));assert.equal(hash(data),c.sha256);assert.equal(data.length,c.bytes);const rows=[];let summary;
  for(const line of data.toString().split('\n')){let at=line.indexOf('OVERLAY_ROW ');if(at>=0)rows.push(JSON.parse(line.slice(at+12)));at=line.indexOf('OVERLAY_SUMMARY ');if(at>=0){assert.equal(summary,undefined);summary=JSON.parse(line.slice(at+16));}}
  if(c.status!==0){failed.push({key,label:c.label,status:c.status,rows:rows.length});keyed.set(key,{failed:true});continue;}
  assert.equal(rows.length,128);assert.ok(summary);assert.equal(summary.dimension,c.dimension);assert.equal(summary.initial,c.initial);assert.equal(summary.final_records,c.initial+128);assert.equal(summary.whole_goal_validation,false);assert.match(data.toString(),/--- PASS: TestResearchOverlayCostV1/);
  rows.forEach((r,i)=>{assert.equal(r.step,i);assert.equal(r.records,c.initial+i+1);assert.ok(r.commit_ns>0&&r.query_ns>0);});
  const metrics=rows.map(quality),cell={key,arm:c.arm,repeat:c.repeat,dimension:c.dimension,initial:c.initial,rows:128,meanTieRecall:mean(metrics.map(x=>x.tieRecall)),meanStrictRecall:mean(metrics.map(x=>x.strictRecall)),meanRegret:mean(metrics.map(x=>x.regret)),maxNativeScoreError:Math.max(...metrics.map(x=>x.nativeError)),commitMeanMS:mean(rows.map(x=>x.commit_ns))/1e6,commitP99MS:quantile(rows.map(x=>x.commit_ns),.99)/1e6,commitMaxMS:Math.max(...rows.map(x=>x.commit_ns))/1e6,queryMeanMS:mean(rows.map(x=>x.query_ns))/1e6,queryP99MS:quantile(rows.map(x=>x.query_ns),.99)/1e6,initialMS:summary.initial_ns/1e6,totalMS:summary.total_ns/1e6,totalAllocBytes:summary.total_alloc_bytes};
  cell.qualityPass=cell.meanTieRecall>=.95&&cell.meanRegret<=.005;cells.push(cell);keyed.set(key,{cell,exactDigest:hash(JSON.stringify(rows.map(x=>x.exact_top)))});
 }
 const paired=[];
 for(let r=0;r<3;r++)for(const d of[32,128])for(const n of[256,1024]){const a=keyed.get([r,d,n,'control'].join('/')),b=keyed.get([r,d,n,'candidate'].join('/'));if(!a?.cell||!b?.cell)continue;assert.equal(a.exactDigest,b.exactDigest);paired.push({repeat:r,dimension:d,initial:n,commitMeanRatio:b.cell.commitMeanMS/a.cell.commitMeanMS,queryP99Ratio:b.cell.queryP99MS/a.cell.queryP99MS,tieRecallDifference:b.cell.meanTieRecall-a.cell.meanTieRecall});}
 return {verified:true,completed:run.results.length===24&&failed.length===0,commands:run.results.length,completedRows:cells.length*128,failed,cells,paired,qualityPass:cells.length===24&&cells.every(x=>x.qualityPass),peakRetainedPayload:'not captured',peakRSS:'not captured',finiteSyntheticGeometryOnly:true,untouchedAgentUtility:false,loadedFreshness:false,wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){const root=path.dirname(fileURLToPath(import.meta.url)),result=evaluate(root);fs.writeFileSync(path.join(root,'evaluation.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({...result,cells:undefined,paired:undefined}));}
