import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root=path.dirname(fileURLToPath(import.meta.url));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const read=name=>fs.readFileSync(path.join(root,name));
const verify=[];
for (const name of ["../frame-mask-cache-load-v1/verify.mjs","verify.mjs","verify-worker-cap.mjs"]) verify.push(JSON.parse(execFileSync(process.execPath,[path.resolve(root,name)],{encoding:"utf8"})));
const diagnosticManifest=JSON.parse(read("diagnostic-manifest.json"));
const diagnostic=JSON.parse(read("diagnostic-results.json"));
assert.equal(hash(read("DIAGNOSTIC_PROTOCOL.md")),diagnosticManifest.protocolSHA256);
assert.equal(hash(read("profile.mjs")),diagnosticManifest.runnerSHA256);
assert.equal(hash(read("diagnostic-test.txt")),diagnostic.transcriptSHA256);
for (const report of diagnostic.reports) assert.equal(hash(read(report.file)),report.sha256);
const publicationPath=path.join(root,"PUBLICATION.json");
const publication=fs.existsSync(publicationPath)?JSON.parse(fs.readFileSync(publicationPath)):null;
const omittedProfiles=[];
for (const [name,expected] of Object.entries(diagnostic.profiles)) {
  if (fs.existsSync(path.join(root,name))) assert.equal(hash(read(name)),expected);
  else {
    assert.equal(publication?.omittedRawProfiles?.[name],expected,"unexplained missing profile");
    omittedProfiles.push(name);
  }
}
if (publication) assert.deepEqual(omittedProfiles.sort(),Object.keys(publication.omittedRawProfiles).sort());
assert.equal(JSON.parse(read("diagnostic-test.txt").toString().match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/)[1]).functional,true);
const validationManifest=JSON.parse(read("validation-v2-manifest.json"));
const validation=JSON.parse(read("validation-v2-results.json"));
assert.equal(hash(read("validation-v2.mjs")),validationManifest.runnerSHA256);
const text=read("guard-race-v2.txt");
assert.equal(hash(text),validation.transcriptSHA256);
assert.equal(validation.status,0);
assert.equal(validation.skipped,false);
assert.equal(validation.dataRaceReported,false);
assert.equal(validation.passed,true);
for (const name of validation.requiredSuites) assert.ok(text.toString().includes("--- PASS: "+name+" "));
const first=JSON.parse(read("validation-results.json"));
assert.equal(first.commands[0].status,0);
assert.equal(hash(read("vet.txt")),first.commands[0].sha256);
assert.match(read("guard-race.txt").toString(),/--- SKIP: TestResearchDurableIdentityAudit/);
const result={verified:true,comparisons:verify,adjacentGuardRacePassed:true,fullLoadRaceValidated:false,
  primaryMetadataErratum:"independentFactCount=288 is design capture-template count, NOT independent facts",
  originalValidationErratum:"opt-in identity audit skipped; only v2 establishes all named guard regressions executed",
  diagnosticOnly:true,wholeGoalValidation:false,allSevenGoals:"OPEN",productionTouched:false,privateDataUsed:false,
  rawProfilesVerified:omittedProfiles.length===0,omittedRawProfiles:omittedProfiles};
fs.writeFileSync(path.join(root,"CHECKPOINT_VERIFICATION.json"),JSON.stringify(result,null,2)+"\n");
for (const directory of [root,path.resolve(root,"../frame-mask-cache-load-v1")]) {
  const files=fs.readdirSync(directory).filter(name=>name!=="SHA256SUMS"&&fs.statSync(path.join(directory,name)).isFile()).sort();
  fs.writeFileSync(path.join(directory,"SHA256SUMS"),files.map(name=>hash(fs.readFileSync(path.join(directory,name)))+"  "+name).join("\n")+"\n");
}
console.log(JSON.stringify({verified:true,adjacentGuardRacePassed:true,fullLoadRaceValidated:false,wholeGoalValidation:false}));
