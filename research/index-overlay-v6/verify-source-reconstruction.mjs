import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),base=JSON.parse(fs.readFileSync(path.join(out,'../index-overlay-v1/manifest.json'))).inheritedRoot,root=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-overlay-reconstruction-')),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'SOURCE_RECONSTRUCTION.json')))throw Error('preserve reconstruction');
const arms=[];
for(let v=1;v<=6;v++) {
 const dir=path.join(out,'../index-overlay-v'+v),m=JSON.parse(fs.readFileSync(path.join(dir,'manifest.json'))),pins=JSON.parse(fs.readFileSync(path.join(dir,'SOURCE_PINS.json'))),a=JSON.parse(fs.readFileSync(path.join(dir,'SOURCE_ARCHIVE.json')));
 for(const arm of ['control','candidate']) {
  const dest=path.join(root,'v'+v+'-'+arm);fs.cpSync(base,dest,{recursive:true});
  for(const [file,p]of Object.entries(a.sourceFiles))if(file.startsWith(arm+'-')) {
   const target=path.join(dest,p.target);assert.equal(fs.existsSync(target)?hash(fs.readFileSync(target)):null,p.beforeSHA256);
   assert.equal(hash(fs.readFileSync(path.join(dir,file))),p.patchSHA256);
   execFileSync('patch',['-p1','-i',path.join(dir,file)],{cwd:dest});assert.equal(hash(fs.readFileSync(target)),p.afterSHA256);
  }
  for(const [file,f]of Object.entries(a.fixtures))if(Object.hasOwn(pins.files,path.join(m.paths[arm],f.target))) {
   assert.equal(hash(fs.readFileSync(path.join(dir,file))),f.sha256);fs.copyFileSync(path.join(dir,file),path.join(dest,f.target));
  }
  let checked=0;
  for(const [file,h]of Object.entries(pins.files))if(file.startsWith(m.paths[arm]+path.sep)) { assert.equal(hash(fs.readFileSync(path.join(dest,path.relative(m.paths[arm],file)))),h,'reconstructed source drift '+v+'/'+arm+'/'+file);checked++; }
  arms.push({version:v,arm,checkedSourceFiles:checked,reconstructedByteIdentical:true});
 }
}
fs.writeFileSync(path.join(out,'SOURCE_RECONSTRUCTION.json'),JSON.stringify({verified:true,root,arms,scope:'all prospectively pinned dependency source files; execution and system/dependency environment not reproduced',wholeGoalValidation:false},null,2)+'\n');console.log(JSON.stringify({verified:true,arms:arms.length,checked:arms.reduce((n,x)=>n+x.checkedSourceFiles,0)}));
