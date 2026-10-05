// Supplemental byte-level provenance check, independent of statistical readback.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const stage=process.argv[2];assert(['diagnostic','normal'].includes(stage));
const root='research/hybrid-v48-study-'+stage;
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));
const done=JSON.parse(fs.readFileSync(root+'/completed.json'));
assert(done.sourceUnchanged&&done.stage===stage&&done.checks.every(c=>c.code===0));
for(const[p,h]of Object.entries(freeze.files)){
 assert.equal(hash(fs.readFileSync(p)),h,'current source '+p);
 assert.equal(hash(fs.readFileSync(root+'/source/'+p)),h,'frozen copy '+p);
}
for(const[p,h]of Object.entries(done.artifacts))assert.equal(await digest(root+'/'+p),h,'artifact '+p);
for(const c of done.checks){
 assert.deepEqual(JSON.parse(fs.readFileSync(root+'/'+c.name+'-command.json')),c,'command identity');
 assert.equal(await digest(root+'/'+c.name+'.log'),c.logSHA256,'command log');
}
const cost=JSON.parse(fs.readFileSync(root+'/cost-report.json'));
assert(Number.isSafeInteger(cost.constructorMaxBytes)&&cost.constructorMaxBytes>0);
assert.equal(cost.preliminaryAllocationPass,cost.constructorMaxBytes<=8<<20);
assert.equal(cost.loops.length,9);
const seen=new Set();let loop=true;
for(const r of cost.loops){
 const key=r.mode+'/'+r.schedule;assert(!seen.has(key));seen.add(key);
 assert(['static','slow','round'].includes(r.mode)&&['immediate','fixed150','uniform299'].includes(r.schedule));
 const c=r.costs;for(const v of Object.values(c))assert(Number.isSafeInteger(v)&&v>0);
 assert.equal(c.AccountedNS,c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS);assert(c.ElapsedNS>=c.AccountedNS);
 loop&&=c.ElapsedNS<=400000000;
}
assert.equal(cost.preliminaryLoopPass,cost.preliminaryAllocationPass&&loop);
const r={time:new Date().toISOString(),stage,sourceCopies:Object.keys(freeze.files).length,artifacts:Object.keys(done.artifacts).length,commands:done.checks.length,costChecks:true,sourceAndArtifactHashes:true,allSevenWholeGoals:'OPEN',productionChanged:false};
const fd=fs.openSync(root+'/artifact-audit.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(r,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify(r,null,2));

