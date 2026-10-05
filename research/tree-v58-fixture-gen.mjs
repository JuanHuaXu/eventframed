import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const sources={},outputs={};
function clone(src,dest,transform){assert(!fs.existsSync(dest));const b=fs.readFileSync(src);sources[src]=hash(b);const value=transform(b.toString());fs.writeFileSync(dest,value,{flag:'wx',mode:0o600});outputs[dest]=hash(value)}
clone('internal/researchdispersion/paired_v57_generated_test.go','internal/researchdispersion/tree_v58_generated_test.go',s=>s.replaceAll('V57','V58').replaceAll('researchpairedvalue','researchtree').replaceAll('researchpairedvalueref','researchtreeref').replaceAll('paired.Config{Strength: 2, Hazard: 1. / 16}','paired.Config{Depth: 7, Window: 600, Strength: 2}'));
clone('research/paired-v57-readback.mjs','research/tree-v58-readback.mjs',s=>s.replaceAll('paired-v57','tree-v58'));
fs.writeFileSync('research/tree-v58-fixture-generation.json',JSON.stringify({sources,outputs,behavioralPatchesSeparate:true},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(outputs).length}));
