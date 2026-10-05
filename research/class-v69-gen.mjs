// Mechanical isolated clones. V68 implementation and artifacts stay immutable.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),files={};
for(const[from,to]of[['internal/researchsharedsequence','internal/researchclasssequence'],['internal/researchsharedsequenceref','internal/researchclasssequenceref']]){
 assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});
 for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go'))){const source=from+'/'+n,raw=fs.readFileSync(source),out=to+'/'+n,b=raw.toString().replaceAll('researchsharedsequenceref','researchclasssequenceref').replaceAll('researchsharedsequence','researchclasssequence').replaceAll('EVENTFRAME_SHARED_V68','EVENTFRAME_CLASS_V69');fs.writeFileSync(out,b,{flag:'wx',mode:0o600});files[out]={source,sourceSHA256:hash(raw),initialOutputSHA256:hash(b)}}
}
const source='internal/researchdispersion/shared_v68_test.go',raw=fs.readFileSync(source);let b=raw.toString().replaceAll('SharedV68','ClassV69').replaceAll('researchsharedsequenceref','researchclasssequenceref').replaceAll('researchsharedsequence','researchclasssequence').replaceAll('EVENTFRAME_SHARED_V68','EVENTFRAME_CLASS_V69');
const start=b.indexOf('var modesClassV69 = '),end=b.indexOf('\n',start);assert(start>0);
b=b.slice(0,start)+'var modesClassV69 = []string{"full", "adaptive", "hybrid_no_pair", "hybrid_random", "hybrid_uncertainty", "hybrid_information", "hybrid_falsification", "hybrid_predictive", "hybrid_model_class", "hybrid_noise_class"}'+b.slice(end);
const out='internal/researchdispersion/class_v69_test.go';fs.writeFileSync(out,b,{flag:'wx',mode:0o600});files[out]={source,sourceSHA256:hash(raw),initialOutputSHA256:hash(b)};
fs.writeFileSync('research/class-v69-generation.json',JSON.stringify({files,stage:'isolated clones before manually adding class acquisition and its independent checker'},null,2)+'\n',{flag:'wx',mode:0o600});
