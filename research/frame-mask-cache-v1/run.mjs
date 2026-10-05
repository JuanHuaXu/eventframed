import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import crypto from "node:crypto";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(output, "../..");
const revision = "f3231fab2244c5c6bca8f7f822e5669ac58d13cd";
const hash = bytes => crypto.createHash("sha256").update(bytes).digest("hex");
const git = args => execFileSync("git", args, { cwd: root, maxBuffer: 16 * 1024 * 1024 });
const dir = fs.mkdtempSync(path.join(os.tmpdir(), "eventframe-mask-reuse-run-"));
const files = git(["ls-tree", "-r", "--name-only", revision, "--", "internal/frame", "internal/model"])
  .toString().trim().split("\n").filter(name => name.endsWith(".go"));
const write = (base, name, bytes) => {
  fs.mkdirSync(path.dirname(path.join(base, name)), { recursive: true });
  fs.writeFileSync(path.join(base, name), bytes);
};
const manifest = { revision, productionTouched: false, privateDataUsed: false,
  wholeGoalValidation: false, sources: {}, goVersion: execFileSync("go", ["version"]).toString().trim(),
  patchSHA256: hash(fs.readFileSync(path.join(output, "candidate.patch"))),
  protocolSHA256: hash(fs.readFileSync(path.join(output, "PROTOCOL.md"))),
  additionalTestsSHA256: hash(fs.readFileSync(path.join(output, "mask_reuse_test.go.txt"))),
};
for (const arm of ["control", "candidate"]) {
  const target = path.join(dir, arm);
  write(target, "go.mod", "module github.com/JuanHuaXu/eventframed\n\ngo 1.25.7\n");
  for (const name of files) write(target, name, git(["show", revision + ":" + name]));
  if (arm === "candidate") {
    execFileSync("git", ["apply", path.join(output, "candidate.patch")], { cwd: target });
    for (const name of files.filter(n => n.startsWith("internal/frame/") && !n.endsWith("_test.go"))) {
      write(target, "internal/framecontrol/" + path.basename(name), git(["show", revision + ":" + name]));
    }
    write(target, "internal/frame/mask_reuse_test.go", fs.readFileSync(path.join(output, "mask_reuse_test.go.txt")));
  }
  const names = [...files, ...(arm === "candidate" ? ["internal/frame/mask_reuse_test.go"] : [])];
  manifest.sources[arm] = Object.fromEntries(names.map(name => [name, hash(fs.readFileSync(path.join(target, name)))]));
}
// Source identities are frozen before any semantic outcome or benchmark run.
fs.writeFileSync(path.join(output, "manifest.json"), JSON.stringify(manifest, null, 2) + "\n");
const gitDir = git(["rev-parse", "--absolute-git-dir"]).toString().trim();
const run = (arm, name, args) => {
  let bytes;
  try {
    bytes = execFileSync("go", args, { cwd: path.join(dir, arm),
      env: { ...process.env, GIT_DIR: gitDir }, maxBuffer: 16 * 1024 * 1024 });
  } catch (error) {
    fs.writeFileSync(path.join(output, name), Buffer.concat([error.stdout ?? Buffer.alloc(0), error.stderr ?? Buffer.alloc(0)]));
    throw error;
  }
  fs.writeFileSync(path.join(output, name), bytes);
  return bytes.toString();
};
run("control", "control-tests.txt", ["test", "-mod=readonly", "-count=1", "./internal/frame"]);
const listing = run("candidate", "candidate-test-list.txt", ["test", "-mod=readonly", "-list", "^Test", "./internal/frame"]);
const tests = listing.split("\n").filter(name => /^Test\w+$/.test(name));
const excluded = "TestTurnFallbackAuditBaselineFidelity";
if (!tests.includes(excluded) || tests.filter(n => n.startsWith("TestMaskReuse")).length !== 3) throw Error("Unexpected semantic suite");
const selector = "^(" + tests.filter(n => n !== excluded).join("|") + ")$";
run("candidate", "candidate-tests.txt", ["test", "-mod=readonly", "-count=1", "-v", "-run", selector, "./internal/frame"]);
run("candidate", "candidate-race.txt", ["test", "-mod=readonly", "-race", "-count=1", "-run", selector, "./internal/frame"]);
run("candidate", "candidate-vet.txt", ["vet", "./internal/frame", "./internal/framecontrol", "./internal/model"]);
const raw = run("candidate", "benchmarks.txt", ["test", "-mod=readonly", "-run", "^$", "-bench", "^BenchmarkMaskReuse(Forward|Reverse)$",
  "-benchmem", "-benchtime=100ms", "-count=3", "-cpu=1", "./internal/frame"]);
const samples = {};
for (const line of raw.split("\n")) {
  const m = line.match(/^BenchmarkMaskReuse(?:Forward|Reverse)\/(turn|text|query)\/(early|late|no-match|quoted|collective)\/(256|2048|16384)B\/(control|candidate)(?:-\d+)?\s+\d+\s+([\d.]+) ns\/op\s+[\d.]+ MB\/s\s+(\d+) B\/op\s+(\d+) allocs\/op/);
  if (m) (samples[m.slice(1, 5).join("/")] ??= []).push({ ns: Number(m[5]), bytes: Number(m[6]), allocs: Number(m[7]) });
}
if (Object.keys(samples).length !== 90 || Object.values(samples).some(rows => rows.length !== 6)) throw Error("Incomplete benchmark cells");
const median = xs => { const sorted = [...xs].sort((a, b) => a - b); return (sorted[2] + sorted[3]) / 2; };
const cells = Object.keys(samples).filter(name => name.endsWith("/control")).map(name => {
  const before = samples[name], after = samples[name.replace(/control$/, "candidate")];
  const oldNS = median(before.map(r => r.ns)), newNS = median(after.map(r => r.ns));
  return { cell: name.replace(/\/control$/, ""), oldNS, newNS, changePercent: 100 * (newNS / oldNS - 1),
    oldBytes: median(before.map(r => r.bytes)), newBytes: median(after.map(r => r.bytes)),
    oldAllocations: median(before.map(r => r.allocs)), newAllocations: median(after.map(r => r.allocs)) };
});
const workflows = Object.fromEntries(["turn", "text", "query"].map(name => {
  const rows = cells.filter(row => row.cell.startsWith(name + "/"));
  const factor = Math.exp(rows.reduce((sum, row) => sum + Math.log(row.newNS / row.oldNS), 0) / rows.length);
  return [name, { geometricMeanImprovementPercent: 100 * (1 - factor), worstRegressionPercent: Math.max(...rows.map(row => row.changePercent)),
    pass: factor <= .95 && rows.every(row => row.changePercent <= 10) }];
}));
const result = { revision, semanticComparisons: 8192, parallelCallComparisons: 384,
  semanticTestsRun: tests.length - 1, historicalSourceIdentityTest: "Control only; candidate is a distinct source-hashed patch",
  scientificConfirmation: false, wholeGoalValidation: false, loadedLatencyTested: false,
  preflightPass: Object.values(workflows).every(row => row.pass), workflows, cells, samples };
fs.writeFileSync(path.join(output, "results.json"), JSON.stringify(result, null, 2) + "\n");
console.log(JSON.stringify({ completed: true, preflightPass: result.preflightPass, workflows }));
