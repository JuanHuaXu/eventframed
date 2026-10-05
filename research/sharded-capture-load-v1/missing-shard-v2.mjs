import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const output=path.dirname(fileURLToPath(import.meta.url)),read=n=>fs.readFileSync(path.join(output,n));
const m=JSON.parse(read("repaired-manifest.json")),hash=b=>crypto.createHash("sha256").update(b).digest("hex");
const check=()=>{for(const[arm,files]of Object.entries(m.sources))for(const[file,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(m.paths[arm],file)))!==h)throw Error("source drift: "+arm+"/"+file);};
check();
const test="libravdb/research_missing_shard_test.go",commands=[];
fs.writeFileSync(path.join(output,"missing-shard-manifest.json"),JSON.stringify({runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),protocolSHA256:hash(read("SUPPLEMENTAL_PROTOCOL.md")),sourceManifestSHA256:hash(read("repaired-manifest.json")),test,testSHA256:hash(fs.readFileSync(path.join(m.paths.fork,test))),requiredSuite:"TestResearchMissingShardFailClosedV2"},null,2)+"\n");
for(const race of[false,true]){
  check();const args=["test","-mod=readonly",...(race?["-race"]:[]),"-count=1","-v","-timeout","45s","-run","^TestResearchMissingShardFailClosedV2$","./libravdb"];
  const p=spawnSync("go",args,{cwd:m.paths.fork,env:{...process.env,GOMAXPROCS:"10"},encoding:"utf8",maxBuffer:8*1024*1024});
  const text=(p.stdout??"")+(p.stderr??""),file=race?"missing-shard-race.txt":"missing-shard.txt";fs.writeFileSync(path.join(output,file),text);check();
  commands.push({command:["go",...args],status:p.status,signal:p.signal,error:p.error?.message??null,transcript:file,sha256:hash(text),passed:p.status===0&&text.includes("--- PASS: TestResearchMissingShardFailClosedV2")&&!/--- SKIP:|WARNING: DATA RACE|panic:/.test(text)});
  console.log(text);
}
fs.writeFileSync(path.join(output,"missing-shard-results.json"),JSON.stringify({commands,passed:commands.every(c=>c.passed),productionTouched:false,wholeGoalValidation:false},null,2)+"\n");
