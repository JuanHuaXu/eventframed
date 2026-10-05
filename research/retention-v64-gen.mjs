// Extract the existing consumed-world lifecycle; keep the control implementation
// unchanged. Semantic mixture/acquisition patches follow in the isolated file.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const src='internal/researchdispersion/tree_v60_generated_test.go',out='internal/researchdispersion/retention_v64_test.go';
assert(!fs.existsSync(out));const raw=fs.readFileSync(src),text=raw.toString();
let body=text.slice(text.indexOf('func runPairedV60('),text.indexOf('func sourcesPairedV60('));assert(body.startsWith('func runPairedV60'));
body=body.replaceAll('runPairedV60','runRetentionV64').replaceAll('choosePairedV60','chooseRetentionV64').replaceAll('paired.Ticket','bank.BankTicket').replaceAll('m.RequestAudit','m.RequestSecond');
body=body.replace('paired.New(w.Base, 1, 2800, paired.Config{Depth: 7, Window: 600})','bank.NewBank(w.Base, 1, 2800, 7, [3]int{600,1200,2400})');
body=body.replace('\t\t\tweights := m.NoiseWeights()\n','').replace('\t\t\ta.NoiseIssued[tick] = weights\n','').replace('\t\t\ta.NoiseSnapshots = append(a.NoiseSnapshots, m.NoiseWeights())\n','');
body=body.replace(', NoiseIssued: make([][3]float64, 2400)','');
body=body.replace('c.AccountedNS = c.SetupNS +','c.AccountedNS = c.ScheduleNS + c.SetupNS +');
const header=`// Isolated V64 quality ablation: actual working window mixture, consumed worlds.\npackage researchdispersion\n\nimport (\n "bufio"\n "encoding/json"\n "fmt"\n "io"\n "math"\n "math/rand"\n "os"\n "path/filepath"\n "reflect"\n "sort"\n "testing"\n "time"\n bank "github.com/JuanHuaXu/eventframed/internal/researchwindowjournal"\n ref "github.com/JuanHuaXu/eventframed/internal/researchwindowjournalref"\n selector "github.com/JuanHuaXu/eventframed/internal/researchretention"\n law "github.com/JuanHuaXu/eventframed/internal/researchretentionlaw"\n)\n\n`;
fs.writeFileSync(out,header+body,{flag:'wx',mode:0o600});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.writeFileSync('research/retention-v64-generation.json',JSON.stringify({source:src,sourceSHA256:hash(raw),output:out,outputSHA256:hash(header+body),stage:'mechanical extraction, not final fixture'},null,2)+'\n',{flag:'wx',mode:0o600});
