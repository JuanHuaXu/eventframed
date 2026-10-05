import fs from 'node:fs';
const read=p=>fs.readFileSync(p,'utf8');
const once=(s,a,b)=>{if(s.split(a).length!==2)throw Error(`nonunique ${a}`);return s.replace(a,b)};
let run=read('internal/observationgate/age_learning_run_test.go');
run=once(run,'func ageLearningRun(','func ageBreadthRun(');
run=once(run,'schedule feedbackSchedule) (ageRecord, error)','schedule feedbackSchedule, cfg subsetBreadthCase) (ageRecord, error)');
run=once(run,'name := observationpreserved.Scenarios[scenario]','name := cfg.Mode');
run=once(run,'Scenario: name','Scenario: cfg.Name');
run=once(run,'x, rx := uint16(live.Intn(512)), uint16(ref.Intn(512))','x, rx := breadthInput(uint16(live.Intn(512)), cfg.Dependent), breadthInput(uint16(ref.Intn(512)), cfg.Dependent)');
run=once(run,'integrationLabel(x, local, name == "null", live), integrationLabel(rx, name == "common_shift" && step >= 256, name == "null", ref)','breadthLabel(x, local, name == "null", live, cfg.Target), breadthLabel(rx, name == "common_shift" && step >= 256, name == "null", ref, cfg.Target)');
fs.writeFileSync('internal/observationgate/age_breadth_run_test.go',run,{flag:'wx'});
let src=read('internal/observationgate/age_challenger_test.go');
let test=src.slice(src.indexOf('func TestAgeChallengerV87('));
test=test.replaceAll('TestAgeChallengerV87','TestAgeBreadthV88').replaceAll('EVENTFRAME_AGE_','EVENTFRAME_AGE_BREADTH_').replaceAll('v87','v88').replaceAll('age-challenger','age-breadth');
test=once(test,'for scenario, name := range observationpreserved.Scenarios {','for scenario, cfg := range subsetBreadthCases {\n name := cfg.Name');
test=once(test,'base, e := observationpreserved.Base(name == "null")\n\t\t\tif e != nil {\n\t\t\t\tt.Fatal(e)\n\t\t\t}','');
test=once(test,'range learningSchedules {','range []feedbackSchedule{learningSchedules[0], learningSchedules[3]} {');
test=once(test,'r, e := ageLearningRun(base, split, scenario, i, int64(2026118701+phase), schedule)',`trainSeed := int64(2026118800)*1000000 + int64(scenario*10000+phase*1000+i)
 base,e := breadthBase(cfg,trainSeed)
 if e != nil { t.Fatal(e) }
 result, e := ageBreadthRun(base, split, scenario, i, int64(2026118801+10*scenario+phase), schedule, cfg)
 r := ageBreadthRecord{ageRecord:result, Config:cfg, TrainSeed:trainSeed}`);
test=once(test,'var want ageRecord','var want ageBreadthRecord');
const imports=`package observationgate
import (
 "bytes"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "io"
 "os"
 "path/filepath"
 "testing"
 "github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)
type ageBreadthRecord struct { ageRecord; Config subsetBreadthCase; TrainSeed int64 }
func TestAgeBreadthContracts(t *testing.T){
 seen:=map[int64]bool{}
 add:=func(seed int64){if seen[seed]{t.Fatal("seed collision",seed)};seen[seed]=true}
 for phase:=0;phase<2;phase++ {for j:=range subsetBreadthCases {for i:=0;i<64;i++ {
  add(int64(2026118800)*1000000+int64(j*10000+phase*1000+i))
  for role:=0;role<5;role++ {add(observationpreserved.Seed(int64(2026118801+10*j+phase),j,i,role))}
 }}}
 cfg:=subsetBreadthCases[1]
 base,e:=breadthBase(cfg,2026118899);if e!=nil{t.Fatal(e)}
 for _,schedule:=range []feedbackSchedule{learningSchedules[0],learningSchedules[3]} {
  a,e:=ageLearningRun(base,"unit",1,0,2026118899,schedule);if e!=nil{t.Fatal(e)}
  b,e:=ageBreadthRun(base,"unit",1,0,2026118899,schedule,cfg);if e!=nil{t.Fatal(e)}
  b.Scenario=a.Scenario
  if a!=b{t.Fatal("unchanged bit integration parity")}
 }
}
`;
fs.writeFileSync('internal/observationgate/age_breadth_test.go',imports+test,{flag:'wx'});
