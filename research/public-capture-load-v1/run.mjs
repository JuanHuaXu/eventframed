import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import crypto from "node:crypto";
import { spawnSync, execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(output, "../..");
const prior = path.resolve(output, "../frame-mask-cache-v1");
const revision = "f3231fab2244c5c6bca8f7f822e5669ac58d13cd";
const hash = b => crypto.createHash("sha256").update(b).digest("hex");
const patch = fs.readFileSync(path.join(prior, "candidate.patch"));
const original = JSON.parse(fs.readFileSync(path.join(prior, "manifest.json")));
if (hash(patch) !== original.patchSHA256) throw Error("Preflight candidate changed");
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), "eventframe-public-capture-load-"));
const sources = {};
for (const arm of ["control", "candidate"]) {
  const checkout = path.join(temporary, arm);
  execFileSync("git", ["clone", "--shared", "--no-checkout", root, checkout], { stdio: "pipe" });
  execFileSync("git", ["checkout", "--detach", revision], { cwd: checkout, stdio: "pipe" });
  if (arm === "candidate") execFileSync("git", ["apply", path.join(prior, "candidate.patch")], { cwd: checkout });
  fs.copyFileSync(path.join(output, "public_capture_load_test.go.txt"), path.join(checkout, "internal/service/public_capture_load_test.go"));
  execFileSync("gofmt", ["-w", "internal/service/public_capture_load_test.go"], { cwd: checkout });
  sources[arm] = Object.fromEntries(["internal/frame/turn.go", "internal/frame/query.go", "internal/frame/text.go", "internal/frame/identity.go",
    "internal/service/public_capture_load_test.go", "internal/service/service.go", "testdata/text-public-facts/corpus.jsonl"].map(name =>
    [name, hash(fs.readFileSync(path.join(checkout, name)))]));
}
for (const name of Object.keys(sources.control).filter(n => !n.startsWith("internal/frame/"))) {
  if (sources.control[name] !== sources.candidate[name]) throw Error("Non-frame source changed");
}
fs.writeFileSync(path.join(output, "manifest.json"), JSON.stringify({ revision, patchSHA256: hash(patch),
  protocolSHA256: hash(fs.readFileSync(path.join(output, "PROTOCOL.md"))), runnerSHA256: hash(fs.readFileSync(fileURLToPath(import.meta.url))),
  fixtureSHA256: hash(fs.readFileSync(path.join(output, "public_capture_load_test.go.txt"))),
  goVersion: execFileSync("go", ["version"]).toString().trim(), architecture: process.arch,
  sources, runOrder: [["control", "candidate"], ["candidate", "control"]], frontiers: [50, 200],
  temporary, privateDataUsed: false, productionTouched: false, wholeGoalValidation: false,
}, null, 2) + "\n");
const quantile = (values, fraction) => [...values].sort((a,b) => a-b)[Math.ceil(fraction * values.length)-1] / 1e6;
const commands = [];
for (const frontier of [50, 200]) for (const [pair, order] of [[0, ["control", "candidate"]], [1, ["candidate", "control"]]]) {
  for (const arm of order) {
    console.log(JSON.stringify({ started: true, frontier, pair, arm }));
    const args = ["test", "-mod=readonly", "-count=1", "-v", "-timeout", "4m", "-run", "^TestResearchPublicCaptureLoadV1$", "./internal/service"];
    const processResult = spawnSync("go", args, { cwd: path.join(temporary, arm),
      env: { ...process.env, EVENTFRAME_PUBLIC_FRONTIER: String(frontier) }, encoding: "utf8", maxBuffer: 32 * 1024 * 1024 });
    const text = (processResult.stdout ?? "") + (processResult.stderr ?? "");
    const file = frontier + "-" + pair + "-" + arm + ".txt";
    fs.writeFileSync(path.join(output, file), text);
    const match = text.match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/);
    const trace = match ? JSON.parse(match[1]) : null;
    const measurements = trace?.functional ? Object.fromEntries(["recall", "feedback", "live_age", "capture"].map(name => [name,
      Object.fromEntries([["p50", .5], ["p95", .95], ["p99", .99], ["max", 1]].map(([key, fraction]) => [key + "_ms", quantile(trace[name + "_ns"], fraction)]))])) : null;
    commands.push({ frontier, pair, arm, command: ["go", ...args], status: processResult.status,
      signal: processResult.signal, error: processResult.error?.message ?? null,
      transcript: file, transcriptSHA256: hash(Buffer.from(text)), trace, measurements });
    fs.writeFileSync(path.join(output, "progress.json"), JSON.stringify({ completed: false, commands }, null, 2) + "\n");
    console.log(JSON.stringify({ completedCommand: true, frontier, pair, arm, status: processResult.status, functional: trace?.functional ?? false, measurements }));
  }
}
const paired = [];
for (const frontier of [50,200]) for (const pair of [0,1]) {
  const control = commands.find(c => c.frontier === frontier && c.pair === pair && c.arm === "control");
  const candidate = commands.find(c => c.frontier === frontier && c.pair === pair && c.arm === "candidate");
  const ratio = control.measurements && candidate.measurements ? candidate.measurements.recall.p99_ms / control.measurements.recall.p99_ms : null;
  paired.push({ frontier, pair, ratio, pass: ratio !== null && ratio <= 1.10 });
}
const result = { completed: true, commands, paired, functionalPass: commands.every(c => c.trace?.functional),
  absoluteTimingPass: commands.every(c => c.status === 0), pairedNonRegressionPass: paired.every(p => p.pass),
  wholeGoalValidation: false, populationTailGuarantee: false, independentFactCount: 288,
  fullCorpusLearningTested: false, realAgentQualityTested: false };
fs.writeFileSync(path.join(output, "results.json"), JSON.stringify(result, null, 2) + "\n");
console.log(JSON.stringify({ completed: true, functionalPass: result.functionalPass, absoluteTimingPass: result.absoluteTimingPass, pairedNonRegressionPass: result.pairedNonRegressionPass }));
