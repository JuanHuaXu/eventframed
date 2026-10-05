// Mechanical scaffolding only; new inference is implemented in separate files.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),sources={},outputs={};
function write(p,s){assert(!fs.existsSync(p));fs.mkdirSync(p.substring(0,p.lastIndexOf('/')),{recursive:true});fs.writeFileSync(p,s,{flag:'wx',mode:0o600});outputs[p]=hash(s)}
function read(p){const b=fs.readFileSync(p);sources[p]=hash(b);return b.toString()}
const model=read('internal/researchtree/model.go');
let tail=model.slice(model.indexOf('func (m *Model) Issue'));
tail=tail.replaceAll('m.nodes, m.weights = p.nodes, p.weights','m.commit(&p)').replaceAll('yes.nodes[0].tree[eta]','m.nodeAt(0, &yes).tree[eta]').replaceAll('no.nodes[0].tree[eta]','m.nodeAt(0, &no).tree[eta]').replaceAll('m.predict(i, &yes.nodes, yes.weights)','m.predict(i, &yes, yes.weights)').replaceAll('m.predict(i, &no.nodes, no.weights)','m.predict(i, &no, no.weights)');
write('internal/researchanchor/lifecycle.go','// Isolated V59 lifecycle, mechanically retained from audited V58.\npackage researchanchor\nimport("errors";"math")\n'+tail);
const reference=read('internal/researchtreeref/reference.go');
write('internal/researchanchorref/lifecycle.go','// Ledger lifecycle retained mechanically; inference rebuilt independently.\npackage researchanchorref\nimport("errors";"math";tree "github.com/JuanHuaXu/eventframed/internal/researchanchor")\n'+reference.slice(reference.indexOf('func (r *Reference) Predict')));
write('internal/researchanchor/benchmark_test.go',read('internal/researchtree/benchmark_test.go').replaceAll('package researchtree','package researchanchor').replaceAll(', Strength: 2',''));
write('internal/researchdispersion/tree_v59_generated_test.go',read('internal/researchdispersion/tree_v58_generated_test.go').replaceAll('V58','V59').replaceAll('researchtreeref','researchanchorref').replaceAll('researchtree','researchanchor').replaceAll(', Strength: 2',''));
write('research/tree-v59-readback.mjs',read('research/tree-v58-readback.mjs').replaceAll('tree-v58','tree-v59'));
write('research/tree-v59-comparison.mjs',read('research/tree-v58-comparison.mjs').replaceAll('tree-v58','tree-v59'));
let runner=read('research/tree-v58-run.mjs').replaceAll('tree-v58','tree-v59').replaceAll('V58','V59').replaceAll('researchtreeref','researchanchorref').replaceAll('researchtree','researchanchor').replaceAll('research/checkpoint-2026-10-04-paired-v57/manifest.json','research/checkpoint-2026-10-04-tree-v58/manifest.json').replaceAll('50abe81a49f9d973bd063a533f8d2cf5bdc8da10b2a6a536c77046a50068b4ed','e6dcbb9a2eed68b4e7fb4866ecaa6878306805913b6b21e2c6d7d4678cdcb0d7').replaceAll('strength:2','identityWeight:.8');
runner=runner.replace("'research/tree-v59-fixture-gen.mjs','research/tree-v59-fixture-generation.json','research/tree-v59-direction.md'","'research/tree-v59-gen.mjs','research/tree-v59-generation.json','research/tree-v59-anchor-direction.md','research/tree-v59-anchor-identities.mjs','research/tree-v59-anchor-identities.json'");
write('research/tree-v59-run.mjs',runner);
write('research/tree-v59-generation.json',JSON.stringify({sources,outputs,manualInferenceAndPatchesSeparate:true},null,2)+'\n');
console.log(JSON.stringify({outputs:Object.keys(outputs).length}));
