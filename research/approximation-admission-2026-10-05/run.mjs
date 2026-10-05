import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import crypto from "node:crypto";
import { execFileSync } from "node:child_process";

// Mechanical source snapshots keep matched controls out of the live checkout.
const root = process.cwd();
const output = path.join(root, "research/approximation-admission-2026-10-05");
const revision = "7a7b9357c06ce855795b947ffcde2fd61a82d3ca";
const packages = ["researchregimelogsummary", "researchregimelog", "researchregimeledger", "researchregimeprotected"];
const git = (...args) => execFileSync("git", args, { cwd: root, maxBuffer: 32 * 1024 * 1024 });
const oldNames = git("ls-tree", "-r", "--name-only", revision, ...packages.map(p => `internal/${p}`)).toString().trim().split("\n").filter(p => p.endsWith(".go"));
const temp = fs.mkdtempSync(path.join(os.tmpdir(), "eventframe-bound-bench-"));
const manifest = { controlRevision: revision, syntheticMechanismsOnly: true, untouchedConfirmation: false, wholeGoalValidation: false, sources: {} };
const put = (dir, name, bytes) => { fs.mkdirSync(path.dirname(path.join(dir,name)), {recursive:true}); fs.writeFileSync(path.join(dir,name), bytes); };
for (const arm of ["before", "after"]) {
  const dir = path.join(temp, arm);
  put(dir, "go.mod", "module github.com/JuanHuaXu/eventframed\n\ngo 1.25.7\n");
  const hashes = {};
  const names = [...oldNames];
  if (arm === "after") {
    for (const pkg of ["researchbounds", "researchregimelogsummary", "researchstats"]) {
      for (const name of fs.readdirSync(path.join(root,"internal",pkg))) {
        const rel = `internal/${pkg}/${name}`;
        if (name.endsWith(".go") && !names.includes(rel)) names.push(rel);
      }
    }
  }
  for (const name of names) {
    const bytes = arm === "before" ? git("show", `${revision}:${name}`) : fs.readFileSync(path.join(root,name));
    put(dir, name, bytes);
    hashes[name] = crypto.createHash("sha256").update(bytes).digest("hex");
  }
  manifest.sources[arm] = hashes;
}
const run = (arm, name, args, env = {}) => {
  const result = execFileSync("go", args, {cwd:path.join(temp,arm), env:{...process.env,...env},maxBuffer:32*1024*1024});
  fs.writeFileSync(path.join(output,name),result);
  return result.toString();
};
// Complete reference audit reuses synthetic development fixtures, not fresh data.
run("after", "audit-tests.txt", ["test","-mod=readonly","-count=1","./internal/researchregimelogsummary"], {EVENTFRAME_REGIME_V83_REPORT:path.join(output,"audit.json")});
const benchArgs = ["test","-mod=readonly","-run","^$","-bench","BenchmarkLedgerOperations/.*/.*/(PendingFirst|Issue)$","-benchmem","-benchtime=100ms","-count=3","-cpu=1","./internal/researchregimelogsummary"];
// Alternate ordering across two rounds; runs remain descriptive, not randomized.
for (const round of [0,1]) for (const arm of round === 0 ? ["before","after"] : ["after","before"]) run(arm,`${arm}-benchmark-${round}.txt`,benchArgs);
run("after","bounds-benchmark.txt",["test","-mod=readonly","-run","^$","-bench","BenchmarkStepTV","-benchmem","-count=3","-cpu=1","./internal/researchbounds"]);
run("after","admission-benchmark.txt",["test","-mod=readonly","-run","^$","-bench","BenchmarkBoundedIssue","-benchmem","-benchtime=200ms","-count=3","-cpu=1","./internal/researchregimelogsummary"]);
// Final extraction runs use the real checkout only to verify the immutable
// test-only baseline and candidate hashes; no service or data imports run.
const frameOutput = execFileSync("go", ["test","-mod=readonly","-run","^$",
  "-bench","^BenchmarkTurnFallbackAuditMatched","-benchmem","-benchtime=100ms",
  "-count=3","-cpu=1","./internal/frame"], {cwd:root,maxBuffer:32*1024*1024});
fs.writeFileSync(path.join(output,"frame-benchmark.txt"),frameOutput);
const frameResults = {};
for (const line of frameOutput.toString().split("\n")) {
  const m = line.match(/^BenchmarkTurnFallbackAuditMatched(?:Reverse)?\/(early|late|no-match|quoted|collective)\/(\d+)B-per-role\/(HEAD|current)(?:-\d+)?\s+\d+\s+([\d.]+) ns\/op\s+[\d.]+ MB\/s\s+(\d+) B\/op\s+(\d+) allocs\/op/);
  if (m) (frameResults[m[1]+"/"+m[2]+"/"+m[3]] ??= [])
    .push({ns:Number(m[4]),bytes:Number(m[5]),allocs:Number(m[6])});
}
if (Object.keys(frameResults).length !== 30 || Object.values(frameResults).some(r => r.length !== 6)) {
  throw new Error("Incomplete final extraction benchmark");
}
fs.writeFileSync(path.join(output,"frame-benchmark-results.json"),JSON.stringify({
  scope:"Full FromTurn only; synthetic public fixtures, not ingestion/retrieval/loaded latency",
  controlRevision:"1a7edb62b6be4031fd01ebeab8071b17303a7815", results:frameResults,
},null,2)+"\n");
manifest.frameSources = Object.fromEntries(fs.readdirSync(path.join(root,"internal/frame"))
  .filter(name => name.endsWith(".go")).map(name => [name,
    crypto.createHash("sha256").update(fs.readFileSync(path.join(root,"internal/frame",name))).digest("hex")]));
manifest.goVersion = execFileSync("go",["version"]).toString().trim();
manifest.architecture = process.arch;
fs.writeFileSync(path.join(output,"manifest.json"), JSON.stringify(manifest,null,2)+"\n");
const results = {};
for (const arm of ["before","after"]) {
  const rows = {};
  for (const round of [0,1]) for (const line of fs.readFileSync(path.join(output,`${arm}-benchmark-${round}.txt`),"utf8").split("\n")) {
    const match = line.match(/^(Benchmark\S+)\s+\d+\s+([\d.]+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op/);
    if (match) (rows[match[1]] ??= []).push({ns:Number(match[2]),bytes:Number(match[3]),allocs:Number(match[4])});
  }
  results[arm] = rows;
  if (Object.keys(rows).length !== 12 || Object.values(rows).some(r => r.length !== 6)) {
    throw new Error(`Incomplete matched benchmark in ${arm}: expected 12 operation profiles and six samples each`);
  }
}
fs.writeFileSync(path.join(output,"benchmark-results.json"),JSON.stringify({scope:"prepared serial operations, not serving latency",runOrder:"before-after, after-before",results},null,2)+"\n");
console.log(JSON.stringify({completed:true,controlRevision:revision,artifacts:Object.keys(results),untouchedConfirmation:false}));
