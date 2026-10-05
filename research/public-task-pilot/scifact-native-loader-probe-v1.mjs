// Controlled env-transmission probe only; no library injection or config writes.
import fs from'node:fs';import{spawnSync}from'node:child_process';import assert from'node:assert/strict';
const root='research/public-task-pilot/native-loader-probe-v1';fs.mkdirSync(root,{mode:0o700});
const node=process.execPath,env={HOME:process.cwd()+'/'+root,PATH:'/usr/bin:/bin',DYLD_LIBRARY_PATH:'/owned-test-library-path'};
const code='console.log(JSON.stringify({dyld:process.env.DYLD_LIBRARY_PATH??null}))';
const cases=[['direct',node,['-e',code]],['nice-wrapper','/usr/bin/nice',['-n','10',node,'-e',code]]];
const results=[];for(const[name,cmd,args]of cases){const r=spawnSync(cmd,args,{env,encoding:'utf8',timeout:10000});assert.equal(r.status,0);results.push({name,command:[cmd,...args],code:r.status,signal:r.signal,stdout:r.stdout,stderr:r.stderr,parsed:JSON.parse(r.stdout)});}
assert.equal(results[0].parsed.dyld,env.DYLD_LIBRARY_PATH);assert.equal(results[1].parsed.dyld,null);
fs.writeFileSync(root+'/result.json',JSON.stringify({time:new Date().toISOString(),env,results,wrapperDropsDyldVerified:true,noProtectedProgramLibraryInjection:true},null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(results.map(r=>({case:r.name,...r.parsed}))));
