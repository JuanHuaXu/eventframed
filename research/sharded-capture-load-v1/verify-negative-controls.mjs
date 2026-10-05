import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import assert from "node:assert/strict";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const root=path.dirname(fileURLToPath(import.meta.url)),temporary=fs.mkdtempSync(path.join(os.tmpdir(),"eventframe-shard-verifier-"));
fs.cpSync(root,temporary,{recursive:true,filter:file=>!file.endsWith("candidate-live-sample.txt")});
const run=()=>spawnSync(process.execPath,[path.join(temporary,"verify-v3.mjs")],{encoding:"utf8",maxBuffer:8*1024*1024});
const positive=run();assert.equal(positive.status,0,positive.stderr);
const cases=[
  {name:"fabricated paired gain",file:"v3-results.json",mutate:r=>{r.paired[0].ratio/=2;}},
  {name:"fabricated durable rows",file:"v3-results.json",mutate:r=>{r.commands[0].trace.ledger_rows++;}},
  {name:"fabricated latency summary",file:"v3-results.json",mutate:r=>{r.commands[1].measurements.recall.p99_ms--; }},
  {name:"fabricated executed coverage",file:"v3-validation-results.json",mutate:r=>{r.commands[0].passNames.push("InventedPassedSuite");}}
];
const results=[];
for(const c of cases){const file=path.join(temporary,c.file),original=fs.readFileSync(file),value=JSON.parse(original);c.mutate(value);fs.writeFileSync(file,JSON.stringify(value));const p=run();fs.writeFileSync(file,original);assert.notEqual(p.status,0,c.name);assert.match(p.stderr,/AssertionError/,"failure must be an assertion, not a setup error");results.push({name:c.name,rejected:true,status:p.status});}
const result={positivePass:true,negativeControls:results,verified:true,historicalArtifactsMutated:false};
fs.writeFileSync(path.join(root,"verifier-negative-results.json"),JSON.stringify(result,null,2)+"\n");
console.log(JSON.stringify(result));
