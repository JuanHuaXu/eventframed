import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),base=JSON.parse(fs.readFileSync(path.join(out,'../index-overlay-v1/manifest.json'))).inheritedRoot;
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const receipts=[];
for(let v=1;v<=4;v++) {
  const root=path.join(out,'../index-overlay-v'+v),m=JSON.parse(fs.readFileSync(path.join(root,'manifest.json'))),pins=JSON.parse(fs.readFileSync(path.join(root,'SOURCE_PINS.json')));
  for(const [file,h] of Object.entries(pins.files))if(hash(fs.readFileSync(file))!==h)throw Error('source drift '+file);
  const sourceFiles={};
  for(const arm of ['control','candidate']) {
    const changed=['internal/index/interfaces.go','libravdb/tx.go'];
    if(arm==='candidate')changed.push('internal/index/research_overlay.go');
    for(const file of changed) {
      const target=path.join(m.paths[arm],file),original=path.join(base,file),actual=fs.readFileSync(target);
      const exists=fs.existsSync(original),before=exists?fs.readFileSync(original):Buffer.alloc(0);
      if(before.equals(actual))continue;
      const r=spawnSync('diff',['-u',exists?original:'/dev/null',target],{encoding:'utf8'});
      if(r.status!==1||!r.stdout.startsWith('--- '))throw Error('diff failure '+file);
      const lines=r.stdout.split('\n');lines[0]='--- '+(exists?'a/'+file:'/dev/null');lines[1]='+++ b/'+file;
      const patch=lines.join('\n'),name=arm+'-'+file.replaceAll('/','-')+'.patch',dest=path.join(root,name);
      if(fs.existsSync(dest))throw Error('preserve source archive '+dest);
      fs.writeFileSync(dest,patch);sourceFiles[name]={target:file,beforeSHA256:exists?hash(before):null,afterSHA256:hash(actual),patchSHA256:hash(Buffer.from(patch))};
    }
  }
  const fixtures={};
  for(const [file,h]of Object.entries(pins.files))if(file.startsWith(m.paths.candidate+path.sep)&&/research.*_test\.go$/.test(file)) {
    const relative=path.relative(m.paths.candidate,file),name='compiled-'+relative.replaceAll('/','-')+'.txt',dest=path.join(root,name),data=fs.readFileSync(file);
    if(fs.existsSync(dest))throw Error('preserve fixture '+dest);
    fs.writeFileSync(dest,data);fixtures[name]={target:relative,sha256:h};
  }
  const manifest={base:'inherited sharded-capture-load-v1 dependency-v3 fork of LibraVDB v1.6.13',sourceFiles,fixtures,wholeGoalValidation:false};
  fs.writeFileSync(path.join(root,'SOURCE_ARCHIVE.json'),JSON.stringify(manifest,null,2)+'\n');receipts.push({version:v,patches:Object.keys(sourceFiles).length,fixtures:Object.keys(fixtures).length});
}
console.log(JSON.stringify({archived:true,receipts}));
