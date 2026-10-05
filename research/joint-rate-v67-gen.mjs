// Mechanical copies of frozen V66 lifecycle and independent replay only.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const src='internal/researchdispersion/joint_v66_test.go',raw=fs.readFileSync(src),s=raw.toString();
let run=s.slice(s.indexOf('func runJointV66('),s.indexOf('var modesJointV66'));
run=run.replace('func runJointV66(p populationPairedV60, mode, schedule string)', 'func runJointRateV67(p populationPairedV60, mode, schedule string, hazard float64)').replace('joint.Config{Hazard: 1. / 16}', 'joint.Config{Hazard: hazard}');
let audit=s.slice(s.indexOf('func independentJointV66('),s.indexOf('func TestJointV66Audit('));
audit=audit.replace('func independentJointV66(p populationPairedV60, a armPairedV60)', 'func independentJointRateV67(p populationPairedV60, a armPairedV60, hazard float64)').replace('ref.New(p.World.Base, 1./16)', 'ref.New(p.World.Base, hazard)');
assert(run.includes('Hazard: hazard'));assert(audit.includes('ref.New(p.World.Base, hazard)'));
const header='// Frozen rate-transition ablation; no new model or production changes.\npackage researchdispersion\nimport("bufio";"encoding/json";"fmt";"io";"math/rand";"os";"reflect";"sort";"testing";"time";joint "github.com/JuanHuaXu/eventframed/internal/researchjointsequence";ref "github.com/JuanHuaXu/eventframed/internal/researchjointsequenceref")\n';
const out='internal/researchdispersion/joint_rate_v67_test.go',b=header+run+audit;
fs.writeFileSync(out,b,{flag:'wx',mode:0o600});const h=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.writeFileSync('research/joint-rate-v67-generation.json',JSON.stringify({source:src,sourceSHA256:h(raw),output:out,initialOutputSHA256:h(b),stage:'lifecycle and replay copied; prospective experiment appended separately'},null,2)+'\n',{flag:'wx',mode:0o600});
