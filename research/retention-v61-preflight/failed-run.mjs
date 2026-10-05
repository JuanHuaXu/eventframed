// Prospective isolated component freeze. Failure artifacts are never overwritten.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync, spawn} from 'node:child_process';
const root = 'research/retention-v61-component';
assert(!fs.existsSync(root));
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p) {const h = crypto.createHash('sha256'); for await (const b of fs.createReadStream(p)) h.update(b); return h.digest('hex');}
const v60 = JSON.parse(fs.readFileSync('research/tree-v60-diagnostic/freeze.json'));
for (const [p,h] of Object.entries({...v60.files, ...v60.protectedFiles})) assert.equal(await fileHash(p), h, p);
const raw = execFileSync('go', ['run', './cmd/research-go-list-closure', './internal/researchretention'], {maxBuffer: 32*1024*1024});
const closure = JSON.parse(raw), compiler = new Set();
for (const p of closure) {
 if (!p.Dir?.startsWith(process.cwd() + '/')) continue;
 for (const key of ['GoFiles','CgoFiles','EmbedFiles']) for (const n of p[key] ?? []) {
  const absolute = path.resolve(p.Dir,n); assert(absolute.startsWith(process.cwd() + '/'));
  compiler.add(path.relative(process.cwd(),absolute));
 }
}
const extras = ['go.mod','go.sum','research/retention-v61-component-run.mjs','research/tree-v61-retention-direction.md'];
const files = Object.fromEntries([...new Set([...compiler,...extras])].sort().map(p => [p,hash(fs.readFileSync(p))]));
fs.mkdirSync(root,{mode:0o700});
function save(p,x) {fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});}
save('compiler-closure.json',closure);
for (const [p,h] of Object.entries(files)) {
 const b = fs.readFileSync(p); assert.equal(hash(b),h);
 const dest = root+'/source/'+p; fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,b,{flag:'wx',mode:0o600});
}
save('freeze.json',{time:new Date().toISOString(),files,compilerFiles:[...compiler].sort(),compilerClosureSHA256:hash(raw),protectedFiles:v60.protectedFiles,
 toolchain:JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOFLAGS','GOTOOLCHAIN'],{encoding:'utf8'})),node:process.version,
 host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},scope:'standalone selector, no base-model integration, scientific outcome or loaded serving claim',goals:Array(7).fill('OPEN')});
const commands = [['race',['test','-race','./internal/researchretention','-count=1','-v','-timeout=10m']],
 ['vet',['vet','./internal/researchretention']],
 ['benchmark',['test','./internal/researchretention','-run','^$','-bench','Benchmark','-benchtime=100ms','-count=3']]];
const checks=[];
async function unchanged() {for (const [p,h] of Object.entries({...v60.files,...v60.protectedFiles,...files})) assert.equal(await fileHash(p),h,p);}
try {
 for (const [name,args] of commands) {
  await unchanged(); const start = new Date().toISOString(), begin = performance.now(),fd = fs.openSync(root+'/'+name+'.log','wx',0o600);
  const env = {...process.env};for (const k of Object.keys(env)) if (k.startsWith('EVENTFRAME_')) delete env[k];
  let code;
  try {code = await new Promise((resolve,reject) => {
   const child = spawn('go',args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>child.kill('SIGTERM'),15*60*1000);
   child.on('error',e=>{clearTimeout(timer);reject(e)});
   for (const stream of [child.stdout,child.stderr]) stream.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});
   child.on('close',c=>{clearTimeout(timer);resolve(c??-1)});
  });} finally {fs.fsyncSync(fd);fs.closeSync(fd);}
  const row={name,args,start,end:new Date().toISOString(),wallMS:performance.now()-begin,exitCode:code,logSHA256:await fileHash(root+'/'+name+'.log')};
  checks.push(row);save(name+'-command.json',row);await unchanged();assert.equal(code,0,name);
 }
 const artifacts={};for (const p of fs.readdirSync(root).filter(p=>/\.(json|log)$/.test(p))) artifacts[p]=await fileHash(root+'/'+p);
 save('completed.json',{checks,artifacts,goals:Array(7).fill('OPEN'),goal:'ACTIVE',productionChanged:false,notScientificValidation:true});
} catch(e) {save('failure.json',{error:e.message,checks,goal:'ACTIVE',goals:Array(7).fill('OPEN')});throw e;}
