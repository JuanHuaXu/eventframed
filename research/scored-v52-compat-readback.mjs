// Early exact compatibility audit. No quality inference or source mutation.
import fs from 'node:fs';import assert from 'node:assert/strict';import readline from 'node:readline';import crypto from 'node:crypto';
const root='research/scored-v52-diagnostic',sha=b=>crypto.createHash('sha256').update(b).digest('hex');
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));for(const[p,h]of Object.entries(freeze.files))assert.equal(sha(fs.readFileSync(p)),h);
for(const name of ['legacy','log_mean'])assert.equal(JSON.parse(fs.readFileSync(root+'/'+name+'-command.json')).code,0);
async function* lines(name){const h=crypto.createHash('sha256'),input=fs.createReadStream(root+'/'+name+'.jsonl');input.on('data',b=>h.update(b));for await(const l of readline.createInterface({input,crlfDelay:Infinity}))yield JSON.parse(l);digests[name]=h.digest('hex');}
const digests={},a=lines('legacy')[Symbol.asyncIterator](),b=lines('log_mean')[Symbol.asyncIterator]();let worlds=0,pairs=0;
for(;;){const x=await a.next(),y=await b.next();assert.equal(x.done,y.done);if(x.done)break;if(worlds===0){assert.equal(x.value.Style,'legacy');assert.equal(y.value.Style,'log_mean');const{Style:_,...u}=x.value,{Style:__,...v}=y.value;assert.deepEqual(u,v);worlds++;continue;}assert.equal(x.value.Seed,y.value.Seed);assert.equal(x.value.Arms.length,9);assert.equal(y.value.Arms.length,9);for(let j=0;j<9;j++){const{Costs:_,...u}=x.value.Arms[j],{Costs:__,...v}=y.value.Arms[j];assert.deepEqual(u,v);pairs++;}worlds++;}
assert.equal(worlds-1,28);assert.equal(pairs,252);
const out={time:new Date().toISOString(),worlds:28,exactNonCostPairs:pairs,rawSHA256:digests,sourceHashesChecked:true,qualityAdoption:false};fs.writeFileSync(root+'/compat-readback.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out));
