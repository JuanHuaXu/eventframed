import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const m=JSON.parse(fs.readFileSync(path.join(output,"repaired-manifest.json")));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const check=()=>{for(const [arm,files] of Object.entries(m.sources))for(const [file,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(m.paths[arm],file)))!==h)throw Error("source drift: "+arm+"/"+file);};
const regressions=["TestResearchTransactionSharedShardAdmissionV1","TestResearchLogicalParentDiscoveryV1","TestResearchQuantizationDeclarationReopenV1","TestResearchQuantizationOptionalCodecV1"];
const guards=["TestResearchGuardedAdmissionPersistentBoundary","TestResearchBatchValidationParity","TestResearchTemporalFeedbackBoundaries"];
const plan=[];
for(const race of [false,true])plan.push({arm:"fork",name:race?"regression-race":"regression",required:regressions,args:["test","-mod=readonly",...(race?["-race"]:[]),"-count=1","-v","-timeout","90s","-run","^("+regressions.join("|")+")$","./libravdb","./internal/storage/singlefile"]});
plan.push({arm:"fork",name:"storage",required:[],args:["test","-mod=readonly","-count=1","-v","-timeout","4m","./internal/storage/singlefile"]});
plan.push({arm:"fork",name:"adjacent-library",required:[],args:["test","-mod=readonly","-count=1","-v","-timeout","4m","-run","(Tx|Transaction|Sharded|Sharding|Config|Persistence|Lifecycle)","./libravdb"]});
for(const arm of ["control","candidate"]){
  plan.push({arm,name:"store",required:arm==="candidate"?["TestResearchShardedTopologyReopenV1"]:[],args:["test","-mod=readonly","-count=1","-v","-timeout","6m","./internal/store/libravdbstore"]});
  plan.push({arm,name:"vet",required:[],args:["vet","-mod=readonly","./internal/store/libravdbstore","./internal/service"]});
  plan.push({arm,name:"guard-race",required:guards,args:["test","-mod=readonly","-race","-count=1","-v","-timeout","3m","-run","^("+guards.join("|")+")$","./internal/service"]});
}
plan.push({arm:"candidate",name:"topology-race",required:["TestResearchShardedTopologyReopenV1"],args:["test","-mod=readonly","-race","-count=1","-v","-timeout","90s","-run","^TestResearchShardedTopologyReopenV1$","./internal/store/libravdbstore"]});
check();
fs.writeFileSync(path.join(output,"repaired-validation-manifest.json"),JSON.stringify({runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),sourceManifestSHA256:hash(fs.readFileSync(path.join(output,"repaired-manifest.json"))),plan,knownOptInSkipsNotExecuted:true,fullPublicLoadRaceValidated:false},null,2)+"\n");
const commands=[];
for(const c of plan){
  check();console.log(JSON.stringify({started:true,arm:c.arm,name:c.name}));
  const r=spawnSync("go",c.args,{cwd:m.paths[c.arm],env:{...process.env,GOMAXPROCS:"10"},encoding:"utf8",maxBuffer:64*1024*1024});
  const text=(r.stdout??"")+(r.stderr??""),file=`repaired-${c.arm}-${c.name}.txt`;
  fs.writeFileSync(path.join(output,file),text);check();
  const passNames=[...text.matchAll(/^--- PASS: (\S+)/gm)].map(x=>x[1]),skipNames=[...text.matchAll(/^--- SKIP: (\S+)/gm)].map(x=>x[1]),failNames=[...text.matchAll(/^--- FAIL: (\S+)/gm)].map(x=>x[1]);
  const row={...c,command:["go",...c.args],status:r.status,signal:r.signal,error:r.error?.message??null,transcript:file,sha256:hash(text),passNames,skipNames,failNames,dataRaceReported:/WARNING: DATA RACE/.test(text)};
  row.passed=row.status===0&&!row.dataRaceReported&&row.failNames.length===0&&c.required.every(name=>passNames.includes(name))&&(c.name==="vet"||passNames.length>0);
  commands.push(row);fs.writeFileSync(path.join(output,"repaired-validation-progress.json"),JSON.stringify({completed:false,commands},null,2)+"\n");
  console.log(JSON.stringify({finished:true,arm:c.arm,name:c.name,status:row.status,passed:row.passed,passCount:passNames.length,skipCount:skipNames.length,failNames}));
}
const result={completed:true,commands,passed:commands.every(c=>c.passed),fullPublicLoadRaceValidated:false,wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
fs.writeFileSync(path.join(output,"repaired-validation-results.json"),JSON.stringify(result,null,2)+"\n");
console.log(JSON.stringify({completed:true,passed:result.passed}));
