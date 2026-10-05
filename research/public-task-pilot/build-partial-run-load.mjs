import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-run-load/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw Error(a);s=s.replace(a,b);}
const start=s.indexOf('\tbase, err := researchindex.BuildImmutableRun'),end=s.indexOf('\tpool, err :=',start);
if(start<0||end<0)throw Error('setup');
s=s.slice(0,start)+` var bases []*researchindex.ImmutableRun
 for part:=0;part<8;part++ {var entries []researchindex.Mutation;for _,m:=range seed {owner,e:=researchindex.ResearchPartition(m.ID,8);check(e);if owner==part {entries=append(entries,m)}}
  base,e:=researchindex.BuildImmutableRun(ctx,entries,768,filepath.Join(dir,fmt.Sprintf("base-initial-%d",part)));check(e);bases=append(bases,base)
 }
 w,err:=researchindex.NewDurableRunWriter(ctx,bases,1,768,64,10,persist);check(err)
`+s.slice(end);
change('if v.Delta < 32 {','if v.Delta < 32 && v.Runs < 10 {');
change('if v.Runs >= 2 {','if v.Runs >= 10 {');
change('mode = "consolidate"','mode = "partial"');
change('w.PrepareRunConsolidation(','w.PreparePartialRunMerge(');
change('pool.PublishConsolidation(ctx, c)','pool.PublishPartialRunMerge(ctx, c)');
change('"cmd/research-run-load/main.go"','"cmd/research-partial-run-load/main.go", "internal/researchindex/run_partial.go", "research/public-task-pilot/PARTIAL_RUN_LOAD_PROTOCOL.md"');
fs.mkdirSync('cmd/research-partial-run-load');fs.writeFileSync('cmd/research-partial-run-load/main.go',s,{flag:'wx'});
