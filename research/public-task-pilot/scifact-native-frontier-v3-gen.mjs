import fs from 'node:fs';import path from 'node:path';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='cmd/research-public-native-resume/main.go',dest='cmd/research-public-native-frontier/main.go';
const raw=fs.readFileSync(source,'utf8'),freeze=JSON.parse(fs.readFileSync('research/public-task-pilot/scifact-native-resume-v2/freeze.json'));
assert.equal(hash(raw),freeze.sources[path.resolve(source)].sha256);
const changes=[
 ['"github.com/JuanHuaXu/eventframed/internal/researchpublicresume"','"github.com/JuanHuaXu/eventframed/internal/researchpublicresume"\n\t"github.com/JuanHuaXu/eventframed/internal/researchpublicrankguard"'],
 ['research-public-native-resume CORPUS.json','research-public-native-frontier CORPUS.json'],
 ['\t\tranked := []retrieval.Candidate{}','\t\tnativeRanked := []retrieval.Candidate{}\n\t\tnativeComplete := len(rows)==0\n\t\tranked := []retrieval.Candidate{}'],
 ['\t\t\tbound, err = r.BindRanked(ctx, rows, ranked, clock, 200)',
  '\t\t\tnativeRanked = ranked\n\t\t\tselection, selectErr := researchpublicrankguard.CompleteOrSearch(ctx,r,rows,nativeRanked,clock,200)\n\t\t\tif selectErr != nil {\n\t\t\t\t_ = write(map[string]any{"kind":"query_error","index":i,"id":q.ID,"text":q.Text,"search":rows,"nativeRanked":nativeRanked,"error":selectErr.Error()})\n\t\t\t\treturn selectErr\n\t\t\t}\n\t\t\tranked = selection.Candidates\n\t\t\tnativeComplete = selection.NativeComplete\n\t\t\tbound, err = r.BindRanked(ctx, rows, ranked, clock, 200)'],
 ['"search": rows, "ranked": ranked, "sources": sources,',
  '"search": rows, "nativeRanked":nativeRanked,"nativeComplete":nativeComplete,"nativeCalled":len(rows)>0,"ranked": ranked, "sources": sources,'],
];
let out=raw;for(const[a,b]of changes){assert.equal(out.split(a).length,2);out=out.replace(a,b);}let inverse=out;for(const[a,b]of changes.toReversed())inverse=inverse.replace(b,a);assert.equal(inverse,raw);
fs.mkdirSync(path.dirname(dest),{mode:0o700});fs.writeFileSync(dest,out,{flag:'wx',mode:0o600});
const testSource='cmd/research-public-native-resume/main_test.go',testDest='cmd/research-public-native-frontier/main_test.go',tests=fs.readFileSync(testSource);assert.equal(hash(tests),freeze.sources[path.resolve(testSource)].sha256);
fs.writeFileSync(testDest,tests,{flag:'wx',mode:0o600});fs.writeFileSync('research/public-task-pilot/scifact-native-frontier-v3-generation.json',JSON.stringify({source,dest,
 sourceSHA256:hash(raw),preformatDestSHA256:hash(out),completeInverseEquality:true,testCopySHA256:hash(tests),
 changed:'journal native advice; incomplete valid advice rejected, emit whole unchanged search frontier; invalid advice remains failure'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log('complete inverse client generation PASS; complete final frontier retained');
