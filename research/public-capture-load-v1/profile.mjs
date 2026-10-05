import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { execFileSync, spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output = path.dirname(fileURLToPath(import.meta.url));
const primary = JSON.parse(fs.readFileSync(path.join(output,"manifest.json")));
const checkout = path.join(primary.temporary,"control");
const hash = b => crypto.createHash("sha256").update(b).digest("hex");
for (const [name, expected] of Object.entries(primary.sources.control)) {
  if (hash(fs.readFileSync(path.join(checkout,name))) !== expected) throw Error("Diagnostic source changed: "+name);
}
const binary = path.join(primary.temporary,"profile-service.test");
const args = ["test","-mod=readonly","-count=1","-v","-timeout","4m","-run","^TestResearchPublicCaptureLoadV1$",
  "-o",binary,"-cpuprofile",path.join(output,"cpu.pprof"),"-mutexprofile",path.join(output,"mutex.pprof"),
  "-blockprofile",path.join(output,"block.pprof"),"./internal/service"];
const manifest = { revision:primary.revision, sourceHashes:primary.sources.control,
  protocolSHA256:hash(fs.readFileSync(path.join(output,"DIAGNOSTIC_PROTOCOL.md"))),
  runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))), frontier:200,
  command:["go",...args], primaryTimingVerdictChanged:false, diagnosticOnly:true, productionTouched:false, privateDataUsed:false };
fs.writeFileSync(path.join(output,"diagnostic-manifest.json"),JSON.stringify(manifest,null,2)+"\n");
const command = spawnSync("go",args,{cwd:checkout,env:{...process.env,EVENTFRAME_PUBLIC_FRONTIER:"200"},encoding:"utf8",maxBuffer:32*1024*1024});
const text=(command.stdout??"")+(command.stderr??"");
fs.writeFileSync(path.join(output,"diagnostic-test.txt"),text);
const reports=[];
for (const [name,args] of [["cpu-top",["-top","-cum","-nodecount=30","cpu.pprof"]],
  ["cpu-recall",["-top","-cum","-nodecount=30","-focus=Service.*Recall","cpu.pprof"]],
  ["mutex-top",["-top","-cum","-nodecount=30","mutex.pprof"]],
  ["block-top",["-top","-cum","-nodecount=30","block.pprof"]]]) {
  const report=execFileSync("go",["tool","pprof",...args],{cwd:output,encoding:"utf8"});
  const file=name+".txt";
  fs.writeFileSync(path.join(output,file),report);
  reports.push({file,sha256:hash(Buffer.from(report))});
}
fs.writeFileSync(path.join(output,"diagnostic-results.json"),JSON.stringify({status:command.status,signal:command.signal,
  error:command.error?.message??null,transcriptSHA256:hash(Buffer.from(text)),reports,
  profiles:Object.fromEntries(["cpu.pprof","mutex.pprof","block.pprof"].map(file=>[file,hash(fs.readFileSync(path.join(output,file)))])),
  diagnosticOnly:true,wholeGoalValidation:false},null,2)+"\n");
console.log(JSON.stringify({completed:true,status:command.status,diagnosticOnly:true}));
