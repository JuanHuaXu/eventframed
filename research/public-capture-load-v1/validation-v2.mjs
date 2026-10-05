import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const primary=JSON.parse(fs.readFileSync(path.join(output,"manifest.json")));
const checkout=path.join(primary.temporary,"control");
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const names=["TestResearchGuardedAdmissionPersistentBoundary","TestResearchBatchValidationParity","TestResearchTemporalFeedbackBoundaries"];
const sourceNames=["internal/service/research_durable_persistent_guard_test.go","internal/service/research_batch_validation_test.go","internal/service/research_feedback_temporal_test.go"];
const sourceHashes={...primary.sources.control,...Object.fromEntries(sourceNames.map(name=>[name,hash(fs.readFileSync(path.join(checkout,name)))]))};
const args=["test","-mod=readonly","-race","-count=1","-v","-run","^("+names.join("|")+")$","./internal/service"];
fs.writeFileSync(path.join(output,"validation-v2-manifest.json"),JSON.stringify({sourceHashes,
  runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),command:["go",...args],
  reason:"Original opt-in identity characterization skipped; require non-optional guard regression execution",
  scope:"Adjacent persistent guard, batch parity and temporal guard regression; not full public-load race coverage",
  requireNoSkips:true,productionTouched:false},null,2)+"\n");
const command=spawnSync("go",args,{cwd:checkout,encoding:"utf8",maxBuffer:16*1024*1024});
const text=(command.stdout??"")+(command.stderr??"");
fs.writeFileSync(path.join(output,"guard-race-v2.txt"),text);
const results={status:command.status,signal:command.signal,error:command.error?.message??null,
  transcriptSHA256:hash(Buffer.from(text)),dataRaceReported:/WARNING: DATA RACE/.test(text),
  skipped:/--- SKIP:/.test(text),requiredSuites:names,
  executedSuites:names.filter(name=>text.includes("--- PASS: "+name+" ")),
  fullLoadRaceValidated:false,wholeGoalValidation:false};
results.passed=results.status===0&&!results.dataRaceReported&&!results.skipped&&results.executedSuites.length===names.length;
fs.writeFileSync(path.join(output,"validation-v2-results.json"),JSON.stringify(results,null,2)+"\n");
console.log(JSON.stringify({completed:true,passed:results.passed,skipped:results.skipped,executedSuites:results.executedSuites}));
