import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const root=path.dirname(fileURLToPath(import.meta.url)),hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const run=JSON.parse(fs.readFileSync(path.join(root,"run-results.json"))),rawNames=new Set(run.commands.filter(c=>c.full).map(c=>c.transcript));
assert.ok(fs.existsSync(path.join(root,"ARCHIVES.json")),"archive first so test copies need not duplicate oversize plaintext");
const protectedFiles=["evaluation.json","run-results.json","manifest.json","quality_test.go.txt"],before=Object.fromEntries(protectedFiles.map(f=>[f,hash(fs.readFileSync(path.join(root,f)))]));
const sandbox=fs.mkdtempSync(path.join(os.tmpdir(),"eventframe-quality-verifier-"));
fs.cpSync(root,sandbox,{recursive:true,filter:f=>!rawNames.has(path.basename(f))});
const originals=Object.fromEntries(protectedFiles.map(f=>[f,fs.readFileSync(path.join(sandbox,f))]));
const verify=()=>spawnSync(process.execPath,[path.join(sandbox,"verify.mjs")],{encoding:"utf8",maxBuffer:1024*1024});
const positive=verify();assert.equal(positive.status,0,positive.stderr);
assert.equal(JSON.parse(positive.stdout).verified,true,"an empty successful process is not verified execution");
const controls=[];
for(const name of["fabricated_paired_quality","hidden_future_law_failure","missing_executed_arm","substituted_harness"]){
  for(const[f,b]of Object.entries(originals))fs.writeFileSync(path.join(sandbox,f),b);
  if(name==="substituted_harness")fs.appendFileSync(path.join(sandbox,"quality_test.go.txt"),"\n// unverifiable replacement\n");
  else if(name==="missing_executed_arm"){const r=JSON.parse(originals["run-results.json"]);r.commands.pop();fs.writeFileSync(path.join(sandbox,"run-results.json"),JSON.stringify(r));}
  else{const r=JSON.parse(originals["evaluation.json"]);if(name==="fabricated_paired_quality")r.paired[0].tieRecallDifference+=.1;else r.counterfactualEqualityPass=!r.counterfactualEqualityPass;fs.writeFileSync(path.join(sandbox,"evaluation.json"),JSON.stringify(r));}
  const r=verify(),rejected=r.status!==0&&(r.stderr??"").includes("AssertionError");assert.ok(rejected,`${name}: ${r.status} ${r.stderr}`);controls.push({name,status:r.status,rejected,setupError:false});
}
for(const[f,h]of Object.entries(before))assert.equal(hash(fs.readFileSync(path.join(root,f))),h);
const result={verified:true,positivePass:true,negativeControls:controls,historicalInputsMutated:false,independentArithmeticSeparatelyChecked:true};
fs.writeFileSync(path.join(root,"negative-results.json"),JSON.stringify(result,null,2)+"\n");console.log(JSON.stringify(result));
