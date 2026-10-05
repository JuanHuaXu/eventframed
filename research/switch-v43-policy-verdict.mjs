// Post-freeze verifier: one FIXED policy must pass BOTH normal cohorts.
// This supplements, never overwrites, the frozen readback's convenience flag.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const modes=['static','slow','round'];
function commonPass(splits){return modes.filter(m=>splits.every(s=>s.failedCells[m]===0))}
const negatives=[
 [{failedCells:{static:0,slow:1,round:1}},{failedCells:{static:1,slow:0,round:1}}],
 [{failedCells:{static:1,slow:1,round:0}},{failedCells:{static:0,slow:1,round:1}}],
];for(const pair of negatives)assert.deepEqual(commonPass(pair),[],'cannot choose a different policy per split');
assert.deepEqual(commonPass([{failedCells:{static:1,slow:0,round:1}},{failedCells:{static:1,slow:0,round:1}}]),['slow']);
const stage=process.argv[2];assert(['diagnostic','normal'].includes(stage));
const root='research/switch-v43-study-'+stage,r=JSON.parse(fs.readFileSync(root+'/readback.json'));
assert.equal(r.stage,stage);assert.equal(r.goal,'ACTIVE');
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80,'current authoritative weekly usage required');
const result={time:new Date().toISOString(),stage,sourceSHA256:crypto.createHash('sha256').update(fs.readFileSync('research/switch-v43-policy-verdict.mjs')).digest('hex'),negativeControls:2,positiveControls:1,qualityAdoption:false,passedFixedPolicies:[],allSevenWholeGoals:'OPEN',goal:'ACTIVE',frozenSourcesChanged:false,weeklyUsage:usage,originalReadbackUsage:r.weeklyUsage,usageMetadataCorrected:r.weeklyUsage!==usage};
if(stage==='normal'){
 const splits=['design','confirmation'].map(k=>r.summaries[k]);
 for(const s of splits){assert.equal(s.worlds,448);assert.equal(Object.keys(s.cells).length,84);for(const m of modes){const count=Object.values(s.cells).filter(c=>!c[m].pass).length;assert.equal(count,s.failedCells[m])}}
 result.passedFixedPolicies=commonPass(splits);result.qualityAdoption=result.passedFixedPolicies.length>0;
 result.frozenConvenienceFlag=r.qualityAdoption;result.convenienceFlagAgreement=r.qualityAdoption===result.qualityAdoption;
}else{assert.equal(r.qualityAdoption,false);result.diagnosticOnly=true}
const fd=fs.openSync(root+'/policy-verdict.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(result,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}console.log(JSON.stringify(result,null,2));
