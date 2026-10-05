import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';

const out=path.dirname(fileURLToPath(import.meta.url));
const read=f=>fs.readFileSync(path.join(out,f));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'manifest.json')))throw Error('preserve prior preparation');
const parent=JSON.parse(read('../index-overlay-v6/service-manifest.json'));
const dep=JSON.parse(read('../index-overlay-v6/manifest.json'));
const depPins=JSON.parse(read('../index-overlay-v6/SOURCE_PINS.json'));
for(const[arm,files]of Object.entries(parent.sources))for(const[f,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(parent.paths[arm],f)))!==h)throw Error('parent service drift '+arm+'/'+f);
// Dependency pins name all original arms, not just the two source changes.
for(const[f,h]of Object.entries(depPins.files))if(hash(fs.readFileSync(f))!==h)throw Error('parent dependency drift '+f);
const root=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-overlay-shard-factorial-'));
const paths={},sources={};
function pin(dir){const files={};function visit(rel){for(const d of fs.readdirSync(path.join(dir,rel),{withFileTypes:true})){if(d.name==='.git'||d.name==='node_modules')continue;const f=path.join(rel,d.name);if(d.isDirectory())visit(f);else if(d.name.endsWith('.go')||d.name==='go.mod'||d.name==='go.sum'||f==='testdata/text-public-facts/corpus.jsonl')files[f]=hash(fs.readFileSync(path.join(dir,f)));}}visit('');return files;}
for(const arm of['control','candidate']){const dest=path.join(root,'library-'+arm);fs.cpSync(dep.paths[arm],dest,{recursive:true});paths['library-'+arm]=dest;}
// The copied library's sibling lexer replacement must keep exactly its source.
fs.cpSync(path.join(path.dirname(dep.paths.control),'lexer'),path.join(root,'lexer'),{recursive:true});
paths.lexer=path.join(root,'lexer');
for(const arm of['rebuild','overlay','shard','both']){
 const dest=path.join(root,arm);fs.cpSync(parent.paths.control,dest,{recursive:true});paths[arm]=dest;
 const overlay=arm==='overlay'||arm==='both',shard=arm==='shard'||arm==='both';
 execFileSync('go',['mod','edit','-replace=github.com/xDarkicex/libravdb='+paths['library-'+(overlay?'candidate':'control')]],{cwd:dest});
 if(shard){const f=path.join(dest,'internal/store/libravdbstore/store.go'),s=fs.readFileSync(f,'utf8'),needle='\t\tlibra.WithHNSW(16, 200, 100),\n';if(s.split(needle).length!==2)throw Error('unique unchanged shard option location required');fs.writeFileSync(f,s.replace(needle,needle+'\t\tlibra.WithSharding(true),\n'));
  fs.writeFileSync(path.join(dest,'internal/store/libravdbstore/sharded_topology_test.go'),execFileSync('gofmt',[],{input:read('../sharded-capture-load-v1/topology_test.go.txt')}));
 }
}
for(const[k,p]of Object.entries(paths))sources[k]=pin(p);
const manifest={paths,sources,protocolSHA256:hash(read('PROTOCOL.md')),preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),parentServiceManifestSHA256:hash(read('../index-overlay-v6/service-manifest.json')),parentDependencyPinsSHA256:hash(read('../index-overlay-v6/SOURCE_PINS.json')),qualityFixtureSHA256:parent.sources.control['internal/service/sharded_quality_test.go'],loadFixtureSHA256:parent.sources.control['internal/service/public_capture_load_test.go'],arms:{rebuild:{overlay:false,shards:1},overlay:{overlay:true,shards:1},shard:{overlay:false,shards:4},both:{overlay:true,shards:4}},workers:10,wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
fs.writeFileSync(path.join(out,'manifest.json'),JSON.stringify(manifest,null,2)+'\n');console.log(JSON.stringify({prepared:true,root,files:Object.values(sources).reduce((n,x)=>n+Object.keys(x).length,0)}));
