import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import crypto from "node:crypto";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(output, "../..");
const revision = "f3231fab2244c5c6bca8f7f822e5669ac58d13cd";
const checkout = path.join(fs.mkdtempSync(path.join(os.tmpdir(), "eventframe-mask-integration-")), "candidate");
execFileSync("git", ["clone", "--shared", "--no-checkout", root, checkout], { stdio: "pipe" });
execFileSync("git", ["checkout", "--detach", revision], { cwd: checkout, stdio: "pipe" });
execFileSync("git", ["apply", path.join(output, "candidate.patch")], { cwd: checkout });
const selector = "^(TestCaptureTurnEnrichesInsideServiceAndObservePreservesAuthoredFields|TestCaptureIdentityLookupBudgetAndIncompleteRoster|TestCaptureIdentityDurableAndRetryIndexesCommittedFrame|TestCaptureIdentityRetryKeepsCommittedUnresolvedReferences|TestCaptureIdentityContextAndFutureCardBoundaries)$";
const args = ["test", "-mod=readonly", "-count=1", "-v", "-run", selector, "./internal/service"];
let raw;
try {
  raw = execFileSync("go", args, { cwd: checkout, maxBuffer: 16 * 1024 * 1024 });
} catch (error) {
  fs.writeFileSync(path.join(output, "service-integration.txt"), Buffer.concat([error.stdout ?? Buffer.alloc(0), error.stderr ?? Buffer.alloc(0)]));
  throw error;
}
fs.writeFileSync(path.join(output, "service-integration.txt"), raw);
const hash = bytes => crypto.createHash("sha256").update(bytes).digest("hex");
fs.writeFileSync(path.join(output, "service-integration-manifest.json"), JSON.stringify({
  revision, command: ["go", ...args], patchSHA256: hash(fs.readFileSync(path.join(output, "candidate.patch"))),
  transcriptSHA256: hash(raw), productionTouched: false, privateDataUsed: false,
  loadedLatencyTested: false, wholeGoalValidation: false,
  sources: Object.fromEntries(["turn.go", "text.go", "query.go", "identity.go"].map(name => [name,
    hash(fs.readFileSync(path.join(checkout, "internal/frame", name)))])),
}, null, 2) + "\n");
console.log(JSON.stringify({ completed: true, integrationTestFamilies: 5, loadedLatencyTested: false }));
