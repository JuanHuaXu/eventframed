import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const read=name=>fs.readFileSync(path.join(output,name));
const m=JSON.parse(read("manifest.json"));
const validation=JSON.parse(read("validation-results.json"));
if (!validation.completed) throw Error("Preflight still incomplete");
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
for (const arm of ["control","candidate"]) for (const [name,expected] of Object.entries(m.sources[arm])) {
  if (hash(fs.readFileSync(path.join(m.temporary,arm,name)))!==expected) throw Error("source changed: "+name);
}
fs.writeFileSync(path.join(output,"run-manifest.json"),JSON.stringify({runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),
  sourceHashes:m.sources,preflightSHA256:hash(read("validation-results.json")),preflightPassed:validation.passed,
  frontiers:[50,200],runOrder:[["control","candidate"],["candidate","control"]],workers:10,
  functionalPreflightFailureCannotBeOverruledByTiming:true,productionTouched:false,privateDataUsed:false},null,2)+"\n");
const q=(values,fraction)=>[...values].sort((a,b)=>a-b)[Math.ceil(values.length*fraction)-1]/1e6;
const commands=[];
for (const frontier of [50,200]) for (const [pair,order] of [[0,["control","candidate"]],[1,["candidate","control"]]]) for (const arm of order) {
  console.log(JSON.stringify({started:true,frontier,pair,arm}));
  const args=["test","-mod=readonly","-count=1","-v","-timeout","4m","-run","^TestResearchPublicCaptureLoadV1$","./internal/service"];
  const command=spawnSync("go",args,{cwd:path.join(m.temporary,arm),env:{...process.env,EVENTFRAME_PUBLIC_FRONTIER:String(frontier),GOMAXPROCS:"10"},encoding:"utf8",maxBuffer:32*1024*1024});
  const text=(command.stdout??"")+(command.stderr??"");
  const file=`${frontier}-${pair}-${arm}.txt`;
  fs.writeFileSync(path.join(output,file),text);
  const match=text.match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/);
  const trace=match?JSON.parse(match[1]):null;
  const measurements=trace?.functional?Object.fromEntries(["recall","feedback","live_age","capture"].map(name=>[name,
    Object.fromEntries([["p50",.5],["p95",.95],["p99",.99],["max",1]].map(([key,fraction])=>[key+"_ms",q(trace[name+"_ns"],fraction)]))])):null;
  commands.push({frontier,pair,arm,command:["go",...args],status:command.status,signal:command.signal,error:command.error?.message??null,
    transcript:file,sha256:hash(Buffer.from(text)),trace,measurements});
  fs.writeFileSync(path.join(output,"progress.json"),JSON.stringify({completed:false,commands},null,2)+"\n");
  console.log(JSON.stringify({finished:true,frontier,pair,arm,status:command.status,functional:trace?.functional??false,measurements}));
}
const paired=[];
for (const frontier of [50,200]) for (const pair of [0,1]) {
  const get=arm=>commands.find(c=>c.frontier===frontier&&c.pair===pair&&c.arm===arm);
  const control=get("control"),candidate=get("candidate");
  const ratio=control.measurements&&candidate.measurements?candidate.measurements.recall.p99_ms/control.measurements.recall.p99_ms:null;
  paired.push({frontier,pair,ratio,pass:ratio!==null&&ratio<=.90});
}
const result={completed:true,commands,paired,preflightPassed:validation.passed,functionalLoadPass:commands.every(c=>c.trace?.functional),
  candidateAbsoluteTimingPass:commands.filter(c=>c.arm==="candidate").every(c=>c.status===0),pairedRescuePass:paired.every(p=>p.pass),
  wholeGoalValidation:false,populationTailGuarantee:false,designCaptureTemplates:288,independentFactCount:"not established",
  productionTouched:false,oldStoreMigrationTested:false,untouchedAgentUtilityTested:false};
result.pilotPass=result.preflightPassed&&result.functionalLoadPass&&result.candidateAbsoluteTimingPass&&result.pairedRescuePass;
fs.writeFileSync(path.join(output,"results.json"),JSON.stringify(result,null,2)+"\n");
console.log(JSON.stringify({completed:true,pilotPass:result.pilotPass,preflightPassed:result.preflightPassed,functionalLoadPass:result.functionalLoadPass,
  candidateAbsoluteTimingPass:result.candidateAbsoluteTimingPass,pairedRescuePass:result.pairedRescuePass}));
