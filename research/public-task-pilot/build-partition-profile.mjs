import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-derived-nosync/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw Error(`ambiguous: ${a}`);s=s.replace(a,b);}
change('"runtime"','"runtime"\n"runtime/pprof"');
change('type arm struct {','type arm struct {\n MemBefore, MemAfter runtime.MemStats');
change('stop, stopped :=',`allocBefore,e:=os.Create(filepath.Join(dir,"alloc-before.pprof"));check(e)
 check(pprof.Lookup("allocs").WriteTo(allocBefore,0));check(allocBefore.Close())
 cpu,e:=os.Create(filepath.Join(dir,"cpu.pprof"));check(e);check(pprof.StartCPUProfile(cpu))
 runtime.ReadMemStats(&a.MemBefore)
 stop, stopped :=`);
change('defer close(stopped)','defer close(stopped)\n pprof.SetGoroutineLabels(pprof.WithLabels(ctx,pprof.Labels("phase","compact")))');
change('id := fmt.Sprintf("seed-%d", i%n)','pprof.SetGoroutineLabels(pprof.WithLabels(ctx,pprof.Labels("phase","read")))\n id := fmt.Sprintf("seed-%d", i%n)');
change('\t\t\tid := fmt.Sprintf("write-%d", i)','\t\t\tpprof.SetGoroutineLabels(pprof.WithLabels(ctx,pprof.Labels("phase","write")))\n id := fmt.Sprintf("write-%d", i)');
change('a.WallNS = time.Since(start).Nanoseconds()',`a.WallNS = time.Since(start).Nanoseconds()
 runtime.ReadMemStats(&a.MemAfter)
 pprof.StopCPUProfile();check(cpu.Close())
 runtime.GC()
 allocAfter,e:=os.Create(filepath.Join(dir,"alloc-after.pprof"));check(e)
 check(pprof.Lookup("allocs").WriteTo(allocAfter,0));check(allocAfter.Close())`);
change('r < 2','r < 1');change('[]int{800, 3200, 6400}','[]int{3200, 6400}');
change('"cmd/research-derived-nosync/main.go"','"cmd/research-partition-profile/main.go", "research/public-task-pilot/PARTITION_PROFILE_PROTOCOL.md"');
fs.mkdirSync('cmd/research-partition-profile');fs.writeFileSync('cmd/research-partition-profile/main.go',s,{flag:'wx'});
