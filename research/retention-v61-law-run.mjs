// Independently freeze/test the joint-law adapter without changing the selector.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const root='research/retention-v61-law',hash=b=>crypto.createHash('sha256').update(b).digest('hex');assert(!fs.existsSync(root));
const base=JSON.parse(fs.readFileSync('research/retention-v61-component/freeze.json'));
const files=Object.fromEntries(['internal/researchretentionlaw/law.go','internal/researchretentionlaw/law_test.go','research/retention-v61-law-run.mjs'].map(p=>[p,hash(fs.readFileSync(p))]));
for(const[p,h]of Object.entries({...base.files,...base.protectedFiles}))assert.equal(hash(fs.readFileSync(p)),h,p);
fs.mkdirSync(root,{mode:0o700});
for(const[p,h]of Object.entries(files)){const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(p,dest,fs.constants.COPYFILE_EXCL);assert.equal(hash(fs.readFileSync(dest)),h)}
const save=(name,value)=>fs.writeFileSync(root+'/'+name,JSON.stringify(value,null,2)+'\n',{flag:'wx',mode:0o600});
save('freeze.json',{files,selectorInputs:base.files,protectedFiles:base.protectedFiles,scope:'joint-law adapter/component integration; NOT actual base-expert integration or empirical rescue'});
const checks=[];
try {
 for(const[name,args]of [['race',['test','-race','./internal/researchretentionlaw','-count=1','-v','-timeout=5m']],['vet',['vet','./internal/researchretentionlaw']]]) {
  const start=new Date().toISOString(),begin=performance.now(),raw=execFileSync('go',args,{timeout:360000,stdio:['ignore','pipe','pipe']});
  fs.writeFileSync(root+'/'+name+'.log',raw,{flag:'wx',mode:0o600});checks.push({name,args,exitCode:0,start,end:new Date().toISOString(),wallMS:performance.now()-begin,logSHA256:hash(raw)});
 }
 for(const[p,h]of Object.entries({...base.files,...base.protectedFiles,...files}))assert.equal(hash(fs.readFileSync(p)),h,p);
 save('completed.json',{checks,allTestsPass:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE',notScientificValidation:true,productionChanged:false});console.log(JSON.stringify(checks,null,2));
}catch(e){save('failure.json',{message:e.message,stdout:e.stdout?.toString(),stderr:e.stderr?.toString(),checks});throw e}
