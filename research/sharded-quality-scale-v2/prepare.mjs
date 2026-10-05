import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import crypto from "node:crypto";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const out=path.dirname(fileURLToPath(import.meta.url));
const old=JSON.parse(fs.readFileSync(path.join(out,"../sharded-capture-load-v1/v3-manifest.json")));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
for(const [arm,files]of Object.entries(old.sources))for(const [file,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(old.paths[arm],file)))!==h)throw Error("inherited source drift: "+arm+"/"+file);
if(fs.existsSync(path.join(out,"manifest.json")))throw Error("preserve prior preparation; do not overwrite manifest");
const root=fs.mkdtempSync(path.join(os.tmpdir(),"eventframe-sharded-quality-"));
const template=fs.readFileSync(path.join(out,"quality_test.go.txt"));
const formatted=execFileSync("gofmt",[],{input:template});
const paths={fork:old.paths.fork};const sources={fork:old.sources.fork};
sources.fork={...sources.fork,"libravdb/collection_sharding.go":hash(fs.readFileSync(path.join(old.paths.fork,"libravdb/collection_sharding.go")))};
for(const arm of ["control","candidate"]){
  paths[arm]=path.join(root,arm);fs.cpSync(old.paths[arm],paths[arm],{recursive:true});
  execFileSync("go",["mod","edit","-replace=github.com/xDarkicex/libravdb="+old.paths.fork],{cwd:paths[arm]});
  fs.writeFileSync(path.join(paths[arm],"internal/service/sharded_quality_test.go"),formatted);
  sources[arm]=Object.fromEntries([...Object.keys(old.sources[arm]),"internal/service/sharded_quality_test.go"].map(f=>[f,hash(fs.readFileSync(path.join(paths[arm],f)))]));
}
const manifest={paths,sources,inheritedManifestSHA256:hash(fs.readFileSync(path.join(out,"../sharded-capture-load-v1/v3-manifest.json"))),protocolSHA256:hash(fs.readFileSync(path.join(out,"PROTOCOL.md"))),templateSHA256:hash(template),preparedTestSHA256:hash(formatted),preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),revision:old.revision,workers:10,productionTouched:false,privateDataUsed:false,wholeGoalValidation:false};
fs.writeFileSync(path.join(out,"manifest.json"),JSON.stringify(manifest,null,2)+"\n");
console.log(JSON.stringify({prepared:true,root,paths}));
