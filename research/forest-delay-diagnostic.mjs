import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),rows=bytes.toString().trim().split('\n').slice(1).map(JSON.parse);
assert.equal(rows.length,192);
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const cells=[];
for(const phase of ['design','confirmation'])for(const name of ['copy_bit','copy_xor2','noise10_bit','reverses_xor2'])for(const schedule of ['Immediate','Delayed'])for(const start of [256,320,384,448]){
 const rs=rows.filter(r=>r.Phase===phase&&r.Case===name);assert.equal(rs.length,16);
 const scores=rs.map(r=>{const frames=r[schedule].Frames.slice(start,start+64);assert.equal(frames.length,64);return Array.from({length:5},(_,j)=>mean(frames.map(f=>{const p=f.Predictions[0];const v=j===4?p.p:p.experts[j];assert(Number.isFinite(v)&&v>=0&&v<=1);return (v-Number(f.Y))**2;})));});
 cells.push({phase,case:name,schedule,start,end:start+64,meanBrier:Array.from({length:5},(_,j)=>mean(scores.map(s=>s[j])))});
}
const out={sourceHash:crypto.createHash('sha256').update(bytes).digest('hex'),columns:['base','short_nested','long','neutral','served_uniform'],cells,limits:'Consumed-data diagnostic on actually acquired masks. Cohort means of pure experts do not bound arbitrary mixtures or establish the cause of recovery failure. No policy was tuned or promoted.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({cells:cells.length,sourceHash:out.sourceHash}));
