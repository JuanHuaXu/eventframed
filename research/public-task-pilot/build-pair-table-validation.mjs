import fs from 'node:fs';
const dir='research/public-task-pilot/build-pair-table-overlay-v1';
const overlay=JSON.parse(fs.readFileSync(`${dir}/overlay.json`));
const target=Object.keys(overlay.Replace).find(x=>x.endsWith('/neighbors.go'));
let s=fs.readFileSync(overlay.Replace[target],'utf8').replace('"fmt"','"fmt"\n "math"');
const a='value,hit:=index.researchPairTable.Lookup(candidate.ID,selected[j].ID)';
if(s.split(a).length!==2)throw Error('validation anchor');
s=s.replace(a,a+'\n          if hit && math.Float32bits(value)!=math.Float32bits(index.distance(candidateVector,selectedVector)){panic("pair table changed real distance")}');
const out=`${dir}/verified_neighbors.go.txt`;
fs.writeFileSync(out,s,{flag:'wx'});overlay.Replace[target]=`${process.cwd()}/${out}`;
fs.writeFileSync(`${dir}/verify-overlay.json`,JSON.stringify(overlay,null,2),{flag:'wx'});
let test=fs.readFileSync('internal/researchindex/build_pair_integration_test.go','utf8');
const start=test.indexOf('\tvar calls, hits');
const end=test.indexOf('\n\tdb, err :=',start);
if(start<0||end<0)throw Error('test anchors');
test=test.slice(0,start)+` var calls,hits atomic.Uint64
 ctx := libra.ResearchTableBuildPairs(context.Background(), func(n int)(interface{Lookup(uint32,uint32)(float32,bool);Store(uint32,uint32,float32)},func()){
  table,err:=NewBuildPairTable(65536);if err!=nil{t.Fatal(err)}
  observed:=&checkedPairTable{BuildPairTable:table,t:t,calls:&calls}
  return observed,func(){observed.ended.Store(true);hits.Add(table.Stats().Hits)}
 })
`+test.slice(end);
test=test.replace('research_pair_cache','research_pair_table').replace('TestBuildPairCacheRealDistanceAndLifetime','TestBuildPairTableRealDistanceAndLifetime').replace('\n\t"math"','');
test+=`
type checkedPairTable struct {
 *BuildPairTable
 t *testing.T
 calls *atomic.Uint64
 ended atomic.Bool
}
func(c *checkedPairTable) Lookup(a,b uint32)(float32,bool){
 if c.ended.Load(){c.t.Error("lookup after build")};c.calls.Add(1);return c.BuildPairTable.Lookup(a,b)
}
func(c *checkedPairTable) Store(a,b uint32,v float32){
 if c.ended.Load(){c.t.Error("store after build")};c.BuildPairTable.Store(a,b,v)
}
`;
fs.writeFileSync('internal/researchindex/build_pair_table_integration_test.go',test,{flag:'wx'});
