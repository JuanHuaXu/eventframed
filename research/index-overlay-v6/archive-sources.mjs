import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),base=JSON.parse(fs.readFileSync(path.join(out,'../index-overlay-v1/manifest.json'))).inheritedRoot,m=JSON.parse(fs.readFileSync(path.join(out,'manifest.json'))),pins=JSON.parse(fs.readFileSync(path.join(out,'SOURCE_PINS.json'))),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'SOURCE_ARCHIVE.json')))throw Error('preserve archive');
for(const[f,h]of Object.entries(pins.files))if(hash(fs.readFileSync(f))!==h)throw Error('source drift '+f);
const sourceFiles={},fixtures={};
for(const arm of ['control','candidate']) {
 for(const file of ['internal/index/interfaces.go','libravdb/tx.go',...(arm==='candidate'?['internal/index/research_overlay.go']:[])]) {
  const original=path.join(base,file),target=path.join(m.paths[arm],file),actual=fs.readFileSync(target),exists=fs.existsSync(original),before=exists?fs.readFileSync(original):Buffer.alloc(0);if(before.equals(actual))continue;
  const r=spawnSync('diff',['-u',exists?original:'/dev/null',target],{encoding:'utf8'});if(r.status!==1||!r.stdout.startsWith('--- '))throw Error('diff failure');
  const lines=r.stdout.split('\n');lines[0]='--- '+(exists?'a/'+file:'/dev/null');lines[1]='+++ b/'+file;const patch=lines.join('\n'),name=arm+'-'+file.replaceAll('/','-')+'.patch';fs.writeFileSync(path.join(out,name),patch);sourceFiles[name]={target:file,beforeSHA256:exists?hash(before):null,afterSHA256:hash(actual),patchSHA256:hash(Buffer.from(patch))};
 }
}
for(const[file,h]of Object.entries(pins.files))if(file.startsWith(m.paths.candidate+path.sep)&&/research.*_test\.go$/.test(file)) {
 const target=path.relative(m.paths.candidate,file),name='compiled-'+target.replaceAll('/','-')+'.txt';fs.copyFileSync(file,path.join(out,name));fixtures[name]={target,sha256:h};
}
fs.writeFileSync(path.join(out,'SOURCE_ARCHIVE.json'),JSON.stringify({base:'inherited sharded-capture-load-v1 dependency-v3 fork of LibraVDB v1.6.13',sourceFiles,fixtures,wholeGoalValidation:false},null,2)+'\n');console.log(JSON.stringify({archived:true,patches:Object.keys(sourceFiles).length,fixtures:Object.keys(fixtures).length}));
