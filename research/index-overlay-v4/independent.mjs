import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
// Independent arithmetic, no import of the primary evaluator.
export function independent(root){
 const run=JSON.parse(fs.readFileSync(path.join(root,'cost-results.json'))),primary=JSON.parse(fs.readFileSync(path.join(root,'evaluation.json')));
 let checkedRows=0,checkedCells=0;
 for(const c of run.results){if(c.status!==0)continue;let n=0,hitsTotal=0,strictTotal=0,regretTotal=0;
  for(const line of fs.readFileSync(path.join(root,c.label+'.txt'),'utf8').split('\n')){
   const at=line.indexOf('OVERLAY_ROW ');if(at<0)continue;const r=JSON.parse(line.slice(at+12));const actual=new Set(r.actual.map(x=>x.id)),cutoff=r.exact_top[49].exact;
   let must=0,have=0,boundary=0,strict=0,bestSum=0,actualSum=0;
   for(const x of r.exact_top){bestSum+=x.exact;if(actual.has(x.id))strict++;if(x.exact>cutoff+1e-6){must++;if(actual.has(x.id))have++;}}
   for(const x of r.actual){actualSum+=x.exact;if(Math.abs(x.exact-cutoff)<=1e-6)boundary++;}
   hitsTotal+=(have+Math.min(50-must,boundary))/50;strictTotal+=strict/50;regretTotal+=Math.max(0,(bestSum-actualSum)/50);n++;
  }
  const p=primary.cells.find(x=>x.key===[c.repeat,c.dimension,c.initial,c.arm].join('/'));assert.ok(p);assert.equal(n,128);
  for(const[a,b]of[[hitsTotal/n,p.meanTieRecall],[strictTotal/n,p.meanStrictRecall],[regretTotal/n,p.meanRegret]])assert.ok(Number.isFinite(a)&&Math.abs(a-b)<1e-12);
  assert.equal(p.qualityPass,hitsTotal/n>=.95&&regretTotal/n<=.005);checkedRows+=n;checkedCells++;
 }
 assert.equal(checkedRows,primary.completedRows);assert.equal(checkedCells,primary.cells.length);
 return {verified:true,checkedRows,checkedCells,independentArithmetic:true,independentVectorOracleRecomputation:false,wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){const root=path.dirname(fileURLToPath(import.meta.url)),r=independent(root);fs.writeFileSync(path.join(root,'independent-results.json'),JSON.stringify(r,null,2)+'\n');console.log(JSON.stringify(r));}
