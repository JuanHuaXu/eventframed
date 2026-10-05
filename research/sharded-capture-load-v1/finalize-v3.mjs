import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const root=path.dirname(fileURLToPath(import.meta.url)),read=n=>fs.readFileSync(path.join(root,n)),json=n=>JSON.parse(read(n));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const verified=["verify-v2.mjs","verify-v3.mjs"].map(n=>JSON.parse(execFileSync(process.execPath,[path.join(root,n)],{encoding:"utf8"})));
const missing=json("missing-shard-results.json"),missingManifest=json("missing-shard-manifest.json");
assert.equal(hash(read("missing-shard-v2.mjs")),missingManifest.runnerSHA256);
assert.equal(hash(read("SUPPLEMENTAL_PROTOCOL.md")),missingManifest.protocolSHA256);
assert.equal(hash(execFileSync("gofmt",[],{input:read("missing_shard_test.go.txt")})),missingManifest.testSHA256);
for(const c of missing.commands){const text=read(c.transcript);assert.equal(hash(text),c.sha256);assert.equal(c.status,0);assert.match(text.toString(),/--- PASS: TestResearchMissingShardFailClosedV2/);assert.doesNotMatch(text.toString(),/WARNING: DATA RACE|--- SKIP:|panic:/);}
assert.equal(missing.passed,true);
const training=json("training-after-results.json");
assert.equal(hash(execFileSync("gofmt",[],{input:read("training_regression_test.go.txt")})),training.testSHA256);
for(const c of training.commands){const text=read(c.transcript);assert.equal(hash(text),c.sha256);assert.equal(c.status,0);assert.match(text.toString(),/--- PASS: TestResearchParallelQuantizerTrainingV3/);assert.doesNotMatch(text.toString(),/WARNING: DATA RACE|--- SKIP:|panic:/);}
assert.equal(training.passed,true);assert.equal(training.originalBeforeUnitReproducedRace,false);
const before=json("training-before.json");assert.equal(hash(read("training-before.txt")),before.transcriptSHA256);assert.equal(before.status,0);assert.equal(before.dataRaceReported,false);
const originalRace=json("race-load-results.json");assert.equal(hash(read(originalRace.transcript)),originalRace.transcriptSHA256);assert.equal(originalRace.functionalRacePass,false);assert.equal(originalRace.dataRaceReported,true);
const negative=json("verifier-negative-results.json");assert.equal(negative.verified,true);assert.equal(negative.positivePass,true);assert.equal(negative.historicalArtifactsMutated,false);assert.equal(negative.negativeControls.length,4);assert.ok(negative.negativeControls.every(c=>c.rejected&&c.status!==0));
const localOnly={omittedFiles:[{file:"candidate-live-sample.txt",reason:"Local native process capture; Go timeout and race transcripts are authoritative."}]};
if(fs.existsSync(path.join(root,"candidate-live-sample.txt")))localOnly.omittedFiles[0].sha256=hash(read("candidate-live-sample.txt"));
else if(fs.existsSync(path.join(root,"LOCAL_ONLY.json")))localOnly.omittedFiles[0].sha256=json("LOCAL_ONLY.json").omittedFiles[0].sha256;
fs.writeFileSync(path.join(root,"LOCAL_ONLY.json"),JSON.stringify(localOnly,null,2)+"\n");
const result={verified:true,pilots:verified,originalOneLineCandidatePass:false,v2PromotionAllowed:false,v2CompleteLoadRacePass:false,missingShardNegativeControlPass:true,v3ComponentTestsPass:true,
  completeCandidate200FunctionalRacePass:verified[1].completeCandidate200FunctionalRacePass,allArmsFullLoadRaceValidated:false,verifierNegativeControlsPass:true,
  allSevenWholeGoals:"OPEN",wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
fs.writeFileSync(path.join(root,"CHECKPOINT_VERIFICATION.json"),JSON.stringify(result,null,2)+"\n");
const files=fs.readdirSync(root).filter(n=>n!=="SHA256SUMS"&&n!=="candidate-live-sample.txt"&&fs.statSync(path.join(root,n)).isFile()).sort();
fs.writeFileSync(path.join(root,"SHA256SUMS"),files.map(n=>hash(read(n))+"  "+n).join("\n")+"\n");
console.log(JSON.stringify(result));
