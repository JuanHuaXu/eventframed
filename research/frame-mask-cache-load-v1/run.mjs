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
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), "eventframe-mask-load-"));
const sources = {};
for (const arm of ["control", "candidate"]) {
  const checkout = path.join(temporary, arm);
  execFileSync("git", ["clone", "--shared", "--no-checkout", root, checkout], { stdio: "pipe" });
  execFileSync("git", ["checkout", "--detach", revision], { cwd: checkout, stdio: "pipe" });
  if (arm === "candidate") execFileSync("git", ["apply", path.join(prior, "candidate.patch")], { cwd: checkout });
  sources[arm] = Object.fromEntries(["internal/frame/turn.go", "internal/frame/query.go", "internal/frame/text.go", "internal/frame/identity.go",
    "internal/service/research_durable_mixed_load_test.go", "internal/service/research_durable_live_freshness_test.go"].map(name =>
    [name, hash(fs.readFileSync(path.join(checkout, name)))]));
}
for (const name of Object.keys(sources.control).filter(n => n.startsWith("internal/service/"))) {
  if (sources.control[name] !== sources.candidate[name]) throw Error("Old loaded fixture changed");
}
fs.writeFileSync(path.join(output, "manifest.json"), JSON.stringify({ revision, patchSHA256: hash(patch),
  protocolSHA256: hash(fs.readFileSync(path.join(output, "PROTOCOL.md"))), runnerSHA256: hash(fs.readFileSync(fileURLToPath(import.meta.url))),
  goVersion: execFileSync("go", ["version"]).toString().trim(), architecture: process.arch,
  sources, runOrder: [["control", "candidate"], ["candidate", "control"]], privateDataUsed: false,
  productionTouched: false, wholeGoalValidation: false,
}, null, 2) + "\n");
const durationMS = value => {
  const match = value.match(/^([\d.]+)(ns|µs|us|ms|s)$/);
  if (!match) throw Error("Unrecognized duration " + value);
  return Number(match[1]) * { ns: .000001, "µs": .001, us: .001, ms: 1, s: 1000 }[match[2]];
};
const commands = [];
for (const [pair, order] of [[0, ["control", "candidate"]], [1, ["candidate", "control"]]]) {
  for (const arm of order) {
    const args = ["test", "-mod=readonly", "-count=1", "-v", "-run", "^TestResearchGuardedDurableLiveFreshnessV2$", "./internal/service"];
    const processResult = spawnSync("go", args, { cwd: path.join(temporary, arm),
      env: { ...process.env, EVENTFRAME_RESEARCH_PERF_GATE: "1" }, encoding: "utf8", maxBuffer: 16 * 1024 * 1024 });
    const text = (processResult.stdout ?? "") + (processResult.stderr ?? "");
    const file = pair + "-" + arm + ".txt";
    fs.writeFileSync(path.join(output, file), text);
    const measurements = [];
    for (const line of text.split("\n").filter(line => /arm=(quiet|future-writer) calls=/.test(line))) {
      const values = Object.fromEntries([...line.matchAll(/(\w+)=([^\s]+)/g)].map(m => [m[1], m[2]]));
      const row = { arm: values.arm, calls: Number(values.calls), labels: Number(values.live_labels),
        writes: Number(values.writes), overlaps: Number(values.overlaps) };
      for (const key of ["recall_p50", "recall_p95", "recall_p99", "feedback_p99", "live_age_p50", "live_age_p95", "live_age_p99", "live_age_max"]) row[key + "_ms"] = durationMS(values[key]);
      measurements.push(row);
    }
    commands.push({ pair, arm, command: ["go", ...args], status: processResult.status,
      signal: processResult.signal, error: processResult.error?.message ?? null,
      transcript: file, transcriptSHA256: hash(Buffer.from(text)), measurements });
    fs.writeFileSync(path.join(output, "progress.json"), JSON.stringify({ completed: false, commands }, null, 2) + "\n");
  }
}
const complete = commands.every(command => command.measurements.length === 2);
const paired = [];
if (complete) for (const pair of [0, 1]) for (const arm of ["quiet", "future-writer"]) {
  const baseline = commands.find(c => c.pair === pair && c.arm === "control").measurements.find(m => m.arm === arm);
  const candidate = commands.find(c => c.pair === pair && c.arm === "candidate").measurements.find(m => m.arm === arm);
  paired.push({ pair, arm, controlRecallP99MS: baseline.recall_p99_ms, candidateRecallP99MS: candidate.recall_p99_ms,
    ratio: candidate.recall_p99_ms / baseline.recall_p99_ms,
    pass: candidate.recall_p99_ms <= 1.10 * baseline.recall_p99_ms });
}
const result = { completed: true, completeMeasurements: complete, commands, paired,
  functionalAndAbsoluteGatePass: commands.every(c => c.status === 0),
  pairedNonRegressionPass: complete && paired.every(p => p.pass),
  wholeGoalValidation: false, largeFrontierTested: false, captureWriterTested: false,
  candidateAttributionEstablished: false, populationTailGuarantee: false,
};
fs.writeFileSync(path.join(output, "results.json"), JSON.stringify(result, null, 2) + "\n");
console.log(JSON.stringify({ completed: true, completeMeasurements: complete,
  functionalAndAbsoluteGatePass: result.functionalAndAbsoluteGatePass, pairedNonRegressionPass: result.pairedNonRegressionPass }));
