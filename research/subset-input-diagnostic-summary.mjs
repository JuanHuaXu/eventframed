import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import path from 'node:path';
const [input, output] = process.argv.slice(2); assert(input && output);
const bytes=fs.readFileSync(input), data=JSON.parse(bytes);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
for(const [name,digest] of Object.entries(data.Hashes)) {
  assert.equal(hash(data.Sources[name]),digest);
  assert.equal(hash(fs.readFileSync(path.join('internal/observationlearners',name))),digest);
}
assert.equal(data.Records.length,64);
assert.deepEqual(data.Arms,['uniform','empirical','oracle_input']);
assert.equal(new Set(data.Records.map(r=>r.Seed)).size,64);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const generator of ['fair','copy_bit2_to_bit0']) for(const n of [64,128]) {
  const rs=data.Records.filter(r=>r.Generator===generator && r.N===n);
  assert.equal(rs.length,16);
  assert.deepEqual(rs.map(r=>r.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
  for(const r of rs) {
    assert.equal(r.Seed,2026092140+n*100+r.Index+(generator==='fair'?0:100000));
    for(const v of [...r.Partial,...r.Full]) assert(Number.isFinite(v)&&v>=0&&v<=1);
    assert(Math.abs(r.Full[0]-r.Full[1])<1e-12 && Math.abs(r.Full[0]-r.Full[2])<1e-12);
    for(const v of r.Partial) assert(v>=r.OraclePartialFloor-1e-12);
  }
  cells.push({generator,n,partial:data.Arms.map((_,i)=>mean(rs.map(r=>r.Partial[i]))),full:mean(rs.map(r=>r.Full[0])),
    empiricalGain:interval(rs.map(r=>r.Partial[0]-r.Partial[1])),oracleGain:interval(rs.map(r=>r.Partial[0]-r.Partial[2])),floor:rs[0].OraclePartialFloor});
}
const result={sourceHash:hash(bytes),cells,limits:'Exact finite-population risk, uncertainty across 16 training fits; fixed-mask component diagnostic only. Oracle input distribution is evaluation-only.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(result,null,2));
