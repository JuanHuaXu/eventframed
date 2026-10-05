import fs from 'node:fs';
const dir='research/public-task-pilot/pair-observation-overlay-v1';fs.mkdirSync(dir);
const mirror=`${process.cwd()}/research/public-task-pilot/candidate-libravdb-v1.6.13`;
const replace={...JSON.parse(fs.readFileSync('research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json')).Replace};
function edit(rel,base,fn){let s=fs.readFileSync(base,'utf8');s=fn(s);const out=`${dir}/${rel.replaceAll('/','_')}.txt`;fs.writeFileSync(out,s,{flag:'wx'});replace[`${mirror}/${rel}`]=`${process.cwd()}/${out}`;}
edit('internal/index/hnsw/hnsw.go',`${mirror}/internal/index/hnsw/hnsw.go`,s=>s.replace('type Index struct {','type Index struct {\n researchPairObserve func(uint32,uint32)'));
edit('internal/index/hnsw/neighbors.go',`${mirror}/internal/index/hnsw/neighbors.go`,s=>{
 const a='shouldSelect := !index.rejectBySelectedHeuristic(';if(s.split(a).length!==2)throw Error('heuristic anchor');
 s=s.replace(a,'shouldSelect := !index.rejectBySelectedHeuristic(');
 const start=s.indexOf('\t\tshouldSelect := !index.rejectBySelectedHeuristic('),end=s.indexOf('\n\n\t\tif shouldSelect',start);
 if(start<0||end<0)throw Error('selection span');const original=s.slice(start,end);
 return s.slice(0,start)+`        shouldSelect:=true
        if index.researchPairObserve!=nil && !usePtrSIMD {
         cutoff:=index.relaxedHeuristicCutoff(candidate.Distance)
         for j,selectedVector:=range selectedVectors {
          if selectedVector==nil {continue}
          index.researchPairObserve(candidate.ID,selected[j].ID)
          if index.distance(candidateVector,selectedVector)<cutoff {shouldSelect=false;break}
         }
        } else {
${original.replace('shouldSelect :=','shouldSelect =')}
        }`+s.slice(end);
});
// Setter is used only before workers start and cleared after they join.
const hp=`${dir}/internal_index_hnsw_hnsw.go.txt`;
fs.appendFileSync(hp,'\nfunc(h *Index) ResearchSetPairObserver(observe func(uint32,uint32)){h.researchPairObserve=observe}\n');
edit('internal/index/interfaces.go',`${mirror}/internal/index/interfaces.go`,s=>s+'\nfunc(w *hnswWrapper) ResearchSetPairObserver(observe func(uint32,uint32)){w.index.ResearchSetPairObserver(observe)}\n');
edit('libravdb/collection.go','research/public-task-pilot/candidate-only-overlay-v1/collection.go.txt',s=>{
 const anchor='\tif err := insertEntriesIntoIndex(ctx, idx, config.Metric, entries); err != nil {';if(s.split(anchor).length!==2)throw Error('fresh build anchor');
 s=s.replace(anchor,` if factory,ok:=ctx.Value(researchPairObserverKey{}).(func(int)(func(uint32,uint32),func()));ok {
  if observable,ok:=idx.(interface{ResearchSetPairObserver(func(uint32,uint32))});ok {
   observe,finish:=factory(len(entries));observable.ResearchSetPairObserver(observe)
   defer func(){observable.ResearchSetPairObserver(nil);finish()}()
  }
 }
${anchor}`);
 return s+'\ntype researchPairObserverKey struct{}\nfunc ResearchObserveBuildPairs(ctx context.Context,factory func(int)(func(uint32,uint32),func()))context.Context{return context.WithValue(ctx,researchPairObserverKey{},factory)}\n';
});
fs.writeFileSync(`${dir}/overlay.json`,JSON.stringify({Replace:replace},null,2),{flag:'wx'});
let r=fs.readFileSync('cmd/research-candidate-screen/main.go','utf8');
function change(a,b){if(r.split(a).length!==2)throw Error(a);r=r.replace(a,b);}
change('"time"','"time"\n"sync"\n libra "github.com/xDarkicex/libravdb/libravdb"');
change('ctx := context.Background()',`type pairStats struct{Records int;Calls,Repeated,Untracked uint64;Entries int}
 var observations []pairStats
 statsFile,e:=os.OpenFile(os.Args[1]+".pairs.json",os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);check(e)
 defer func(){check(json.NewEncoder(statsFile).Encode(observations));check(statsFile.Close())}()
 ctx := libra.ResearchObserveBuildPairs(context.Background(),func(n int)(func(uint32,uint32),func()){
  var mu sync.Mutex;seen:=map[uint64]bool{};stats:=pairStats{Records:n}
  return func(a,b uint32){mu.Lock();defer mu.Unlock();stats.Calls++;key:=uint64(a)<<32|uint64(b);if seen[key]{stats.Repeated++}else if len(seen)<65536{seen[key]=true}else{stats.Untracked++}},func(){stats.Entries=len(seen);observations=append(observations,stats)}
 })`);
change('"cmd/research-candidate-screen/main.go"',`"cmd/research-pair-observation/main.go", "${dir}/overlay.json", "${dir}/internal_index_hnsw_hnsw.go.txt", "${dir}/internal_index_hnsw_neighbors.go.txt", "${dir}/internal_index_interfaces.go.txt", "${dir}/libravdb_collection.go.txt"`);
fs.mkdirSync('cmd/research-pair-observation');fs.writeFileSync('cmd/research-pair-observation/main.go',r,{flag:'wx'});
