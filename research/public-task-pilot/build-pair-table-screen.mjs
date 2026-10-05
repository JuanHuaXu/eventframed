import fs from 'node:fs';
const dir='research/public-task-pilot/build-pair-table-overlay-v1';
fs.mkdirSync(dir);
const old='research/public-task-pilot/build-pair-cache-overlay-v1';
const original=JSON.parse(fs.readFileSync(`${old}/overlay.json`));
const shape='interface{Lookup(uint32,uint32)(float32,bool);Store(uint32,uint32,float32)}';
const replace={};
for(const [target,source] of Object.entries(original.Replace)){
 if(!source.includes('build-pair-cache-overlay-v1')){replace[target]=source;continue;}
 let s=fs.readFileSync(source,'utf8');
 if(target.endsWith('neighbors.go')){
  const a='value:=index.researchPairCache(candidate.ID,selected[j].ID,func()float32{return index.distance(candidateVector,selectedVector)})';
  if(s.split(a).length!==2)throw Error('distance anchor');
  s=s.replace(a,'value,hit:=index.researchPairCache.Lookup(candidate.ID,selected[j].ID)\n          if !hit {value=index.distance(candidateVector,selectedVector);index.researchPairCache.Store(candidate.ID,selected[j].ID,value)}');
 }
 s=s.replaceAll('func(uint32,uint32,func()float32)float32',shape)
  .replaceAll('researchPairCache','researchPairTable')
  .replaceAll('ResearchSetPairCache','ResearchSetPairTable')
  .replaceAll('ResearchCacheBuildPairs','ResearchTableBuildPairs');
 const out=`${dir}/${source.split('/').at(-1)}`;
 fs.writeFileSync(out,s,{flag:'wx'});replace[target]=`${process.cwd()}/${out}`;
}
fs.writeFileSync(`${dir}/overlay.json`,JSON.stringify({Replace:replace},null,2),{flag:'wx'});
let r=fs.readFileSync('cmd/research-build-pair-cache/main.go','utf8')
 .replaceAll('research-build-pair-cache','research-build-pair-table')
 .replaceAll('build_pair_cache.go','build_pair_table.go')
 .replaceAll('BUILD_PAIR_CACHE_PROTOCOL','BUILD_PAIR_TABLE_PROTOCOL')
 .replaceAll('build-pair-cache-overlay-v1','build-pair-table-overlay-v1')
 .replaceAll('ResearchCacheBuildPairs','ResearchTableBuildPairs')
 .replaceAll('NewBuildPairCache','NewBuildPairTable')
 .replaceAll('func(uint32, uint32, func() float32) float32',shape)
 .replaceAll('return cache.Distance,','return cache,');
fs.mkdirSync('cmd/research-build-pair-table');
fs.writeFileSync('cmd/research-build-pair-table/main.go',r,{flag:'wx'});
