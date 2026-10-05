import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-generation-bulk/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw new Error(`ambiguous anchor:${a}`);s=s.replace(a,b);}
change(`	h := sha256.Sum256([]byte(id))
	v := make([]float32, 32)
	for i, b := range h {
		v[i] = float32(b) - 127.5
	}`,`	v := make([]float32, 768)
 for block:=0;block<24;block++ {
  h:=sha256.Sum256([]byte(fmt.Sprintf("%s/block-%d",id,block)))
  for i,b:=range h { v[block*32+i]=float32(b)-127.5 }
 }`);
change('"records", 32, libra.WithFlat()','"records", 768, libra.WithFlat()');
change('researchindex.RestoreDurable(32, 64','researchindex.RestoreDurable(768, 64');
change('"cmd/research-generation-bulk/main.go"','"cmd/research-generation-wide/main.go", "research/public-task-pilot/WIDE_VECTOR_PROTOCOL.md"');
change('Hashes map[string]string\n\t\tArms','Dimension int\n\t\tHashes map[string]string\n\t\tArms');
change('}{hashes, arms}))','}{768, hashes, arms}))');
fs.mkdirSync('cmd/research-generation-wide');fs.writeFileSync('cmd/research-generation-wide/main.go',s,{flag:'wx',mode:0o600});
