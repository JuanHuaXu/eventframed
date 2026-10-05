import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";

const out=path.dirname(fileURLToPath(import.meta.url)),read=n=>fs.readFileSync(path.join(out,n));
const m=JSON.parse(read("manifest.json"));
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
function check(){for(const[arm,files]of Object.entries(m.sources))for(const[f,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(m.paths[arm],f)))!==h)throw Error("source drift "+arm+"/"+f);}
check();if(fs.existsSync(path.join(out,"run-manifest.json")))throw Error("preserve prior run");
fs.writeFileSync(path.join(out,"run-manifest.json"),JSON.stringify({manifestSHA256:hash(read("manifest.json")),runnerSHA256:hash(read("run.mjs")),protocolSHA256:hash(read("PROTOCOL.md")),workers:10,orders:[["control","candidate"],["candidate","control"]],coldLawOnly:true,wholeGoalValidation:false},null,2)+"\n");
const commands=[];
function run(arm,pair,race,full){
  check();const args=["test","-mod=readonly","-count=1","-v","-timeout","5m"];
  if(race)args.push("-race");args.push("-run",full?"^TestResearchShardedQualityV1$":"^TestResearchQualityOracleV1$","./internal/service");
  const name=full?`quality-${pair}-${arm}.txt`:`oracle-${arm}-${race?"race":"ordinary"}.txt`;
  console.log(JSON.stringify({started:true,arm,pair,race,full,transcript:name}));
  const fd=fs.openSync(path.join(out,name),"wx");let result;
  try{result=spawnSync("go",args,{cwd:m.paths[arm],env:{...process.env,GOMAXPROCS:"10"},stdio:["ignore",fd,fd],timeout:330000});}finally{fs.closeSync(fd);}
  check();const text=read(name).toString(),row={arm,pair,race,full,command:["go",...args],status:result.status,signal:result.signal,error:result.error?.message??null,transcript:name,sha256:hash(text),bytes:Buffer.byteLength(text),dataRaceReported:text.includes("WARNING: DATA RACE")};
  commands.push(row);fs.writeFileSync(path.join(out,"progress.json"),JSON.stringify({completed:false,commands},null,2)+"\n");
  console.log(JSON.stringify({finished:true,...row}));return row.status===0&&!row.dataRaceReported;
}
let preflight=true;for(const arm of["control","candidate"])for(const race of[false,true])preflight=run(arm,-1,race,false)&&preflight;
if(preflight)for(const[pair,order]of[[0,["control","candidate"]],[1,["candidate","control"]]])for(const arm of order)run(arm,pair,false,true);
const result={completed:true,preflightPassed:preflight,commands,functionalPass:preflight&&commands.filter(c=>c.full).length===4&&commands.every(c=>c.status===0&&!c.dataRaceReported),wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
fs.writeFileSync(path.join(out,"run-results.json"),JSON.stringify(result,null,2)+"\n");
console.log(JSON.stringify({completed:true,preflightPassed:preflight,functionalPass:result.functionalPass}));
