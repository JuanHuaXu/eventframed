import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const paths=process.argv.slice(2);assert.equal(paths.length,2);
const bytes=paths.map(p=>readFileSync(p)),[strict,moving]=bytes.map(b=>JSON.parse(b));
assert.equal(strict.artifactSHA256,moving.artifactSHA256);
assert.equal(strict.records.length,2688);assert.equal(moving.records.length,2688);
strict.records.forEach((r,i)=>{const s=moving.records[i];assert.deepEqual([r.phase,r.case,r.index,r.schedule],[s.phase,s.case,s.index,s.schedule]);});
const mean=a=>a.reduce((v,x)=>v+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++)for(let arm=0;arm<4;arm++){
  const indices=strict.records.flatMap((r,i)=>r.phase===phase&&r.case===c&&r.schedule===schedule?[i]:[]);
  groups.push({phase,case:c,schedule,segment,arm,gain:interval(indices.map(i=>moving.records[i].scores[arm][segment]-strict.records[i].scores[arm][segment]))});
}
const support=[];
for(let schedule=0;schedule<2;schedule++)for(let arm=0;arm<4;arm++){
  const fs=strict.records.filter(r=>r.schedule===schedule).flatMap(r=>r.fits[arm]);
  support.push({schedule,arm,fits:fs.length,meanLabels:mean(fs.map(f=>f.origins.length)),empty:fs.filter(f=>f.origins.length===0).length});
}
console.log(JSON.stringify({scope:'Consumed paired cadence-matched comparison; positive gain favors strict version restriction',hashes:Object.fromEntries(paths.map((p,i)=>[p,createHash('sha256').update(bytes[i]).digest('hex')])),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),groups,support},null,2));
