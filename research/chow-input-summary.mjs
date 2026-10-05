import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';import path from 'node:path';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[h,...rows]=bytes.toString().trim().split('\n').map(JSON.parse),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(rows.length,576);assert.deepEqual(h.Masks,[0,1,3,31,511]);assert.deepEqual(h.Arms,['uniform','histogram','tree']);
for(const [name,digest]of Object.entries(h.Hashes)){assert.equal(hash(h.Sources[name]),digest);assert.equal(hash(fs.readFileSync(path.join('internal/observationlearners',name))),digest);}
assert.equal(new Set(rows.map(r=>r.Seed)).size,576);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const ci=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const [li,law]of ['uniform','copy','xor'].entries())for(const [ti,target]of ['bit','parity4'].entries())for(const [ni,noise]of [.05,.25,.5].entries())for(const n of [64,128]){
 const rs=rows.filter(r=>r.Law===law&&r.Target===target&&r.Noise===noise&&r.N===n);assert.equal(rs.length,16);
 assert.deepEqual(rs.map(r=>r.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
 for(const r of rs){assert.equal(r.Seed,2026092200+li*1000000+ti*100000+ni*10000+n*20+r.Index);assert.equal(r.Risk.length,3);for(const a of r.Risk){assert.equal(a.length,5);for(const v of a)assert(Number.isFinite(v)&&v>=noise*(1-noise)-1e-12&&v<=1);}for(const a of [1,2])assert(Math.abs(r.Risk[0][4]-r.Risk[a][4])<1e-12);}
 for(const [j,mask]of h.Masks.entries()){
  const treeGain=ci(rs.map(r=>r.Risk[0][j]-r.Risk[2][j])),histogramGain=ci(rs.map(r=>r.Risk[0][j]-r.Risk[1][j])),treeVsHistogram=ci(rs.map(r=>r.Risk[1][j]-r.Risk[2][j]));
  cells.push({law,target,noise,n,mask,risks:h.Arms.map((_,a)=>mean(rs.map(r=>r.Risk[a][j]))),treeGain,histogramGain,treeVsHistogram,nonHarm:treeGain.lower>=-.01,primary:law==='copy'&&target==='bit'&&noise===.05&&mask===1,gain:treeGain.mean>=.005&&treeGain.lower>0});
 }
}
const result={sourceHash:hash(bytes),capturedSources:Object.keys(h.Hashes).length,records:576,primaryPass:cells.filter(c=>c.primary&&c.gain).length,nonHarmFailures:cells.filter(c=>!c.nonHarm).length,cells,limits:'Exploratory fixed-mask fit uncertainty, not adaptive or real-task validation.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({...result,cells:cells.filter(c=>c.primary||!c.nonHarm)},null,2));
