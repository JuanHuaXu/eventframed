import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const root=path.dirname(fileURLToPath(import.meta.url));
const read=name=>fs.readFileSync(path.join(root,name));
const json=name=>JSON.parse(read(name));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const m=json("v3-manifest.json"),v=json("v3-validation-results.json"),vm=json("v3-validation-manifest.json"),r=json("v3-results.json"),rm=json("v3-run-manifest.json");
assert.equal(hash(read("V3_PILOT_PROTOCOL.md")),m.protocolSHA256);
assert.equal(hash(read("stage-v3.mjs")),m.preparationSHA256);
assert.equal(hash(read("dependency-v3.patch")),m.patchSHA256);
assert.equal(hash(read("manifest.json")),m.inheritedFailedManifestSHA256);
assert.equal(hash(read("repaired-manifest.json")),m.priorV2ManifestSHA256);
assert.equal(hash(read("race-load-results.json")),m.priorV2RaceSHA256);
assert.equal(hash(read("validation-v3.mjs")),vm.runnerSHA256);
assert.equal(hash(read("v3-manifest.json")),vm.sourceManifestSHA256);
assert.equal(hash(read("run-v3.mjs")),rm.runnerSHA256);
assert.equal(hash(read("v3-manifest.json")),rm.sourceManifestSHA256);
assert.equal(hash(read("v3-validation-results.json")),rm.preflightSHA256);
const templates={"libravdb/research_admission_regression_test.go":"admission_regression_test.go.txt","libravdb/research_logical_discovery_test.go":"discovery_regression_test.go.txt","libravdb/research_quantization_reopen_test.go":"quantization_regression_test.go.txt","internal/storage/singlefile/research_quantization_codec_test.go":"quantization_codec_test.go.txt","internal/index/hnsw/research_training_race_test.go":"training_regression_test.go.txt","libravdb/research_missing_shard_test.go":"missing_shard_test.go.txt"};
for(const[file,template]of Object.entries(templates))assert.equal(hash(execFileSync("gofmt",[],{input:read(template)})),m.sources.fork[file]);
assert.equal(hash(execFileSync("gofmt",[],{input:read("topology_test.go.txt")})),m.sources.candidate["internal/store/libravdbstore/sharded_topology_test.go"]);
assert.equal(v.completed,true);assert.equal(v.commands.length,vm.plan.length);assert.equal(v.commands.length,12);
for(let i=0;i<v.commands.length;i++){
  const c=v.commands[i],p=vm.plan[i],text=read(c.transcript).toString();assert.equal(hash(text),c.sha256);
  for(const key of["arm","name","required","args"])assert.deepEqual(c[key],p[key]);
  const passes=[...text.matchAll(/^--- PASS: (\S+)/gm)].map(x=>x[1]),skips=[...text.matchAll(/^--- SKIP: (\S+)/gm)].map(x=>x[1]),fails=[...text.matchAll(/^--- FAIL: (\S+)/gm)].map(x=>x[1]);
  assert.deepEqual(c.passNames,passes);assert.deepEqual(c.skipNames,skips);assert.deepEqual(c.failNames,fails);
  assert.equal(c.dataRaceReported,/WARNING: DATA RACE/.test(text));
  if(c.name==="load-race"){
    const match=text.match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/);assert.ok(match);
    assert.deepEqual(c.trace,JSON.parse(match[1]));assert.equal(c.trace.functional,true);
    assert.equal(c.trace.candidate_counts.length,64);assert.ok(c.trace.candidate_counts.every(n=>n===200));
    assert.equal(c.trace.capture_ns.length,256);assert.equal(c.trace.completed,64);assert.equal(c.trace.replay_completed,64);assert.equal(c.trace.ledger_rows,128);
    const errors=[...text.matchAll(/^\s+public_capture_load_test.go:\d+: (.+)$/gm)].map(x=>x[1]).filter(x=>!x.startsWith("PUBLIC_CAPTURE_TRACE="));
    assert.deepEqual(c.testErrors,errors);
    const timingOnly=c.status===1&&errors.length>0&&errors.every(x=>["frozen recall p99 <100ms screen failed","frozen publication p99 <250ms screen failed"].includes(x));
    assert.equal(c.instrumentedTimingOnly,timingOnly);assert.equal(c.functionalOnly,true);
    assert.equal(c.passed,c.trace.functional===true&&!c.dataRaceReported&&!/panic:|fatal error:|--- SKIP:/.test(text)&&(c.status===0||timingOnly));
  }else assert.equal(c.passed,c.status===0&&!c.dataRaceReported&&fails.length===0&&p.required.every(n=>passes.includes(n))&&(p.name==="vet"||passes.length>0));
}
assert.equal(v.passed,v.commands.every(c=>c.passed));assert.equal(v.passed,true);
const baselineSkips=json("validation-results.json").commands.find(c=>c.arm==="control"&&c.name==="store").skipNames;
for(const c of v.commands)if(c.name==="store")assert.deepEqual(c.skipNames,baselineSkips);else assert.equal(c.skipNames.length,0);
for(const name of["admission-before-actual.json","discovery-before.json","quantization-before.json","candidate-discovery-topology.json"]){
  const before=json(name);assert.notEqual(before.status,0);
  const transcript=name.replace(".json",".txt");assert.equal(hash(read(transcript)),before.transcriptSHA256??before.sha256);
}
assert.equal(json("validation-results.json").passed,false);assert.notEqual(json("topology-only-result.json").status,0);
const oldRace=json("race-load-results.json");assert.equal(oldRace.dataRaceReported,true);assert.equal(oldRace.functionalRacePass,false);assert.equal(hash(read(oldRace.transcript)),oldRace.transcriptSHA256);
assert.equal(r.completed,true);assert.equal(r.commands.length,8);
assert.deepEqual(r.commands.map(c=>[c.frontier,c.pair,c.arm]),[50,200].flatMap(k=>[[0,["control","candidate"]],[1,["candidate","control"]]].flatMap(([pair,order])=>order.map(arm=>[k,pair,arm]))));
const q=(values,f)=>[...values].sort((a,b)=>a-b)[Math.ceil(values.length*f)-1]/1e6;
for(const c of r.commands){
  const text=read(c.transcript).toString();assert.equal(hash(text),c.sha256);
  const match=text.match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/);assert.ok(match);const trace=JSON.parse(match[1]);assert.deepEqual(trace,c.trace);assert.equal(trace.functional,true);
  for(const key of["recall_ns","recall_start_ns","recall_end_ns","feedback_ns","live_age_ns","candidate_counts","packed_counts"])assert.equal(trace[key].length,64);
  for(const key of["capture_ns","capture_start_ns","capture_end_ns"])assert.equal(trace[key].length,256);
  assert.ok(trace.candidate_counts.every(n=>n===c.frontier));assert.ok(trace.packed_counts.every(n=>n>=0&&n<=10));
  assert.equal(trace.completed,64);assert.equal(trace.replay_completed,64);assert.equal(trace.ledger_rows,128);
  for(const key of["init_ns","init_bytes","capture_bytes"])assert.ok(trace[key]>0);
  let overlap=0;for(let j=0;j<256;j++){assert.equal(trace.capture_ns[j],trace.capture_end_ns[j]-trace.capture_start_ns[j]);if(trace.recall_start_ns.some((start,i)=>trace.capture_start_ns[j]<trace.recall_end_ns[i]&&trace.capture_end_ns[j]>start))overlap++;}
  assert.equal(trace.overlaps,overlap);assert.ok(overlap>0);
  for(let i=0;i<64;i++)assert.equal(trace.recall_ns[i],trace.recall_end_ns[i]-trace.recall_start_ns[i]);
  for(const name of["recall","feedback","live_age","capture"]){assert.ok(trace[name+"_ns"].every(n=>Number.isSafeInteger(n)&&n>=0));for(const[key,f]of[["p50",.5],["p95",.95],["p99",.99],["max",1]])assert.equal(c.measurements[name][key+"_ms"],q(trace[name+"_ns"],f));}
  const absolute=c.measurements.recall.p99_ms<100&&c.measurements.live_age.p99_ms<250;assert.equal(c.status,absolute?0:1);
}
assert.equal(r.paired.length,4);
for(const p of r.paired){const get=arm=>r.commands.find(c=>c.frontier===p.frontier&&c.pair===p.pair&&c.arm===arm).measurements.recall.p99_ms;assert.equal(p.ratio,get("candidate")/get("control"));assert.equal(p.pass,p.ratio<=.90);}
assert.equal(r.preflightPassed,v.passed);assert.equal(r.functionalLoadPass,r.commands.every(c=>c.trace.functional));assert.equal(r.candidateAbsoluteTimingPass,r.commands.filter(c=>c.arm==="candidate").every(c=>c.status===0));assert.equal(r.pairedRescuePass,r.paired.every(p=>p.pass));assert.equal(r.pilotPass,r.preflightPassed&&r.functionalLoadPass&&r.candidateAbsoluteTimingPass&&r.pairedRescuePass);
for(const key of["wholeGoalValidation","populationTailGuarantee","oldStoreMigrationTested","untouchedAgentUtilityTested","fullPublicLoadRaceValidated","productionTouched","privateDataUsed"])assert.equal(r[key],false);
console.log(JSON.stringify({verified:true,preflightPass:v.passed,loadFunctionalPass:r.functionalLoadPass,candidateAbsoluteTimingPass:r.candidateAbsoluteTimingPass,pairedRescuePass:r.pairedRescuePass,pilotPass:r.pilotPass,completeCandidate200FunctionalRacePass:v.commands.find(c=>c.name==="load-race").passed,allArmsFullLoadRaceValidated:false,wholeGoalValidation:false}));
