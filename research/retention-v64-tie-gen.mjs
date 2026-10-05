// Independent arithmetic may induce a near-tie. Preserve and diagnose before
// changing the checker; the collected learner and protocol remain unchanged.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const src='internal/researchdispersion/retention_v64_test.go',out='internal/researchdispersion/retention_v64_tie_audit_test.go',root='research/retention-v64b-diagnostic';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),raw=fs.readFileSync(src),freeze=JSON.parse(fs.readFileSync(root+'/freeze.json'));
assert.equal(hash(raw),freeze.files[src]);assert.equal(hash(fs.readFileSync(root+'/source/'+src)),freeze.files[src]);assert.equal(JSON.parse(fs.readFileSync(root+'/failure.json')).checks.at(-1).exitCode,1);
let s=raw.toString(),body=s.slice(s.indexOf('func independentRetentionV64('),s.indexOf('func TestRetentionV64FutureAndCorruptions('));
body=body.replaceAll('independentRetentionV64','independentRetentionV64TieAudit').replaceAll('TestRetentionV64Audit','TestRetentionV64TieAudit');
body=body.replace('sources := sourcesRetentionV64(t)',`old:=os.Getenv("EVENTFRAME_RETENTION_V64_FREEZE")\n if e:=os.Setenv("EVENTFRAME_RETENTION_V64_FREEZE",os.Getenv("EVENTFRAME_RETENTION_V64_DATA_FREEZE"));e!=nil{t.Fatal(e)}\n sources := sourcesRetentionV64(t)\n if e:=os.Setenv("EVENTFRAME_RETENTION_V64_FREEZE",old);e!=nil{t.Fatal(e)}`);
const target='order = order[:25]';assert(body.includes(target));
body=body.replace(target,target+'\n if !reflect.DeepEqual(order,d.Members) {return fmt.Errorf("selection diagnostic round=%d expected=%v actual=%v referenceScores=%v recordedOptions=%v",nextRound,order,d.Members,scores,d.Options)}');
const header=`// Independent replay ordering diagnostic; original failed artifact retained.\npackage researchdispersion\nimport (\n "bufio"\n "encoding/json"\n "fmt"\n "io"\n "math"\n "math/rand"\n "os"\n "reflect"\n "sort"\n "testing"\n bank "github.com/JuanHuaXu/eventframed/internal/researchwindowjournal"\n ref "github.com/JuanHuaXu/eventframed/internal/researchwindowjournalref"\n selector "github.com/JuanHuaXu/eventframed/internal/researchretention"\n law "github.com/JuanHuaXu/eventframed/internal/researchretentionlaw"\n)\n`;
fs.writeFileSync(out,header+body,{flag:'wx',mode:0o600});fs.writeFileSync('research/retention-v64-tie-generation.json',JSON.stringify({source:src,sourceSHA256:hash(raw),output:out,outputSHA256:hash(header+body),failedLogSHA256:hash(fs.readFileSync(root+'/audit.log')),stage:'diagnostic only; no acceptance change'},null,2)+'\n',{flag:'wx',mode:0o600});
