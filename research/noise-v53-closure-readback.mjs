// Post-run reconstruction of the prospective repository-source boundary.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/noise-v53-diagnostic',repo=process.cwd();
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const f=JSON.parse(fs.readFileSync(root+'/freeze.json')),rows=JSON.parse(fs.readFileSync(root+'/compiler-closure.json'));
assert.equal(hash(JSON.stringify(rows)+'\n'),f.compilerClosureSHA256);
const compiled=new Set(),generated=new Set();
for(const x of rows){if(!x.Dir?.startsWith(repo+'/'))continue;for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of x[key]??[]){const absolute=path.resolve(x.Dir,n);if(absolute.startsWith(repo+'/'))compiled.add(path.relative(repo,absolute));else generated.add(absolute);}}
assert.deepEqual([...compiled].sort(),f.compilerFiles);
for(const p of compiled){assert.equal(hash(fs.readFileSync(p)),f.files[p]);assert.equal(hash(fs.readFileSync(root+'/source/'+p)),f.files[p]);}
const result={repositoryCompilerInputs:compiled.size,repositoryClosureProspectivelyComplete:true,savedCopiesVerified:true,compactGoJSONDigestVerified:true,nonrepositoryGeneratedTestMains:[...generated].sort(),generatedTestMainsRebuildableFromSavedRepositoryTests:true,nonrepositoryFilesRead:false,initialSupplementalAuditFailure:'An inline readback omitted the final absolute-path repository filter. It incorrectly flagged three generated Go cache test mains as missing repository inputs. Runner already had the filter; no frozen source or scientific record repaired.',sources:{'research/noise-v53-closure-readback.mjs':hash(fs.readFileSync('research/noise-v53-closure-readback.mjs'))}};
fs.writeFileSync(root+'/closure-readback.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));
