import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import crypto from "node:crypto";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const root=path.resolve(output,"../..");
const old=path.resolve(output,"../public-capture-load-v1");
const original=JSON.parse(fs.readFileSync(path.join(old,"manifest.json")));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const patch=path.join(output,"candidate.patch");
const fixture=path.join(old,"public_capture_load_test.go.txt");
if (hash(fs.readFileSync(fixture))!==original.fixtureSHA256) throw Error("Old fixture changed");
const temporary=fs.mkdtempSync(path.join(os.tmpdir(),"eventframe-sharded-load-"));
const sources={};
for (const arm of ["control","candidate"]) {
  const checkout=path.join(temporary,arm);
  execFileSync("git",["clone","--shared","--no-checkout",root,checkout],{stdio:"pipe"});
  execFileSync("git",["checkout","--detach",original.revision],{cwd:checkout,stdio:"pipe"});
  if (arm==="candidate") {
    execFileSync("git",["apply",patch],{cwd:checkout});
    fs.copyFileSync(path.join(output,"topology_test.go.txt"),path.join(checkout,"internal/store/libravdbstore/sharded_topology_test.go"));
    execFileSync("gofmt",["-w","internal/store/libravdbstore/sharded_topology_test.go"],{cwd:checkout});
  }
  fs.copyFileSync(fixture,path.join(checkout,"internal/service/public_capture_load_test.go"));
  execFileSync("gofmt",["-w","internal/service/public_capture_load_test.go"],{cwd:checkout});
  sources[arm]=Object.fromEntries(["internal/store/libravdbstore/store.go","internal/service/service.go","internal/frame/turn.go",
    "internal/service/public_capture_load_test.go","testdata/text-public-facts/corpus.jsonl"].map(name=>[name,hash(fs.readFileSync(path.join(checkout,name)))]));
  if (sources[arm]["internal/service/public_capture_load_test.go"]!==original.sources.control["internal/service/public_capture_load_test.go"]) throw Error("Formatted old fixture changed");
}
fs.writeFileSync(path.join(output,"manifest.json"),JSON.stringify({revision:original.revision,temporary,sources,
  protocolSHA256:hash(fs.readFileSync(path.join(output,"PROTOCOL.md"))),patchSHA256:hash(fs.readFileSync(patch)),
  preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),topologyTemplateSHA256:hash(fs.readFileSync(path.join(output,"topology_test.go.txt"))),
  topologyCompiledSHA256:hash(fs.readFileSync(path.join(temporary,"candidate/internal/store/libravdbstore/sharded_topology_test.go"))),
  oldFixtureSHA256:original.fixtureSHA256,privateDataUsed:false,productionTouched:false,wholeGoalValidation:false,
  goVersion:execFileSync("go",["version"],{encoding:"utf8"}).trim(),runtimeWorkers:10,
  expectedLibraryShardCount:4},null,2)+"\n");
console.log(JSON.stringify({prepared:true,temporary}));
