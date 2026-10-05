import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {verify as geometry} from './verify.mjs';
import {verify as service} from './service-verify.mjs';
const root=path.dirname(fileURLToPath(import.meta.url)),read=f=>fs.readFileSync(path.join(root,f)),json=f=>JSON.parse(read(f)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(root,'CHECKPOINT_VERIFICATION.json')))throw Error('preserve verification');
const v=await geometry(root),s=await service(root);assert.equal(v.qualityPass,true);assert.equal(s.counterfactualEqualityPass,true);
assert.equal(json('SOURCE_RECONSTRUCTION.json').verified,true);assert.equal(json('SOURCE_RECONSTRUCTION.json').arms.length,12);
assert.equal(json('evaluator-audit.json').status,0);assert.equal(json('service-evaluator-audit.json').status,0);
const lineage=[];
for(let n=1;n<=6;n++) {
 const dir=path.join(root,'../index-overlay-v'+n),r=f=>JSON.parse(fs.readFileSync(path.join(dir,f))),pins=r('SOURCE_PINS.json');
 for(const [file,h]of Object.entries(pins.files))assert.equal(hash(fs.readFileSync(file)),h,'source drift '+file);
 const e=r('evaluation.json');let serviceResult=null;
 if(fs.existsSync(path.join(dir,'service-evaluation.json'))) {const x=r('service-evaluation.json');serviceResult={functional:x.functionalPass,ann:x.annAbsolutePass,packet:x.packetQualityPass,counterfactual:x.counterfactualEqualityPass};}
 else if(fs.existsSync(path.join(dir,'service-run-results.json')))serviceResult={completed:r('service-run-results.json').completed};
 lineage.push({version:n,geometryCompleted:e.completed,geometryQuality:e.qualityPass,geometryCommands:e.commands,service:serviceResult});
}
assert.equal(lineage[0].geometryCompleted,false);assert.equal(lineage[1].service.completed,false);assert.equal(lineage[2].service.completed,false);assert.equal(lineage[3].service.counterfactual,false);assert.equal(lineage[4].service.ann,false);assert.equal(lineage[5].service.counterfactual,true);
fs.writeFileSync(path.join(root,'CHECKPOINT_VERIFICATION.json'),JSON.stringify({verified:true,geometry:v,service:s,lineage,sourceReconstructionFiles:6590,originalSevenGoals:'all OPEN',weeklyUsedPercent:50,goalStatus:'active',peakRSS:'not captured',loadedFreshnessValidated:false,compactionStillSynchronous:true,productionTouched:false,privateDataUsed:false,automaticPush:false,wholeGoalValidation:false},null,2)+'\n');
for(let n=1;n<=6;n++) {
 const dir=path.join(root,'../index-overlay-v'+n),excluded=new Set(['SHA256SUMS']);
 if(fs.existsSync(path.join(dir,'LOCAL_ONLY.json')))for(const x of JSON.parse(fs.readFileSync(path.join(dir,'LOCAL_ONLY.json'))).omittedFiles)excluded.add(x.file);
 const rows=[];for(const e of fs.readdirSync(dir,{withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name)))if(e.isFile()&&!excluded.has(e.name))rows.push(hash(fs.readFileSync(path.join(dir,e.name)))+'  '+e.name);
 fs.writeFileSync(path.join(dir,'SHA256SUMS'),rows.join('\n')+'\n');
}
console.log(JSON.stringify({verified:true,lineage,localCheckpointOnly:true,wholeGoalValidation:false}));
