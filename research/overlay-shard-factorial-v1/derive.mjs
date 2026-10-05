import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
const root=path.dirname(fileURLToPath(import.meta.url)),hash=b=>crypto.createHash('sha256').update(b).digest('hex'),derivations=[];
function derive(src,dst,rules){let text=fs.readFileSync(path.join(root,src),'utf8'),original=text;for(const[a,b]of rules){if(!text.includes(a))throw Error('missing mechanical derivation '+a);text=text.replaceAll(a,b);}if(fs.existsSync(path.join(root,dst)))throw Error('preserve derivation');fs.writeFileSync(path.join(root,dst),text);derivations.push({src,dst,sourceSHA256:hash(original),destinationSHA256:hash(text),rules});}
derive('../index-overlay-v6/service-evaluate.mjs','quality-core.mjs',[['async function parse(root,c){','export async function parse(root,c){']]);
derive('../index-overlay-v6/service-independent.mjs','quality-independent.mjs',[
 ['service-run-results.json','run-results.json'],['service-evaluation.json','quality-evaluation.json'],['service-independent-results.json','quality-independent-results.json'],
 ['run.commands.filter(c=>c.full)','run.commands.filter(c=>c.kind==="quality")'],['checkedRows,3072','checkedRows,6144'],['checkedCells,32','checkedCells,64'],['checkedPhasePairs,48','checkedPhasePairs,96']
]);
fs.writeFileSync(path.join(root,'EVALUATOR_DERIVATION.json'),JSON.stringify({mechanical:true,derivations},null,2)+'\n');
