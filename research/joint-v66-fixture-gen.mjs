// Mechanical lifecycle extraction from immutable V64; new law/model, unchanged
// consumed worlds, delay queues, controls and scientific gates.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const src='internal/researchdispersion/retention_v64_test.go',out='internal/researchdispersion/joint_v66_test.go',raw=fs.readFileSync(src);let s=raw.toString();
let body=s.slice(s.indexOf('func runRetentionV64('),s.indexOf('var modesRetentionV64='));
if(body.length===0||body.includes('func independentRetentionV64'))body=s.slice(s.indexOf('func runRetentionV64('),s.indexOf('var modesRetentionV64 ='));
assert(body.startsWith('func runRetentionV64('));assert(!body.includes('independentRetentionV64'));
body=body.replaceAll('runRetentionV64','runJointV66').replaceAll('chooseRetentionV64','chooseJointV66').replaceAll('bank.BankTicket','joint.Ticket');
body=body.replace('bank.NewBank(w.Base, 1, 2800, 7, [3]int{600, 1200, 2400})','joint.New(w.Base, 1, 2800, joint.Config{Hazard:1./16})');
body=body.replace('Expired: k < min(tick+1, 2400)-600','Expired: false');
const header=`// Isolated coherent sequence study; no production or truth certification.\npackage researchdispersion\nimport("bufio";"encoding/json";"fmt";"io";"math";"math/rand";"os";"path/filepath";"reflect";"sort";"testing";"time";joint "github.com/JuanHuaXu/eventframed/internal/researchjointsequence";ref "github.com/JuanHuaXu/eventframed/internal/researchjointsequenceref";paired "github.com/JuanHuaXu/eventframed/internal/researchpaired")\n`;
fs.writeFileSync(out,header+body,{flag:'wx',mode:0o600});const hash=b=>crypto.createHash('sha256').update(b).digest('hex');fs.writeFileSync('research/joint-v66-fixture-generation.json',JSON.stringify({source:src,sourceSHA256:hash(raw),output:out,initialOutputSHA256:hash(header+body),stage:'mechanical extraction; acquisition/audit added separately'},null,2)+'\n',{flag:'wx',mode:0o600});
