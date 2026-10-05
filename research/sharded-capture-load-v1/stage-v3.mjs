import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url)),read=n=>fs.readFileSync(path.join(output,n));
const old=JSON.parse(read("repaired-manifest.json")),hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const temporary=path.dirname(old.paths.fork);
const paths={control:path.join(temporary,"control-repaired-v3"),candidate:path.join(temporary,"candidate-repaired-v3"),fork:path.join(temporary,"libravdb-fork-v3")};
for(const arm of["control","candidate"]){if(fs.existsSync(paths[arm]))throw Error("V3 arm exists; preserve it");fs.cpSync(old.paths[arm],paths[arm],{recursive:true});execFileSync("go",["mod","edit","-replace","github.com/xDarkicex/libravdb=../libravdb-fork-v3"],{cwd:paths[arm]});}
const libraryFiles=[...old.libraryFiles,"internal/index/hnsw/hnsw.go"],libraryTests=[...old.libraryTests,"libravdb/research_missing_shard_test.go","internal/index/hnsw/research_training_race_test.go"];
const sources={};
for(const arm of["control","candidate"]){sources[arm]=Object.fromEntries(Object.keys(old.sources[arm]).map(file=>[file,hash(fs.readFileSync(path.join(paths[arm],file)))]));for(const[file,h]of Object.entries(old.sources[arm]))if(file!=="go.mod"&&sources[arm][file]!==h)throw Error("unchanged V3 source drift "+arm+"/"+file);}
sources.fork=Object.fromEntries(["go.mod","go.sum",...libraryFiles,...libraryTests].map(file=>[file,hash(fs.readFileSync(path.join(paths.fork,file)))]));
for(const[file,h]of Object.entries(old.sources.fork))if(sources.fork[file]!==h)throw Error("V3 changed prior repaired source "+file);
const patch=Buffer.concat([read("dependency-v2.patch"),read("training-v3.patch")]);fs.writeFileSync(path.join(output,"dependency-v3.patch"),patch);
const m={...old,paths,sources,libraryFiles,libraryTests,originalLibrarySources:{...old.originalLibrarySources,"internal/index/hnsw/hnsw.go":JSON.parse(read("upstream-training-check.json")).sha256},protocolSHA256:hash(read("V3_PILOT_PROTOCOL.md")),preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),patchSHA256:hash(patch),priorV2ManifestSHA256:hash(read("repaired-manifest.json")),priorV2RaceSHA256:hash(read("race-load-results.json")),wholeGoalValidation:false};
fs.writeFileSync(path.join(output,"v3-manifest.json"),JSON.stringify(m,null,2)+"\n");
console.log(JSON.stringify({staged:true,paths,patchSHA256:m.patchSHA256}));
