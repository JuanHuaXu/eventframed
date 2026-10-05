import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {evaluate} from './evaluate.mjs';
import {independent} from './independent.mjs';
export function verify(root,{sources=true}={}){
 const read=f=>fs.readFileSync(path.join(root,f)),json=f=>JSON.parse(read(f)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
 const pins=json('SOURCE_PINS.json');assert.equal(pins.prospective,true);
 if(sources)for(const[f,h]of Object.entries(pins.files))assert.equal(hash(fs.readFileSync(f)),h,'source drift '+f);
 const actual=evaluate(root);assert.deepEqual(actual,json('evaluation.json'));const math=independent(root);assert.deepEqual(math,json('independent-results.json'));
 const expectedComplete=json('cost-results.json').results.length===24;
 if(expectedComplete){assert.equal(actual.completed,true);assert.equal(actual.commands,24);assert.equal(actual.completedRows,3072);assert.equal(actual.paired.length,12);}
 assert.equal(actual.wholeGoalValidation,false);return {verified:true,commands:actual.commands,completed:actual.completed,qualityPass:actual.qualityPass,independentRows:math.checkedRows,currentSourceReadback:sources,wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url)))console.log(JSON.stringify(verify(path.dirname(fileURLToPath(import.meta.url)))));
