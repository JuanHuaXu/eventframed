import assert from "node:assert/strict";
import fs from "node:fs/promises";
import http from "node:http";
import path from "node:path";
import { spawn, execFile } from "node:child_process";
import { promisify } from "node:util";
import { createHash, randomUUID } from "node:crypto";
import { startAgencyWorker } from "../src/agency-worker.js";
import type { AdapterConfig } from "../src/types.js";

const execute = promisify(execFile);
const required = (name: string): string => { const value=process.env[name];if(!value)throw new Error(`explicit ${name} required`);return value; };
const cliPath = required("EVENTFRAME_LAB_CLI");
const configPath = required("EVENTFRAME_LAB_CONFIG");
const stateDir = required("EVENTFRAME_LAB_STATE");
const gatewayURL = required("EVENTFRAME_LAB_GATEWAY");
const root = await fs.mkdtemp(path.join(process.env.EVENTFRAME_LAB_ARTIFACT_ROOT ?? "/tmp", "ef-tau-"));
await fs.chmod(root, 0o700);
const socketPath = path.join(root, "daemon.sock");
const configHash = () => fs.readFile(configPath).then(data => createHash("sha256").update(data).digest("hex"));
const beforeHash = await configHash();
const protocol_version = "eventframe.v1alpha1";
const tenant = `tau-lab-${randomUUID()}`;
const session = `agent:main:eventframe-tau-lab-${randomUUID()}`;
const daemonArgs = ["-listen", `unix://${socketPath}`, "-database", path.join(root, "events.db"), "-rank-delta-sqlite", path.join(root, "ranks.sqlite"), "-dimension", "8", "-quantization", "none", "-agency-enabled", "-agency-private-key", path.join(root, "private"), "-agency-public-key", path.join(root, "public"), "-agency-issuer-token", path.join(root, "issuer"), "-agency-authority-token", path.join(root, "authority"), "-background-fuzz=false"];
let daemon: ReturnType<typeof spawn> | undefined;
let stopAuthority: (() => void) | undefined;
let createdJob: string | undefined;
const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));
const call = async (method: string, params: unknown) => {
  assert(["cron.runs", "cron.remove", "cron.status", "models.list"].includes(method));
  const result = await execute(process.execPath, [cliPath, "gateway", "call", method, "--expect-url", gatewayURL, "--json", "--params", JSON.stringify(params), "--timeout", "10000"], { env: { ...process.env, OPENCLAW_STATE_DIR: stateDir, OPENCLAW_CONFIG_PATH: configPath }, timeout: 15000, maxBuffer: 2 << 20 });
  return JSON.parse(result.stdout);
};
const request = (route: string, body?: unknown): Promise<{status:number;value:any;elapsed:number}> => new Promise((resolve,reject) => {
  const start=performance.now();const bytes=body===undefined?undefined:JSON.stringify(body);
  const req=http.request({socketPath,path:route,method:bytes?"POST":"GET",headers:bytes?{"content-type":"application/json","content-length":Buffer.byteLength(bytes)}:{}},res=>{
    let data="";res.on("data",chunk=>{data+=chunk;if(data.length>2<<20)req.destroy(new Error("response cap"))});res.on("end",()=>{try{resolve({status:res.statusCode!,value:JSON.parse(data),elapsed:performance.now()-start})}catch(error){reject(error)}});
  });req.on("error",reject);req.setTimeout(5000,()=>req.destroy(new Error("daemon timeout")));if(bytes)req.end(bytes);else req.end();
});
const startDaemon = async () => {
  daemon=spawn(path.resolve("../bin/eventframed"),daemonArgs,{stdio:["ignore","ignore","pipe"]});
  daemon.stderr?.on("data",data=>void fs.appendFile(path.join(root,"daemon.log"),data));
  for(let i=0;i<100;i++){try{if((await request("/v1/health")).status===200)return}catch{}if(daemon.exitCode!==null)throw new Error("daemon exited");await sleep(50)}throw new Error("daemon startup timeout");
};
const stopDaemon=async()=>{if(!daemon || daemon.exitCode!==null)return;const child=daemon;await new Promise<void>(resolve=>{child.once("exit",()=>resolve());child.kill("SIGTERM")});daemon=undefined};
const frame=async(id:string,kind:string,occurred:number,eventSession=session)=>{
  const text=`Public synthetic ${kind} observation ${id}`;const observed=occurred+1;const available=occurred+2;
  const field=(value:string)=>({value,source:"synthetic",confidence:1});
  const result=await request("/v1/events:observe",{protocol_version,idempotency_key:id,event:{id,tenant_id:tenant,session_id:eventSession,sequence:occurred,kind,content:text,occurred_at:new Date(occurred).toISOString(),observed_at:new Date(observed).toISOString(),available_at:new Date(available).toISOString(),who:field("public test participant"),what:field(text),where:field("isolated lab"),when:field(new Date(occurred).toISOString()),why:field("public mechanism test"),how:field("controlled fixture"),priority:.5,provenance:{producer:"public-tau-lab",source_event_ids:[id]},embedding_model:""}});
  assert.equal(result.status,200,JSON.stringify(result.value));return {id,occurred,available};
};

try {
  console.log(JSON.stringify({stage:"start",root,session,gatewayURL}));
  assert.equal((await call("cron.status",{})).enabled,true);
  const catalogue=await call("models.list",{});
  assert(catalogue.models.some((m:any)=>m.id==="gpt-6-luna" && m.provider==="openai" && m.available),"approved model unavailable");
  await startDaemon();
  const issuer_token=(await fs.readFile(path.join(root,"issuer"),"utf8")).trim();
  const authority_token=(await fs.readFile(path.join(root,"authority"),"utf8")).trim();
  const origin=await frame("control-origin","task_start",Date.now()-100);
  const policy={target_kind:"complete",benefit:1,false_cost:.1,turn_cost:.01,min_probability:.4,margin:.01,lead_ms:2000,window_ms:15000,slack_ms:10000,min_delay_ms:1000,cooldown_ms:60000};
  const makeRequest=async(plan:string)=>{
    const base=(await request("/v1/health")).value.snapshot;
    return {protocol_version,issuer_token,tenant_id:tenant,session_id:session,plan_id:plan,origin_event_id:origin.id,evidence_ids:[origin.id],base,priority:.5,reason:"Public mechanism check: a test completion is predicted soon. Acknowledge EVENTFRAME_TAU_LAB_OK without tools.",forecast:{id:`law-${plan}`,base,origin:new Date(origin.occurred).toISOString(),issued_at:new Date().toISOString(),valid_until:new Date(origin.occurred+60000).toISOString(),horizon_ms:60000,points:[{mark:"complete",delay_ms:30000,probability:.9}],no_event:.1,tv_radius:0},policy:structuredClone(policy)};
  };
  const controls:any[]=[];
  const noEvent=await makeRequest("none");noEvent.forecast.points=[];noEvent.forecast.no_event=1;
  const n=await request("/v1/agency/tau:plan",noEvent);assert.equal(n.status,200);assert.equal(n.value.decision.schedule,false);controls.push({name:"no-event",pass:true});
  const costly=await makeRequest("costly");costly.policy={...policy,turn_cost:1};const c=await request("/v1/agency/tau:plan",costly);assert.equal(c.status,200);assert.equal(c.value.decision.schedule,false);controls.push({name:"negative-utility",pass:true});
  const empty=await makeRequest("empty-candidates");(empty.policy as any).candidate_delays_ms=[];const e=await request("/v1/agency/tau:plan",empty);assert.equal(e.status,200);assert.equal(e.value.decision.schedule,false);controls.push({name:"empty-candidates",pass:true});
  const unsure=await makeRequest("unsure");unsure.forecast.tv_radius=.95;const u=await request("/v1/agency/tau:plan",unsure);assert.equal(u.status,200);assert.equal(u.value.decision.schedule,false);controls.push({name:"uncertain-law",pass:true});
  const missing=await makeRequest("missing");delete (missing.forecast as any).tv_radius;assert.equal((await request("/v1/agency/tau:plan",missing)).status,400);controls.push({name:"missing-error-budget",pass:true});
  const unauth=await makeRequest("unauth");unauth.issuer_token="incorrect";assert.equal((await request("/v1/agency/tau:plan",unauth)).status,400);controls.push({name:"unauthenticated",pass:true});
  const stale=await makeRequest("stale");const st=await request("/v1/agency/tau:plan",stale);assert.equal(st.status,200);assert(st.value.proposal);
  await frame("state-change","new_fact",Date.now()-10);
  const claims=await request("/v1/agency/proposals:claim",{protocol_version,authority_token,tenant_id:tenant,consumer_id:"stale-control",limit:10});assert.equal(claims.status,200);assert.equal(claims.value.records.length,0);controls.push({name:"stale-plan",pass:true});
  const perf:number[]=[];
  noEvent.base=(await request("/v1/health")).value.snapshot;noEvent.forecast.base=noEvent.base;noEvent.forecast.issued_at=new Date().toISOString();
  for(let i=0;i<100;i++){const r=await request("/v1/agency/tau:plan",noEvent);assert.equal(r.status,200);perf.push(r.elapsed)}
  perf.sort((a,b)=>a-b);
  const past=Date.now()-300000;
  const episodes=[];
  for(let i=0;i<3;i++){
    const start=past+i*80000;const eventSession=`public-episode-${i}`;
    await frame(`anchor-${i}`,"task_start",start,eventSession);
    await frame(`outcome-${i}`,i===2?"tau_no_event":"complete",start+(i===2?60000:20000),eventSession);
    episodes.push({anchor_id:`anchor-${i}`,outcome_id:`outcome-${i}`,no_event:i===2});
  }
  const current=await frame("live-origin","task_start",Date.now()-10);
  const base=(await request("/v1/health")).value.snapshot;
  const fit={protocol_version,issuer_token,tenant_id:tenant,session_id:session,plan_id:"live-fit",origin_event_id:current.id,base,priority:.5,reason:"Public mechanism check: a synthetic task completion is predicted. Acknowledge EVENTFRAME_TAU_LAB_OK without tools.",forecast:{id:"fitted-law",base,origin:new Date(current.occurred).toISOString(),issued_at:new Date().toISOString(),valid_until:new Date(current.occurred+60000).toISOString(),horizon_ms:60000,points:[],no_event:0,tv_radius:0},policy,episodes,bin_delays_ms:[20000,60000]};
  const fitted=await request("/v1/agency/tau:plan",fit);assert.equal(fitted.status,200,JSON.stringify(fitted.value));assert(fitted.value.proposal);assert.equal(fitted.value.forecast.points[0].probability,.5);assert.equal(fitted.value.forecast.no_event,1/3);
  const repeated=await request("/v1/agency/tau:plan",fit);assert.equal(repeated.value.proposal.duplicate,true);controls.push({name:"durable-idempotency",pass:true});
  const proposalID=fitted.value.proposal.record.proposal.id;
  await stopDaemon();await startDaemon();
  const config:AdapterConfig={socketPath,tenantId:tenant,recallK:50,packK:10,tokenBudget:2000,capture:false,agencyEnabled:true,agencyKillSwitch:false,agencyPublicKeyPath:path.join(root,"public"),agencyAuthorityTokenPath:path.join(root,"authority"),agencyCapabilities:["eventframe.agency.schedule"],agencyConsentActions:["schedule"],agencyConsumerId:"tau-lab-authority",agencyPollIntervalMs:1000,agencyMaxClaims:10,agencyMaxChainDepth:3,agencyCriticalThreshold:.9,agencyAllowedSessionPrefixes:[session]};
  stopAuthority=await startAgencyWorker(config,{cliPath,configPath,stateDir,url:gatewayURL,silent:true,model:"openai/gpt-6-luna"});
  console.log(JSON.stringify({stage:"waiting-for-predicted-turn",nominal:fitted.value.decision.deliver_at,latest:fitted.value.decision.latest_dispatch,controls:controls.length}));
  let record:any;
  for(let i=0;i<45;i++){
    const read=await request("/v1/agency/proposals:get",{protocol_version,authority_token,tenant_id:tenant,proposal_id:proposalID});record=read.value.record;
    if(record?.status==="approved"){createdJob=record.execution_ref;break}
    if(["rejected","expired","uncertain"].includes(record?.status))throw new Error(`live proposal ${record.status}: ${record.resolution_reason}`);
    await sleep(1000);
  }
  assert(createdJob,"authority did not approve a scheduled job");
  stopAuthority();stopAuthority=undefined;
  let run:any;
  for(let i=0;i<35;i++){const history=await call("cron.runs",{id:createdJob,limit:5});run=history.entries?.find((e:any)=>e.action==="finished");if(run)break;await sleep(2000)}
  assert(run,"gateway did not report a terminal scheduled turn");
  const summary={kind:"tau-driven-automatic-turn-lab-v1",controls,forecast:fitted.value.forecast,decision:fitted.value.decision,daemonRestarted:true,proposalStatus:record.status,jobID:createdJob,session,run:{status:run.status,error:run.error,runAtMs:run.runAtMs,durationMs:run.durationMs,model:run.model,provider:run.provider,markerPresent:typeof run.summary==="string"&&run.summary.includes("EVENTFRAME_TAU_LAB_OK")},planHTTPEligibleMS:fitted.elapsed,planHTTPNoTurn:{n:100,p50MS:perf[49],p95MS:perf[94],p99MS:perf[98]},configUnchanged:beforeHash===await configHash(),claimScope:"mechanism and one live lab turn; no utility or calibration claim"};
  await fs.writeFile(path.join(root,"SUMMARY.json"),JSON.stringify(summary,null,2)+"\n",{mode:0o600});
  console.log(JSON.stringify(summary));
  assert.equal(summary.configUnchanged,true);
  assert.equal(run.status,"ok",run.error);
  assert.equal(run.provider,"openai");assert.equal(run.model,"gpt-6-luna");
  assert(run.runAtMs>=Date.parse(fitted.value.decision.deliver_at) && run.runAtMs<Date.parse(fitted.value.decision.latest_dispatch),"observed model start exceeded signed dispatch window");
} finally {
  stopAuthority?.();
  if(createdJob){try{await call("cron.remove",{id:createdJob})}catch{}}
  await stopDaemon();
  console.log(JSON.stringify({stage:"stopped",root}));
}
