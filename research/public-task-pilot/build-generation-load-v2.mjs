import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-generation-load/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw new Error(`ambiguous anchor:${a}`);s=s.replace(a,b);}
change('check(persist(ctx, 1, seed))',`// Setup is unserved; all bounded chunks belong to initial revision1.
 for i:=0;i<len(seed);i+=16 { check(persist(ctx,1,seed[i:min(i+16,len(seed))])) }`);
change('"cmd/research-generation-load/main.go"','"cmd/research-generation-load-v2/main.go", "research/public-task-pilot/COMPACTION_LOAD_SETUP_CORRECTION.md"');
change('arms = append(arms, a)',`arms = append(arms, a)
 partial,e:=os.OpenFile(fmt.Sprintf("%s.arm-n%d-gap%d-r%d.json",out,n,gap,r),os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);check(e)
 check(json.NewEncoder(partial).Encode(a));check(partial.Close())`);
fs.mkdirSync('cmd/research-generation-load-v2');fs.writeFileSync('cmd/research-generation-load-v2/main.go',s,{flag:'wx',mode:0o600});
