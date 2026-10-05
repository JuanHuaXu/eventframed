import fs from 'node:fs';
const dir='research/public-task-pilot/build-pair-cache-overlay-v1';fs.mkdirSync(dir);
const original=JSON.parse(fs.readFileSync('research/public-task-pilot/pair-observation-overlay-v1/overlay.json'));
const replace={};
for(const [target,source]of Object.entries(original.Replace)){
 if(!source.includes('pair-observation-overlay-v1')){replace[target]=source;continue;}
 let s=fs.readFileSync(source,'utf8').replaceAll('ResearchObserveBuildPairs','ResearchCacheBuildPairs').replaceAll('ResearchSetPairObserver','ResearchSetPairCache').replaceAll('researchPairObserverKey','researchPairCacheKey').replaceAll('researchPairObserve','researchPairCache').replaceAll('func(uint32,uint32)','func(uint32,uint32,func()float32)float32');
 if(target.endsWith('neighbors.go')){
  const a='index.researchPairCache(candidate.ID,selected[j].ID)\n          if index.distance(candidateVector,selectedVector)<cutoff';
  if(s.split(a).length!==2)throw Error('distance anchor');
  s=s.replace(a,'value:=index.researchPairCache(candidate.ID,selected[j].ID,func()float32{return index.distance(candidateVector,selectedVector)})\n          if value<cutoff');
 }
 const out=`${dir}/${source.split('/').at(-1)}`;fs.writeFileSync(out,s,{flag:'wx'});replace[target]=`${process.cwd()}/${out}`;
}
fs.writeFileSync(`${dir}/overlay.json`,JSON.stringify({Replace:replace},null,2),{flag:'wx'});
let r=fs.readFileSync('cmd/research-candidate-screen/main.go','utf8');
function change(a,b){if(r.split(a).length!==2)throw Error(a);r=r.replace(a,b);}
change('"time"','"time"\n libra "github.com/xDarkicex/libravdb/libravdb"');
change('if len(os.Args) != 2 {','if len(os.Args) != 3 || (os.Args[2]!="on"&&os.Args[2]!="off") {');
change('ctx := context.Background()',`type pairStats struct{Records int;Stats ri.BuildPairStats}
 var observations []pairStats
 statsFile,e:=os.OpenFile(os.Args[1]+".pairs.json",os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);check(e)
 defer func(){check(json.NewEncoder(statsFile).Encode(struct{Mode string;Builds []pairStats}{os.Args[2],observations}));check(statsFile.Close())}()
 ctx := context.Background()
 if os.Args[2]=="on" {ctx=libra.ResearchCacheBuildPairs(ctx,func(n int)(func(uint32,uint32,func()float32)float32,func()){
  cache,e:=ri.NewBuildPairCache(65536);check(e)
  return cache.Distance,func(){observations=append(observations,pairStats{n,cache.Stats()})}
 })}`);
change('"cmd/research-candidate-screen/main.go"',`"cmd/research-build-pair-cache/main.go", "internal/researchindex/build_pair_cache.go", "research/public-task-pilot/BUILD_PAIR_CACHE_PROTOCOL.md", "${dir}/overlay.json", "${dir}/internal_index_hnsw_hnsw.go.txt", "${dir}/internal_index_hnsw_neighbors.go.txt", "${dir}/internal_index_interfaces.go.txt", "${dir}/libravdb_collection.go.txt"`);
fs.mkdirSync('cmd/research-build-pair-cache');fs.writeFileSync('cmd/research-build-pair-cache/main.go',r,{flag:'wx'});
