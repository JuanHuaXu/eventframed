import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),parent=path.join(out,'../sharded-quality-v1'),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const derivations=[];
for(const [source,dest]of [['evaluate.mjs','service-evaluate.mjs'],['independent.mjs','service-independent.mjs'],['archive.mjs','service-archive.mjs']]){
 const raw=fs.readFileSync(path.join(parent,source)),text=raw.toString().replaceAll('"run-results.json"','"service-run-results.json"').replaceAll('"evaluation.json"','"service-evaluation.json"').replaceAll('"independent-results.json"','"service-independent-results.json"');
 const target=path.join(out,dest);if(fs.existsSync(target))throw Error('preserve derivation');fs.writeFileSync(target,text);derivations.push({source:'../sharded-quality-v1/'+source,sourceSHA256:hash(raw),destination:dest,destinationSHA256:hash(Buffer.from(text)),change:'only local artifact basenames; arithmetic and phase gates unchanged'});
}
fs.writeFileSync(path.join(out,'SERVICE_EVALUATOR_DERIVATION.json'),JSON.stringify({derivations,wholeGoalValidation:false},null,2)+'\n');console.log(JSON.stringify({prepared:true}));
