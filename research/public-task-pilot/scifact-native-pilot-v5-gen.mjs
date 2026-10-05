// Local loader repair only; installed libraries/configuration remain immutable.
import fs from'node:fs';import assert from'node:assert/strict';import crypto from'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='research/public-task-pilot/scifact-native-pilot-v4-run.mjs',dest='research/public-task-pilot/scifact-native-pilot-v5-run.mjs';
const raw=fs.readFileSync(source,'utf8'),f=JSON.parse(fs.readFileSync('research/public-task-pilot/nativepool-v4/manifest.json'));assert.equal(hash(raw),f.inputs[process.cwd()+'/'+source].sha256);
const oldConfig='research/public-task-pilot/scifact-native-pilot-v4-config.yaml',config='research/public-task-pilot/scifact-native-pilot-v5-config.yaml';
const c=fs.readFileSync(oldConfig,'utf8').replaceAll('nativepool-v4','nativepool-v5');fs.writeFileSync(config,c,{flag:'wx',mode:0o600});
let out=raw.replaceAll('nativepool-v4','nativepool-v5').replaceAll('scifact-native-pilot-v4-config','scifact-native-pilot-v5-config').replaceAll('scifact-native-pilot-v4-run','scifact-native-pilot-v5-run');
const marker="const env={HOME:root+'/home',TMPDIR:root+'/tmp',PATH:'/usr/bin:/bin:/usr/sbin:/sbin',";
const replacement=marker+"\n DYLD_LIBRARY_PATH:llama,";assert.equal(out.split(marker).length,2);out=out.replace(marker,replacement);
assert.equal(out.replace(replacement,marker).replaceAll('nativepool-v5','nativepool-v4').replaceAll('scifact-native-pilot-v5-config','scifact-native-pilot-v4-config').replaceAll('scifact-native-pilot-v5-run','scifact-native-pilot-v4-run'),raw);
fs.writeFileSync(dest,out,{flag:'wx',mode:0o600});
fs.writeFileSync('research/public-task-pilot/scifact-native-pilot-v5-generation.json',JSON.stringify({source,dest,sourceSHA256:hash(raw),destSHA256:hash(out),configSHA256:hash(c),completeInverseEquality:true,
 changed:'owned child DYLD_LIBRARY_PATH for bare libllama lookup',scientificClientBackendAndRssCeilingUnchanged:true},null,2)+'\n',{flag:'wx',mode:0o600});console.log('runner inverse PASS; owned child loader path only, no installed changes');
