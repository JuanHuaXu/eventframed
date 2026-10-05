import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const root=path.dirname(fileURLToPath(import.meta.url)),target=process.argv[2],hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.ok(target&&path.isAbsolute(target));assert.ok(fs.existsSync(path.join(target,'.git')));
assert.equal(execFileSync('git',['status','--porcelain'],{cwd:target,encoding:'utf8'}),'');
const verdict=JSON.parse(fs.readFileSync(path.join(root,'CHECKPOINT_VERIFICATION.json')));assert.equal(verdict.verified,true);assert.equal(verdict.currentSourceReadback,true);assert.equal(verdict.commands,42);
const repair=JSON.parse(fs.readFileSync(path.join(root,'VERIFIER_REPAIR.json'))),pins=JSON.parse(fs.readFileSync(path.join(root,'EVALUATOR_PINS.json')));
for(const[f,h]of Object.entries(pins.files)){
 const r=repair.files.find(x=>x.file===f);assert.equal(hash(fs.readFileSync(path.join(root,f))),r?r.afterSHA256:h);
 if(r){assert.equal(r.beforeSHA256,h);assert.equal(hash(fs.readFileSync(path.join(root,r.before))),h);}
}
const archives=JSON.parse(fs.readFileSync(path.join(root,'ARCHIVES.json')));assert.equal(archives.archives.length,8);assert.equal(archives.losslessVerified,true);
for(const a of archives.archives)assert.equal(hash(fs.readFileSync(path.join(root,a.archive))),a.archiveSHA256);
const excluded=new Set(JSON.parse(fs.readFileSync(path.join(root,'LOCAL_ONLY.json'))).files),dest=path.join(target,'research/overlay-shard-factorial-v1');assert.equal(fs.existsSync(dest),false);
const names=fs.readdirSync(root).filter(f=>!excluded.has(f)&&f!=='SHA256SUMS').sort();
for(const f of names){assert.ok(fs.lstatSync(path.join(root,f)).isFile());assert.ok(fs.statSync(path.join(root,f)).size<100*1024*1024);}
const sums=names.map(f=>hash(fs.readFileSync(path.join(root,f)))+'  '+f).join('\n')+'\n';fs.writeFileSync(path.join(root,'SHA256SUMS'),sums);
fs.mkdirSync(dest,{recursive:true});for(const f of[...names,'SHA256SUMS']){fs.copyFileSync(path.join(root,f),path.join(dest,f));assert.equal(hash(fs.readFileSync(path.join(dest,f))),hash(fs.readFileSync(path.join(root,f))));}
fs.copyFileSync(path.join(root,'../../research-direction.md'),path.join(target,'research-direction.md'));
console.log(JSON.stringify({stagedCopy:true,files:names.length+1,rawNumericalOriginalsExcluded:excluded.size,target:dest,productionTouched:false,automaticPush:false}));
