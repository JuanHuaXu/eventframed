import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const manifest=JSON.parse(fs.readFileSync(path.join(output,"manifest.json")));
const checkout=path.join(manifest.temporary,"control");
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
for (const [name,expected] of Object.entries(manifest.sources.control)) {
  if (hash(fs.readFileSync(path.join(checkout,name)))!==expected) throw Error("Validation source changed: "+name);
}
const protocol={sourceHashes:manifest.sources.control,
  runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),
  scope:"Static validation of new fixture, adjacent durable identity and temporal guard race regressions; NOT full public-load race coverage or timing",productionTouched:false};
fs.writeFileSync(path.join(output,"validation-manifest.json"),JSON.stringify(protocol,null,2)+"\n");
const commands=[];
for (const [name,args] of [["vet",["vet","./internal/service"]],
  ["guard-race",["test","-mod=readonly","-race","-count=1","-v","-run","^(TestResearchDurableIdentityAudit|TestResearchTemporalFeedbackBoundaries)$","./internal/service"]]]) {
  const result=spawnSync("go",args,{cwd:checkout,encoding:"utf8",maxBuffer:16*1024*1024});
  const text=(result.stdout??"")+(result.stderr??"");
  const file=name+".txt";
  fs.writeFileSync(path.join(output,file),text);
  commands.push({command:["go",...args],status:result.status,signal:result.signal,error:result.error?.message??null,
    transcript:file,sha256:hash(Buffer.from(text)),dataRaceReported:/WARNING: DATA RACE/.test(text)});
}
const results={commands,passed:commands.every(c=>c.status===0&&!c.dataRaceReported),fullLoadRaceValidated:false,wholeGoalValidation:false};
fs.writeFileSync(path.join(output,"validation-results.json"),JSON.stringify(results,null,2)+"\n");
console.log(JSON.stringify({completed:true,passed:results.passed,fullLoadRaceValidated:false}));
