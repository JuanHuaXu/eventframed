import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";
import {evaluate} from "./evaluate.mjs";

export async function verify(root){
  const read=n=>fs.readFileSync(path.join(root,n)),json=n=>JSON.parse(read(n)),hash=b=>crypto.createHash("sha256").update(b).digest("hex");
  const manifest=json("manifest.json"),runManifest=json("run-manifest.json"),run=json("run-results.json");
  assert.equal(hash(read("PROTOCOL.md")),manifest.protocolSHA256);assert.equal(hash(read("prepare.mjs")),manifest.preparationSHA256);
  assert.equal(hash(read("quality_test.go.txt")),manifest.templateSHA256);assert.equal(hash(execFileSync("gofmt",[],{input:read("quality_test.go.txt")})),manifest.preparedTestSHA256);
  for(const arm of["control","candidate"])assert.equal(manifest.sources[arm]["internal/service/sharded_quality_test.go"],manifest.preparedTestSHA256);
  assert.equal(hash(read("manifest.json")),runManifest.manifestSHA256);assert.equal(hash(read("PROTOCOL.md")),runManifest.protocolSHA256);assert.equal(hash(read("run.mjs")),runManifest.runnerSHA256);
  assert.equal(manifest.workers,10);assert.equal(runManifest.workers,10);assert.equal(run.commands.length,8);
  const preflight=run.commands.filter(x=>!x.full);assert.equal(preflight.length,4);
  for(const arm of["control","candidate"])for(const race of[false,true]){
    const c=preflight.find(x=>x.arm===arm&&x.race===race);assert.ok(c);assert.equal(c.status,0);assert.equal(c.signal,null);assert.equal(c.error,null);assert.equal(c.dataRaceReported,false);
    const b=read(c.transcript),t=b.toString();assert.equal(hash(b),c.sha256);assert.equal(b.length,c.bytes);assert.match(t,/--- PASS: TestResearchQualityOracleV1/);assert.doesNotMatch(t,/WARNING: DATA RACE|--- SKIP:|panic:/);
  }
  assert.deepEqual(run.commands.filter(x=>x.full).map(x=>[x.pair,x.arm]),[[0,"control"],[0,"candidate"],[1,"candidate"],[1,"control"]]);
  if(fs.existsSync(path.join(root,"ARCHIVES.json"))){const archives=json("ARCHIVES.json");assert.equal(archives.losslessVerified,true);assert.equal(archives.archives.length,4);for(const x of archives.archives){assert.equal(hash(read(x.archive)),x.archiveSHA256);assert.equal(read(x.archive).length,x.archiveBytes);const c=run.commands.find(c=>c.transcript===x.original);assert.equal(x.originalSHA256,c.sha256);assert.equal(x.originalBytes,c.bytes);}}
  const actual=await evaluate(root);assert.deepEqual(json("evaluation.json"),actual);
  return {verified:true,commands:8,ordinaryServiceRecalls:3072,exactReferenceRecalls:3072,...Object.fromEntries(Object.entries(actual).filter(([k])=>!["cells","comparisons","paired"].includes(k)))};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url)))console.log(JSON.stringify(await verify(path.dirname(fileURLToPath(import.meta.url)))));
