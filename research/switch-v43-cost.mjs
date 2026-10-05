// Prospective cost screen; refusal to overlap V41 is part of the protocol.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync} from 'node:child_process';

const label=process.argv[2];assert(/^[a-z0-9-]+$/.test(label??''));
const finish=JSON.parse(fs.readFileSync('research/moment-v41-normal/completed.json'));
assert(finish.source_unchanged&&finish.stage==='normal','V41 must be terminal first');
let live='';try{live=execFileSync('pgrep',['-fl','TestMoment(Experiment|StudyAudit)V41'],{encoding:'utf8'})}catch(e){assert.equal(e.status,1)}
assert.equal(live,'','no concurrent V41 collector/audit');
const root='research/switch-v43-cost-'+label;assert(!fs.existsSync(root));fs.mkdirSync(root,{mode:0o700});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const prior=JSON.parse(fs.readFileSync('research/moment-v41-normal/freeze.json')).files;
for(const[p,h]of Object.entries(prior))assert.equal(hash(fs.readFileSync(p)),h,'unchanged underlying learner '+p);
const paths=[...new Set([...Object.keys(prior),...fs.readdirSync('internal/researchswitch').filter(p=>p.endsWith('.go')).map(p=>'internal/researchswitch/'+p),'internal/researchswitchref/reference.go','docs/experiments/mmm-switch-v43-contract.md','docs/experiments/mmm-switch-v43-cost-contract.md','docs/experiments/mmm-switch-v43-span-rescue.md','research/switch-v43-cost.mjs'])].sort();
const files=Object.fromEntries(paths.map(p=>[p,hash(fs.readFileSync(p))]));
function save(name,x){const f=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(f,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(f)}finally{fs.closeSync(f)}}
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
save('freeze.json',{time:new Date().toISOString(),files,protocol:'prospective computational screen only',go:execFileSync('go',['env','GOVERSION','GOOS','GOARCH'],{encoding:'utf8'}).trim(),host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,arch:os.arch(),memory:os.totalmem(),load:os.loadavg()},constructorGateBytes:8<<20,learnerLoopGateNS:400000000});
const checks=[];
function run(name,args){unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync('go',args,{encoding:'utf8',timeout:310000,maxBuffer:8<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
const f=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(f,log);fs.fsyncSync(f)}finally{fs.closeSync(f)}
const row={name,args,start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,logSHA256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name);return log}
try{
run('race',['test','-race','./internal/researchswitch','./internal/researchswitchref','-v','-count=1']);
run('vet',['vet','./internal/researchswitch','./internal/researchswitchref']);
const log=run('benchmarks',['test','./internal/researchswitch','-run','^$','-bench','^(BenchmarkSwitch|BenchmarkPool)','-benchmem','-benchtime=1x','-count=3','-timeout=5m']);
const rows=[...log.matchAll(/^(Benchmark\S+)\s+(\d+)\s+([\d.]+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/gm)].map(m=>({name:m[1],iterations:+m[2],ns:+m[3],bytes:+m[4],allocs:+m[5]}));
assert.equal(rows.length,18,'all six benchmarks three repetitions');assert(rows.every(r=>r.iterations===1&&Number.isFinite(r.ns)&&r.ns>0));
const constructor=rows.filter(r=>r.name.startsWith('BenchmarkPoolConstructor150'));
const loops=rows.filter(r=>r.name.startsWith('BenchmarkPoolLoop150x16'));
assert.equal(constructor.length,3);assert.equal(loops.length,9);
const report={time:new Date().toISOString(),rows,constructorMaxBytes:Math.max(...constructor.map(r=>r.bytes)),loopMaxNS:Math.max(...loops.map(r=>r.ns)),preliminaryAllocationPass:constructor.every(r=>r.bytes<=8*1024*1024),preliminaryLoopPass:loops.every(r=>r.ns<=400000000),qualityOrProductionAdoption:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
save('cost-report.json',report);unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,qualityOrProductionAdoption:false});console.log(JSON.stringify(report,null,2));
}catch(e){save('failure.json',{time:new Date().toISOString(),checks,error:e.message,qualityOrProductionAdoption:false});throw e}
