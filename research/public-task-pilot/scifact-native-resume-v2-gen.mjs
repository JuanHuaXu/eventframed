// The original failed fit client stays frozen; only a new offline-resume client
// gains prefix validation/readback before its otherwise identical import/query path.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import path from 'node:path';
import crypto from 'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='cmd/research-public-native-fit/main.go',dest='cmd/research-public-native-resume/main.go';
const raw=fs.readFileSync(source,'utf8'),freeze=JSON.parse(fs.readFileSync('research/public-task-pilot/scifact-native-fit-v1/freeze.json'));
assert.equal(hash(raw),freeze.sources[path.resolve(source)].sha256);
let out=raw;
const changes=[
 ['"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"','"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"\n\t"github.com/JuanHuaXu/eventframed/internal/researchpublicresume"'],
 ['if len(os.Args) != 5 {','if len(os.Args) != 6 {'],
 ['usage: research-public-native-fit CORPUS.json FIT_ONLY.json OWNED_ROOT','usage: research-public-native-resume CORPUS.json FIT_ONLY.json OWNED_ROOT TRACE.ndjson PREFIX.json'],
 ['\tclient, err := retrieval.OpenLibraVDBContractsWithConfig(',
  '\tvar prefix struct { IDs []string `json:"ids"` }\n\tpb, err := os.ReadFile(os.Args[5]); if err != nil { return err }\n\tif err = json.Unmarshal(pb, &prefix); err != nil { return err }\n\tif err = researchpublicresume.ValidatePlan(r,prefix.IDs); err != nil { return err }\n\treader, err := researchpublicresume.Open(root,cwd); if err != nil { return err }; defer reader.Close()\n\tclient, err := retrieval.OpenLibraVDBContractsWithConfig('],
 ['\tfor i, e := range r.Entries() {',
  '\tif err = researchpublicresume.Verify(ctx,r,prefix.IDs,clock,reader,func(i int,e researchpublicpool.Entry,rows []retrieval.Candidate,ns int64)error {\n\t\treturn write(map[string]any{"kind":"verify","index":i,"id":e.Candidate.ID,"source":e.SourceID,"rows":rows,"ns":ns,"ok":true})\n\t}); err != nil { return err }\n\tfor i, e := range r.Entries() {\n\t\tif i < len(prefix.IDs) { continue }'],
 ['"totalNS": time.Since(started).Nanoseconds(), "labelsRead": false, "confirmationPredictions": 0}',
  '"totalNS": time.Since(started).Nanoseconds(), "labelsRead": false, "confirmationPredictions": 0,"verifiedPrefix":len(prefix.IDs)}'],
];
for(const[a,b]of changes){assert.equal(out.split(a).length,2);out=out.replace(a,b);}
let inverse=out;for(const[a,b]of changes.toReversed())inverse=inverse.replace(b,a);assert.equal(inverse,raw);
fs.mkdirSync(path.dirname(dest),{mode:0o700});fs.writeFileSync(dest,out,{flag:'wx',mode:0o600});
const testSource='cmd/research-public-native-fit/main_test.go',testDest='cmd/research-public-native-resume/main_test.go';
const tests=fs.readFileSync(testSource);assert.equal(hash(tests),freeze.sources[path.resolve(testSource)].sha256);
fs.writeFileSync(testDest,tests,{flag:'wx',mode:0o600});
fs.writeFileSync('research/public-task-pilot/scifact-native-resume-v2-generation.json',JSON.stringify({source,dest,sourceSHA256:hash(raw),
 preformatDestSHA256:hash(out),completeInverseEquality:true,testSource,testDest,testCopySHA256:hash(tests),
 queryAlgorithmAndFullCorpusUnchanged:true,changed:'extra immutable prefix plan, complete native readback, skip only all-verified prefix'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log('complete inverse client generation PASS; original query algorithm unchanged');
