import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),rows=JSON.parse(bytes),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(rows.length,192);assert.equal(new Set(rows.map(r=>r.Seed)).size,192);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const ci=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const dep of [false,true])for(const target of ['bit','parity4'])for(const n of [64,128,4096]) {
 const rs=rows.filter(r=>r.Dependent===dep&&r.Target===target&&r.N===n);assert.equal(rs.length,16);
 assert.deepEqual(rs.map(r=>r.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
 for(const r of rs){assert.equal(r.Seed,2026092170+Number(dep)*1000000+Number(target==='parity4')*100000+n*20+r.Index);for(const a of [r.Old,r.Coherent]){assert.equal(a.length,4);for(const v of a)assert(Number.isFinite(v)&&v>=.0475-1e-12&&v<=1);}assert.equal(r.Old[0],r.Coherent[0]);}
 cells.push({dependent:dep,target,n,masks:[0,1,31,511].map((mask,i)=>({mask,old:mean(rs.map(r=>r.Old[i])),coherent:mean(rs.map(r=>r.Coherent[i])),gain:ci(rs.map(r=>r.Old[i]-r.Coherent[i]))}))});
}
const sources={},hashes={};for(const name of ['internal/observationlearners/coherent_count_test.go','internal/observationlearners/conditional.go','internal/observation/controller.go','internal/observation/forecast.go','docs/experiments/mmm-coherent-count-v1-contract.md','research/coherent-count-summary.mjs']){const b=fs.readFileSync(name);sources[name]=b.toString();hashes[name]=hash(b);}
const out={sourceHash:hash(bytes),sources,hashes,cells,limits:'Fixed-mask finite-population diagnostic, descriptive paired mean +/-3.5SE over training fits; no adaptive or real-task benefit established.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({sourceHash:out.sourceHash,cells},null,2));
