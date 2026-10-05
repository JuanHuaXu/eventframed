import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-generation-growth/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw Error(`ambiguous anchor: ${a}`);s=s.replace(a,b);}
change('type build struct {','type build struct {\n Partition, BeforeTotal, AfterTotal, BeforeLocal, AfterLocal int\n BeforeRevision, AfterRevision uint64');
change('researchindex.RestoreDurable(768, 64, 1, seed, persist)','researchindex.RestorePartitioned(768, 64, 8, 1, seed, persist)');
change('researchindex.NewServingBounded(ctx, d,','researchindex.NewPartitionServing(ctx, d,');
change('count++\n\t\t\tstart := time.Now()',`partition, largest := 0, -1
            for p:=0;p<v.Count();p++ {pv,e:=v.Partition(p);check(e);if pv.DeltaCount()>largest {partition,largest=p,pv.DeltaCount()}}
            count++
            start := time.Now()`);
change('s.Compact(ctx, filepath.Join','s.Compact(ctx, partition, filepath.Join');
change('b := build{NS: time.Since(start).Nanoseconds()}',`b := build{NS: time.Since(start).Nanoseconds(),Partition:partition,BeforeTotal:v.DeltaCount(),BeforeLocal:largest,BeforeRevision:v.Revision()}
            after,viewErr:=d.View(ctx)
            if viewErr==nil {b.AfterTotal=after.DeltaCount();b.AfterRevision=after.Revision();pv,e:=after.Partition(partition);check(e);b.AfterLocal=pv.DeltaCount()} else if err==nil {err=viewErr}`);
change('"cmd/research-generation-growth/main.go"','"cmd/research-partition-load/main.go", "research/public-task-pilot/PARTITION_LOAD_PROTOCOL.md", "internal/researchindex/partition.go", "internal/researchindex/partition_durable.go", "internal/researchindex/partition_serving.go"');
fs.mkdirSync('cmd/research-partition-load',{recursive:true});fs.writeFileSync('cmd/research-partition-load/main.go',s,{flag:'wx'});
