// Research comparator: no fixture IDs, labels, relation dictionary or model calls.
const stop=new Set('a an the on in at to of for from with by and is was were did do does it that this which what when why how before after not request outcome recorded'.split(' '));
function terms(text){
 const out=new Set();
 for(let w of text.toLowerCase().match(/[a-z]+/g)??[]){
  if(stop.has(w)||w.length<3)continue;
  if(w.length>5&&w.endsWith('ing'))w=w.slice(0,-3);
  else if(w.length>4&&w.endsWith('ed'))w=w.slice(0,-2);
  else if(w.length>4&&w.endsWith('s'))w=w.slice(0,-1);
  out.add(w);
 }
 return out;
}
export function lexicalOrder(query,candidates,plan=null){
 if(typeof query!=='string'||query.length>4096||!candidates.length||candidates.length>200)throw Error('bounded lexical input required');
 const seen=new Set();
 const docs=candidates.map(c=>{
  if(!c.ID||seen.has(c.ID)||typeof c.Text!=='string'||c.Text.length>8192||!Number.isFinite(c.Score))throw Error('invalid lexical candidate');
  seen.add(c.ID);
  const lines=c.Text.split('\n');
  const fields=lines.filter(l=>l.startsWith('what: '));
  if(lines[0]!=='representation: eventframe-5w1h-v1'||fields.length!==1)throw Error('one canonical what field required');
  return terms(fields[0].slice(6));
 });
 const q=terms(query),df=new Map();
 for(const d of docs)for(const w of d)df.set(w,(df.get(w)??0)+1);
 const weight=w=>1+Math.log((docs.length+1)/((df.get(w)??0)+1));
 const qnorm=Math.sqrt([...q].reduce((s,w)=>s+weight(w)**2,0));
 const scores=docs.map(d=>{
  const norm=Math.sqrt([...d].reduce((s,w)=>s+weight(w)**2,0));
  const dot=[...d].reduce((s,w)=>s+(q.has(w)?weight(w)**2:0),0);
  return norm&&qnorm?dot/(norm*qnorm):0;
 });
 const states=new Map(plan?.Decisions.map(d=>[d.EventID,d.State])??[]);
 if(plan&&(states.size!==candidates.length||candidates.some(c=>!states.has(c.ID))))throw Error('plan identity mismatch');
 const order=candidates.map((_,i)=>i);
 const group=i=>plan&&!plan.AllContradicted&&states.get(candidates[i].ID)==='contradicted'?1:0;
 order.sort((a,b)=>group(a)-group(b)||scores[b]-scores[a]||candidates[b].Score-candidates[a].Score||a-b);
 return {order,scores,method:'what-lexical-v1'};
}
