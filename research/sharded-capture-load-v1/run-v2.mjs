import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const read=name=>fs.readFileSync(path.join(output,name));
const m=JSON.parse(read("repaired-manifest.json"));
const validation=JSON.parse(read("repaired-validation-results.json"));
if(!validation.completed||!validation.passed)throw Error("repaired functional preflight must pass before load");
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const check=()=>{for(const[arm,files]of Object.entries(m.sources))for(const[file,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(m.paths[arm],file)))!==h)throw Error("source drift: "+arm+"/"+file);};
check();
fs.writeFileSync(path.join(output,"repaired-run-manifest.json"),JSON.stringify({runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),sourceManifestSHA256:hash(read("repaired-manifest.json")),preflightSHA256:hash(read("repaired-validation-results.json")),preflightPassed:validation.passed,frontiers:[50,200],runOrder:[["control","candidate"],["candidate","control"]],workers:10,postFailureDesignPilot:true,productionTouched:false,privateDataUsed:false},null,2)+"\n");
const q=(values,fraction)=>[...values].sort((a,b)=>a-b)[Math.ceil(values.length*fraction)-1]/1e6;
const commands=[];
for(const frontier of[50,200])for(const[pair,order]of[[0,["control","candidate"]],[1,["candidate","control"]]])for(const arm of order){
  check();console.log(JSON.stringify({started:true,frontier,pair,arm}));
  const args=["test","-mod=readonly","-count=1","-v","-timeout","4m","-run","^TestResearchPublicCaptureLoadV1$","./internal/service"];
  const r=spawnSync("go",args,{cwd:m.paths[arm],env:{...process.env,EVENTFRAME_PUBLIC_FRONTIER:String(frontier),GOMAXPROCS:"10"},encoding:"utf8",maxBuffer:32*1024*1024});
  const text=(r.stdout??"")+(r.stderr??""),file=`repaired-${frontier}-${pair}-${arm}.txt`;fs.writeFileSync(path.join(output,file),text);check();
  const match=text.match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/),trace=match?JSON.parse(match[1]):null;
  const measurements=trace?.functional?Object.fromEntries(["recall","feedback","live_age","capture"].map(name=>[name,Object.fromEntries([["p50",.5],["p95",.95],["p99",.99],["max",1]].map(([key,fraction])=>[key+"_ms",q(trace[name+"_ns"],fraction)]))])):null;
  const row={frontier,pair,arm,command:["go",...args],status:r.status,signal:r.signal,error:r.error?.message??null,transcript:file,sha256:hash(text),trace,measurements};
  commands.push(row);fs.writeFileSync(path.join(output,"repaired-progress.json"),JSON.stringify({completed:false,commands},null,2)+"\n");
  console.log(JSON.stringify({finished:true,frontier,pair,arm,status:row.status,functional:trace?.functional??false,measurements}));
}
const paired=[];
for(const frontier of[50,200])for(const pair of[0,1]){
  const get=arm=>commands.find(c=>c.frontier===frontier&&c.pair===pair&&c.arm===arm);
  const control=get("control"),candidate=get("candidate");
  const ratio=control.measurements&&candidate.measurements?candidate.measurements.recall.p99_ms/control.measurements.recall.p99_ms:null;
  paired.push({frontier,pair,ratio,pass:ratio!==null&&ratio<=.90});
}
const result={completed:true,commands,paired,preflightPassed:validation.passed,functionalLoadPass:commands.every(c=>c.trace?.functional),candidateAbsoluteTimingPass:commands.filter(c=>c.arm==="candidate").every(c=>c.status===0),pairedRescuePass:paired.every(p=>p.pass),wholeGoalValidation:false,populationTailGuarantee:false,designCaptureTemplates:288,independentFactCount:"not established",postFailureDesignPilot:true,productionTouched:false,privateDataUsed:false,oldStoreMigrationTested:false,untouchedAgentUtilityTested:false,fullPublicLoadRaceValidated:false};
result.pilotPass=result.preflightPassed&&result.functionalLoadPass&&result.candidateAbsoluteTimingPass&&result.pairedRescuePass;
fs.writeFileSync(path.join(output,"repaired-results.json"),JSON.stringify(result,null,2)+"\n");
console.log(JSON.stringify({completed:true,pilotPass:result.pilotPass,preflightPassed:result.preflightPassed,functionalLoadPass:result.functionalLoadPass,candidateAbsoluteTimingPass:result.candidateAbsoluteTimingPass,pairedRescuePass:result.pairedRescuePass}));
