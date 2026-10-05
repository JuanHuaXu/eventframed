// Asset-complete owned working directory; the scientific client is unchanged.
import fs from'node:fs';import assert from'node:assert/strict';import crypto from'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='research/public-task-pilot/scifact-native-pilot-v2-run.mjs',dest='research/public-task-pilot/scifact-native-pilot-v3-run.mjs';
const raw=fs.readFileSync(source,'utf8'),f=JSON.parse(fs.readFileSync('research/public-task-pilot/nativepool-v2/manifest.json'));
assert.equal(hash(raw),f.inputs[process.cwd()+'/'+source].sha256);
const cfg='research/public-task-pilot/scifact-native-pilot-v2-config.yaml',nextCfg='research/public-task-pilot/scifact-native-pilot-v3-config.yaml';
const config=fs.readFileSync(cfg,'utf8').replaceAll('nativepool-v2','nativepool-v3');fs.writeFileSync(nextCfg,config,{flag:'wx',mode:0o600});
let converted=raw.replaceAll('nativepool-v2','nativepool-v3').replaceAll('scifact-native-pilot-v2-config','scifact-native-pilot-v3-config').replaceAll('scifact-native-pilot-v2-run','scifact-native-pilot-v3-run');
const injection="const scanner=model+'/cognitive/cognitive_scanner.bin';\npaths.push(scanner);\nfs.copyFileSync(scanner,root+'/cognitive_scanner.bin',fs.constants.COPYFILE_EXCL);\npaths.push(root+'/cognitive_scanner.bin');\n";
const marker='const inputs={};';assert.equal(converted.split(marker).length,2);converted=converted.replace(marker,injection+marker);
const old="['-n','10',binary,'serve'],{env,stdio:",next="['-n','10',binary,'serve'],{env,cwd:root,stdio:";assert.equal(converted.split(old).length,2);converted=converted.replace(old,next);
const inverse=converted.replace(injection,'').replace(next,old).replaceAll('nativepool-v3','nativepool-v2').replaceAll('scifact-native-pilot-v3-config','scifact-native-pilot-v2-config').replaceAll('scifact-native-pilot-v3-run','scifact-native-pilot-v2-run');assert.equal(inverse,raw);
fs.writeFileSync(dest,converted,{flag:'wx',mode:0o600});
fs.writeFileSync('research/public-task-pilot/scifact-native-pilot-v3-generation.json',JSON.stringify({source,dest,sourceSHA256:hash(raw),destSHA256:hash(converted),configSHA256:hash(config),completeInverseEquality:true,scientificClientUnchanged:true,additionalInput:'installed cognitive scanner copy in owned cwd'},null,2)+'\n',{flag:'wx',mode:0o600});console.log('runner inverse PASS; only root, config and explicit owned scanner asset/cwd');
