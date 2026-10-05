// Preserve exact isolated research source bytes, not merely their hashes.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const label=process.argv[2];assert(/^[a-z0-9-]+$/.test(label??''),'new label required');
const root='research/moment-switch-checkpoint-'+label;
const selected=process.argv[3];
const record=selected==='cost'?'research/switch-v43-cost-initial/freeze.json':selected==='study-diagnostic'?'research/switch-v43-study-diagnostic/freeze.json':selected==='unit-spans'?'research/switch-v43-cost-unit-spans/freeze.json':'research/switch-v43-log-domain-exact-oracle/freeze.json';
const records=['research/moment-v41-normal/freeze.json',record];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const files={};
for(const p of records){const freeze=JSON.parse(fs.readFileSync(p));for(const[q,h]of Object.entries(freeze.files)){assert(files[q]===undefined||files[q]===h,'inconsistent research sources');files[q]=h}}
for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,'research source changed '+p);
assert(!fs.existsSync(root),'exclusive checkpoint root');fs.mkdirSync(root,{mode:0o700});
function save(p,b){fs.mkdirSync(path.dirname(p),{recursive:true,mode:0o700});const f=fs.openSync(p,'wx',0o600);try{fs.writeFileSync(f,b);fs.fsyncSync(f)}finally{fs.closeSync(f)}}
for(const[p,h]of Object.entries(files)){const b=fs.readFileSync(p);assert.equal(hash(b),h);save(root+'/source/'+p,b);assert.equal(hash(fs.readFileSync(root+'/source/'+p)),h)}
for(const p of records)save(root+'/records/'+p,fs.readFileSync(p));
for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,'source changed while copying '+p);
save(root+'/manifest.json',Buffer.from(JSON.stringify({time:new Date().toISOString(),files,records,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionEdited:false,qualityAdoption:false,completedRecordPresent:fs.existsSync(path.join(path.dirname(record),'completed.json')),resume:'Inspect specific live process/session and completed artifacts; never restart collection solely because output is buffered. A completed diagnostic is not a normal quality result.'},null,2)+'\n'));
console.log(JSON.stringify({root,sources:Object.keys(files).length,verified:true,allSevenWholeGoals:'OPEN'}));
