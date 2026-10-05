// Mechanical runner copy: same experiment/client, repaired owned configuration.
import fs from'node:fs';import assert from'node:assert/strict';import crypto from'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='research/public-task-pilot/scifact-native-pilot-v1-run.mjs',dest='research/public-task-pilot/scifact-native-pilot-v2-run.mjs';
const raw=fs.readFileSync(source,'utf8'),f=JSON.parse(fs.readFileSync('research/public-task-pilot/nativepool-v1/manifest.json'));
assert.equal(hash(raw),f.inputs[process.cwd()+'/'+source].sha256);
const converted=raw.replaceAll('nativepool-v1','nativepool-v2').replaceAll('scifact-native-pilot-v1-config','scifact-native-pilot-v2-config').replaceAll('scifact-native-pilot-v1-run','scifact-native-pilot-v2-run');
assert.equal(converted.replaceAll('nativepool-v2','nativepool-v1').replaceAll('scifact-native-pilot-v2-config','scifact-native-pilot-v1-config').replaceAll('scifact-native-pilot-v2-run','scifact-native-pilot-v1-run'),raw);
fs.writeFileSync(dest,converted,{flag:'wx',mode:0o600});
fs.writeFileSync('research/public-task-pilot/scifact-native-pilot-v2-generation.json',JSON.stringify({source,dest,sourceSHA256:hash(raw),destSHA256:hash(converted),completeInverseEquality:true,scientificClientUnchanged:true},null,2)+'\n',{flag:'wx',mode:0o600});
console.log('runner inverse comparison PASS; only owned root/config/self-fingerprint changes');
