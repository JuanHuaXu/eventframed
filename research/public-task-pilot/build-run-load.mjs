import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-candidate-load/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw Error(a);s=s.replace(a,b);}
change('type build struct {','type build struct {\n Mode string\n BeforeRuns,AfterRuns int');
const a=s.indexOf('\td, err := researchindex.RestorePartitioned('),b=s.indexOf('\tstop, stopped :=',a);
if(a<0||b<0)throw Error('setup');
s=s.slice(0,a)+` base,err:=researchindex.BuildImmutableRun(ctx,seed,768,filepath.Join(dir,"base-initial"));check(err)
 w,err:=researchindex.NewDurableRunWriter(ctx,[]*researchindex.ImmutableRun{base},1,768,64,10,persist);check(err)
 pool,err:=researchindex.NewRunLeasePool(8,2);check(err)
`+s.slice(b);
const x=s.indexOf('\t\t\tv, err := d.View(ctx)'),y=s.indexOf('\t\t\tif err != nil {\n\t\t\t\tb.Error',x);
if(x<0||y<0)throw Error('build body');
s=s.slice(0,x)+` v,err:=w.Status(ctx)
 if err!=nil {a.Builds=append(a.Builds,build{Error:err.Error()});return}
 if v.Delta<32 {continue}
 count++;start:=time.Now();mode:="flush"
 if v.Runs>=2 {
  mode="consolidate"
  c,e:=w.PrepareRunConsolidation(ctx,filepath.Join(dir,fmt.Sprintf("base-%d",count)));err=e
  if err==nil {err=pool.PublishConsolidation(ctx,c);if err!=nil {err=errors.Join(err,c.Abort())}}
 } else {
  f,e:=w.PrepareRunFlush(ctx,filepath.Join(dir,fmt.Sprintf("base-%d",count)));err=e
  if err==nil {err=f.Publish(ctx)}
 }
 b:=build{Mode:mode,BeforeRuns:v.Runs,NS:time.Since(start).Nanoseconds(),BeforeTotal:v.Delta,BeforeLocal:v.Delta,BeforeRevision:v.Revision}
 after,viewErr:=w.Status(ctx)
 if viewErr==nil {b.AfterTotal=after.Delta;b.AfterLocal=after.Delta;b.AfterRevision=after.Revision;b.AfterRuns=after.Runs}else if err==nil {err=viewErr}
`+s.slice(y);
change('s.Acquire(c)','pool.AcquireCurrent(c,w)');
change('lease.Search(c, vector(id), 10)','lease.Search(c, vector(id))');
change('s.Apply(c,','w.Apply(c,');
change('cleanupErr := s.Close()','cleanupErr := pool.Close()\n if cleanupErr==nil {cleanupErr=w.CloseRunsAfterReaders(ctx)}');
change('"cmd/research-candidate-load/main.go"','"cmd/research-run-load/main.go", "research/public-task-pilot/RUN_LOAD_PROTOCOL.md", "internal/researchindex/immutable_run.go", "internal/researchindex/run_merge.go", "internal/researchindex/run_delta.go", "internal/researchindex/run_writer.go", "internal/researchindex/run_flush.go", "internal/researchindex/run_consolidation.go", "internal/researchindex/run_leases.go", "internal/researchindex/run_publication.go", "internal/researchindex/run_load_support.go"');
fs.mkdirSync('cmd/research-run-load');fs.writeFileSync('cmd/research-run-load/main.go',s,{flag:'wx'});
