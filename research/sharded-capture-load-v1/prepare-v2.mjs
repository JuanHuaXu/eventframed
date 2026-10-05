import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url));
const old=JSON.parse(fs.readFileSync(path.join(output,"manifest.json")));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const paths={control:path.join(old.temporary,"control-repaired-v2"),candidate:path.join(old.temporary,"candidate-admission-v2"),fork:path.join(old.temporary,"libravdb-fork")};
const pinned=execFileSync("go",["env","GOMODCACHE"],{cwd:paths.control,encoding:"utf8"}).trim()+"/github.com/x!darkicex/libravdb@v1.6.13";
const libraryFiles=["libravdb/tx.go","libravdb/database.go","libravdb/collection.go","libravdb/index_persistence.go","internal/storage/interfaces.go","internal/storage/singlefile/codec.go"];
const libraryTests=["libravdb/research_admission_regression_test.go","libravdb/research_logical_discovery_test.go","libravdb/research_quantization_reopen_test.go","internal/storage/singlefile/research_quantization_codec_test.go"];
const source={};
for(const arm of ["control","candidate"]){
  const files=["go.mod","go.sum",...Object.keys(old.sources[arm])];
  if(arm==="candidate")files.push("internal/store/libravdbstore/sharded_topology_test.go");
  source[arm]=Object.fromEntries(files.map(file=>[file,hash(fs.readFileSync(path.join(paths[arm],file)))]));
  for(const [file,h] of Object.entries(old.sources[arm]))if(source[arm][file]!==h)throw Error("old fixture/source drift "+arm+"/"+file);
}
source.fork=Object.fromEntries(["go.mod","go.sum",...libraryFiles,...libraryTests].map(file=>[file,hash(fs.readFileSync(path.join(paths.fork,file)))]));
const original=Object.fromEntries(libraryFiles.map(file=>[file,hash(fs.readFileSync(path.join(pinned,file)))]));
const patchRoot=fs.mkdtempSync(path.join(old.temporary,"fork-patch-v2-"));
for(const file of libraryFiles){fs.mkdirSync(path.dirname(path.join(patchRoot,file)),{recursive:true});fs.copyFileSync(path.join(pinned,file),path.join(patchRoot,file));}
execFileSync("git",["init","-q"],{cwd:patchRoot});
execFileSync("git",["add","--",...libraryFiles],{cwd:patchRoot});
execFileSync("git",["-c","user.name=Research checkpoint","-c","user.email=research@localhost","commit","-qm","Pinned dependency source checkpoint"],{cwd:patchRoot});
for(const file of libraryFiles){fs.chmodSync(path.join(patchRoot,file),0o644);fs.copyFileSync(path.join(paths.fork,file),path.join(patchRoot,file));}
const patch=execFileSync("git",["diff","--",...libraryFiles],{cwd:patchRoot});
fs.writeFileSync(path.join(output,"dependency-v2.patch"),patch);
const manifest={revision:old.revision,paths,sources:source,originalLibrarySources:original,
  libraryFiles,libraryTests,protocolSHA256:hash(fs.readFileSync(path.join(output,"REPAIRED_PILOT_PROTOCOL.md"))),
  preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),patchSHA256:hash(patch),
  inheritedFailedManifestSHA256:hash(fs.readFileSync(path.join(output,"manifest.json"))),
  workers:10,privateDataUsed:false,productionTouched:false,wholeGoalValidation:false};
fs.writeFileSync(path.join(output,"repaired-manifest.json"),JSON.stringify(manifest,null,2)+"\n");
console.log(JSON.stringify({prepared:true,patchSHA256:manifest.patchSHA256}));
