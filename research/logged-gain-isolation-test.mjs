import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const [source,pop,policies,baseline]=process.argv.slice(2);
const file=path.resolve('research/logged-gain-project.mjs');
let code=fs.readFileSync(file,'utf8');
const marker='const origins=policy.selected.slice(1,4)';
assert.equal(code.split(marker).length,2);
// Instrument the current projector, not a copied implementation. Any access
// to teacher quantities or integrated branch risks must fail immediately.
code=code.replace(marker,`
  for(const step of s.Steps)for(const key of Object.keys(step))if(key!=='Y')Object.defineProperty(step,key,{get(){throw Error('hidden step field '+key);}});
  for(const key of Object.keys(s))if(!['Phase','Case','Index','Schedule','Steps'].includes(key))Object.defineProperty(s,key,{get(){throw Error('hidden source field '+key);}});
  for(const branch of p.Branches)for(const key of Object.keys(branch))if(!['Origin','ActualY','Predictions'].includes(key))Object.defineProperty(branch,key,{get(){throw Error('oracle branch field '+key);}});
  ${marker}`).replaceAll('import.meta.filename',JSON.stringify(file));
const output=execFileSync(process.execPath,['--input-type=module','-e',code,'unused',source,pop,policies],{encoding:'utf8',maxBuffer:8*1024*1024});
assert.equal(output,fs.readFileSync(baseline,'utf8'));
console.log(JSON.stringify({status:'PASS',records:JSON.parse(output).records.length,scope:'Current projector with throwing teacher, non-outcome source and integrated-risk getters exactly reproduces baseline'}));
