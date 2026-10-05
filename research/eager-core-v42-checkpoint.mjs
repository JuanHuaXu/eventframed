// Exact preflight snapshot; recover unrelated research revisions from their
// preserved source archive, never revert the user's current runtime files.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const label=process.argv[2];assert(/^[a-z0-9-]+$/.test(label??''));
const prior='research/eager-core-v42-publication-accounting';
const freeze=JSON.parse(fs.readFileSync(prior+'/freeze.json'));
const complete=JSON.parse(fs.readFileSync(prior+'/completed.json'));
assert(complete.sourceUnchanged&&complete.checks.every(c=>c.code===0));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source=new Map(),recovered=[];
for(const[p,h]of Object.entries(freeze.files)){
 let b=fs.readFileSync(p);
 if(hash(b)!==h){const old='research/switch-v43-extreme-weight-regression/source/'+p;assert(fs.existsSync(old),'cannot reproduce changed source '+p);b=fs.readFileSync(old);assert.equal(hash(b),h,'archive does not match '+p);recovered.push(p)}
 source.set(p,b);
}
const root='research/eager-core-v42-checkpoint-'+label;assert(!fs.existsSync(root));fs.mkdirSync(root,{mode:0o700});
function save(p,b){fs.mkdirSync(path.dirname(p),{recursive:true,mode:0o700});const f=fs.openSync(p,'wx',0o600);try{fs.writeFileSync(f,b);fs.fsyncSync(f)}finally{fs.closeSync(f)}}
for(const[p,b]of source){save(root+'/source/'+p,b);assert.equal(hash(fs.readFileSync(root+'/source/'+p)),freeze.files[p])}
for(const c of complete.checks)assert.equal(hash(fs.readFileSync(prior+'/'+c.name+'.log')),c.logSHA256);
for(const name of fs.readdirSync(prior))if(/\.(json|log)$/.test(name))save(root+'/evidence/'+name,fs.readFileSync(prior+'/'+name));
save(root+'/manifest.json',Buffer.from(JSON.stringify({time:new Date().toISOString(),original:prior,files:freeze.files,recoveredFromPreservedResearchSource:recovered,verified:true,productionEdited:false,loadedLatencyAdoption:false,allSevenWholeGoals:'OPEN'},null,2)+'\n'));
console.log(JSON.stringify({root,verified:true,sources:source.size,recovered,productionEdited:false}));
