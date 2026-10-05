import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),rows=bytes.toString().trim().split('\n').slice(1).map(JSON.parse);
assert.equal(rows.length,192);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const cells=['Immediate','Delayed'].map(schedule=>{
 const fits=rows.flatMap(r=>r[schedule].Fits).filter(f=>f.Clock>=256&&f.Clock<512);
 for(const f of fits){assert(f.Train.length>0);assert.deepEqual(f.Train,f.Origins.slice(0,-16));}
 return {schedule,fitCount:fits.length,meanAdditionalNewestTrainingAge:mean(fits.map(f=>f.Origins.at(-1)-f.Train.at(-1))),meanControlNewestAge:mean(fits.map(f=>f.Clock-f.Origins.at(-1))),meanHoldoutNewestAge:mean(fits.map(f=>f.Clock-f.Train.at(-1)))};
});
fs.writeFileSync(output,JSON.stringify({sourceHash:crypto.createHash('sha256').update(bytes).digest('hex'),cells,limits:'Descriptive averages over post-change publication records, not independent replicates or causal attribution of forecast harm.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells));
