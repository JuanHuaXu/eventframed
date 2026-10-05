import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const m=JSON.parse(fs.readFileSync(path.join(output,"manifest.json")));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const suites=["TestResearchGuardedAdmissionPersistentBoundary","TestResearchBatchValidationParity","TestResearchTemporalFeedbackBoundaries"];
const plan=[];
for (const arm of ["control","candidate"]) {
  const checkout=path.join(m.temporary,arm);
  for (const [name,expected] of Object.entries(m.sources[arm])) if(hash(fs.readFileSync(path.join(checkout,name)))!==expected) throw Error("source changed: "+name);
  plan.push({arm,name:"store",args:["test","-mod=readonly","-count=1","-v","-timeout","6m","./internal/store/libravdbstore"]});
  plan.push({arm,name:"vet",args:["vet","./internal/store/libravdbstore","./internal/service"]});
  plan.push({arm,name:"guard-race",args:["test","-mod=readonly","-race","-count=1","-v","-timeout","3m","-run","^("+suites.join("|")+")$","./internal/service"]});
}
fs.writeFileSync(path.join(output,"validation-manifest.json"),JSON.stringify({runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),
  plan,sourceHashes:m.sources,knownStoreOptInSkips:"reported separately, not executed coverage",requiredGuardSuites:suites,
  requireTopology:"candidate store suite must execute and PASS TestResearchShardedTopologyReopenV1",
  fullPublicLoadRaceValidated:false,productionTouched:false},null,2)+"\n");
const commands=[];
for (const command of plan) {
  console.log(JSON.stringify({started:true,arm:command.arm,name:command.name}));
  const r=spawnSync("go",command.args,{cwd:path.join(m.temporary,command.arm),env:{...process.env,GOMAXPROCS:"10"},encoding:"utf8",maxBuffer:32*1024*1024});
  const text=(r.stdout??"")+(r.stderr??"");
  const file=`${command.arm}-${command.name}.txt`;
  fs.writeFileSync(path.join(output,file),text);
  const passNames=[...text.matchAll(/^--- PASS: (\S+)/gm)].map(x=>x[1]);
  const skipNames=[...text.matchAll(/^--- SKIP: (\S+)/gm)].map(x=>x[1]);
  const failNames=[...text.matchAll(/^--- FAIL: (\S+)/gm)].map(x=>x[1]);
  const row={...command,command:["go",...command.args],status:r.status,signal:r.signal,error:r.error?.message??null,
    transcript:file,sha256:hash(Buffer.from(text)),passNames,skipNames,failNames,dataRaceReported:/WARNING: DATA RACE/.test(text)};
  row.requiredExecuted=command.name!=="guard-race"||suites.every(name=>passNames.includes(name));
  row.topologyExecuted=command.arm!=="candidate"||command.name!=="store"||passNames.includes("TestResearchShardedTopologyReopenV1");
  commands.push(row);
  fs.writeFileSync(path.join(output,"validation-progress.json"),JSON.stringify({completed:false,commands},null,2)+"\n");
  console.log(JSON.stringify({finished:true,arm:row.arm,name:row.name,status:row.status,passed:passNames.length,skipped:skipNames.length,failed:failNames}));
}
const result={completed:true,commands,passed:commands.every(c=>c.status===0&&!c.dataRaceReported&&c.requiredExecuted&&c.topologyExecuted),
  fullPublicLoadRaceValidated:false,wholeGoalValidation:false};
fs.writeFileSync(path.join(output,"validation-results.json"),JSON.stringify(result,null,2)+"\n");
console.log(JSON.stringify({completed:true,passed:result.passed}));
