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
  if (hash(fs.readFileSync(path.join(checkout,name))) !== expected) throw Error("Rescue source changed: "+name);
}
const cpus = Number(execFileSync("sysctl",["-n","hw.logicalcpu"],{encoding:"utf8"}).trim());
if (cpus !== 10) throw Error("Frozen host default changed");
fs.writeFileSync(path.join(output,"worker-cap-manifest.json"),JSON.stringify({revision:primary.revision,
  sourceHashes:primary.sources.control, workers:{default:10,cap:2},frontiers:[50,200],
  protocolSHA256:hash(fs.readFileSync(path.join(output,"WORKER_CAP_PROTOCOL.md"))),
  runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),
  privateDataUsed:false,productionTouched:false,wholeGoalValidation:false,maskCandidateApplied:false},null,2)+"\n");
const quantile = (values, fraction) => [...values].sort((a,b)=>a-b)[Math.ceil(values.length*fraction)-1]/1e6;
const commands=[];
for (const frontier of [50,200]) for (const [pair,order] of [[0,["default","cap"]],[1,["cap","default"]]]) for (const arm of order) {
  const workers=arm==="cap"?2:10;
  console.log(JSON.stringify({started:true,frontier,pair,arm,workers}));
  const args=["test","-mod=readonly","-count=1","-v","-timeout","4m","-run","^TestResearchPublicCaptureLoadV1$","./internal/service"];
  const command=spawnSync("go",args,{cwd:checkout,env:{...process.env,EVENTFRAME_PUBLIC_FRONTIER:String(frontier),GOMAXPROCS:String(workers)},encoding:"utf8",maxBuffer:32*1024*1024});
  const text=(command.stdout??"")+(command.stderr??"");
  const file=`worker-cap-${frontier}-${pair}-${arm}.txt`;
  fs.writeFileSync(path.join(output,file),text);
  const match=text.match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/);
  const trace=match?JSON.parse(match[1]):null;
  const measurements=trace?.functional?Object.fromEntries(["recall","feedback","live_age","capture"].map(name=>[name,
    Object.fromEntries([["p50",.5],["p95",.95],["p99",.99],["max",1]].map(([key,fraction])=>[key+"_ms",quantile(trace[name+"_ns"],fraction)]))])):null;
  commands.push({frontier,pair,arm,workers,command:["go",...args],status:command.status,signal:command.signal,
    error:command.error?.message??null,transcript:file,transcriptSHA256:hash(Buffer.from(text)),trace,measurements});
  fs.writeFileSync(path.join(output,"worker-cap-progress.json"),JSON.stringify({completed:false,commands},null,2)+"\n");
  console.log(JSON.stringify({completedCommand:true,frontier,pair,arm,status:command.status,functional:trace?.functional??false,measurements}));
}
const paired=[];
for (const frontier of [50,200]) for (const pair of [0,1]) {
  const get=arm=>commands.find(c=>c.frontier===frontier&&c.pair===pair&&c.arm===arm);
  const control=get("default"),candidate=get("cap");
  const ratio=control.measurements&&candidate.measurements?candidate.measurements.recall.p99_ms/control.measurements.recall.p99_ms:null;
  paired.push({frontier,pair,ratio,pass:ratio!==null&&ratio<=.90});
}
const results={completed:true,commands,paired,functionalPass:commands.every(c=>c.trace?.functional),
  capAbsoluteTimingPass:commands.filter(c=>c.arm==="cap").every(c=>c.status===0),
  pairedRescuePass:paired.every(p=>p.pass),designCaptureTemplates:288,independentFactCount:"not established",
  primaryTimingVerdictChanged:false,productionTouched:false,wholeGoalValidation:false,populationTailGuarantee:false};
fs.writeFileSync(path.join(output,"worker-cap-results.json"),JSON.stringify(results,null,2)+"\n");
console.log(JSON.stringify({completed:true,functionalPass:results.functionalPass,capAbsoluteTimingPass:results.capAbsoluteTimingPass,pairedRescuePass:results.pairedRescuePass}));
